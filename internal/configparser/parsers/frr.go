// frr.go: one FRR config parser, shared by every OS that delegates routing to FRR.
// SPDX-License-Identifier: AGPL-3.0-or-later
package parsers

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
	"strings"

	"nsl-graph/internal/configparser"
)

// FRR is the common denominator of the lab's routed topologies. OpenWrt reaches
// it through the `frr` package, Infix embeds it behind sysrepo, and OPNsense
// ships it as the `os-frr` plugin — but all three end up with the same
// `frr.conf` grammar, so they share one parser rather than three near-copies.
//
// The grammar parsed here is the `vtysh -c "show running-config"` / `frr.conf`
// form: flat lines, `!` comments, one-space indentation inside a `router …`
// block, terminated by `exit` or by the next top-level keyword.

// ParseFRRConfig extracts routing protocol instances from an FRR configuration.
//
// It is deliberately tolerant: an unrecognised line inside a known block is
// skipped rather than failing the parse, because FRR emits a great many
// timers, redistribute/route-map lines that carry no topology information, and a
// device is not misconfigured just because we do not model one of them.
func ParseFRRConfig(raw string) []configparser.ConfigRoutingProtocol {
	var protos []configparser.ConfigRoutingProtocol
	var cur *configparser.ConfigRoutingProtocol
	// OSPF states an area per `network` line, so areas accumulate by ID and are
	// flushed when the block closes.
	var areas map[string]*configparser.ConfigOSPFArea
	var areaOrder []string

	flush := func() {
		if cur == nil {
			return
		}
		for _, id := range areaOrder {
			cur.Areas = append(cur.Areas, *areas[id])
		}
		protos = append(protos, *cur)
		cur, areas, areaOrder = nil, nil, nil
	}

	area := func(id string) *configparser.ConfigOSPFArea {
		if areas == nil {
			areas = map[string]*configparser.ConfigOSPFArea{}
		}
		if a, ok := areas[id]; ok {
			return a
		}
		a := &configparser.ConfigOSPFArea{ID: id}
		areas[id] = a
		areaOrder = append(areaOrder, id)
		return a
	}

	// neighbor collects per-peer lines, which FRR spreads over several
	// statements for the same address.
	neighbor := func(addr string) *configparser.ConfigBGPNeighbor {
		for i := range cur.Neighbors {
			if cur.Neighbors[i].Address == addr {
				return &cur.Neighbors[i]
			}
		}
		cur.Neighbors = append(cur.Neighbors, configparser.ConfigBGPNeighbor{Address: addr})
		return &cur.Neighbors[len(cur.Neighbors)-1]
	}

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		if line == "" || strings.HasPrefix(line, "!") || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Fields(line)

		// `exit` closes the current block; `exit-address-family` does not — it
		// only leaves the AF sub-block, and the router block continues.
		if f[0] == "exit" {
			flush()
			continue
		}
		if f[0] == "exit-address-family" {
			continue
		}

		if f[0] == "router" && len(f) >= 2 {
			flush() // a router block may be closed by the next one, not by `exit`
			p := configparser.ConfigRoutingProtocol{Enabled: true}
			switch f[1] {
			case "ospf":
				p.Type = configparser.RoutingProtoOSPF
			case "ospf6":
				p.Type = configparser.RoutingProtoOSPFv3
			case "bgp":
				p.Type = configparser.RoutingProtoBGP
				if len(f) >= 3 {
					p.LocalAS = f[2]
				}
			case "rip":
				p.Type = configparser.RoutingProtoRIP
			case "isis":
				p.Type = configparser.RoutingProtoISIS
			default:
				continue // unknown protocol: ignore the block entirely
			}
			// `router ospf <instance>` and `router ospf vrf <name>`.
			if p.Type == configparser.RoutingProtoOSPF && len(f) >= 3 {
				if f[2] == "vrf" && len(f) >= 4 {
					p.VRF = f[3]
				} else {
					p.Instance = f[2]
				}
			}
			if p.Type == configparser.RoutingProtoBGP && len(f) >= 5 && f[3] == "vrf" {
				p.VRF = f[4]
			}
			cur = &p
			continue
		}

		if cur == nil {
			continue
		}

		switch {
		// `ospf router-id X` / `bgp router-id X` / bare `router-id X`.
		case len(f) >= 2 && f[len(f)-2] == "router-id":
			cur.RouterID = f[len(f)-1]

		// `network 10.1.4.0/24 area 0`  (OSPF)
		// `network 10.1.4.0/24`          (BGP / RIP)
		case f[0] == "network" && len(f) >= 2:
			if len(f) >= 4 && f[2] == "area" {
				a := area(f[3])
				a.Networks = append(a.Networks, f[1])
			} else {
				cur.Networks = append(cur.Networks, f[1])
			}

		// `area 0 stub` / `area 0.0.0.0 nssa`
		case f[0] == "area" && len(f) >= 3:
			a := area(f[1])
			switch f[2] {
			case "stub", "nssa":
				a.Type = f[2]
			}

		// `neighbor <addr> remote-as 65000` and friends.
		case f[0] == "neighbor" && len(f) >= 3:
			n := neighbor(f[1])
			switch f[2] {
			case "remote-as":
				if len(f) >= 4 {
					n.RemoteAS = f[3]
				}
			case "update-source":
				if len(f) >= 4 {
					n.UpdateSource = f[3]
				}
			case "description":
				if len(f) >= 4 {
					n.Description = strings.Join(f[3:], " ")
				}
			case "next-hop-self":
				n.NextHopSelf = true
			case "local-as":
				if len(f) >= 4 {
					n.LocalAS = f[3]
				}
			}

		case f[0] == "redistribute" && len(f) >= 2:
			cur.Redistribute = append(cur.Redistribute, f[1])

		// A protocol block may be administratively shut down.
		case f[0] == "shutdown":
			cur.Enabled = false
		}
	}
	flush()

	return protos
}

