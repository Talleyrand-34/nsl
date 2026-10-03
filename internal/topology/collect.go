// SPDX-License-Identifier: AGPL-3.0-or-later
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
package topology

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/scanner"
)

// NormalizeMAC lowercases and trims a MAC address for stable comparison.
func NormalizeMAC(mac string) string {
	return strings.ToLower(strings.TrimSpace(mac))
}

// CollectHost gathers all available evidence for one target. The only filter, if
// non-empty, restricts collection to that single source. Every collector failure
// is non-fatal and recorded in HostScan.Errors so partial data is still usable.
func CollectHost(t Target, only string) HostScan {
	hs := HostScan{Host: t.Host, DeviceLabel: t.DeviceLabel, Profile: t.Profile}
	want := func(src string) bool { return only == "" || only == src }

	if t.SNMP != nil && (want(SourceSNMPLLDP) || want(SourceSNMPCDP) || want(SourceSNMPFDB)) {
		collectSNMP(t, only, &hs)
	}
	if t.SSH != nil && want(SourceSSHLLDP) {
		if ev, localMAC, localSysName, err := collectSSHLLDP(t); err != nil {
			hs.Errors = append(hs.Errors, fmt.Sprintf("%s: %v", SourceSSHLLDP, err))
		} else {
			hs.Evidence = append(hs.Evidence, ev...)
			hs.LocalChassisMAC = localMAC
			hs.LocalSysName = localSysName
		}
	}
	if t.SSH != nil && want(SourceSSHFDB) {
		if fdb, err := collectSSHFDB(t); err != nil {
			hs.Errors = append(hs.Errors, fmt.Sprintf("%s: %v", SourceSSHFDB, err))
		} else {
			hs.FDB = append(hs.FDB, fdb...)
		}
	}
	if t.Local && want(SourceLocalLLDP) {
		if ev, err := collectLocalLLDP(t); err != nil {
			hs.Errors = append(hs.Errors, fmt.Sprintf("%s: %v", SourceLocalLLDP, err))
		} else {
			hs.Evidence = append(hs.Evidence, ev...)
		}
	}
	return hs
}

// collectSNMP runs one SNMP scan and derives snmp-lldp / snmp-cdp neighbour
// evidence plus snmp-fdb entries from it (a single connection, reused).
func collectSNMP(t Target, only string, hs *HostScan) {
	want := func(src string) bool { return only == "" || only == src }
	ss := scanner.NewSNMPScanner()
	dev, err := ss.ScanDevice(t.Host, *t.SNMP)
	if err != nil {
		hs.Errors = append(hs.Errors, fmt.Sprintf("snmp: %v", err))
		return
	}
	if !dev.Reachable {
		hs.Errors = append(hs.Errors, "snmp: host did not respond (community/version?)")
		return
	}
	hs.Device = dev
	if hs.LocalSysName == "" {
		hs.LocalSysName = dev.SysName
	}

	for _, n := range dev.Neighbors {
		src := SourceSNMPLLDP
		if n.Protocol == "cdp" {
			src = SourceSNMPCDP
		}
		if !want(src) {
			continue
		}
		hs.Evidence = append(hs.Evidence, NeighborEvidence{
			Source:           src,
			ObservedHost:     t.Host,
			ObservedDevice:   t.DeviceLabel,
			LocalPort:        n.LocalPort,
			RemoteChassisMAC: NormalizeMAC(n.RemoteMAC),
			RemoteSysName:    n.RemoteName,
			RemotePort:       n.RemotePort,
			RemoteIP:         n.RemoteIP,
		})
	}
	if want(SourceSNMPFDB) {
		for _, f := range dev.BridgeFDB {
			hs.FDB = append(hs.FDB, FdbEvidence{
				Source:         SourceSNMPFDB,
				ObservedHost:   t.Host,
				ObservedDevice: t.DeviceLabel,
				MAC:            NormalizeMAC(f.MAC),
				Port:           f.Port,
				VLAN:           f.VLAN,
			})
		}
	}
}

