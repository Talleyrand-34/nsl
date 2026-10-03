// routeros.go: MikroTik RouterOS, read through `print terse`.
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

// RouterOSParser reads a MikroTik through its own CLI rather than a config file.
//
// RouterOS has no /etc: the configuration lives in a proprietary store and the
// only text rendering of it is `export`, which is pretty-printed, line-wrapped
// and full of `add`/`set` verbs — a grammar of its own. `print terse` instead
// returns one record per line as `<index> [FLAGS] key=value …`, which is the
// same shape for every menu in the tree. Paying the cost of one good terse
// parser buys interfaces, addresses, VLANs, routes, OSPF and BGP for free,
// which is why this parser runs ten narrow commands instead of one big export.
type RouterOSParser struct{}

// NewRouterOSParser creates a new RouterOS configuration parser.
func NewRouterOSParser() *RouterOSParser {
	return &RouterOSParser{}
}

func init() { configparser.DefaultRegistry.RegisterParser(NewRouterOSParser()) }

// routerosCommands are the menus fetched, in the order they are written into the
// raw blob. Each is tagged with "# <command>" so ParseConfig can re-split them,
// exactly as the OpenWrt parser does with its uci commands.
var routerosCommands = []string{
	"/system/identity/print",
	"/interface/print terse without-paging",
	"/ip/address/print terse without-paging",
	"/interface/vlan/print terse without-paging",
	"/interface/bridge/port/print terse without-paging",
	"/ip/route/print terse without-paging",
	"/routing/ospf/instance/print terse without-paging",
	"/routing/ospf/area/print terse without-paging",
	"/routing/ospf/interface-template/print terse without-paging",
	"/routing/bgp/connection/print terse without-paging",
}

// GetOsType returns the OS type key used for --os-type and registry lookup.
//
// "routeros" and not "mikrotik": the parser understands an operating system, not
// a vendor, and MikroTik's SwOS boxes speak nothing like this. Auto-detection
// (SupportsDevice) accepts either word, so a device that only ever says
// "MikroTik" in its sysDescr is still matched.
func (p *RouterOSParser) GetOsType() string {
	return "routeros"
}

// SupportsDevice returns true for MikroTik devices running RouterOS.
//
// The sysDescr of a stock RouterOS box is literally "RouterOS <version>" and
// carries no vendor name, while a customised one often carries only "MikroTik";
// both spellings therefore have to match, in either field.
func (p *RouterOSParser) SupportsDevice(device s.SNMPDevice) bool {
	haystack := strings.ToLower(device.SysDescr + " " + device.SysName)
	return strings.Contains(haystack, "routeros") || strings.Contains(haystack, "mikrotik")
}

// Fetch runs each menu's print command and tags its output with "# <command>".
//
// A failing command is deliberately not fatal. /routing/ospf and /routing/bgp
// only exist when the routing package is installed and, on RouterOS 7, the paths
// moved between releases — a box without them is a perfectly valid statically
// routed device, not a fetch failure. The error is recorded in-band so the
// operator can see which menu was missing, and everything else still parses.
func (p *RouterOSParser) Fetch(sess configparser.Session) (string, error) {
	var out strings.Builder
	for _, cmd := range routerosCommands {
		output, err := sess.Execute(cmd)
		if err != nil {
			out.WriteString(fmt.Sprintf("# Error executing %s: %v\n", cmd, err))
			continue
		}
		out.WriteString(fmt.Sprintf("# %s\n", cmd))
		out.WriteString(output)
		out.WriteString("\n")
	}
	return out.String(), nil
}

