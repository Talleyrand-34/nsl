// vyos.go: the VyOS parser, driven by the declarative /config/config.boot.
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
	"fmt"
	"strconv"
	"strings"
	"time"

	"nsl-graph/internal/configparser"
	s "nsl-graph/internal/scanner"
)

// VyOSParser reads a VyOS router's intended configuration rather than its live
// state. VyOS is unusual among the vendors here in that its configuration is
// genuinely declarative: /config/config.boot is the whole device, committed and
// self-consistent, so there is nothing to reconstruct from a dozen show commands
// the way FreeBSD and OpenWrt need. That makes it the single source we go to.
//
// It also runs FRR underneath, but we deliberately do NOT reuse frr.go here: the
// FRR running-config on a VyOS box is generated output, and reading it would
// report what the daemon currently believes instead of what the operator wrote.
// The two disagree exactly when something is wrong, which is when it matters.
type VyOSParser struct{}

// NewVyOSParser creates a new VyOS configuration parser.
func NewVyOSParser() *VyOSParser {
	return &VyOSParser{}
}

func init() { configparser.DefaultRegistry.RegisterParser(NewVyOSParser()) }

// GetOsType returns the OS type key used for --os-type and registry lookup.
func (p *VyOSParser) GetOsType() string {
	return "vyos"
}

// SupportsDevice returns true for VyOS devices. VyOS reports a stock Linux
// sysDescr, so the distinguishing token is the word "vyos" itself — which shows
// up in the description on most builds and in the hostname on the rest.
func (p *VyOSParser) SupportsDevice(device s.SNMPDevice) bool {
	return strings.Contains(strings.ToLower(device.SysDescr), "vyos") ||
		strings.Contains(strings.ToLower(device.SysName), "vyos")
}

// Fetch reads the committed configuration from a VyOS host.
//
// /config/config.boot is the primary source and is a plain file, so a bare `cat`
// gets it without going anywhere near the VyOS CLI. That matters: the operational
// CLI is a shell function set up by an interactive login, so a plain `show
// configuration` over a non-interactive SSH session dies with
// "Invalid command: [show]". The fallback therefore calls the op-mode wrapper by
// its absolute path, which works without the interactive environment, and emits
// the flat `set …` command form that ParseConfig also understands.
//
// The hostname is prefixed the same way freebsd.go does it, so a config.boot with
// no `system host-name` (or an unreadable one) still yields a named device.
func (p *VyOSParser) Fetch(sess configparser.Session) (string, error) {
	hostname, err := sess.Execute("hostname")
	if err != nil {
		hostname = ""
	}
	hostname = strings.TrimSpace(hostname)

	out, err := sess.Execute("cat /config/config.boot")
	if err != nil || !vyosConfigLooksReal(out) {
		fallback, ferr := sess.Execute(
			"/opt/vyatta/bin/vyatta-op-cmd-wrapper show configuration commands")
		if ferr != nil || !vyosConfigLooksReal(fallback) {
			// Report the primary failure: an unreadable config.boot is the
			// actionable problem, the wrapper is only the consolation prize.
			if err != nil {
				return "", fmt.Errorf("failed to read /config/config.boot: %w", err)
			}
			return "", fmt.Errorf("no usable VyOS configuration returned by the device")
		}
		out = fallback
	}

	return fmt.Sprintf("HOSTNAME:%s\n%s", hostname, out), nil
}

// vyosConfigLooksReal reports whether a command's output is a configuration
// rather than a shell or CLI complaint. Both fetch paths can "succeed" at the
// exit-status level while returning "Permission denied" or "Invalid command",
// so the content is what decides.
func vyosConfigLooksReal(out string) bool {
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return false
	}
	low := strings.ToLower(trimmed)
	if strings.Contains(low, "no such file") || strings.Contains(low, "permission denied") ||
		strings.Contains(low, "invalid command") || strings.Contains(low, "command not found") {
		return false
	}
	return strings.Contains(low, "interfaces") || strings.Contains(low, "system") ||
		strings.Contains(low, "protocols")
}