func collectSSHLLDP(t Target) ([]NeighborEvidence, string, string, error) {
	c := configparser.NewSSHClient(*t.SSH)
	if err := c.Connect(t.Host); err != nil {
		return nil, "", "", err
	}
	defer c.Close()
	// Probe each vendor's way of being asked, keeping the first that answers.
	// A probe that the device does not understand is not an error — it just
	// means this is a different vendor — so only the case where *nothing*
	// answered is reported as a failure.
	var (
		ev       []NeighborEvidence
		answered bool
		attempts []string
	)
	note := func(cmd, out string, err error) {
		msg := firstLine(out)
		if msg == "" && err != nil {
			msg = err.Error()
		}
		attempts = append(attempts, fmt.Sprintf("%s: %s", cmd, msg))
	}
	for _, probe := range neighborProbes {
		out, err := c.Execute(probe.command)
		if !probeOutputIsUsable(out) {
			note(probe.command, out, err)
			continue
		}
		parsed, perr := probe.parse(out, t.Host, t.DeviceLabel)
		if perr != nil {
			note(probe.command, out, perr)
			continue
		}
		// The device answered in a language we understand. An empty neighbour
		// list is a legitimate answer and ends the probing: trying the next
		// vendor's command against a box that already replied would at best
		// waste a round trip and at worst misread an error string as data.
		ev, answered = parsed, true
		break
	}
	if !answered {
		return nil, "", "", noNeighborSourceError(attempts)
	}

	localMAC, localSysName := collectLocalChassis(c)
	return ev, localMAC, localSysName, nil
}

// collectLocalChassis records the host's own chassis MAC and sysName, which the
// correlator needs to recognise the observer in other hosts' evidence and to keep
// it out of intermediary detection.
//
// Neither command is fatal: a host that cannot name itself still contributes
// usable adjacencies, it just cannot be matched by name from the far side.
func collectLocalChassis(c interface{ Execute(string) (string, error) }) (mac, sysName string) {
	for _, cmd := range []string{"lldpcli show chassis", "/usr/sbin/lldpcli show chassis"} {
		if out, err := c.Execute(cmd); err == nil && probeOutputIsUsable(out) {
			if mac, sysName = parseLocalChassis(out); mac != "" || sysName != "" {
				return mac, sysName
			}
		}
	}
	// RouterOS has no lldpd, so its own identity comes from the system menu.
	// There is no chassis MAC to report — RouterOS does not expose one — and an
	// empty MAC is correct here rather than a per-interface MAC standing in for
	// a chassis identifier it is not.
	if out, err := c.Execute("/system/identity/print"); err == nil && probeOutputIsUsable(out) {
		return "", parseRouterOSIdentityLine(out)
	}
	return "", ""
}

// parseLocalChassis pulls the "ChassisID: mac .." and "SysName: .." values from
// `lldpcli show chassis` output.
func parseLocalChassis(out string) (mac, sysName string) {
	for _, line := range strings.Split(out, "\n") {
		l := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(l, "ChassisID:") && strings.Contains(l, "mac"):
			f := strings.Fields(l)
			mac = NormalizeMAC(f[len(f)-1])
		case strings.HasPrefix(l, "SysName:"):
			sysName = strings.TrimSpace(strings.TrimPrefix(l, "SysName:"))
		}
	}
	return mac, sysName
}

// collectSSHFDB reads the bridge forwarding database over SSH (`bridge fdb show`)
// and returns the dynamically learned unicast MAC→port entries. This is how a
// transparent switch that doesn't speak LLDP/SNMP still gets detected.
func collectSSHFDB(t Target) ([]FdbEvidence, error) {
	c := configparser.NewSSHClient(*t.SSH)
	if err := c.Connect(t.Host); err != nil {
		return nil, err
	}
	defer c.Close()
	// These images ship `brctl` (not iproute2 `bridge`). brctl showmacs reports a
	// bridge *port number*, so join it to the interface name via sysfs and emit
	// learned (non-local) entries as "<ifname> <mac>" lines.
	// `brctl showstp` lists each port as "<ifname> (<portno>)"; join that to
	// `brctl showmacs` (port no -> mac) to emit "<ifname> <mac>" for learned MACs.
	const cmd = `for d in /sys/class/net/*/bridge; do [ -e "$d" ] || continue; ` +
		`b=$(basename "$(dirname "$d")"); ` +
		`{ brctl showstp "$b" 2>/dev/null | awk '/^[^ ].*\([0-9]+\)/{p=$2; gsub(/[()]/,"",p); print "MAP", p, $1}'; ` +
		`brctl showmacs "$b" 2>/dev/null; } | ` +
		`awk '$1=="MAP"{n[$2]=$3;next} $3=="no"{print n[$1], $2}'; done`
	out, err := c.Execute(cmd)
	if probeOutputIsUsable(out) && err == nil {
		return parseBridgeFDB(out, t), nil
	}

	// The shell pipeline above is meaningless on a vendor CLI. RouterOS keeps the
	// same information in its own bridge host table, and rejecting the shell
	// command is how it says so — not a fault worth reporting as one.
	// An empty answer here is a real one -- a router with no bridge has no
	// bridge hosts -- so only an outright rejection sends us on.
	if rosOut, rosErr := c.Execute("/interface/bridge/host/print terse without-paging"); rosErr == nil && !probeRejected(rosOut) {
		return parseRouterOSFDB(rosOut, t), nil
	}

	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, strings.TrimSpace(out))
	}
	return parseBridgeFDB(out, t), nil
}