// ParseConfig turns the tagged command blocks produced by Fetch into ConfigData.
func (p *RouterOSParser) ParseConfig(rawConfig string, deviceInfo s.SNMPDevice) (*configparser.ConfigData, error) {
	blocks := splitRouterOSBlocks(rawConfig)

	hostname := parseRouterOSIdentity(blocks["/system/identity/print"])
	if hostname == "" {
		hostname = deviceInfo.SysName
	}

	config := &configparser.ConfigData{
		OsType:      p.GetOsType(),
		DeviceModel: "MikroTik RouterOS",
		Hostname:    hostname,
		Source:      configparser.ConfigSourceSSH,
		ParsedAt:    time.Now(),
		Raw:         rawConfig,
	}

	interfaces, vlans := parseRouterOSInterfaces(
		parseTerseRecords(blocks["/interface/print terse without-paging"]),
		parseTerseRecords(blocks["/ip/address/print terse without-paging"]),
		parseTerseRecords(blocks["/interface/vlan/print terse without-paging"]),
		parseTerseRecords(blocks["/interface/bridge/port/print terse without-paging"]),
	)
	config.Interfaces = interfaces
	config.VLANs = vlans

	config.Routes = parseRouterOSRoutes(parseTerseRecords(blocks["/ip/route/print terse without-paging"]))

	config.RoutingProtocols = append(config.RoutingProtocols, parseRouterOSOSPF(
		parseTerseRecords(blocks["/routing/ospf/instance/print terse without-paging"]),
		parseTerseRecords(blocks["/routing/ospf/area/print terse without-paging"]),
		parseTerseRecords(blocks["/routing/ospf/interface-template/print terse without-paging"]),
	)...)
	config.RoutingProtocols = append(config.RoutingProtocols, parseRouterOSBGP(
		parseTerseRecords(blocks["/routing/bgp/connection/print terse without-paging"]),
	)...)

	return config, nil
}

// ValidateConfig performs basic structural validation on parsed RouterOS data.
func (p *RouterOSParser) ValidateConfig(config *configparser.ConfigData) []error {
	var errs []error

	if config.Hostname == "" {
		errs = append(errs, fmt.Errorf("hostname is empty"))
	}

	seen := make(map[string]bool, len(config.Interfaces))
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
		id, err := strconv.Atoi(vlan.ID)
		if err != nil {
			errs = append(errs, fmt.Errorf("non-numeric VLAN ID: %s", vlan.ID))
			continue
		}
		if id < 1 || id > 4094 {
			errs = append(errs, fmt.Errorf("VLAN ID %s is out of valid range (1-4094)", vlan.ID))
		}
	}

	for _, proto := range config.RoutingProtocols {
		switch proto.Type {
		case configparser.RoutingProtoOSPF:
			if proto.RouterID == "" {
				errs = append(errs, fmt.Errorf("OSPF instance %q has no router-id", proto.Instance))
			}
		case configparser.RoutingProtoBGP:
			if proto.LocalAS == "" {
				errs = append(errs, fmt.Errorf("BGP instance has no local AS"))
			}
			for _, n := range proto.Neighbors {
				if n.Address == "" {
					errs = append(errs, fmt.Errorf("BGP neighbor %q has no address", n.Description))
				}
			}
		}
	}

	return errs
}

// ---------------------------------------------------------------------------
// Terse-record parsing
// ---------------------------------------------------------------------------

// terseRecord is one line of `print terse` output: an index, zero or more flag
// letters, and the key=value attributes.
type terseRecord struct {
	Index string
	Flags string
	Attrs map[string]string
}

// Get returns an attribute, or "" when absent.
func (r terseRecord) Get(key string) string { return r.Attrs[key] }

// Disabled reports whether RouterOS flagged the record as administratively down.
// The flag letter is "X" in every menu that has one.
func (r terseRecord) Disabled() bool { return strings.ContainsRune(r.Flags, 'X') }

// ParseRouterOSTerse parses `print terse` output into one attribute map per
// record, discarding the index and flags.
//
// It is exported for the topology collector, which reads `/ip/neighbor` off the
// same CLI to discover adjacencies. The grammar is not obvious — values may
// contain spaces, and a console-wrapped record continues on the next line, even
// mid-token — so having two implementations of it would be two chances to get it
// subtly wrong on a different menu.
func ParseRouterOSTerse(block string) []map[string]string {
	recs := parseTerseRecords(block)
	out := make([]map[string]string, 0, len(recs))
	for _, r := range recs {
		out = append(out, r.Attrs)
	}
	return out
}

// splitRouterOSBlocks re-splits the blob Fetch assembled, keyed by the command
// that produced each block.
//
// "# Error executing …" lines become a block whose key nothing looks up, which
// is exactly the wanted behaviour: a menu that failed simply yields no records.
func splitRouterOSBlocks(raw string) map[string]string {
	blocks := make(map[string]string, len(routerosCommands))

	var current string
	var body strings.Builder
	flush := func() {
		if current != "" {
			blocks[current] = body.String()
		}
		body.Reset()
	}

	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "# ") {
			flush()
			current = strings.TrimSpace(line[2:])
			continue
		}
		if current != "" {
			body.WriteString(line)
			body.WriteByte('\n')
		}
	}
	flush()

	return blocks
}