// ParseConfig parses a VyOS configuration into structured ConfigData. It accepts
// the HOSTNAME-prefixed output of Fetch as well as a bare config.boot read from a
// file, and either of the two grammars VyOS speaks (see parseVyOSConfig).
func (p *VyOSParser) ParseConfig(rawConfig string, deviceInfo s.SNMPDevice) (*configparser.ConfigData, error) {
	// The prefix format is identical to FreeBSD's, so the same splitter serves.
	hostname, body := extractFreeBSDHostname(rawConfig)

	root := parseVyOSConfig(body)

	interfaces, vlans := parseVyOSInterfaces(root.block("interfaces"))
	protocols := root.block("protocols")

	// The device's own host-name wins: it is what the operator committed, while
	// the HOSTNAME prefix is only whatever the shell happened to report.
	var domain string
	if system := root.block("system"); system != nil {
		if h := system.leafValue("host-name"); h != "" {
			hostname = h
		}
		domain = system.leafValue("domain-name")
	}
	if hostname == "" {
		hostname = deviceInfo.SysName
	}

	return &configparser.ConfigData{
		OsType:           p.GetOsType(),
		DeviceModel:      "VyOS Router",
		Hostname:         hostname,
		Domain:           domain,
		Source:           configparser.ConfigSourceSSH,
		Interfaces:       interfaces,
		VLANs:            vlans,
		Routes:           parseVyOSStaticRoutes(protocols.block("static")),
		RoutingProtocols: parseVyOSRoutingProtocols(protocols),
		ParsedAt:         time.Now(),
		Raw:              rawConfig,
	}, nil
}

// ValidateConfig performs basic structural validation on parsed VyOS configuration.
func (p *VyOSParser) ValidateConfig(config *configparser.ConfigData) []error {
	var errs []error

	if config.Hostname == "" {
		errs = append(errs, fmt.Errorf("hostname is empty"))
	}

	seen := make(map[string]bool)
	for _, iface := range config.Interfaces {
		if iface.Name == "" {
			errs = append(errs, fmt.Errorf("interface with empty name"))
			continue
		}
		if seen[iface.Name] {
			errs = append(errs, fmt.Errorf("duplicate interface name: %s", iface.Name))
		}
		seen[iface.Name] = true
	}

	for _, vlan := range config.VLANs {
		if id, err := strconv.Atoi(vlan.ID); err == nil && (id < 1 || id > 4094) {
			errs = append(errs, fmt.Errorf("VLAN ID %s is out of valid range (1-4094)", vlan.ID))
		}
	}

	// A BGP peer without a remote AS never comes up, and an OSPF instance with no
	// router-id picks one at random from the interface addresses — which makes the
	// topology unstable across reboots. Both are worth reporting.
	for _, proto := range config.RoutingProtocols {
		if proto.Type == configparser.RoutingProtoBGP {
			if proto.LocalAS == "" {
				errs = append(errs, fmt.Errorf("bgp instance has no local AS"))
			}
			for _, n := range proto.Neighbors {
				if n.RemoteAS == "" {
					errs = append(errs, fmt.Errorf("bgp neighbor %s has no remote-as", n.Address))
				}
			}
		}
		if proto.Type == configparser.RoutingProtoOSPF && proto.RouterID == "" {
			errs = append(errs, fmt.Errorf("ospf instance has no router-id"))
		}
	}

	return errs
}

// ---------------------------------------------------------------------------
// The configuration tree
// ---------------------------------------------------------------------------

// VyOS presents the same configuration in two syntaxes, and a device can hand us
// either depending on which fetch path succeeded:
//
//	config.boot                      show configuration commands
//	  interfaces {                     set interfaces ethernet eth0 address '10.0.0.1/30'
//	      ethernet eth0 {
//	          address "10.0.0.1/30"
//	      }
//	  }
//
// They are the same tree written down differently, so both are parsed into one
// node type and every extractor below works on either. The only structural
// difference that survives is where a name lands: config.boot puts the keyword
// and the name on one line ("ethernet eth0"), the set form puts them on separate
// levels. namedChildren papers over exactly that, and nothing else has to care.

// vyosLeaf is a statement that carries a value rather than opening a block.
type vyosLeaf struct {
	key   string
	value string
}

// vyosNode is one block. Leaves and children are kept in file order because the
// order of `network` and `neighbor` statements is how an operator reads a config,
// and reordering them makes a diff against the device unreadable.
type vyosNode struct {
	words    []string // the tokens naming this block: ["ethernet", "eth0"]
	leaves   []vyosLeaf
	children []*vyosNode
}

// vyosNamed is a block together with the name it was declared under.
type vyosNamed struct {
	name string
	node *vyosNode
}

// block returns the single-word child block named kw, or nil. Every accessor
// tolerates a nil receiver so that a missing section — most devices have no
// `protocols` at all — reads as an absence rather than needing a guard at each
// call site.
func (n *vyosNode) block(kw string) *vyosNode {
	if n == nil {
		return nil
	}
	for _, c := range n.children {
		if len(c.words) == 1 && c.words[0] == kw {
			return c
		}
	}
	return nil
}

