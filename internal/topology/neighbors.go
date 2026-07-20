// neighbors.go: asking each vendor for its neighbour table in its own language.
package topology

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

import (
	"fmt"
	"strings"

	"nsl-graph/internal/configparser/parsers"
)

// The SSH neighbour collector used to be one hardcoded command,
// `lldpcli -f json0 show neighbors`, which assumed every target was a Linux host
// running lldpd. Three of the five vendors in a routed multi-vendor lab fail that
// assumption, each for a different reason:
//
//   - RouterOS has no lldpd at all. It keeps its own neighbour table, populated
//     by MNDP, LLDP and CDP together, at `/ip/neighbor` — and that table is
//     richer than LLDP alone, because it names the remote *port* as well as the
//     remote host.
//   - VyOS ships lldpcli but the operator's login shell is the VyOS CLI, which
//     answers a bare `lldpcli` with "Invalid command". The binary is only
//     reachable by absolute path.
//   - Infix runs stock lldpd and works with the original command — but only
//     because it permits the unprivileged read; that is worth a probe rather
//     than an assumption.
//
// Rather than dispatch on an OS type, the collector probes. Target carries no OS
// type (it is built from a scan profile, which may not name one), and probing is
// the more honest design anyway: it discovers what the device actually answers
// instead of trusting a label that may be stale or absent.

// neighborProbe is one way of asking a device for its neighbours.
type neighborProbe struct {
	// command is run verbatim over the session.
	command string
	// parse turns its output into evidence. Returning no evidence and no error
	// means "this device answered, but has no neighbours" — which is a real and
	// common answer, so the prober must not mistake it for a failed probe.
	parse func(out, host, deviceLabel string) ([]NeighborEvidence, error)
}

// neighborProbes are tried in order until one produces evidence.
//
// lldpd comes first because it is the standard and the most widely deployed; the
// RouterOS probe is last because its command is meaningless on any other vendor
// and would only ever be reached after the portable ones have failed.
var neighborProbes = []neighborProbe{
	{
		command: "lldpcli -f json0 show neighbors",
		parse: func(out, host, label string) ([]NeighborEvidence, error) {
			return parseLLDPCLI(out, SourceSSHLLDP, host, label)
		},
	},
	{
		// VyOS: lldpcli is not on the CLI shell's PATH.
		command: "/usr/sbin/lldpcli -f json0 show neighbors",
		parse: func(out, host, label string) ([]NeighborEvidence, error) {
			return parseLLDPCLI(out, SourceSSHLLDP, host, label)
		},
	},
	{
		command: "/ip/neighbor/print terse without-paging",
		parse:   parseRouterOSNeighbors,
	},
}

// probeOutputIsUsable reports whether a probe's output is worth parsing.
//
// A shell that does not know the command, and a restricted CLI that rejects it,
// both answer on stdout with a zero or non-zero status depending on the vendor —
// so the exit status alone cannot decide this. These are the shapes seen from the
// lab's five vendors when a probe does not apply.
func probeOutputIsUsable(out string) bool {
	return strings.TrimSpace(out) != "" && !probeRejected(out)
}

// probeRejected reports whether the device refused the command outright, as
// opposed to answering it with nothing.
//
// The distinction matters and is easy to lose. An empty neighbour or bridge
// table is a real answer — a router with no bridge configured has no bridge
// hosts, and that is information. A rejection is not an answer at all and means
// we asked the wrong vendor's question. Conflating the two either reports a
// working device as broken, or stops a probe sequence before the command that
// would have worked.
func probeRejected(out string) bool {
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return false
	}
	low := strings.ToLower(trimmed)
	for _, marker := range []string{
		"not found",        // POSIX shells
		"no such file",     // absolute-path probe on a box without lldpcli
		"invalid command",  // VyOS CLI shell
		"bad command name", // RouterOS
		"syntax error",     // RouterOS
		"expected end of command",
		"unable to connect to socket", // lldpcli present, lldpd not running
		"command not found",
		"command parse error", // FortiOS
		"command fail. return code",
		"unknown action", // FortiOS, for a command it cannot even tokenise
	} {
		if strings.Contains(low, marker) {
			return true
		}
	}
	return false
}