// parseBridgeFDB parses "<ifname> <mac>" lines (learned bridge FDB entries).
func parseBridgeFDB(out string, t Target) []FdbEvidence {
	var ev []FdbEvidence
	seen := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		port := f[0]
		mac := NormalizeMAC(f[1])
		if port == "" || !isUnicastMAC(mac) {
			continue
		}
		if key := mac + "|" + port; !seen[key] {
			seen[key] = true
			ev = append(ev, FdbEvidence{Source: SourceSSHFDB, ObservedHost: t.Host, ObservedDevice: t.DeviceLabel, MAC: mac, Port: port})
		}
	}
	return ev
}

// isUnicastMAC reports whether mac is a valid, globally-meaningful unicast MAC
// (not multicast, broadcast, or all-zero).
func isUnicastMAC(mac string) bool {
	parts := strings.Split(mac, ":")
	if len(parts) != 6 {
		return false
	}
	var first int
	if _, err := fmt.Sscanf(parts[0], "%x", &first); err != nil {
		return false
	}
	if first&1 == 1 { // multicast/broadcast
		return false
	}
	return mac != "00:00:00:00:00:00"
}

func collectLocalLLDP(t Target) ([]NeighborEvidence, error) {
	out, err := exec.Command("lldpcli", "-f", "json0", "show", "neighbors").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	host := t.Host
	if host == "" {
		host = "localhost"
	}
	return parseLLDPCLI(string(out), SourceLocalLLDP, host, t.DeviceLabel)
}

// --- lldpcli JSON (json0) parsing -------------------------------------------
//
// json0 renders every value as a single-element array of objects, e.g.
//   {"lldp":[{"interface":[{"name":"eth0",
//      "chassis":[{"id":[{"type":"mac","value":"..."}],"name":[{"value":"sw1"}],
//                  "mgmt-ip":[{"value":"10.0.0.1"}]}],
//      "port":[{"id":[{"type":"mac","value":"..."}],"descr":[{"value":"br-lan"}]}]}]}]}

type cliValue struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type cliChassis struct {
	ID     []cliValue `json:"id"`
	Name   []cliValue `json:"name"`
	MgmtIP []cliValue `json:"mgmt-ip"`
}

type cliPort struct {
	ID    []cliValue `json:"id"`
	Descr []cliValue `json:"descr"`
}

type cliIface struct {
	Name    string       `json:"name"`
	Chassis []cliChassis `json:"chassis"`
	Port    []cliPort    `json:"port"`
}

type cliRoot struct {
	Lldp []struct {
		Interface []cliIface `json:"interface"`
	} `json:"lldp"`
}

func parseLLDPCLI(data, source, host, deviceLabel string) ([]NeighborEvidence, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return nil, nil
	}
	var root cliRoot
	if err := json.Unmarshal([]byte(data), &root); err != nil {
		return nil, fmt.Errorf("parse lldpcli json: %w", err)
	}
	var ev []NeighborEvidence
	for _, l := range root.Lldp {
		for _, iface := range l.Interface {
			e := NeighborEvidence{
				Source:         source,
				ObservedHost:   host,
				ObservedDevice: deviceLabel,
				LocalPort:      iface.Name,
			}
			if len(iface.Chassis) > 0 {
				ch := iface.Chassis[0]
				e.RemoteChassisMAC = firstMAC(ch.ID)
				e.RemoteSysName = firstValue(ch.Name)
				e.RemoteIP = firstValue(ch.MgmtIP)
			}
			if len(iface.Port) > 0 {
				p := iface.Port[0]
				e.RemotePort = firstValue(p.Descr)
				if e.RemotePort == "" {
					e.RemotePort = firstMAC(p.ID)
				}
			}
			ev = append(ev, e)
		}
	}
	return ev, nil
}

func firstValue(vs []cliValue) string {
	if len(vs) == 0 {
		return ""
	}
	return strings.TrimSpace(vs[0].Value)
}

// firstMAC prefers an id entry of type "mac", normalized.
func firstMAC(vs []cliValue) string {
	for _, v := range vs {
		if strings.EqualFold(v.Type, "mac") {
			return NormalizeMAC(v.Value)
		}
	}
	if len(vs) > 0 {
		return NormalizeMAC(vs[0].Value)
	}
	return ""
}