// namedChildren returns the blocks introduced by the keyword kw, paired with the
// name each was declared under: "ethernet eth0 { … }" yields ("eth0", block).
// This is the one place the two grammars differ, so it handles both — the set
// form splits the keyword and the name across two levels.
func (n *vyosNode) namedChildren(kw string) []vyosNamed {
	if n == nil {
		return nil
	}
	var out []vyosNamed
	for _, c := range n.children {
		switch {
		case len(c.words) >= 2 && c.words[0] == kw:
			out = append(out, vyosNamed{name: c.words[1], node: c})
		case len(c.words) == 1 && c.words[0] == kw:
			for _, gc := range c.children {
				if len(gc.words) >= 1 {
					out = append(out, vyosNamed{name: gc.words[0], node: gc})
				}
			}
		}
	}
	return out
}

// namedValues returns the names declared under kw whether they arrived as blocks
// or as leaf values. VyOS writes `network 10.1.1.0/24 { }` under a BGP address
// family but `network "10.1.1.0/24"` under an OSPF area, and both mean the same
// thing to us.
func (n *vyosNode) namedValues(kw string) []string {
	if n == nil {
		return nil
	}
	out := n.leafValues(kw)
	for _, c := range n.namedChildren(kw) {
		out = append(out, c.name)
	}
	return out
}

// leafValues returns every value assigned to kw, in order; VyOS repeats the key
// for list-valued settings such as `address` and `network`.
func (n *vyosNode) leafValues(kw string) []string {
	if n == nil {
		return nil
	}
	var out []string
	for _, l := range n.leaves {
		if l.key == kw && l.value != "" {
			out = append(out, l.value)
		}
	}
	return out
}

// leafValue returns the first value assigned to kw, or "".
func (n *vyosNode) leafValue(kw string) string {
	if v := n.leafValues(kw); len(v) > 0 {
		return v[0]
	}
	return ""
}

// hasKeyword reports whether kw appears directly inside this block. Valueless
// settings are written as an empty block in config.boot (`disable { }` or just
// `disable`) and as a bare path in the set form, so both shapes count.
func (n *vyosNode) hasKeyword(kw string) bool {
	if n == nil {
		return false
	}
	for _, l := range n.leaves {
		if l.key == kw {
			return true
		}
	}
	return n.block(kw) != nil
}

// hasKeywordDeep reports whether kw appears anywhere beneath this block. Used for
// settings whose depth is a moving target across VyOS releases — `nexthop-self`
// sat directly under a neighbor before address families were mandatory, and sits
// two levels down now.
func (n *vyosNode) hasKeywordDeep(kw string) bool {
	if n == nil {
		return false
	}
	if n.hasKeyword(kw) {
		return true
	}
	for _, c := range n.children {
		if c.hasKeywordDeep(kw) {
			return true
		}
	}
	return false
}

// ensureChild finds or creates a single-word child, for building the set-command
// tree one path token at a time.
func (n *vyosNode) ensureChild(word string) *vyosNode {
	if c := n.block(word); c != nil {
		return c
	}
	c := &vyosNode{words: []string{word}}
	n.children = append(n.children, c)
	return c
}

// parseVyOSConfig parses either grammar into a tree, choosing between them by
// looking for a block opener. config.boot always has one; the flat set-command
// form never does.
func parseVyOSConfig(raw string) *vyosNode {
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		if strings.HasSuffix(strings.TrimSpace(line), "{") {
			root := &vyosNode{}
			i := 0
			parseVyOSBraces(lines, &i, root)
			return root
		}
	}
	return parseVyOSSetCommands(lines)
}

