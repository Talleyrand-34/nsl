/*
Copyright © 2026 Talleyrand-34 (t34@t34.dev)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package application

import (
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	configparser "nsl-graph/internal/configparser"
	e "nsl-graph/internal/repository/entities"
	s "nsl-graph/internal/scanner"
	"nsl-graph/internal/secret"
	"nsl-graph/internal/topology"
)

// ConnectionScanOptions controls a non-interactive connection discovery run (used
// by the HTTP API). Targets are selected by mode; Passphrase decrypts profiles
// that store an SSH password (key-based / SNMP-only profiles need none).
type ConnectionScanOptions struct {
	FromDB      bool   `json:"from_db"`
	Profiles    bool   `json:"profiles"`
	Subnet      string `json:"subnet"`
	Community   string `json:"community"`
	SNMPVersion string `json:"snmp_version"`
	Collector   string `json:"collector"`
	TimeoutSec  int    `json:"timeout_sec"`
	Passphrase  string `json:"passphrase"`
	// Runtime SSH credentials for subnet mode (agnostic — not from a DB profile):
	// when set, SSH-reachable hosts found in the sweep are collected over SSH.
	SSHUser     string `json:"ssh_user"`
	SSHKeyFile  string `json:"ssh_key_file"`
	SSHPassword string `json:"ssh_password"`
	SSHPort     int    `json:"ssh_port"`
	// Runtime SSH credentials brought in agnostically (never stored on the server):
	//   GenericProfile — name of a saved "generic" profile, used as an SSH fallback.
	//   SSHConfig      — raw OpenSSH config content; per-host User/IdentityFile are
	//                    matched by HostName/alias.
	//   SSHKeys        — uploaded private keys, keyed by IdentityFile basename; the
	//                    PEM is held in memory only (never written to disk).
	GenericProfile string            `json:"generic_profile"`
	SSHConfig      string            `json:"ssh_config"`
	SSHKeys        map[string]string `json:"ssh_keys"`
}

// DiscoverConnectionsByMode builds targets from the given options (no interactive
// prompting) and runs discovery — the entry point used by the web API.
func (ns *NetService) DiscoverConnectionsByMode(opts ConnectionScanOptions) (*topology.ConnectionScanResult, error) {
	community := opts.Community
	if community == "" {
		community = "public"
	}
	version := opts.SNMPVersion
	if version == "" {
		version = "v2c"
	}
	timeout := time.Duration(opts.TimeoutSec) * time.Second
	if opts.TimeoutSec == 0 {
		timeout = 10 * time.Second
	}

	if !opts.FromDB && !opts.Profiles && opts.Subnet == "" {
		opts.FromDB = true
	}

	byHost := map[string]*topology.Target{}
	var order []string
	sshNoCreds := 0 // SSH-reachable hosts found in a sweep with no runtime SSH creds
	get := func(host string) *topology.Target {
		if t, ok := byHost[host]; ok {
			return t
		}
		t := &topology.Target{Host: host}
		byHost[host] = t
		order = append(order, host)
		return t
	}
	defaultSNMP := func() *s.ScanOptions {
		return &s.ScanOptions{Timeout: timeout, SNMP: s.SNMPOptions{Community: community, Version: version}}
	}
	applyProfile := func(t *topology.Target, p *e.ScanProfile) {
		comm := p.SNMPCommunity
		if comm == "" {
			comm = community
		}
		ver := p.SNMPVersion
		if ver == "" {
			ver = version
		}
		to := timeout
		if p.TimeoutSec != 0 {
			to = time.Duration(p.TimeoutSec) * time.Second
		}
		t.SNMP = &s.ScanOptions{Timeout: to, SNMP: s.SNMPOptions{Community: comm, Version: ver, Port: uint16(p.SNMPPort)}}
		if p.SSHUser == "" {
			return
		}
		creds := configparser.SSHCredentials{Username: p.SSHUser, Port: p.SSHPort}
		switch {
		case p.SSHKeyFile != "":
			creds.KeyFile = p.SSHKeyFile
		case p.SSHPassword != "":
			pw, err := secret.Decrypt(p.SSHPassword, opts.Passphrase)
			if err != nil {
				return // can't unlock SSH password — skip SSH for this host
			}
			creds.Password = pw
		default:
			return
		}
		t.SSH = &creds
	}

	// Per-host runtime SSH resolver, used by the subnet sweep and the from-db
	// SSH fallback (when no device profile matched). Priority, highest first:
	//   1. ssh-config entry matched by IP (HostName/alias) — the key is taken
	//      from the uploaded SSHKeys by IdentityFile basename, held in memory;
	//   2. inline --ssh-user credentials;
	//   3. a selected generic profile's SSH credentials.
	// All are agnostic — none are written to the server's disk.
	var sshConfigEntries []configparser.SSHConfigHost
	if opts.SSHConfig != "" {
		if entries, err := configparser.ParseSSHConfig(strings.NewReader(opts.SSHConfig)); err == nil {
			sshConfigEntries = entries
		}
	}
	var inlineSSH *configparser.SSHCredentials
	if opts.SSHUser != "" {
		inlineSSH = &configparser.SSHCredentials{
			Username: opts.SSHUser, KeyFile: opts.SSHKeyFile,
			Password: opts.SSHPassword, Port: opts.SSHPort,
		}
	}
	var genericSSH *configparser.SSHCredentials
	if opts.GenericProfile != "" {
		if gp, err := ns.GetScanProfileByName(opts.GenericProfile); err == nil && gp != nil && gp.Kind == "generic" {
			genericSSH = profileSSHCreds(gp, opts.Passphrase)
		}
	}
	var keyMissing []string // ssh-config matches whose IdentityFile wasn't uploaded
	resolveRuntimeSSH := func(ip string) *configparser.SSHCredentials {
		if entry := configparser.MatchSSHConfig(sshConfigEntries, ip); entry != nil && entry.IdentityFile != "" {
			base := filepath.Base(entry.IdentityFile)
			pem, ok := opts.SSHKeys[base]
			if !ok {
				keyMissing = append(keyMissing, fmt.Sprintf("%s: key %q referenced by the ssh-config was not uploaded", ip, base))
				return nil
			}
			return &configparser.SSHCredentials{Username: entry.User, PrivateKey: pem, Port: entry.Port}
		}
		if inlineSSH != nil {
			c := *inlineSSH
			return &c
		}
		if genericSSH != nil {
			c := *genericSSH
			return &c
		}
		return nil
	}

	// device IP candidates + IP->label
	ipToLabel := map[string]string{}
	devName := map[string]string{}
	ipsByDevice := map[string][]string{}
	var deviceOrder []string
	if devs, err := ns.GetDevices(); err == nil {
		for _, d := range devs {
			devName[d.ID] = d.Name
		}
	}
	if ifaces, err := ns.GetAllDeviceInterfaces(); err == nil {
		seenDev := map[string]bool{}
		for _, iface := range ifaces {
			for _, raw := range iface.IPAddresses {
				ip := usableIP(raw)
				if ip == "" {
					continue
				}
				ipToLabel[ip] = devName[iface.DeviceID]
				if !seenDev[iface.DeviceID] {
					seenDev[iface.DeviceID] = true
					deviceOrder = append(deviceOrder, iface.DeviceID)
				}
				ipsByDevice[iface.DeviceID] = append(ipsByDevice[iface.DeviceID], ip)
			}
		}
	}

	if opts.FromDB {
		for _, devID := range deviceOrder {
			ips := uniqSortedStrings(ipsByDevice[devID])
			var chosen string
			var profile *e.ScanProfile
			for _, ip := range ips {
				if p, ok := ns.ResolveScanProfile(ip, ""); ok {
					chosen, profile = ip, p
					break
				}
			}
			if profile == nil {
				ss := s.NewSNMPScanner()
				probe := s.ScanOptions{Timeout: 2 * time.Second, SNMP: s.SNMPOptions{Community: community, Version: version}}
				for _, ip := range ips {
					if ss.Probe(ip, probe) {
						chosen = ip
						break
					}
				}
			}
			if chosen == "" && len(ips) > 0 {
				chosen = ips[0]
			}
			if chosen == "" {
				continue
			}
			t := get(chosen)
			t.DeviceLabel = devName[devID]
			if profile != nil {
				applyProfile(t, profile)
			} else {
				if t.SNMP == nil {
					t.SNMP = defaultSNMP()
				}
				// No device profile pins this host — fall back to runtime SSH
				// credentials (ssh-config / inline / generic profile) if any.
				if t.SSH == nil {
					if creds := resolveRuntimeSSH(chosen); creds != nil {
						t.SSH = creds
					}
				}
			}
		}
	}

	if opts.Profiles {
		profiles, err := ns.GetScanProfiles()
		if err != nil {
			return nil, err
		}
		for _, p := range profiles {
			raw, err := ns.GetScanProfileByName(p.Name)
			if err != nil || raw == nil {
				continue
			}
			host := strings.TrimSpace(raw.Host)
			t := get(host)
			t.Profile = raw.Name
			if lbl := ipToLabel[host]; lbl != "" {
				t.DeviceLabel = lbl
			}
			applyProfile(t, raw)
		}
	}

	if opts.Subnet != "" {
		// Subnet mode is DB-agnostic: discover hosts purely by SNMP-sweeping the
		// segment at runtime — no scan profiles, no DB labels for targeting — so
		// the same scan of two different DBs yields the same gather. (SSH-only
		// hosts that don't answer SNMP aren't reachable here.) Use a short sweep
		// timeout: a live agent answers fast and most of a subnet is empty.
		sweepTimeout := timeout
		if sweepTimeout > 3*time.Second {
			sweepTimeout = 3 * time.Second
		}
		for _, h := range SweepSubnet(opts.Subnet, community, version, sweepTimeout, opts.SSHPort) {
			t := get(h.IP)
			if h.SNMP && t.SNMP == nil {
				t.SNMP = defaultSNMP()
			}
			if h.SSH && t.SSH == nil {
				if creds := resolveRuntimeSSH(h.IP); creds != nil {
					t.SSH = creds
				} else if !h.SNMP {
					sshNoCreds++ // SSH-only host we can't collect from without creds
				}
			}
		}
	}

	targets := make([]topology.Target, 0, len(order))
	for _, h := range order {
		targets = append(targets, *byHost[h])
	}
	if len(targets) == 0 {
		return &topology.ConnectionScanResult{}, nil
	}
	result, err := ns.DiscoverConnections(targets, opts.Collector)
	if err == nil && result != nil && sshNoCreds > 0 {
		result.Discrepancies = append(result.Discrepancies, topology.Discrepancy{
			Kind:   "ssh-no-credentials",
			Detail: fmt.Sprintf("%d host(s) have SSH open but answer no SNMP — provide SSH credentials (ssh_user/ssh_key) to collect their LLDP/FDB and form edges", sshNoCreds),
		})
	}
	if err == nil && result != nil {
		for _, m := range keyMissing {
			result.Discrepancies = append(result.Discrepancies, topology.Discrepancy{
				Kind:   "ssh-key-missing",
				Detail: m,
			})
		}
	}
	return result, err
}

// profileSSHCreds builds SSH credentials from a profile, decrypting an encrypted
// SSH password or in-memory private key with passphrase. Returns nil when the
// profile carries no usable SSH credentials (or the passphrase is wrong).
func profileSSHCreds(p *e.ScanProfile, passphrase string) *configparser.SSHCredentials {
	if p.SSHUser == "" {
		return nil
	}
	creds := configparser.SSHCredentials{Username: p.SSHUser, Port: p.SSHPort}
	switch {
	case p.SSHKeyFile != "":
		creds.KeyFile = p.SSHKeyFile
	case p.SSHKey != "":
		pk, err := secret.Decrypt(p.SSHKey, passphrase)
		if err != nil {
			return nil
		}
		creds.PrivateKey = pk
	case p.SSHPassword != "":
		pw, err := secret.Decrypt(p.SSHPassword, passphrase)
		if err != nil {
			return nil
		}
		creds.Password = pw
	default:
		return nil
	}
	return &creds
}

// SubnetHost is a live host found by a runtime sweep and how it can be reached.
type SubnetHost struct {
	IP   string
	SNMP bool // answered SNMP
	SSH  bool // TCP port (SSH) is open
}

// SweepSubnet probes every IP in cidr at runtime for SNMP (a lightweight Get) and
// an open SSH port, concurrently and bounded, so one slow host can't stall it.
// The result depends only on the live network, not the DB. A host is returned if
// it answers SNMP OR has SSH open — so LLDP-only hosts (lldpd over SSH, no SNMP)
// are not discarded.
func SweepSubnet(cidr, community, version string, timeout time.Duration, sshPort int) []SubnetHost {
	if sshPort == 0 {
		sshPort = 22
	}
	ss := s.NewSNMPScanner()
	snmpOpts := s.ScanOptions{Timeout: timeout, SNMP: s.SNMPOptions{Community: community, Version: version}}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var hosts []SubnetHost
	sem := make(chan struct{}, 64)
	for _, ip := range enumerateCIDR(cidr) {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			snmpOK := ss.Probe(ip, snmpOpts)
			sshOK := tcpOpen(ip, sshPort, timeout)
			if snmpOK || sshOK {
				mu.Lock()
				hosts = append(hosts, SubnetHost{IP: ip, SNMP: snmpOK, SSH: sshOK})
				mu.Unlock()
			}
		}(ip)
	}
	wg.Wait()
	sort.Slice(hosts, func(i, j int) bool { return hosts[i].IP < hosts[j].IP })
	return hosts
}

// tcpOpen reports whether a TCP connection to ip:port succeeds within timeout.
func tcpOpen(ip string, port int, timeout time.Duration) bool {
	c, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}

// enumerateCIDR lists the host addresses across one or more comma/space-separated
// CIDRs (capped, de-duplicated). A bare IP (no /) contributes itself.
func enumerateCIDR(cidr string) []string {
	const max = 4096
	var ips []string
	seen := map[string]bool{}
	add := func(ip string) {
		if ip != "" && !seen[ip] {
			seen[ip] = true
			ips = append(ips, ip)
		}
	}
	for _, tok := range s.SplitSubnets(cidr) {
		if len(ips) >= max {
			break
		}
		if !strings.Contains(tok, "/") {
			add(tok)
			continue
		}
		_, ipnet, err := net.ParseCIDR(tok)
		if err != nil {
			continue
		}
		ip := make(net.IP, len(ipnet.IP))
		copy(ip, ipnet.IP)
		for ; ipnet.Contains(ip) && len(ips) < max; incIP(ip) {
			add(ip.String())
		}
	}
	return ips
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// usableIP strips a CIDR suffix and rejects IPs unusable as a scan target.
func usableIP(raw string) string {
	ip := stripCIDR(raw)
	if ip == "" || strings.Contains(ip, ":") {
		return ""
	}
	if strings.HasPrefix(ip, "127.") || strings.HasPrefix(ip, "169.254.") || ip == "0.0.0.0" {
		return ""
	}
	return ip
}

// uniqSortedStrings returns the unique, sorted, non-empty members of in.
func uniqSortedStrings(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range in {
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}

// DiscoverConnections collects multi-source L2/L1 evidence from every target and
// correlates it into connection edges. The full gather (each host's raw device
// data, evidence and FDB) is preserved in the result; only the derived edges are
// candidates for import.
func (ns *NetService) DiscoverConnections(targets []topology.Target, only string) (*topology.ConnectionScanResult, error) {
	// Collect from hosts concurrently: per-host SNMP/SSH timeouts overlap instead
	// of summing, so a multi-host scan finishes in ~one host's time, not N×.
	result := &topology.ConnectionScanResult{Hosts: make([]topology.HostScan, len(targets))}
	var wg sync.WaitGroup
	sem := make(chan struct{}, 64) // bound concurrency for large subnet sweeps
	const hostDeadline = 30 * time.Second
	for i, t := range targets {
		wg.Add(1)
		go func(i int, t topology.Target) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			// Bound each host: a broken SNMP agent can hang a walk indefinitely;
			// one bad host must not stall the whole scan.
			done := make(chan topology.HostScan, 1)
			go func() { done <- topology.CollectHost(t, only) }()
			select {
			case hs := <-done:
				result.Hosts[i] = hs
			case <-time.After(hostDeadline):
				result.Hosts[i] = topology.HostScan{
					Host:        t.Host,
					DeviceLabel: t.DeviceLabel,
					Errors:      []string{fmt.Sprintf("collection timed out after %s", hostDeadline)},
				}
			}
		}(i, t)
	}
	wg.Wait()

	if err := ns.correlate(result); err != nil {
		return result, err
	}
	return result, nil
}

// ImportConnectionEdges persists each edge as a Connection (with provenance). An
// endpoint that isn't yet a device port but names one (e.g. a "possible" edge
// like OpenWrt:eth0) has that port created on the fly. Endpoints that are only a
// device (no port name) or unknown can't be imported and are reported.
func (ns *NetService) ImportConnectionEdges(edges []topology.ConnectionEdge) (int, error) {
	n := 0
	var errs []string
	for _, edge := range edges {
		fromID, ferr := ns.resolveOrCreatePort(edge.FromDevicePortID, edge.FromLabel)
		toID, terr := ns.resolveOrCreatePort(edge.ToDevicePortID, edge.ToLabel)
		if ferr != nil || terr != nil {
			err := ferr
			if err == nil {
				err = terr
			}
			errs = append(errs, fmt.Sprintf("%s <-> %s: %v", edge.FromLabel, edge.ToLabel, err))
			continue
		}
		if err := ns.AddConnection(fromID, toID, edge.Provenance...); err != nil {
			errs = append(errs, fmt.Sprintf("%s <-> %s: %v", edge.FromLabel, edge.ToLabel, err))
			continue
		}
		n++
	}
	if len(errs) > 0 {
		return n, fmt.Errorf("%d edge(s) skipped/failed:\n  %s", len(errs), strings.Join(errs, "\n  "))
	}
	return n, nil
}

// resolveOrCreatePort returns a device-port id for an edge endpoint. If portID is
// already set it is returned; otherwise the "device:port" label is parsed and the
// device port (and its model port) is created if missing. An endpoint that is only
// a device (no port) or unknown returns an error.
func (ns *NetService) resolveOrCreatePort(portID, label string) (string, error) {
	if portID != "" {
		return portID, nil
	}
	if label == "" || strings.HasPrefix(label, "unknown(") {
		return "", fmt.Errorf("unidentified endpoint %q", label)
	}
	i := strings.LastIndexByte(label, ':')
	if i < 0 || i+1 >= len(label) {
		return "", fmt.Errorf("device-level endpoint %q (no port to connect)", label)
	}
	devLabel, portName := label[:i], label[i+1:]

	devs, err := ns.GetDevices()
	if err != nil {
		return "", err
	}
	var dev *e.Device
	for k := range devs {
		if devs[k].Name == devLabel {
			dev = &devs[k]
			break
		}
	}
	if dev == nil {
		return "", fmt.Errorf("device %q not found", devLabel)
	}

	ports, _ := ns.GetDevicePorts()
	for _, p := range ports {
		if p.DevLabel == devLabel && p.PortName == portName {
			return p.ID, nil
		}
	}

	// Find or create the model port, then the device port.
	modelPortID := ""
	if mports, err := ns.GetModelPorts(); err == nil {
		for _, mp := range mports {
			if mp.Model == dev.Model && mp.Name == portName {
				modelPortID = mp.ID
				break
			}
		}
	}
	if modelPortID == "" {
		if err := ns.AddModelPort(portName, "0", "0", dev.Model, false, "ethernet", ""); err != nil {
			return "", fmt.Errorf("create model port %q: %w", portName, err)
		}
		mports, _ := ns.GetModelPorts()
		for _, mp := range mports {
			if mp.Model == dev.Model && mp.Name == portName {
				modelPortID = mp.ID
				break
			}
		}
		if modelPortID == "" {
			return "", fmt.Errorf("model port %q not found after creation", portName)
		}
	}
	id, err := ns.AddDevicePort(dev.ID, modelPortID, "", nil)
	if err != nil {
		return "", fmt.Errorf("create device port %s:%s: %w", devLabel, portName, err)
	}
	return id, nil
}

// endpoint is a device port resolved from evidence.
type endpoint struct {
	portID string
	label  string // "device:port"
	ok     bool
}

// edgeAgg accumulates evidence for one unordered device-port pair.
type edgeAgg struct {
	fromID, toID       string
	fromLabel, toLabel string
	provenance         []string
	observedFrom       map[string]bool // port IDs that reported this neighbour
	direct             bool            // any LLDP/CDP (non-FDB) evidence
}

// correlate turns the collected evidence into edges and discrepancies, resolving
// endpoints against the device ports in the DB.
func (ns *NetService) correlate(result *topology.ConnectionScanResult) error {
	ports, err := ns.GetDevicePorts()
	if err != nil {
		return err
	}
	var deviceNames []string
	devLabelByID := map[string]string{}
	if devs, err := ns.GetDevices(); err == nil {
		for _, d := range devs {
			deviceNames = append(deviceNames, d.Name)
			devLabelByID[d.ID] = d.Name
		}
	}
	// IP (LLDP MgmtIP) -> device label, for resolving remote endpoints whose DB
	// label doesn't match their advertised sysname. An IP held by more than one
	// device (e.g. a default 10.1.1.1 on several OpenWrt boxes) is ambiguous and
	// dropped, so it can't cause a misresolution.
	ipToLabel := map[string]string{}
	ambiguousIP := map[string]bool{}
	if ifaces, err := ns.GetAllDeviceInterfaces(); err == nil {
		for _, iface := range ifaces {
			for _, raw := range iface.IPAddresses {
				ip := stripCIDR(raw)
				if ip == "" {
					continue
				}
				lbl := devLabelByID[iface.DeviceID]
				if prev, ok := ipToLabel[ip]; ok && prev != lbl {
					ambiguousIP[ip] = true
				} else {
					ipToLabel[ip] = lbl
				}
			}
		}
	}
	for ip := range ambiguousIP {
		delete(ipToLabel, ip)
	}
	correlateEvidence(result, ports, deviceNames, ipToLabel)
	return nil
}

// stripCIDR removes a trailing /prefix and whitespace from an address.
func stripCIDR(raw string) string {
	ip := strings.TrimSpace(raw)
	if i := strings.IndexByte(ip, '/'); i >= 0 {
		ip = ip[:i]
	}
	return ip
}

// correlateEvidence is the pure (DB-free) core of correlation: given the gathered
// evidence in result.Hosts and the known device ports / device names, it derives
// result.Edges and result.Discrepancies. Exposed at package level so it can be
// unit-tested with synthetic evidence.
func correlateEvidence(result *topology.ConnectionScanResult, ports []e.DevicePort, deviceNames []string, ipToLabel map[string]string) {
	norm := topology.NormalizeMAC
	lc := func(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

	portByMAC := map[string]e.DevicePort{}     // normalized MAC -> port
	portByDevPort := map[string]e.DevicePort{} // "devlabel|portname" -> port
	labelByMAC := map[string]string{}          // normalized MAC -> device label
	labelByPortID := map[string]string{}       // port id -> "device:port"
	for _, p := range ports {
		labelByPortID[p.ID] = p.DevLabel + ":" + p.PortName
		if p.MacAddress != "" {
			m := norm(p.MacAddress)
			portByMAC[m] = p
			labelByMAC[m] = p.DevLabel
		}
		portByDevPort[lc(p.DevLabel)+"|"+lc(p.PortName)] = p
	}

	devLabels := map[string]string{} // lowercased label -> canonical
	for _, name := range deviceNames {
		devLabels[lc(name)] = name
	}

	asEndpoint := func(p e.DevicePort) endpoint {
		return endpoint{portID: p.ID, label: p.DevLabel + ":" + p.PortName, ok: true}
	}

	resolveLocal := func(ev topology.NeighborEvidence) endpoint {
		if p, ok := portByDevPort[lc(ev.ObservedDevice)+"|"+lc(ev.LocalPort)]; ok {
			return asEndpoint(p)
		}
		return endpoint{}
	}

	remoteLabel := func(ev topology.NeighborEvidence) string {
		if ev.RemoteChassisMAC != "" {
			if l, ok := labelByMAC[norm(ev.RemoteChassisMAC)]; ok {
				return l
			}
		}
		if ev.RemoteSysName != "" {
			if l, ok := devLabels[lc(ev.RemoteSysName)]; ok {
				return l
			}
		}
		if ev.RemoteIP != "" {
			if l, ok := ipToLabel[stripCIDR(ev.RemoteIP)]; ok && l != "" {
				return l
			}
		}
		return ""
	}

	resolveRemote := func(ev topology.NeighborEvidence) endpoint {
		if ev.RemoteChassisMAC != "" {
			if p, ok := portByMAC[norm(ev.RemoteChassisMAC)]; ok {
				return asEndpoint(p)
			}
		}
		if l := remoteLabel(ev); l != "" && ev.RemotePort != "" {
			if p, ok := portByDevPort[lc(l)+"|"+lc(ev.RemotePort)]; ok {
				return asEndpoint(p)
			}
		}
		return endpoint{}
	}

	edges := map[string]*edgeAgg{}
	neighborsByPort := map[string]map[string]bool{}   // local port -> set of remote ports (discrepancy detection)
	possible := map[string]*topology.ConnectionEdge{} // possible connections inferred from stored port MACs

	touch := func(a, b endpoint, prov string, direct bool) {
		from, to := a, b
		if from.portID > to.portID {
			from, to = to, from
		}
		key := from.portID + "|" + to.portID
		agg := edges[key]
		if agg == nil {
			agg = &edgeAgg{fromID: from.portID, toID: to.portID, fromLabel: from.label, toLabel: to.label, observedFrom: map[string]bool{}}
			edges[key] = agg
		}
		agg.provenance = append(agg.provenance, prov)
		if direct {
			agg.direct = true
		}
	}

	// canonLabel gives an observing host a consistent name: its DB label, else its
	// scanned SysName mapped to a DB device, else the host address.
	canonLabel := func(hs topology.HostScan) string {
		if hs.DeviceLabel != "" {
			return hs.DeviceLabel
		}
		if hs.Device != nil {
			if c, ok := devLabels[lc(hs.Device.SysName)]; ok {
				return c
			}
		}
		return hs.Host
	}

	// Network-identity maps, so edges can be derived from the evidence itself
	// (DB-agnostic) and merged across observers: each scanned host gets a canonical
	// node name, reachable by its IP, sysName or chassis MAC.
	hostCanon := map[string]string{}    // host address -> canonical node name
	identToCanon := map[string]string{} // sysname/mac/ip (lowercased) -> canonical
	for _, hs := range result.Hosts {
		canon := hs.LocalSysName
		if canon == "" {
			canon = hs.DeviceLabel
		}
		if canon == "" && hs.Device != nil {
			canon = hs.Device.SysName
		}
		if canon == "" {
			canon = hs.Host
		}
		hostCanon[hs.Host] = canon
		if hs.Host != "" {
			identToCanon[lc(hs.Host)] = canon
		}
		if hs.LocalSysName != "" {
			identToCanon[lc(hs.LocalSysName)] = canon
		}
		if hs.LocalChassisMAC != "" {
			identToCanon[norm(hs.LocalChassisMAC)] = canon
		}
		if hs.DeviceLabel != "" {
			identToCanon[lc(hs.DeviceLabel)] = canon
		}
		if hs.Device != nil {
			for _, iface := range hs.Device.Interfaces {
				for _, ip := range iface.IPAddresses {
					identToCanon[lc(stripCIDR(ip))] = canon
				}
			}
		}
	}
	// resolveCanon maps a remote identity (chassis MAC / mgmt IP / sysName) to a
	// scanned host's canonical name; the bool is false for an external (un-scanned)
	// device, where the best-available identity string is returned instead.
	resolveCanon := func(sysname, mac, ip string) (string, bool) {
		if mac != "" {
			if c, ok := identToCanon[norm(mac)]; ok {
				return c, true
			}
		}
		if ip != "" {
			if c, ok := identToCanon[lc(stripCIDR(ip))]; ok {
				return c, true
			}
		}
		if sysname != "" {
			if c, ok := identToCanon[lc(sysname)]; ok {
				return c, true
			}
		}
		return firstNonEmpty(sysname, mac, ip), false
	}

	// Warn about hosts that responded but are not devices in the DB (e.g. turned
	// up by a subnet sweep): their connections can't resolve until they're added.
	for _, hs := range result.Hosts {
		if hs.Host == "" || hs.Host == "localhost" || hs.DeviceLabel != "" {
			continue
		}
		if hs.Device == nil && len(hs.Evidence) == 0 && len(hs.FDB) == 0 {
			continue // didn't actually respond
		}
		if hs.Device != nil {
			if _, ok := devLabels[lc(hs.Device.SysName)]; ok {
				continue // matched a DB device by sysname
			}
		}
		who := hs.Host
		if hs.Device != nil && hs.Device.SysName != "" {
			who = hs.Host + " (" + hs.Device.SysName + ")"
		}
		result.Discrepancies = append(result.Discrepancies, topology.Discrepancy{
			Kind:   "host-not-in-db",
			Detail: fmt.Sprintf("%s responded but is not a device in the DB — add it first (e.g. `scan host %s`) so its connections can resolve", who, hs.Host),
		})
	}

	// Count distinct MACs per physical port (collapsing VLAN sub-interfaces like
	// eth0.2 onto eth0). A port that has learned many MACs is an uplink/trunk —
	// MACs behind it are NOT directly attached, so FDB matches there are ignored.
	fdbPortMACs := map[string]map[string]bool{}
	for _, hs := range result.Hosts {
		for _, f := range hs.FDB {
			k := hs.Host + "|" + physPort(f.Port)
			if fdbPortMACs[k] == nil {
				fdbPortMACs[k] = map[string]bool{}
			}
			fdbPortMACs[k][norm(f.MAC)] = true
		}
	}

	for _, hs := range result.Hosts {
		hostLabel := canonLabel(hs)

		for _, ev := range hs.Evidence {
			local := resolveLocal(ev)
			remote := resolveRemote(ev)
			prov := provString(ev)

			if local.ok && remote.ok {
				touch(local, remote, prov, true)
				edges[edgeKey(local.portID, remote.portID)].observedFrom[local.portID] = true
				if neighborsByPort[local.portID] == nil {
					neighborsByPort[local.portID] = map[string]bool{}
				}
				neighborsByPort[local.portID][remote.portID] = true
				continue
			}
			if local.ok && !remote.ok {
				// Known local port sees a neighbour not in the DB (e.g. an
				// unmanaged switch / PoE injector). Report it; never import.
				id := firstNonEmpty(ev.RemoteSysName, ev.RemoteChassisMAC, ev.RemoteIP, "unknown")
				result.Discrepancies = append(result.Discrepancies, topology.Discrepancy{
					Kind:       "unknown-remote",
					Detail:     fmt.Sprintf("%s sees neighbour %q (port %q) which is not in the DB", local.label, id, ev.RemotePort),
					Provenance: []string{prov},
				})
				result.Edges = append(result.Edges, topology.ConnectionEdge{
					FromDevicePortID: local.portID,
					FromLabel:        local.label,
					ToLabel:          "unknown(" + id + ")",
					Confidence:       topology.ConfidenceCandidate,
					Provenance:       []string{prov},
					RemoteResolved:   false,
				})
			}
		}

		// Bridge FDB: a learned MAC that matches a device port MAC stored in the DB
		// (from the host-scan phase) reveals that device's adjacency — even for
		// devices that expose no LLDP/FDB of their own (e.g. OPNsense).
		for _, f := range hs.FDB {
			rp, rok := portByMAC[norm(f.MAC)]
			if !rok || rp.DevLabel == hostLabel {
				continue
			}
			pp := physPort(f.Port)
			// Skip MACs learned via an uplink/trunk port (many MACs behind it):
			// they are not directly attached to the observing device.
			if len(fdbPortMACs[hs.Host+"|"+pp]) > fdbDirectMax {
				continue
			}
			who := hostLabel
			prov := fmt.Sprintf("%s@%s:%s matched stored MAC %s (%s:%s)", f.Source, who, pp, norm(f.MAC), rp.DevLabel, rp.PortName)
			if lp, lok := portByDevPort[lc(hostLabel)+"|"+lc(pp)]; lok {
				// Both ends are known device ports -> a weak, importable edge.
				if lp.ID == rp.ID || lp.DevLabel == rp.DevLabel {
					continue
				}
				touch(asEndpoint(lp), asEndpoint(rp), prov, false)
				continue
			}
			// The observing switch port isn't a DB port, but the learned MAC is a
			// known device port -> a *possible* connection (shown, not imported).
			// If the MAC belongs to a logical/management interface (e.g. a switch's
			// Vlan-interface), it doesn't identify a physical port — report it at
			// device level rather than claiming that interface.
			toLabel := rp.DevLabel + ":" + rp.PortName
			toPortID := rp.ID
			dedupKey := rp.ID
			if isLogicalPortName(rp.PortName) {
				toLabel = rp.DevLabel
				toPortID = ""
				dedupKey = rp.DevLabel
			}
			key := who + "|" + pp + "|" + dedupKey
			if possible[key] == nil {
				possible[key] = &topology.ConnectionEdge{
					ToDevicePortID: toPortID,
					FromLabel:      who + ":" + pp,
					ToLabel:        toLabel,
					Confidence:     topology.ConfidenceWeak,
					RemoteResolved: false,
				}
			}
			possible[key].Provenance = append(possible[key].Provenance, prov)
		}
	}

	for _, pe := range possible {
		pe.Provenance = dedup(pe.Provenance)
		result.Edges = append(result.Edges, *pe)
	}

	// DB-agnostic edges: for LLDP evidence whose local endpoint isn't a DB port
	// (e.g. a subnet scan against an empty DB), derive the edge from the network
	// identities themselves so the adjacency is still detected. The two directions
	// merge via the canonical node names. These aren't tied to DB ports, so they're
	// shown but not directly importable.
	agn := map[string]*topology.ConnectionEdge{}
	agnObserved := map[string]map[string]bool{}
	for _, hs := range result.Hosts {
		observer := hostCanon[hs.Host]
		for _, ev := range hs.Evidence {
			if resolveLocal(ev).ok {
				continue // already handled by the DB-resolved path
			}
			rCanon, _ := resolveCanon(ev.RemoteSysName, ev.RemoteChassisMAC, ev.RemoteIP)
			if observer == "" || rCanon == "" || observer == rCanon {
				continue
			}
			epA := observer
			if ev.LocalPort != "" {
				epA = observer + ":" + ev.LocalPort
			}
			epB := rCanon
			if ev.RemotePort != "" {
				epB = rCanon + ":" + ev.RemotePort
			}
			key := epA + "|" + epB
			if epA > epB {
				key = epB + "|" + epA
			}
			if agn[key] == nil {
				from, to := epA, epB
				agn[key] = &topology.ConnectionEdge{FromLabel: from, ToLabel: to, Confidence: topology.ConfidenceCandidate}
				agnObserved[key] = map[string]bool{}
			}
			agn[key].Provenance = append(agn[key].Provenance, provString(ev))
			agnObserved[key][epA] = true
		}
	}
	for key, e := range agn {
		if len(agnObserved[key]) >= 2 {
			e.Confidence = topology.ConfidenceConfirmed
		}
		e.Provenance = dedup(e.Provenance)
		result.Edges = append(result.Edges, *e)
	}

	for _, agg := range edges {
		conf := topology.ConfidenceWeak
		if agg.direct {
			conf = topology.ConfidenceCandidate
		}
		if agg.observedFrom[agg.fromID] && agg.observedFrom[agg.toID] {
			conf = topology.ConfidenceConfirmed
		}
		result.Edges = append(result.Edges, topology.ConnectionEdge{
			FromDevicePortID: agg.fromID,
			ToDevicePortID:   agg.toID,
			FromLabel:        agg.fromLabel,
			ToLabel:          agg.toLabel,
			Confidence:       conf,
			Provenance:       dedup(agg.provenance),
			RemoteResolved:   true,
		})
	}

	// A port reporting more than one distinct neighbour is suspicious.
	for portID, set := range neighborsByPort {
		if len(set) > 1 {
			var remotes []string
			for r := range set {
				remotes = append(remotes, labelByPortID[r])
			}
			sort.Strings(remotes)
			result.Discrepancies = append(result.Discrepancies, topology.Discrepancy{
				Kind:   "multi-neighbor",
				Detail: fmt.Sprintf("%s reports %d distinct neighbours: %s", labelByPortID[portID], len(set), strings.Join(remotes, ", ")),
			})
		}
	}

	// Detect intermediary ("middle") devices: a MAC learned in the forwarding
	// tables of two or more hosts that never appears as an LLDP endpoint and is
	// not a known device port — e.g. an unmanaged / Netgear "Plus" switch that
	// bridges hosts transparently (no LLDP/SNMP of its own).
	// Known MACs to ignore: LLDP endpoints (real neighbours), our own scanned
	// devices' chassis MACs, and any DB device port MAC.
	known := map[string]bool{}
	for _, hs := range result.Hosts {
		if hs.LocalChassisMAC != "" {
			known[norm(hs.LocalChassisMAC)] = true
		}
		for _, ev := range hs.Evidence {
			if ev.RemoteChassisMAC != "" {
				known[norm(ev.RemoteChassisMAC)] = true
			}
		}
	}
	fdbHosts := map[string]map[string]bool{} // mac -> set of observing hosts
	fdbWhere := map[string][]string{}        // mac -> ["device:port", ...]
	for _, hs := range result.Hosts {
		for _, f := range hs.FDB {
			m := norm(f.MAC)
			// Only flag real network-equipment hardware: a globally-administered
			// MAC with a recognised vendor OUI (filters out randomised client
			// MACs and unknown endpoints, leaving switches/APs/routers).
			if m == "" || known[m] || !isGlobalMAC(m) || ouiVendor(m) == "" {
				continue
			}
			if _, ok := portByMAC[m]; ok {
				continue // a known device port
			}
			if fdbHosts[m] == nil {
				fdbHosts[m] = map[string]bool{}
			}
			if !fdbHosts[m][f.ObservedHost] {
				fdbHosts[m][f.ObservedHost] = true
				fdbWhere[m] = append(fdbWhere[m], canonLabel(hs)+":"+f.Port)
			}
		}
	}
	for m, hosts := range fdbHosts {
		if len(hosts) < 2 {
			continue
		}
		result.Intermediaries = append(result.Intermediaries, topology.Intermediary{
			MAC:    m,
			Vendor: ouiVendor(m),
			SeenBy: dedup(fdbWhere[m]),
		})
	}
	sort.Slice(result.Intermediaries, func(i, j int) bool {
		return len(result.Intermediaries[i].SeenBy) > len(result.Intermediaries[j].SeenBy)
	})

	sort.Slice(result.Edges, func(i, j int) bool {
		if result.Edges[i].FromLabel != result.Edges[j].FromLabel {
			return result.Edges[i].FromLabel < result.Edges[j].FromLabel
		}
		return result.Edges[i].ToLabel < result.Edges[j].ToLabel
	})
}

// ouiVendor returns a network-equipment vendor name for a MAC's OUI, or "" if the
// OUI isn't a recognised switch/AP/router vendor (so it can be filtered out).
func ouiVendor(mac string) string {
	oui := mac
	if len(mac) >= 8 {
		oui = mac[:8]
	}
	switch oui {
	case "54:07:7d", "6c:cd:d6", "20:4e:7f", "a0:04:60", "08:02:8e", "9c:3d:cf", "3c:37:86", "28:80:88", "b0:39:56", "c0:3f:0e", "04:a1:51":
		return "Netgear"
	case "60:22:32", "74:83:c2", "44:d9:e7", "fc:ec:da", "78:8a:20", "e0:63:da", "24:5a:4c", "68:d7:9a", "f0:9f:c2", "dc:9f:db":
		return "Ubiquiti"
	case "50:c7:bf", "54:af:97", "ac:84:c6", "c4:e9:0a", "98:da:c4", "00:31:92":
		return "TP-Link"
	case "00:0c:29", "00:1b:0d", "00:1e:14", "00:24:14", "f4:cf:e2", "00:1a:a1":
		return "Cisco"
	case "4c:5e:0c", "48:8f:5a", "dc:2c:6e", "cc:2d:e0":
		return "MikroTik"
	}
	return ""
}

// isGlobalMAC reports whether a MAC is globally administered (real burned-in
// hardware), i.e. the locally-administered bit is clear. Randomised/virtual
// client MACs set that bit and are excluded.
func isGlobalMAC(mac string) bool {
	parts := strings.Split(mac, ":")
	if len(parts) != 6 {
		return false
	}
	var first int
	if _, err := fmt.Sscanf(parts[0], "%x", &first); err != nil {
		return false
	}
	return first&0x02 == 0
}

// fdbDirectMax is the most MACs a port may have learned for an FDB match on it
// to count as a direct attachment (more than this = an uplink/trunk).
const fdbDirectMax = 6

// isLogicalPortName reports whether a port name is a logical/management interface
// (VLAN interface, bridge, bond, loopback, tunnel) rather than a physical port.
func isLogicalPortName(name string) bool {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "vlan"),
		strings.HasPrefix(n, "br-"), strings.HasPrefix(n, "bridge"),
		n == "lo", n == "loopback",
		strings.HasPrefix(n, "bond"), strings.HasPrefix(n, "lag"),
		strings.HasPrefix(n, "tun"), strings.HasPrefix(n, "tap"), strings.HasPrefix(n, "gre"):
		return true
	}
	return false
}

// physPort collapses a VLAN sub-interface name (eth0.2) to its physical port
// (eth0); other names are returned unchanged.
func physPort(p string) string {
	if i := strings.IndexByte(p, '.'); i >= 0 {
		return p[:i]
	}
	return p
}

func edgeKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "|" + b
}

func provString(ev topology.NeighborEvidence) string {
	who := ev.ObservedDevice
	if who == "" {
		who = ev.ObservedHost
	}
	return fmt.Sprintf("%s@%s:%s->%s", ev.Source, who, ev.LocalPort, ev.RemotePort)
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func dedup(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