// parseRouterOSNeighbors turns `/ip/neighbor/print terse` into evidence.
//
// This table is the best adjacency source any vendor in the lab offers. Where
// LLDP alone yields the remote chassis and port, RouterOS additionally reports
// the neighbour's management address and the protocol that found it, and it
// merges MNDP, LLDP and CDP into one view — so a MikroTik facing a MikroTik is
// discovered even where neither speaks LLDP.
//
// A record is kept only if it names the local interface, which is the one field
// without which the observation cannot be placed on the graph.
func parseRouterOSNeighbors(out, host, deviceLabel string) ([]NeighborEvidence, error) {
	// Refuse output that is not RouterOS at all.
	//
	// Blocklisting each vendor's rejection wording is a losing game — FortiOS
	// alone says "command parse error" for one command and "Unknown action 0"
	// for another. Without this check, any prose the probe list has not seen
	// before parses to zero records, which is indistinguishable from a device
	// that genuinely has no neighbours: the last probe "succeeds" with an empty
	// result and a box that rejected every question is reported as simply having
	// no links. Recognising the grammar positively is the check that does not
	// need updating for the next vendor.
	if !looksLikeTerseOutput(out) {
		return nil, fmt.Errorf("not RouterOS terse output")
	}

	var evidence []NeighborEvidence

	for _, rec := range parsers.ParseRouterOSTerse(out) {
		local := rec["interface"]
		if local == "" {
			continue
		}
		// `interface` may list several ports when the neighbour is heard on a
		// bridge; the first is the one that received the advertisement.
		if i := strings.IndexAny(local, ","); i >= 0 {
			local = local[:i]
		}

		ev := NeighborEvidence{
			Source:           SourceSSHLLDP,
			Discovery:        strings.ToLower(rec["discovered-by"]),
			ObservedHost:     host,
			ObservedDevice:   deviceLabel,
			LocalPort:        strings.TrimSpace(local),
			RemoteChassisMAC: NormalizeMAC(rec["mac-address"]),
			RemoteSysName:    rec["identity"],
			// RouterOS calls the *remote* port `interface-name`, next to the
			// local `interface`. Confusing the two silently reverses every edge.
			RemotePort: rec["interface-name"],
			RemoteIP:   firstNonEmpty(rec["address4"], rec["address"]),
		}
		evidence = append(evidence, ev)
	}

	return evidence, nil
}

// parseRouterOSIdentityLine pulls the device's own name out of
// `/system/identity/print`, whose output is `  name: X` rather than key=value.
func parseRouterOSIdentityLine(out string) string {
	for _, line := range strings.Split(out, "\n") {
		if _, rest, ok := strings.Cut(strings.TrimSpace(line), "name:"); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// parseRouterOSFDB reads `/interface/bridge/host/print terse` — RouterOS's
// equivalent of a bridge forwarding database.
//
// Only dynamically learned, non-local entries are adjacency evidence. A `local`
// entry is one of the bridge's own addresses, and reporting it would claim the
// device is its own neighbour.
func parseRouterOSFDB(out string, t Target) []FdbEvidence {
	var ev []FdbEvidence
	seen := map[string]bool{}

	for _, rec := range parsers.ParseRouterOSTerse(out) {
		if strings.EqualFold(rec["local"], "yes") {
			continue
		}
		mac := NormalizeMAC(rec["mac-address"])
		port := strings.TrimSpace(rec["interface"])
		if mac == "" || port == "" || !isUnicastMAC(mac) {
			continue
		}
		key := port + "|" + mac
		if seen[key] {
			continue
		}
		seen[key] = true

		ev = append(ev, FdbEvidence{
			Source:         SourceSSHFDB,
			ObservedHost:   t.Host,
			ObservedDevice: t.DeviceLabel,
			MAC:            mac,
			Port:           port,
			VLAN:           rec["vid"],
		})
	}
	return ev
}

// noNeighborSourceError explains why no probe answered.
//
// Reporting the last probe's complaint is actively misleading: the probes are
// ordered, the last one is RouterOS's, and telling the operator of a VyOS box
// that `/ip/neighbor/print` does not exist says nothing about their device. The
// diagnosis worth surfacing is the one that names a fixable cause, so a host
// that has lldpcli but no running lldpd is reported as exactly that — the single
// most common reason a Linux-family device yields nothing.
func noNeighborSourceError(attempts []string) error {
	for _, a := range attempts {
		low := strings.ToLower(a)
		if strings.Contains(low, "unable to connect to socket") || strings.Contains(low, "lldpd.socket") {
			return fmt.Errorf("lldpd is installed but not running — enable LLDP on the device " +
				"(VyOS: `set service lldp`); this tool only collects, it never configures a target")
		}
	}
	if len(attempts) == 0 {
		return fmt.Errorf("no neighbour source answered")
	}
	return fmt.Errorf("no neighbour source answered; tried %s", strings.Join(attempts, "; "))
}

// firstLine returns the first non-blank line of s, trimmed — enough to identify
// a failure without pasting a whole error dump into a per-host message.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if l := strings.TrimSpace(line); l != "" {
			return l
		}
	}
	return ""
}

// looksLikeTerseOutput reports whether text has the shape of RouterOS
// `print terse` output.
//
// Every meaningful line is either a record (an index, then key=value pairs), a
// column/flag legend, or the continuation of a wrapped record. Anything else —
// a shell error, another vendor's prose — is not this grammar.
func looksLikeTerseOutput(out string) bool {
	sawRecord := false
	for _, line := range strings.Split(out, "\n") {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		if strings.HasPrefix(l, "Flags:") || strings.HasPrefix(l, "Columns:") {
			continue
		}
		if startsWithIndex(l) {
			sawRecord = true
			continue
		}
		// A continuation only makes sense after a record has started, and it
		// still has to carry the key=value grammar.
		if sawRecord && strings.Contains(l, "=") {
			continue
		}
		return false
	}
	return sawRecord
}

// startsWithIndex reports whether a line opens with a RouterOS record index.
func startsWithIndex(l string) bool {
	i := 0
	for i < len(l) && l[i] >= '0' && l[i] <= '9' {
		i++
	}
	return i > 0 && i < len(l) && (l[i] == ' ' || l[i] == '\t')
}