// parseVyOSBraces fills parent from lines[*i:] until the block's closing brace,
// recursing for each nested block. A trailing `// vyos-config-version: …` banner
// and blank lines are skipped.
func parseVyOSBraces(lines []string, i *int, parent *vyosNode) {
	for *i < len(lines) {
		line := strings.TrimSpace(strings.TrimRight(lines[*i], "\r"))
		*i++

		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "}" {
			return
		}
		if strings.HasSuffix(line, "{") {
			child := &vyosNode{words: strings.Fields(strings.TrimSuffix(line, "{"))}
			parent.children = append(parent.children, child)
			parseVyOSBraces(lines, i, child)
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		leaf := vyosLeaf{key: fields[0]}
		if len(fields) > 1 {
			leaf.value = stripVyOSQuotes(strings.Join(fields[1:], " "))
		}
		parent.leaves = append(parent.leaves, leaf)
	}
}

// parseVyOSSetCommands builds the same tree from the flat `set …` form.
//
// The quoting is what tells a value from another path component: VyOS quotes leaf
// values and never quotes path tokens, so `set … eth0 address '10.0.0.1/30'` ends
// in a leaf while `set … route 0.0.0.0/0 next-hop 10.0.50.1` is nothing but path.
func parseVyOSSetCommands(lines []string) *vyosNode {
	root := &vyosNode{}

	for _, raw := range lines {
		line := strings.TrimSpace(strings.TrimRight(raw, "\r"))
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		line = strings.TrimPrefix(line, "set ")

		// Peel a trailing quoted value off the end. Scanning back to the matching
		// quote (rather than splitting on spaces) keeps values such as a
		// description with spaces in one piece.
		value, hasValue := "", false
		if last := len(line) - 1; last >= 0 && (line[last] == '\'' || line[last] == '"') {
			if open := strings.LastIndexByte(line[:last], line[last]); open >= 0 {
				value, hasValue = line[open+1:last], true
				line = strings.TrimSpace(line[:open])
			}
		}

		path := strings.Fields(line)
		if len(path) == 0 {
			continue
		}

		if !hasValue {
			// A valueless setting: the whole path is blocks, the last of which is
			// the flag itself (`… eth5 disable`).
			node := root
			for _, tok := range path {
				node = node.ensureChild(tok)
			}
			continue
		}

		node := root
		for _, tok := range path[:len(path)-1] {
			node = node.ensureChild(tok)
		}
		node.leaves = append(node.leaves, vyosLeaf{key: path[len(path)-1], value: value})
	}

	return root
}

// stripVyOSQuotes removes a matched surrounding pair of single or double quotes.
// config.boot double-quotes values, the set form single-quotes them, and older
// releases quoted neither.
func stripVyOSQuotes(v string) string {
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}
	return v
}

// ---------------------------------------------------------------------------
// interfaces
// ---------------------------------------------------------------------------

// vyosIfaceType maps a VyOS interface keyword to the Type vocabulary the other
// parsers already use. A `dummy` is reported as a loopback rather than inventing
// a type for it: it exists for the same reason lo does — to hold an address that
// no physical link can take down — and downstream consumers should treat it alike.
func vyosIfaceType(kind string) string {
	switch kind {
	case "ethernet":
		return "physical"
	case "loopback", "dummy":
		return "loopback"
	case "bridge":
		return "bridge"
	case "tunnel", "vti", "wireguard", "vxlan", "geneve", "l2tpv3":
		return "tunnel"
	default:
		return "logical"
	}
}

// parseVyOSInterfaces walks the `interfaces` section, emitting one
// ConfigInterface per interface plus one per VLAN subinterface.
//
// VyOS nests VLANs inside their parent (`ethernet eth0 { vif 10 { … } }`) rather
// than declaring them alongside it, but the rest of the tool expects the flat
// Linux-style view — eth0.10 as an interface in its own right, with eth0 as its
// parent — so the nesting is unrolled here.
func parseVyOSInterfaces(ifaces *vyosNode) ([]configparser.ConfigInterface, []configparser.ConfigVLAN) {
	if ifaces == nil {
		return nil, nil
	}

	var out []configparser.ConfigInterface
	var vlans []configparser.ConfigVLAN
	seenVLAN := make(map[string]bool)

	// The interface kinds are whatever the config declares, so they are collected
	// from the tree rather than from a fixed list — a device with a `wireguard`
	// section should not be silently truncated to its ethernets.
	for _, kind := range vyosChildKeywords(ifaces) {
		for _, decl := range ifaces.namedChildren(kind) {
			iface := vyosInterfaceFrom(decl.name, vyosIfaceType(kind), decl.node)
			out = append(out, iface)

			for _, vif := range decl.node.namedChildren("vif") {
				sub := vyosInterfaceFrom(decl.name+"."+vif.name, "vlan", vif.node)
				sub.Parent = decl.name
				sub.VLANs = []configparser.ConfigVLAN{
					{ID: vif.name, Tagged: true, Enabled: sub.Enabled},
				}
				out = append(out, sub)

				if seenVLAN[vif.name] {
					continue
				}
				seenVLAN[vif.name] = true
				vlan := configparser.ConfigVLAN{
					ID:          vif.name,
					Name:        sub.Name,
					Description: sub.Description,
					Enabled:     sub.Enabled,
					Tagged:      true,
				}
				if len(sub.IPAddresses) > 0 {
					vlan.IPRange = sub.IPAddresses[0]
				}
				vlans = append(vlans, vlan)
			}
		}
	}

	return out, vlans
}