// parseTerseRecords parses a block of `print terse` output into records.
//
// The continuation handling is the defensive part. Driven from a console,
// RouterOS wraps print output at the terminal width and will happily break a
// record mid-token, so the tail arrives on the following line with no marker of
// any kind. Over SSH — the transport this parser actually uses — the pty is
// wide enough that each record is a single line, but the cost of being wrong
// here is a silently truncated record, so the rule is applied anyway: a line
// that does not start with a record index continues the previous one and is
// concatenated with NO separator, because the break may have landed inside a
// token and inserting a space would corrupt the value.
func parseTerseRecords(block string) []terseRecord {
	var records []terseRecord
	var pending string

	flush := func() {
		if pending == "" {
			return
		}
		if rec, ok := parseTerseRecord(pending); ok {
			records = append(records, rec)
		}
		pending = ""
	}

	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimRight(line, " \t\r")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}
		// Legends and column headers that `print` emits when it feels chatty are
		// not records, and must not be mistaken for continuations either.
		if head := strings.TrimSpace(trimmed); strings.HasPrefix(head, "#") ||
			strings.HasPrefix(head, "Flags:") || strings.HasPrefix(head, "Columns:") {
			continue
		}

		if startsWithRecordIndex(trimmed) {
			flush()
			pending = strings.TrimLeft(trimmed, " \t")
			continue
		}
		if pending != "" {
			pending += strings.TrimLeft(trimmed, " \t")
		}
	}
	flush()

	return records
}

// startsWithRecordIndex reports whether a line opens a new terse record, i.e.
// begins with the decimal index RouterOS prints in the first column.
func startsWithRecordIndex(line string) bool {
	line = strings.TrimLeft(line, " \t")
	digits := 0
	for digits < len(line) && line[digits] >= '0' && line[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return false
	}
	// A bare index with nothing after it is not a record; a wrapped value that
	// happens to start with digits is caught by requiring the separator.
	return digits < len(line) && (line[digits] == ' ' || line[digits] == '\t')
}

// parseTerseRecord splits one assembled record into index, flags and attributes.
//
// The index/flags prefix is found by locating the first "=" and walking back to
// the start of its token: everything before that is the prefix. Splitting on
// whitespace instead would break, because attribute values may contain spaces
// (last-link-up-time=jul/20/2026 00:01:13).
func parseTerseRecord(line string) (terseRecord, bool) {
	eq := strings.IndexByte(line, '=')
	prefix, rest := line, ""
	if eq >= 0 {
		keyStart := strings.LastIndexAny(line[:eq], " \t") + 1
		prefix, rest = line[:keyStart], line[keyStart:]
	}

	fields := strings.Fields(prefix)
	if len(fields) == 0 {
		return terseRecord{}, false
	}

	rec := terseRecord{
		Index: fields[0],
		Flags: strings.Join(fields[1:], ""),
		Attrs: parseTerseAttrs(rest),
	}
	return rec, true
}

// parseTerseAttrs scans a "key=value key=value …" run.
//
// Values are unquoted and may contain spaces, so a value ends only where the
// next key demonstrably begins: at whitespace followed by a bare key token and
// an "=". Values that RouterOS does quote (comments, anything containing an "=")
// are read to their closing quote instead.
func parseTerseAttrs(text string) map[string]string {
	attrs := make(map[string]string)

	for i := 0; i < len(text); {
		for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
			i++
		}
		if i >= len(text) {
			break
		}

		eq := strings.IndexByte(text[i:], '=')
		if eq < 0 {
			break
		}
		key := text[i : i+eq]
		i += eq + 1

		var value string
		if i < len(text) && text[i] == '"' {
			j := i + 1
			for j < len(text) && text[j] != '"' {
				if text[j] == '\\' {
					j++
				}
				j++
			}
			value = text[i+1 : min(j, len(text))]
			i = min(j+1, len(text))
		} else {
			end := nextTerseKeyStart(text, i)
			value = strings.TrimRight(text[i:end], " \t")
			i = end
		}

		if key != "" {
			attrs[key] = value
		}
	}

	return attrs
}

