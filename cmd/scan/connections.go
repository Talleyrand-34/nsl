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
package cmd_scan

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"

	configparser "nsl-graph/internal/configparser"
	q "nsl-graph/internal/repository/application"
	e "nsl-graph/internal/repository/entities"
	s "nsl-graph/internal/scanner"
	"nsl-graph/internal/secret"
	"nsl-graph/internal/topology"
)

var (
	connFromDB     bool
	connSubnet     string
	connProfiles   bool
	connSource     string
	connCommunity  string
	connSNMPVer    string
	connLocal      bool
	connLocalDev   string
	connYes        bool
	connDryRun     bool
	connOutput     string
	connTimeout    int
	connHuman      bool
)

// ConnectionsCmd implements "scan connections": multi-source L2/L1 link discovery.
var ConnectionsCmd = &cobra.Command{
	Use:   "connections",
	Short: "Discover layer-2/1 links between hosts and import them as connections",
	Long: `Gather all available adjacency evidence from a set of hosts (LLDP via SNMP or
SSH, CDP, and bridge MAC tables), correlate it into connection edges, review any
discrepancies, then commit the confirmed links to the DB. The full per-host
gather is also emitted as JSON for other uses.

Targets (combine freely; default is --from-db):
  --from-db    every device with a management IP and a resolvable scan profile
  --subnet     SNMP-sweep a CIDR, then collect from each responder
  --profiles   the host of every saved scan profile

Sources: by default every available source per host is used and merged, with
discrepancies surfaced for review. Use --collector to restrict to exactly one of:
  snmp-lldp, snmp-cdp, snmp-fdb, ssh-lldp, local-lldp

Output: JSON to stdout by default (status messages go to stderr, so it pipes
cleanly); pass -H/--human for the readable summary plus interactive review.

Prerequisite: this tool only collects — it never configures the targets. LLDP
(lldpd) and/or SNMP must already be enabled on each host (set up out of band over
SSH). Hosts without them simply yield no evidence (a non-fatal per-host error).

Examples:
  nsl-graph scan connections --from-db
  nsl-graph scan connections --subnet 10.0.0.0/24 --community public
  nsl-graph scan connections --collector ssh-lldp --dry-run
  nsl-graph scan connections --from-db --yes --output gather.json`,
	Run: func(cmd *cobra.Command, args []string) {
		if connSource != "" && !validSource(connSource) {
			fmt.Fprintf(os.Stderr, "Error: unknown --collector %q (valid: %s)\n", connSource, strings.Join(topology.AllSources, ", "))
			os.Exit(1)
		}
		// Default target mode when none was selected.
		if !connFromDB && connSubnet == "" && !connProfiles {
			connFromDB = true
		}

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}

		targets, err := buildTargets(service)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building targets: %v\n", err)
			os.Exit(1)
		}
		if len(targets) == 0 {
			fmt.Fprintln(os.Stderr, "No targets to scan. Use --from-db, --subnet or --profiles (and ensure devices have IPs/profiles).")
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Scanning %d target(s)", len(targets))
		if connSource != "" {
			fmt.Fprintf(os.Stderr, " (collector: %s)", connSource)
		}
		fmt.Fprintln(os.Stderr, "...")

		result, err := service.DiscoverConnections(targets, connSource)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Discovery error: %v\n", err)
			os.Exit(1)
		}

		// Output format: JSON by default (scriptable); -H for the readable summary.
		if connHuman {
			printGatherSummary(result)
			reviewAndImport(service, result)
			if connOutput != "" {
				if err := writeJSONFile(result, connOutput); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to write gather: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "Full gather written to %s\n", connOutput)
				}
			}
		} else {
			importNonInteractive(service, result) // status to stderr
			if err := emitJSON(result); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write gather: %v\n", err)
			}
		}
	},
}