// vyosChildKeywords lists the distinct leading keywords of a block's children, in
// first-appearance order.
func vyosChildKeywords(n *vyosNode) []string {
	var out []string
	seen := make(map[string]bool)
	for _, c := range n.children {
		if len(c.words) == 0 || seen[c.words[0]] {
			continue
		}
		seen[c.words[0]] = true
		out = append(out, c.words[0])
	}
	return out
}

// vyosInterfaceFrom builds a ConfigInterface from an interface or vif block.
//
// VyOS states only the departures from the default, so an interface is enabled
// unless it says `disable`.
func vyosInterfaceFrom(name, ifaceType string, node *vyosNode) configparser.ConfigInterface {
	iface := configparser.ConfigInterface{
		Name:        name,
		Type:        ifaceType,
		Description: node.leafValue("description"),
		MACAddress:  node.leafValue("hw-id"),
		Enabled:     !node.hasKeyword("disable"),
	}

	for _, addr := range node.leafValues("address") {
		// "dhcp"/"dhcpv6" are address *methods*, not addresses; the resulting
		// lease is live state and is not in this file.
		if addr == "dhcp" || addr == "dhcpv6" {
			continue
		}
		iface.IPAddresses = append(iface.IPAddresses, addr)
	}

	if mtu, err := strconv.Atoi(node.leafValue("mtu")); err == nil {
		iface.MTU = mtu
	}

	return iface
}

// ---------------------------------------------------------------------------
// routing
// ---------------------------------------------------------------------------

// parseVyOSRoutingProtocols extracts the dynamic routing instances from the
// `protocols` section. Static routes are not included: they are not a protocol
// instance and live in ConfigData.Routes instead.
func parseVyOSRoutingProtocols(protocols *vyosNode) []configparser.ConfigRoutingProtocol {
	if protocols == nil {
		return nil
	}

	var out []configparser.ConfigRoutingProtocol
	if ospf := parseVyOSOSPF(protocols); ospf != nil {
		out = append(out, *ospf)
	}
	if bgp := parseVyOSBGP(protocols); bgp != nil {
		out = append(out, *bgp)
	}
	return out
}

// parseVyOSOSPF reads `protocols ospf`.
//
// The area's networks are the interesting part: VyOS 1.3 and earlier enable OSPF
// by matching interface addresses against `network` prefixes inside an area,
// while 1.4 also allows naming an interface directly. Both are recorded, because
// which one a config uses tells you how it was written.
func parseVyOSOSPF(protocols *vyosNode) *configparser.ConfigRoutingProtocol {
	ospf := protocols.block("ospf")
	if ospf == nil {
		return nil
	}

	proto := configparser.ConfigRoutingProtocol{
		Type:     configparser.RoutingProtoOSPF,
		Enabled:  true,
		RouterID: ospf.block("parameters").leafValue("router-id"),
	}

	areas := make(map[string]int) // area ID → index in proto.Areas
	areaAt := func(id string) *configparser.ConfigOSPFArea {
		if i, ok := areas[id]; ok {
			return &proto.Areas[i]
		}
		proto.Areas = append(proto.Areas, configparser.ConfigOSPFArea{ID: id})
		areas[id] = len(proto.Areas) - 1
		return &proto.Areas[len(proto.Areas)-1]
	}

	for _, decl := range ospf.namedChildren("area") {
		area := areaAt(decl.name)
		area.Networks = append(area.Networks, decl.node.namedValues("network")...)
		area.Interfaces = append(area.Interfaces, decl.node.namedValues("interface")...)
		// `area-type { stub { } }`, or the flatter `area-type "stub"`.
		if at := decl.node.block("area-type"); at != nil {
			if kinds := vyosChildKeywords(at); len(kinds) > 0 {
				area.Type = kinds[0]
			}
		} else if at := decl.node.leafValue("area-type"); at != "" {
			area.Type = at
		}
	}

	// VyOS 1.4 binds interfaces to areas from the interface side instead.
	for _, decl := range ospf.namedChildren("interface") {
		proto.Interfaces = append(proto.Interfaces, decl.name)
		if id := decl.node.leafValue("area"); id != "" {
			area := areaAt(id)
			area.Interfaces = append(area.Interfaces, decl.name)
		}
	}

	proto.Redistribute = vyosRedistribute(ospf)

	return &proto
}