// nextTerseKeyStart returns the offset of the whitespace that separates the
// current value from the next "key=" token, or len(text) when the value runs to
// the end of the record.
func nextTerseKeyStart(text string, from int) int {
	for i := from; i < len(text); i++ {
		if text[i] != ' ' && text[i] != '\t' {
			continue
		}
		j := i
		for j < len(text) && (text[j] == ' ' || text[j] == '\t') {
			j++
		}
		k := j
		for k < len(text) && isTerseKeyByte(text[k]) {
			k++
		}
		if k > j && k < len(text) && text[k] == '=' {
			return i
		}
	}
	return len(text)
}

// isTerseKeyByte reports whether c may appear in a terse attribute name.
// RouterOS 7 keys are dotted for nested properties (remote.address, output.network).
func isTerseKeyByte(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	case c == '-' || c == '.' || c == '_':
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// Menu-specific parsing
// ---------------------------------------------------------------------------

// parseRouterOSIdentity reads the device name from /system/identity/print.
//
// That menu has no terse form worth using: it prints a single "  name: X" line
// in the colon style, so it is handled apart from every other block. The
// key=value spelling is accepted too, for when the command was run with terse.
func parseRouterOSIdentity(block string) string {
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "name:"); idx >= 0 {
			return strings.TrimSpace(line[idx+len("name:"):])
		}
		if idx := strings.Index(line, "name="); idx >= 0 {
			return strings.Fields(line[idx+len("name="):])[0]
		}
	}
	return ""
}

// parseRouterOSInterfaces builds the interface list from the four interface-ish
// menus: /interface (the netdev list), /ip/address, /interface/vlan and
// /interface/bridge/port.
//
// /interface already lists VLAN and bridge members, so the latter three enrich
// records that usually exist rather than creating them; they only create when
// the device answered one menu and not another.
func parseRouterOSInterfaces(ifaceRecs, addrRecs, vlanRecs, bridgePortRecs []terseRecord) ([]configparser.ConfigInterface, []configparser.ConfigVLAN) {
	var interfaces []configparser.ConfigInterface
	index := make(map[string]int, len(ifaceRecs))

	ensure := func(name string) int {
		if name == "" {
			return -1
		}
		if i, ok := index[name]; ok {
			return i
		}
		interfaces = append(interfaces, configparser.ConfigInterface{
			Name:    name,
			Enabled: true,
			Type:    "logical",
		})
		index[name] = len(interfaces) - 1
		return len(interfaces) - 1
	}

	for _, rec := range ifaceRecs {
		i := ensure(rec.Get("name"))
		if i < 0 {
			continue
		}
		iface := &interfaces[i]
		iface.Enabled = !rec.Disabled()
		iface.Description = rec.Get("comment")
		iface.MACAddress = rec.Get("mac-address")
		iface.Type = routerosIfaceType(rec.Get("type"))
		// "mtu=auto" is common on bridges; the resolved figure is in actual-mtu.
		if mtu, ok := routerosMTU(rec.Get("mtu")); ok {
			iface.MTU = mtu
		} else if mtu, ok := routerosMTU(rec.Get("actual-mtu")); ok {
			iface.MTU = mtu
		}
	}

	for _, rec := range addrRecs {
		addr := rec.Get("address")
		if addr == "" {
			continue
		}
		// actual-interface resolves the address to the netdev it really landed
		// on when "interface" names a bridge member or a disabled interface.
		name := rec.Get("interface")
		if name == "" {
			name = rec.Get("actual-interface")
		}
		i := ensure(name)
		if i < 0 {
			continue
		}
		interfaces[i].IPAddresses = append(interfaces[i].IPAddresses, addr)
	}

	var vlans []configparser.ConfigVLAN
	for _, rec := range vlanRecs {
		name, vid := rec.Get("name"), rec.Get("vlan-id")
		i := ensure(name)
		if i < 0 || vid == "" {
			continue
		}
		iface := &interfaces[i]
		iface.Type = "vlan"
		iface.Parent = rec.Get("interface")
		iface.VLANs = append(iface.VLANs, configparser.ConfigVLAN{
			ID:      vid,
			Name:    name,
			Enabled: !rec.Disabled(),
			Tagged:  true,
		})
		vlans = append(vlans, configparser.ConfigVLAN{
			ID:      vid,
			Name:    name,
			Enabled: !rec.Disabled(),
			Tagged:  true,
		})
	}

	// Bridge membership is the only place RouterOS states the L2 parent of a
	// physical port, so it is what turns a flat interface list into a topology.
	for _, rec := range bridgePortRecs {
		member, bridge := rec.Get("interface"), rec.Get("bridge")
		if member == "" || bridge == "" {
			continue
		}
		if i := ensure(member); i >= 0 && interfaces[i].Parent == "" {
			interfaces[i].Parent = bridge
		}
	}

	return interfaces, vlans
}