func init() {
	cmd_root.ScanCmd.AddCommand(ConnectionsCmd)
	f := ConnectionsCmd.Flags()
	f.BoolVar(&connFromDB, "from-db", false, "Target every DB device with a mgmt IP and a resolvable scan profile (default if no target flag)")
	f.StringVar(&connSubnet, "subnet", "", "SNMP-sweep this CIDR and collect from responders")
	f.BoolVar(&connProfiles, "profiles", false, "Target the host of every saved scan profile")
	f.StringVar(&connSource, "collector", "", "Restrict to a single source (snmp-lldp|snmp-cdp|snmp-fdb|ssh-lldp|local-lldp); default = all, merged")
	f.StringVar(&connCommunity, "community", "public", "SNMP community for --subnet sweeps / fallback")
	f.StringVar(&connSNMPVer, "snmp-version", "v2c", "SNMP version (v1|v2c)")
	f.BoolVar(&connLocal, "local", true, "Also collect LLDP from the machine running the tool (local-lldp)")
	f.StringVar(&connLocalDev, "local-device", "", "DB device label of the local machine (so local-lldp resolves to its ports)")
	f.BoolVar(&connYes, "yes", false, "Commit resolved confirmed/candidate edges without prompting")
	f.BoolVar(&connDryRun, "dry-run", false, "Collect and report, but never write to the DB")
	f.StringVar(&connOutput, "output", "", "Write the full gather (JSON) to this file (default JSON mode prints to stdout)")
	f.IntVar(&connTimeout, "timeout", 10, "Per-host SNMP timeout, seconds")
	f.BoolVarP(&connHuman, "human", "H", false, "Human-readable summary + interactive review (default output is JSON)")
}

// cleanIP strips a CIDR suffix and whitespace, and drops addresses unusable as
// an SNMP target (IPv6, loopback, link-local). Returns "" to skip.
func cleanIP(raw string) string {
	ip := strings.TrimSpace(raw)
	if i := strings.IndexByte(ip, '/'); i >= 0 {
		ip = ip[:i]
	}
	if ip == "" || strings.Contains(ip, ":") { // skip empty and IPv6
		return ""
	}
	if strings.HasPrefix(ip, "127.") || strings.HasPrefix(ip, "169.254.") || ip == "0.0.0.0" {
		return ""
	}
	return ip
}