// extractCommandBlock returns the output of one command from a combined fetch
// blob, which parsers assemble as "# <command>\n<output>" per command.
//
// It stops at the next "# " line that names a command, so an FRR config's own
// "!" comments and a UCI dump's quoted values pass through untouched.
func extractCommandBlock(raw, command string) string {
	marker := "# " + command
	lines := strings.Split(raw, "\n")
	var out []string
	collecting := false
	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		if strings.TrimSpace(trimmed) == marker {
			collecting = true
			continue
		}
		if collecting {
			// The next command's tag ends this block. An error line for a
			// following command ("# Error executing X: …") ends it too.
			if strings.HasPrefix(trimmed, "# ") && !strings.HasPrefix(trimmed, "# !") {
				if strings.HasPrefix(trimmed, "# Error executing ") || looksLikeCommandTag(trimmed) {
					break
				}
			}
			out = append(out, trimmed)
		}
	}
	return strings.Join(out, "\n")
}

// looksLikeCommandTag reports whether a "# …" line is one of our own command
// tags rather than a comment belonging to the command's output. Our tags are
// always a shell command, so they begin with a path, a slash-rooted RouterOS
// path, or a known verb.
func looksLikeCommandTag(line string) bool {
	rest := strings.TrimSpace(strings.TrimPrefix(line, "#"))
	if rest == "" {
		return false
	}
	// Every verb any parser uses to open a command block. A verb missing here is
	// not a cosmetic gap: the block for the *previous* command then runs on
	// through the unrecognised tag and swallows every command after it. FortiOS
	// contributed `get` and `diagnose`.
	for _, verb := range []string{
		"uci ", "cat ", "swconfig ", "sudo ", "ifconfig", "hostname",
		"show ", "get ", "diagnose ", "/",
	} {
		if strings.HasPrefix(rest, verb) {
			return true
		}
	}
	return false
}

// frrConfigLooksReal reports whether output from a `cat`/`vtysh` attempt is an
// actual FRR configuration rather than a shell error ("No such file or
// directory", "command not found"). Callers use it to decide whether the device
// runs FRR at all, which is not an error condition — most devices do not.
func frrConfigLooksReal(out string) bool {
	low := strings.ToLower(out)
	if strings.Contains(low, "no such file") || strings.Contains(low, "not found") {
		return false
	}
	return strings.Contains(low, "router ospf") || strings.Contains(low, "router bgp") ||
		strings.Contains(low, "router rip") || strings.Contains(low, "router isis") ||
		strings.Contains(low, "frr version")
}