// routerosIfaceType maps a RouterOS interface type to the tool's vocabulary.
func routerosIfaceType(t string) string {
	switch t {
	case "ether":
		return "physical"
	case "bridge":
		return "bridge"
	case "vlan":
		return "vlan"
	case "":
		return "logical"
	case "wg", "gre-tunnel", "ipip-tunnel", "eoip", "l2tp-out", "ovpn-out", "vpls":
		return "tunnel"
	default:
		return "logical"
	}
}

// routerosMTU parses an MTU field, rejecting the symbolic values ("auto") that
// RouterOS puts there when the figure is inherited rather than configured.
func routerosMTU(v string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// parseRouterOSRoutes converts /ip/route records into ConfigRoute.
//
// /ip/route is a RIB dump, not a config stanza, so unlike the file-based parsers
// it can say how each route was learned — and that is the flag letters, not any
// attribute. Only records with a destination are emitted: the menu also carries
// placeholder rows for unresolved nexthops.
func parseRouterOSRoutes(recs []terseRecord) []configparser.ConfigRoute {
	var routes []configparser.ConfigRoute

	for _, rec := range recs {
		dst := rec.Get("dst-address")
		if dst == "" {
			continue
		}

		route := configparser.ConfigRoute{
			Network:     dst,
			Gateway:     rec.Get("gateway"),
			Protocol:    routerosRouteProtocol(rec.Flags),
			Description: rec.Get("comment"),
		}

		// A recursive nexthop is printed as "10.0.0.1%ether1"; the part after
		// the % is the resolved outgoing interface.
		if gw, ifname, found := strings.Cut(route.Gateway, "%"); found {
			route.Gateway, route.Interface = gw, ifname
		}
		if route.Interface == "" {
			route.Interface = rec.Get("immediate-gw")
			if _, ifname, found := strings.Cut(route.Interface, "%"); found {
				route.Interface = ifname
			} else if strings.ContainsAny(route.Interface, ".:") {
				route.Interface = "" // an address, not an interface name
			}
		}
		if d, err := strconv.Atoi(rec.Get("distance")); err == nil {
			route.Metric = d
		}

		routes = append(routes, route)
	}

	return routes
}

// routerosRouteProtocol reads the route source out of the terse flag letters.
//
// RouterOS 7 prints them run together ("DAc", "As", "DAo", "DAb"): the uppercase
// letters are state (Dynamic, Active, X=disabled) and the lowercase one is the
// origin. Matching on the lowercase letter alone therefore works regardless of
// which state letters accompany it.
func routerosRouteProtocol(flags string) string {
	switch {
	case strings.ContainsRune(flags, 'c'):
		return "connected"
	case strings.ContainsRune(flags, 'o'):
		return configparser.RoutingProtoOSPF
	case strings.ContainsRune(flags, 'b'):
		return configparser.RoutingProtoBGP
	case strings.ContainsRune(flags, 'i'):
		return configparser.RoutingProtoISIS
	case strings.ContainsRune(flags, 'r'):
		return configparser.RoutingProtoRIP
	case strings.ContainsRune(flags, 's'):
		return configparser.RoutingProtoStatic
	}
	// No origin letter at all is what an older RouterOS prints for a route the
	// operator typed in; static is the only thing it can be.
	return configparser.RoutingProtoStatic
}

// parseRouterOSOSPF assembles OSPF instances from the three menus that describe
// them, and resolving the indirection between those menus is the whole job.
//
// RouterOS 7 splits OSPF three ways: an instance carries the router-id, an area
// belongs to an instance by NAME and carries the dotted area-id, and an
// interface-template attaches networks to an area — again by the area's NAME,
// not its area-id. So a network reaches its router only through
// template.area → area.name → area.instance → instance.name, and taking the
// tempting shortcut of matching template.area against an area-id silently drops
// every network on a box whose areas are named anything but their id.
func parseRouterOSOSPF(instanceRecs, areaRecs, templateRecs []terseRecord) []configparser.ConfigRoutingProtocol {
	protos := make([]configparser.ConfigRoutingProtocol, 0, len(instanceRecs))
	byInstance := make(map[string]int, len(instanceRecs))

	for _, rec := range instanceRecs {
		name := rec.Get("name")
		protos = append(protos, configparser.ConfigRoutingProtocol{
			Type:     configparser.RoutingProtoOSPF,
			Enabled:  !rec.Disabled(),
			Instance: name,
			RouterID: rec.Get("router-id"),
			VRF:      rec.Get("vrf"),
		})
		byInstance[name] = len(protos) - 1
	}

	// areaLoc maps an area's own name to where its ConfigOSPFArea ended up, so
	// the templates can find it in one step.
	type areaLoc struct{ proto, area int }
	byArea := make(map[string]areaLoc, len(areaRecs))

	for _, rec := range areaRecs {
		pi, ok := byInstance[rec.Get("instance")]
		if !ok {
			continue // an area whose instance was not returned has no home
		}
		protos[pi].Areas = append(protos[pi].Areas, configparser.ConfigOSPFArea{
			ID:   rec.Get("area-id"),
			Type: rec.Get("type"),
		})
		byArea[rec.Get("name")] = areaLoc{proto: pi, area: len(protos[pi].Areas) - 1}
	}

	for _, rec := range templateRecs {
		loc, ok := byArea[rec.Get("area")]
		if !ok {
			continue
		}
		area := &protos[loc.proto].Areas[loc.area]
		// A template states either the prefixes it covers or the interfaces it
		// binds; both are comma-separated lists in terse output.
		area.Networks = appendUniqueCSV(area.Networks, rec.Get("networks"))
		area.Interfaces = appendUniqueCSV(area.Interfaces, rec.Get("interfaces"))
	}

	return protos
}

// parseRouterOSBGP assembles BGP instances from /routing/bgp/connection.
//
// RouterOS 7 removed the BGP instance object: every peer is a standalone
// "connection" that repeats the local AS and router-id, so a full-mesh iBGP
// router prints the same as/router-id on every line. Emitting one protocol per
// connection would report a router as running four BGP processes, so the
// connections are grouped back into the instance they came from, keyed by the
// (as, router-id) pair they all share. Insertion order is preserved because the
// first connection listed is the one an operator will recognise.
func parseRouterOSBGP(connRecs []terseRecord) []configparser.ConfigRoutingProtocol {
	var protos []configparser.ConfigRoutingProtocol
	byKey := make(map[string]int, len(connRecs))

	for _, rec := range connRecs {
		as, routerID := rec.Get("as"), rec.Get("router-id")
		key := as + "/" + routerID

		pi, ok := byKey[key]
		if !ok {
			// Enabled starts false and is raised by the first live peer: a
			// single disabled peer must not take the instance down, but an
			// instance whose every peer is disabled is genuinely down.
			protos = append(protos, configparser.ConfigRoutingProtocol{
				Type:     configparser.RoutingProtoBGP,
				LocalAS:  as,
				RouterID: routerID,
			})
			pi = len(protos) - 1
			byKey[key] = pi
		}
		if !rec.Disabled() {
			protos[pi].Enabled = true
		}

		protos[pi].Neighbors = append(protos[pi].Neighbors, configparser.ConfigBGPNeighbor{
			Address:      rec.Get("remote.address"),
			RemoteAS:     rec.Get("remote.as"),
			Description:  rec.Get("name"),
			UpdateSource: rec.Get("local.address"),
		})
	}

	return protos
}

// appendUniqueCSV appends the comma-separated entries of csv to dst, skipping
// blanks and anything already present. Templates repeat prefixes across areas
// often enough that deduplication is worth doing here rather than downstream.
func appendUniqueCSV(dst []string, csv string) []string {
	for _, entry := range strings.Split(csv, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		found := false
		for _, existing := range dst {
			if existing == entry {
				found = true
				break
			}
		}
		if !found {
			dst = append(dst, entry)
		}
	}
	return dst
}