// uniqSorted returns the unique, sorted, non-empty members of in.
func uniqSorted(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, s := range in {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func validSource(src string) bool {
	for _, s := range topology.AllSources {
		if s == src {
			return true
		}
	}
	return false
}

// buildTargets enumerates and de-duplicates the hosts to scan across all
// selected modes, resolving SNMP options and (decrypted) SSH credentials.
func buildTargets(service q.NetServiceInt) ([]topology.Target, error) {
	byHost := map[string]*topology.Target{}
	order := []string{}
	get := func(host string) *topology.Target {
		if t, ok := byHost[host]; ok {
			return t
		}
		t := &topology.Target{Host: host}
		byHost[host] = t
		order = append(order, host)
		return t
	}

	// Per-device cleaned management IP candidates, and a flat IP → label map for
	// labelling subnet/profile targets.
	ipToLabel := map[string]string{}
	devName := map[string]string{}
	ipsByDevice := map[string][]string{}
	deviceOrder := []string{}
	if devs, err := service.GetDevices(); err == nil {
		for _, d := range devs {
			devName[d.ID] = d.Name
		}
	}
	if ifaces, err := service.GetAllDeviceInterfaces(); err == nil {
		seenDev := map[string]bool{}
		for _, iface := range ifaces {
			for _, raw := range iface.IPAddresses {
				ip := cleanIP(raw)
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

	defaultSNMP := func() *s.ScanOptions {
		return &s.ScanOptions{
			Timeout: time.Duration(connTimeout) * time.Second,
			SNMP:    s.SNMPOptions{Community: connCommunity, Version: connSNMPVer},
		}
	}

	// --from-db: ONE target per device, using a single management IP. Prefer an
	// IP that matches a scan profile (the known mgmt address); otherwise the
	// first usable IPv4. Devices with no usable IP are skipped.
	if connFromDB {
		for _, devID := range deviceOrder {
			ips := uniqSorted(ipsByDevice[devID])
			var chosen string
			var profile *e.ScanProfile
			for _, ip := range ips {
				if p, ok := service.ResolveScanProfile(ip, ""); ok {
					chosen, profile = ip, p
					break
				}
			}
			// No profile pins the mgmt IP: probe each candidate and use the first
			// that actually answers SNMP (firewalls bind SNMP to some interfaces
			// only). Fall back to the first IP so the device still appears.
			if profile == nil {
				ss := s.NewSNMPScanner()
				probeOpts := s.ScanOptions{Timeout: 2 * time.Second, SNMP: s.SNMPOptions{Community: connCommunity, Version: connSNMPVer}}
				for _, ip := range ips {
					if ss.Probe(ip, probeOpts) {
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
			} else if t.SNMP == nil {
				t.SNMP = defaultSNMP()
			}
		}
	}

	// --profiles: every saved profile's host.
	if connProfiles {
		profiles, err := service.GetScanProfiles()
		if err != nil {
			return nil, err
		}
		for _, p := range profiles {
			raw, err := service.GetScanProfileByName(p.Name)
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

	// --subnet: profile-matched hosts in the subnet (covers SSH-only devices with
	// no SNMP), then an SNMP sweep for everything else that responds.
	if connSubnet != "" {
		for _, ip := range q.EnumerateCIDR(connSubnet) {
			if p, ok := service.ResolveScanProfile(ip, ""); ok {
				t := get(ip)
				t.Profile = p.Name
				if lbl := ipToLabel[ip]; lbl != "" {
					t.DeviceLabel = lbl
				}
				applyProfile(t, p)
			}
		}
		res, err := s.NewSNMPScanner().Scan(s.ScanOptions{
			Subnet:  connSubnet,
			Timeout: time.Duration(connTimeout) * time.Second,
			SNMP:    s.SNMPOptions{Community: connCommunity, Version: connSNMPVer},
		})
		if err != nil {
			return nil, fmt.Errorf("subnet sweep: %w", err)
		}
		for _, dev := range res.Devices {
			t := get(dev.IP)
			if lbl := ipToLabel[dev.IP]; lbl != "" {
				t.DeviceLabel = lbl
			}
			if t.SNMP == nil {
				t.SNMP = defaultSNMP()
			}
		}
	}

	// local-lldp from this machine.
	if connLocal && (connSource == "" || connSource == topology.SourceLocalLLDP) {
		t := get("localhost")
		t.Local = true
		if connLocalDev != "" {
			t.DeviceLabel = connLocalDev
		}
	}

	targets := make([]topology.Target, 0, len(order))
	for _, h := range order {
		targets = append(targets, *byHost[h])
	}
	return targets, nil
}

// applyProfile fills a target's SNMP options and SSH credentials from a scan
// profile, prompting once (cached) for the passphrase that unlocks an encrypted
// SSH password.
func applyProfile(t *topology.Target, p *e.ScanProfile) {
	community := p.SNMPCommunity
	if community == "" {
		community = connCommunity
	}
	version := p.SNMPVersion
	if version == "" {
		version = connSNMPVer
	}
	timeout := connTimeout
	if p.TimeoutSec != 0 {
		timeout = p.TimeoutSec
	}
	t.SNMP = &s.ScanOptions{
		Timeout: time.Duration(timeout) * time.Second,
		SNMP:    s.SNMPOptions{Community: community, Version: version, Port: uint16(p.SNMPPort)},
	}

	if p.SSHUser == "" {
		return
	}
	creds := configparser.SSHCredentials{Username: p.SSHUser, Port: p.SSHPort}
	switch {
	case p.SSHKeyFile != "":
		creds.KeyFile = p.SSHKeyFile
	case p.SSHPassword != "":
		pw, err := unlockProfilePassword(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  (skipping SSH for %s: %v)\n", p.Host, err)
			return
		}
		creds.Password = pw
	default:
		return // no usable SSH secret
	}
	t.SSH = &creds
}

// passphraseCache memoizes the user's passphrase across profiles in one run.
var passphraseCache string
var passphraseAsked bool

func unlockProfilePassword(p *e.ScanProfile) (string, error) {
	if !passphraseAsked {
		pass, err := readSecret("Passphrase to unlock SSH credentials: ")
		if err != nil {
			return "", err
		}
		passphraseCache = pass
		passphraseAsked = true
	}
	return secret.Decrypt(p.SSHPassword, passphraseCache)
}

func printGatherSummary(result *topology.ConnectionScanResult) {
	fmt.Printf("\n=== Gather (%d host(s)) ===\n", len(result.Hosts))
	for _, hs := range result.Hosts {
		name := hs.DeviceLabel
		if name == "" && hs.Device != nil {
			name = hs.Device.SysName
		}
		who := hs.Host
		if name != "" && name != hs.Host {
			who = hs.Host + " (" + name + ")"
		}
		fmt.Printf("  %-34s evidence=%d fdb=%d", who, len(hs.Evidence), len(hs.FDB))
		if len(hs.Errors) > 0 {
			fmt.Printf("  errors: %s", strings.Join(hs.Errors, "; "))
		}
		fmt.Println()
	}

	if len(result.Intermediaries) > 0 {
		fmt.Printf("\n=== Intermediary devices detected in the middle (%d) ===\n", len(result.Intermediaries))
		for _, in := range result.Intermediaries {
			fmt.Printf("  %s (%s) — seen by %s\n", in.MAC, in.Vendor, strings.Join(in.SeenBy, ", "))
		}
	}

	if len(result.Discrepancies) > 0 {
		fmt.Printf("\n=== Discrepancies to review (%d) ===\n", len(result.Discrepancies))
		for _, d := range result.Discrepancies {
			fmt.Printf("  [%s] %s\n", d.Kind, d.Detail)
		}
	}

	fmt.Printf("\n=== Derived edges (%d) ===\n", len(result.Edges))
	for _, edge := range result.Edges {
		mark := edge.Confidence
		if !edge.RemoteResolved {
			if strings.HasPrefix(edge.ToLabel, "unknown(") {
				mark = "unresolved" // neither end identified in the DB
			} else {
				mark = "possible" // one end identified (by stored MAC), not importable
			}
		}
		fmt.Printf("  [%-10s] %s <-> %s\n", mark, edge.FromLabel, edge.ToLabel)
		if len(edge.Provenance) > 0 {
			fmt.Printf("               via %s\n", strings.Join(edge.Provenance, ", "))
		}
	}

	renderTopology(result)
}

// topoLink is one adjacency in the rendered topology.
type topoLink struct {
	peer, localPort, peerPort, tag string
}

// splitEndpoint splits a "device:port" label into device and port (port "" when
// the label is device-level or an unknown(...) placeholder).
func splitEndpoint(s string) (string, string) {
	if strings.HasPrefix(s, "unknown(") {
		return s, ""
	}
	if i := strings.LastIndexByte(s, ':'); i >= 0 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

// renderTopology prints the discovered edges (and intermediaries) as an ASCII
// tree, rooted at a gateway-like node (one reached only via FDB).
func renderTopology(result *topology.ConnectionScanResult) {
	adj := map[string][]topoLink{}
	hasLLDP := map[string]bool{}
	degree := map[string]int{}
	realEdge := map[string]bool{} // has at least one non-intermediary edge
	add := func(a, ap, b, bp, tag string, lldp, intermediary bool) {
		adj[a] = append(adj[a], topoLink{b, ap, bp, tag})
		adj[b] = append(adj[b], topoLink{a, bp, ap, tag})
		degree[a]++
		degree[b]++
		if lldp {
			hasLLDP[a] = true
			hasLLDP[b] = true
		}
		if !intermediary {
			realEdge[a] = true
			realEdge[b] = true
		}
	}

	srcOf := func(provs []string) string {
		seen := map[string]bool{}
		var order []string
		for _, p := range provs {
			s := p
			if i := strings.IndexByte(p, '@'); i >= 0 {
				s = p[:i]
			}
			if !seen[s] {
				seen[s] = true
				order = append(order, s)
			}
		}
		return strings.Join(order, "+")
	}

	for _, e := range result.Edges {
		aDev, aPort := splitEndpoint(e.FromLabel)
		bDev, bPort := splitEndpoint(e.ToLabel)
		if aDev == "" || bDev == "" || aDev == bDev {
			continue
		}
		mark := e.Confidence
		if !e.RemoteResolved {
			if strings.HasPrefix(e.ToLabel, "unknown(") {
				mark = "unresolved"
			} else {
				mark = "possible"
			}
		}
		src := srcOf(e.Provenance)
		tag := mark
		if src != "" {
			tag = mark + " " + src
		}
		add(aDev, aPort, bDev, bPort, tag, strings.Contains(src, "lldp"), false)
	}

	// Attach intermediary-only devices (no edge of their own) behind the hub port.
	for _, in := range result.Intermediaries {
		hub, hubPort, best := "", "", -1
		for _, sb := range in.SeenBy {
			d, p := splitEndpoint(sb)
			if degree[d] > best {
				best, hub, hubPort = degree[d], d, p
			}
		}
		if hub == "" {
			continue
		}
		tag := "via " + in.Vendor + " switch " + in.MAC
		for _, sb := range in.SeenBy {
			d, _ := splitEndpoint(sb)
			if d != hub && degree[d] == 0 {
				add(hub, hubPort, d, "", tag, false, true)
			}
		}
	}

	if len(adj) == 0 {
		return
	}

	// Root: a node reached only via FDB (gateway-like), most peripheral; else any.
	var nodes []string
	for n := range adj {
		nodes = append(nodes, n)
	}
	sort.Strings(nodes)
	root := nodes[0]
	bestScore := 1 << 30
	for _, n := range nodes {
		if !realEdge[n] {
			continue // intermediary-only leaf — never the root
		}
		score := degree[n]
		if hasLLDP[n] {
			score += 1000 // prefer non-LLDP (router/gateway) roots
		}
		if score < bestScore {
			bestScore, root = score, n
		}
	}

	// Build a spanning tree depth-first, visiting lower-degree neighbours first so
	// a bridge (e.g. a switch) claims a hub before the hub is reached directly —
	// turning a transitive triangle (A–C alongside A–B–C) into a clean chain.
	children := map[string][]topoLink{}
	visited := map[string]bool{root: true}
	order := func(links []topoLink) {
		sort.Slice(links, func(i, j int) bool {
			if di, dj := degree[links[i].peer], degree[links[j].peer]; di != dj {
				return di < dj
			}
			if links[i].localPort != links[j].localPort {
				return links[i].localPort < links[j].localPort
			}
			return links[i].peer < links[j].peer
		})
	}
	var build func(dev string)
	build = func(dev string) {
		links := append([]topoLink{}, adj[dev]...)
		order(links)
		for _, l := range links {
			if visited[l.peer] {
				continue
			}
			visited[l.peer] = true
			children[dev] = append(children[dev], l)
			build(l.peer)
		}
	}
	build(root)

	fmt.Println("\n=== Discovered topology ===")
	fmt.Println(root)
	var render func(dev, prefix string)
	render = func(dev, prefix string) {
		kids := children[dev]
		for i, l := range kids {
			branch, childPrefix := "├── ", prefix+"│   "
			if i == len(kids)-1 {
				branch, childPrefix = "└── ", prefix+"    "
			}
			lp := l.localPort
			if lp == "" {
				lp = "·"
			}
			peer := l.peer
			if l.peerPort != "" {
				peer += ":" + l.peerPort
			}
			fmt.Printf("%s%s%s ─[%s]─ %s\n", prefix, branch, lp, l.tag, peer)
			render(l.peer, childPrefix)
		}
	}
	render(root, "")
}

// reviewAndImport commits the resolved edges, interactively unless --yes/--dry-run.
func reviewAndImport(service q.NetServiceInt, result *topology.ConnectionScanResult) {
	if connDryRun {
		fmt.Println("\n--dry-run: nothing written to the DB.")
		return
	}

	var toImport []topology.ConnectionEdge
	importable := func(edge topology.ConnectionEdge) bool {
		return edge.RemoteResolved && edge.FromDevicePortID != "" && edge.ToDevicePortID != ""
	}

	if connYes {
		for _, edge := range result.Edges {
			if importable(edge) && edge.Confidence != topology.ConfidenceWeak {
				toImport = append(toImport, edge)
			}
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		acceptRest := false
		fmt.Println("\nReview edges to import (y = yes, n = no, a = accept all remaining, q = quit):")
		for _, edge := range result.Edges {
			if !importable(edge) {
				continue
			}
			if acceptRest {
				toImport = append(toImport, edge)
				continue
			}
			def := edge.Confidence == topology.ConfidenceConfirmed || edge.Confidence == topology.ConfidenceCandidate
			fmt.Printf("  Import [%s] %s <-> %s? %s ", edge.Confidence, edge.FromLabel, edge.ToLabel, defaultHint(def))
			line, _ := reader.ReadString('\n')
			switch strings.ToLower(strings.TrimSpace(line)) {
			case "a":
				acceptRest = true
				toImport = append(toImport, edge)
			case "q":
				goto done
			case "y", "yes":
				toImport = append(toImport, edge)
			case "n", "no":
			case "":
				if def {
					toImport = append(toImport, edge)
				}
			}
		}
	}
done:
	if len(toImport) == 0 {
		fmt.Println("\nNo edges selected; nothing imported.")
		return
	}
	n, err := service.ImportConnectionEdges(toImport)
	fmt.Printf("\nImported %d connection(s).\n", n)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Some edges were not imported:\n%v\n", err)
	}
}

func defaultHint(def bool) string {
	if def {
		return "[Y/n]"
	}
	return "[y/N]"
}

// importNonInteractive commits resolved confirmed/candidate edges in JSON mode
// (only with --yes); all status goes to stderr so stdout stays valid JSON.
func importNonInteractive(service q.NetServiceInt, result *topology.ConnectionScanResult) {
	if connDryRun {
		fmt.Fprintln(os.Stderr, "--dry-run: nothing written to the DB.")
		return
	}
	if !connYes {
		fmt.Fprintln(os.Stderr, "note: default JSON mode is non-interactive — pass --yes to import, or -H to review interactively. Nothing written.")
		return
	}
	var toImport []topology.ConnectionEdge
	for _, edge := range result.Edges {
		if edge.RemoteResolved && edge.FromDevicePortID != "" && edge.ToDevicePortID != "" && edge.Confidence != topology.ConfidenceWeak {
			toImport = append(toImport, edge)
		}
	}
	if len(toImport) == 0 {
		fmt.Fprintln(os.Stderr, "No importable edges.")
		return
	}
	n, err := service.ImportConnectionEdges(toImport)
	fmt.Fprintf(os.Stderr, "Imported %d connection(s).\n", n)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Some edges were not imported:\n%v\n", err)
	}
}

// writeJSONFile writes the full result as indented JSON to path.
func writeJSONFile(result *topology.ConnectionScanResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// emitJSON writes the result as JSON to --output, or to stdout (raw, the default).
func emitJSON(result *topology.ConnectionScanResult) error {
	if connOutput != "" {
		if err := writeJSONFile(result, connOutput); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Full gather written to %s\n", connOutput)
		return nil
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