// parseVyOSBGP reads `protocols bgp`.
//
// The local AS moved between releases: it used to be part of the block header
// (`protocols bgp 65000 { … }`) and is now a `system-as` setting inside it. Both
// spellings are accepted, since a lab upgraded in place will have configs of
// either vintage lying around.
func parseVyOSBGP(protocols *vyosNode) *configparser.ConfigRoutingProtocol {
	var bgp *vyosNode
	var localAS string

	if b := protocols.block("bgp"); b != nil {
		bgp = b
	} else if named := protocols.namedChildren("bgp"); len(named) > 0 {
		bgp, localAS = named[0].node, named[0].name
	}
	if bgp == nil {
		return nil
	}
	if as := bgp.leafValue("system-as"); as != "" {
		localAS = as
	}

	proto := configparser.ConfigRoutingProtocol{
		Type:     configparser.RoutingProtoBGP,
		Enabled:  true,
		LocalAS:  localAS,
		RouterID: bgp.block("parameters").leafValue("router-id"),
	}

	// Prefixes the instance originates, stated per address family.
	if af := bgp.block("address-family"); af != nil {
		for _, family := range vyosChildKeywords(af) {
			proto.Networks = append(proto.Networks, af.block(family).namedValues("network")...)
		}
	}

	for _, decl := range bgp.namedChildren("neighbor") {
		proto.Neighbors = append(proto.Neighbors, configparser.ConfigBGPNeighbor{
			Address:      decl.name,
			RemoteAS:     decl.node.leafValue("remote-as"),
			LocalAS:      decl.node.leafValue("local-as"),
			Description:  decl.node.leafValue("description"),
			UpdateSource: decl.node.leafValue("update-source"),
			NextHopSelf:  decl.node.hasKeywordDeep("nexthop-self"),
		})
	}

	proto.Redistribute = vyosRedistribute(bgp)

	return &proto
}

// vyosRedistribute lists the protocols redistributed into an instance, looking
// both directly under the block and one address family down, which is where BGP
// puts them.
func vyosRedistribute(node *vyosNode) []string {
	var out []string
	collect := func(n *vyosNode) {
		if r := n.block("redistribute"); r != nil {
			out = append(out, vyosChildKeywords(r)...)
			for _, l := range r.leaves {
				out = append(out, l.key)
			}
		}
	}

	collect(node)
	if af := node.block("address-family"); af != nil {
		for _, family := range vyosChildKeywords(af) {
			collect(af.block(family))
		}
	}
	return out
}

// parseVyOSStaticRoutes reads `protocols static`, emitting one ConfigRoute per
// next hop — a destination with two next hops is two routes, which is what the
// forwarding table will hold.
//
// A destination with neither a next hop nor an interface (a blackhole, or a
// `reject` route) still produces a route: the point of the entry is that traffic
// for that prefix stops here, and dropping it would hide that.
func parseVyOSStaticRoutes(static *vyosNode) []configparser.ConfigRoute {
	if static == nil {
		return nil
	}

	var routes []configparser.ConfigRoute

	for _, kw := range []string{"route", "route6"} {
		for _, decl := range static.namedChildren(kw) {
			base := configparser.ConfigRoute{
				Network:     decl.name,
				Protocol:    configparser.RoutingProtoStatic,
				Description: decl.node.leafValue("description"),
			}

			emitted := false

			for _, hop := range decl.node.namedChildren("next-hop") {
				route := base
				route.Gateway = hop.name
				// VyOS calls the metric "distance"; it is the same knob.
				if d, err := strconv.Atoi(hop.node.leafValue("distance")); err == nil {
					route.Metric = d
				}
				// A next hop may be pinned to an egress interface.
				route.Interface = hop.node.leafValue("interface")
				routes = append(routes, route)
				emitted = true
			}
			// Older configs write the next hop as a plain value.
			for _, hop := range decl.node.leafValues("next-hop") {
				route := base
				route.Gateway = hop
				routes = append(routes, route)
				emitted = true
			}

			// Interface routes: the destination is reachable directly out of a link.
			for _, ifname := range decl.node.namedValues("interface") {
				route := base
				route.Interface = ifname
				routes = append(routes, route)
				emitted = true
			}

			if !emitted {
				routes = append(routes, base)
			}
		}
	}

	return routes
}
