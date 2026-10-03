// infix.go: the one parser that reads a datastore instead of a CLI dialect.
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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"nsl-graph/internal/configparser"
	s "nsl-graph/internal/scanner"
)

// InfixParser reads configuration from an Infix device.
//
// Every other parser in this package is a line-scraper, because every other OS
// answers questions in a grammar it invented: UCI's dotted assignments, BSD's
// ifconfig blocks, FRR's indented router stanzas, RouterOS's export syntax. The
// scraping is not a shortcut — it is the only interface those devices offer, and
// each dialect had to be learned separately.
//
// Infix is the exception, and that is the whole point of it. It is a YANG-native
// NOS built on sysrepo: there is no vendor grammar to scrape because there is no
// vendor grammar at all. The running configuration *is* IETF YANG data, and
// `sysrepocfg -f json` hands it over as JSON that conforms to RFC 7951. So this
// parser decodes standards-based structures — ietf-interfaces, ietf-ip,
// ietf-routing, ietf-ospf, ietf-system — into typed Go structs, and the only
// vendor-specific knowledge it needs is how to read three identityref namespaces
// (infix-if-type:, infix-routing:). Anything this file learns about ietf-ospf is
// knowledge about the standard, not about Infix, and would transfer unchanged to
// any other YANG-native device.
//
// The practical consequence is that the parse is exact rather than heuristic.
// There is no ambiguity about whether a token is an address or a flag, no
// fallback for a slightly different release's column layout: a prefix length is
// an integer in a leaf named "prefix-length", and it either is there or it is not.
type InfixParser struct{}

// NewInfixParser creates a new Infix configuration parser.
func NewInfixParser() *InfixParser {
	return &InfixParser{}
}

func init() { configparser.DefaultRegistry.RegisterParser(NewInfixParser()) }

// The sysrepocfg invocations Fetch runs, one per YANG module, used verbatim both
// as the command and as the tag that delimits its output in the combined blob.
const (
	infixCmdInterfaces = "sudo sysrepocfg -X -d running -f json -m ietf-interfaces"
	infixCmdRouting    = "sudo sysrepocfg -X -d running -f json -m ietf-routing"
	infixCmdSystem     = "sudo sysrepocfg -X -d running -f json -m ietf-system"
)

// GetOsType returns the OS type key used for --os-type and registry lookup.
func (p *InfixParser) GetOsType() string {
	return "infix"
}

// SupportsDevice returns true for devices that name Infix in their SNMP identity.
func (p *InfixParser) SupportsDevice(device s.SNMPDevice) bool {
	return strings.Contains(strings.ToLower(device.SysDescr), "infix") ||
		strings.Contains(strings.ToLower(device.SysName), "infix")
}

// Fetch dumps the running datastore, one YANG module per command, tagging each
// block with "# <command>" so ParseConfig can re-split the combined output — the
// same convention the OpenWrt parser uses for its uci commands.
//
// Note what is deliberately *not* run here: `ip addr` / `ip route`. On Infix the
// kernel's runtime state is not persistent and is not authoritative — it is
// whatever sysrepo last pushed down, and it is discarded on reboot. The datastore
// is the only honest source, so reading the kernel would at best duplicate it and
// at worst report a transient that no longer reflects the configuration.
//
// A failing command is not fatal. ietf-system may be unreadable to an
// unprivileged login while the interfaces still come through, and a device with
// no routing configured can return an error for ietf-routing rather than an empty
// document. The error is recorded in-band and the remaining modules still parse.
func (p *InfixParser) Fetch(sess configparser.Session) (string, error) {
	commands := []string{infixCmdInterfaces, infixCmdRouting, infixCmdSystem}

	var out strings.Builder
	for _, cmd := range commands {
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

// ParseConfig decodes the tagged sysrepocfg output into structured ConfigData.
func (p *InfixParser) ParseConfig(rawConfig string, deviceInfo s.SNMPDevice) (*configparser.ConfigData, error) {
	blocks := splitInfixBlocks(rawConfig)

	configData := &configparser.ConfigData{
		OsType:      p.GetOsType(),
		DeviceModel: "Infix",
		Hostname:    parseInfixHostname(blocks.get(infixCmdSystem)),
		Source:      configparser.ConfigSourceSSH,
		ParsedAt:    time.Now(),
		Raw:         rawConfig,
	}
	if configData.Hostname == "" {
		configData.Hostname = deviceInfo.SysName
	}

	interfaces, vlans, err := parseInfixInterfaces(blocks.get(infixCmdInterfaces))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ietf-interfaces: %w", err)
	}
	configData.Interfaces = interfaces
	configData.VLANs = vlans

	protocols, routes, err := parseInfixRouting(blocks.get(infixCmdRouting))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ietf-routing: %w", err)
	}
	configData.RoutingProtocols = protocols
	configData.Routes = routes

	return configData, nil
}

// ValidateConfig performs basic structural validation on parsed Infix configuration.
func (p *InfixParser) ValidateConfig(config *configparser.ConfigData) []error {
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
		if id, err := strconv.Atoi(vlan.ID); err == nil && (id < 1 || id > 4094) {
			errs = append(errs, fmt.Errorf("VLAN ID %s is out of valid range (1-4094)", vlan.ID))
		}
	}

	// An OSPF instance that binds no interface is configured but carries nothing,
	// which on Infix means the areas were created and never populated.
	for _, proto := range config.RoutingProtocols {
		if proto.Type != configparser.RoutingProtoOSPF && proto.Type != configparser.RoutingProtoOSPFv3 {
			continue
		}
		bound := len(proto.Interfaces)
		for _, area := range proto.Areas {
			bound += len(area.Interfaces) + len(area.Networks)
		}
		if bound == 0 {
			errs = append(errs, fmt.Errorf("%s instance %q has no interfaces or networks", proto.Type, proto.Instance))
		}
	}

	return errs
}

// ---------------------------------------------------------------------------
// Command-tagged blob handling
// ---------------------------------------------------------------------------

// infixBlocks maps a command tag to the output that followed it.
type infixBlocks struct {
	byCommand map[string]string
	whole     string
}

// get returns the JSON document produced by cmd. When the input carries no tags
// at all — a single module dumped to a file and passed with --config-source file —
// the whole input is offered to every module, and the ones it does not match
// simply decode to an empty document.
func (b infixBlocks) get(cmd string) string {
	if body, ok := b.byCommand[cmd]; ok {
		return body
	}
	if len(b.byCommand) == 0 {
		return b.whole
	}
	return ""
}

// splitInfixBlocks re-splits Fetch's output on its "# <command>" tags. An
// in-band "# Error executing …" line opens a block whose key matches no command,
// so a failed module is silently absent rather than corrupting a good one.
func splitInfixBlocks(raw string) infixBlocks {
	blocks := infixBlocks{byCommand: make(map[string]string), whole: raw}

	var key string
	var body []string
	flush := func() {
		if key != "" {
			blocks.byCommand[key] = strings.Join(body, "\n")
		}
		key, body = "", nil
	}

	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "#") {
			flush()
			key = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			continue
		}
		body = append(body, line)
	}
	flush()

	return blocks
}

// decodeInfixJSON unmarshals one module's document, tolerating an empty block:
// a module that produced no output is not an error, it is a device with nothing
// configured for it.
func decodeInfixJSON(block string, into any) error {
	if strings.TrimSpace(block) == "" {
		return nil
	}
	return json.Unmarshal([]byte(block), into)
}

// stripYANGPrefix drops the module prefix from an identityref or a namespaced
// key: "infix-if-type:ethernet" becomes "ethernet". RFC 7951 omits the prefix
// when the identity lives in the same module as the leaf, so the bare form has to
// be accepted too.
func stripYANGPrefix(v string) string {
	if i := strings.LastIndexByte(v, ':'); i >= 0 {
		return v[i+1:]
	}
	return v
}

// ---------------------------------------------------------------------------
// ietf-interfaces / ietf-ip
// ---------------------------------------------------------------------------

type infixInterfacesDoc struct {
	Interfaces struct {
		Interface []infixInterface `json:"interface"`
	} `json:"ietf-interfaces:interfaces"`
}

type infixInterface struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	// Enabled is a pointer because ietf-interfaces defaults it to true, and
	// sysrepocfg omits leaves left at their default.
	Enabled     *bool             `json:"enabled"`
	PhysAddress string            `json:"phys-address"`
	MTU         int               `json:"mtu"`
	IPv4        *infixIPContainer `json:"ietf-ip:ipv4"`
	IPv6        *infixIPContainer `json:"ietf-ip:ipv6"`
	// ParentExt is the standard ietf-if-extensions binding; ParentInfix is the
	// Infix-native spelling of the same relationship.
	ParentExt   string           `json:"ietf-if-extensions:parent-interface"`
	ParentInfix string           `json:"infix-interfaces:parent-interface"`
	VLAN        *infixVLANConf   `json:"vlan"`
	VLANNS      *infixVLANConf   `json:"infix-interfaces:vlan"`
	BridgePort  *infixBridgePort `json:"bridge-port"`
	BridgePortN *infixBridgePort `json:"infix-interfaces:bridge-port"`
}

type infixIPContainer struct {
	MTU     int `json:"mtu"`
	Address []struct {
		IP           string `json:"ip"`
		PrefixLength int    `json:"prefix-length"`
	} `json:"address"`
}

// infixVLANConf is Infix's VLAN interface container: an 802.1Q tag plus the
// interface it is carried over.
type infixVLANConf struct {
	ID          *int   `json:"id"`
	Tag         *int   `json:"tag"`
	TagType     string `json:"tag-type"`
	LowerLayer  string `json:"lower-layer-if"`
	LowerLayer2 string `json:"lower-layer-interface"`
}

func (v *infixVLANConf) tag() string {
	switch {
	case v == nil:
		return ""
	case v.ID != nil:
		return strconv.Itoa(*v.ID)
	case v.Tag != nil:
		return strconv.Itoa(*v.Tag)
	}
	return ""
}

func (v *infixVLANConf) lower() string {
	if v == nil {
		return ""
	}
	if v.LowerLayer != "" {
		return v.LowerLayer
	}
	return v.LowerLayer2
}

type infixBridgePort struct {
	Bridge string `json:"bridge"`
}

// parseInfixInterfaces decodes ietf-interfaces into ConfigInterfaces, plus the
// device-level VLAN list implied by any VLAN interfaces it declares.
func parseInfixInterfaces(block string) ([]configparser.ConfigInterface, []configparser.ConfigVLAN, error) {
	var doc infixInterfacesDoc
	if err := decodeInfixJSON(block, &doc); err != nil {
		return nil, nil, err
	}

	var interfaces []configparser.ConfigInterface
	var vlans []configparser.ConfigVLAN
	seenVLAN := make(map[string]bool)

	for _, in := range doc.Interfaces.Interface {
		iface := configparser.ConfigInterface{
			Name:        in.Name,
			Description: in.Description,
			Enabled:     in.Enabled == nil || *in.Enabled,
			MACAddress:  in.PhysAddress,
			MTU:         in.MTU,
			Type:        infixInterfaceType(in.Type),
		}
		if iface.MTU == 0 && in.IPv4 != nil {
			iface.MTU = in.IPv4.MTU
		}

		iface.IPAddresses = append(iface.IPAddresses, infixAddresses(in.IPv4)...)
		iface.IPAddresses = append(iface.IPAddresses, infixAddresses(in.IPv6)...)

		vlanConf := in.VLAN
		if vlanConf == nil {
			vlanConf = in.VLANNS
		}
		bridgePort := in.BridgePort
		if bridgePort == nil {
			bridgePort = in.BridgePortN
		}

		switch {
		case in.ParentExt != "":
			iface.Parent = in.ParentExt
		case in.ParentInfix != "":
			iface.Parent = in.ParentInfix
		case vlanConf.lower() != "":
			iface.Parent = vlanConf.lower()
		case bridgePort != nil:
			iface.Parent = bridgePort.Bridge
		}

		if tag := vlanConf.tag(); tag != "" {
			// A VLAN interface is always the tagged end of the trunk — the
			// untagged member is the lower-layer interface itself.
			iface.VLANs = append(iface.VLANs, configparser.ConfigVLAN{
				ID:      tag,
				Name:    in.Name,
				Enabled: iface.Enabled,
				Tagged:  true,
			})
			if iface.Type == "" || iface.Type == "logical" {
				iface.Type = "vlan"
			}
			if !seenVLAN[tag] {
				seenVLAN[tag] = true
				vlans = append(vlans, configparser.ConfigVLAN{
					ID:      tag,
					Name:    in.Name,
					Enabled: iface.Enabled,
					Tagged:  true,
				})
			}
		}

		interfaces = append(interfaces, iface)
	}

	return interfaces, vlans, nil
}

// infixAddresses renders an ietf-ip address list as CIDR strings.
func infixAddresses(c *infixIPContainer) []string {
	if c == nil {
		return nil
	}
	var out []string
	for _, a := range c.Address {
		if a.IP == "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s/%d", a.IP, a.PrefixLength))
	}
	return out
}

// infixInterfaceType maps an iana-if-type/infix-if-type identityref onto the Type
// vocabulary the other parsers use ("physical", "bridge", "vlan", …).
func infixInterfaceType(identity string) string {
	switch stripYANGPrefix(identity) {
	case "ethernet", "ethernetCsmacd":
		return "physical"
	case "bridge", "bridgeVlan":
		return "bridge"
	case "vlan", "l2vlan", "l3ipvlan":
		return "vlan"
	case "loopback", "softwareLoopback":
		return "loopback"
	case "veth", "dummy":
		return "virtual"
	case "gre", "gretap", "ipip", "tunnel", "vxlan", "wireguard":
		return "tunnel"
	case "":
		return "logical"
	}
	return "logical"
}

// ---------------------------------------------------------------------------
// ietf-routing / ietf-ospf
// ---------------------------------------------------------------------------

type infixRoutingDoc struct {
	Routing struct {
		ControlPlaneProtocols struct {
			ControlPlaneProtocol []infixControlPlaneProtocol `json:"control-plane-protocol"`
		} `json:"control-plane-protocols"`
	} `json:"ietf-routing:routing"`
}

type infixControlPlaneProtocol struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// OSPF covers both ospfv2 and ospfv3: ietf-ospf models them in one container
	// and distinguishes them by the protocol's identityref, not by the container.
	OSPF *infixOSPF `json:"ietf-ospf:ospf"`
	// The static-routes container is namespaced when it augments a foreign
	// module and bare when sysrepocfg dumps ietf-routing itself.
	StaticRoutes   *infixStaticRoutes `json:"static-routes"`
	StaticRoutesNS *infixStaticRoutes `json:"ietf-routing:static-routes"`
}

type infixOSPF struct {
	RouterID string `json:"router-id"`
	Areas    struct {
		Area []infixOSPFArea `json:"area"`
	} `json:"areas"`
}

type infixOSPFArea struct {
	AreaID     string `json:"area-id"`
	AreaType   string `json:"area-type"`
	Interfaces struct {
		Interface []struct {
			Name    string `json:"name"`
			Enabled *bool  `json:"enabled"`
		} `json:"interface"`
	} `json:"interfaces"`
}

type infixStaticRoutes struct {
	IPv4 *infixStaticAF `json:"ipv4"`
	IPv6 *infixStaticAF `json:"ipv6"`
}

type infixStaticAF struct {
	Route []struct {
		DestinationPrefix string `json:"destination-prefix"`
		Description       string `json:"description"`
		NextHop           struct {
			NextHopAddress    string `json:"next-hop-address"`
			OutgoingInterface string `json:"outgoing-interface"`
		} `json:"next-hop"`
	} `json:"route"`
}

// parseInfixRouting decodes ietf-routing's control-plane-protocol list.
//
// Static routes are reported as ConfigData.Routes rather than as a
// ConfigRoutingProtocol entry: a static "protocol" has no state beyond the routes
// themselves, and ConfigData.ControlPlane() already infers "static" from a
// non-empty route list. Dynamic protocols are the ones worth an instance record.
func parseInfixRouting(block string) ([]configparser.ConfigRoutingProtocol, []configparser.ConfigRoute, error) {
	var doc infixRoutingDoc
	if err := decodeInfixJSON(block, &doc); err != nil {
		return nil, nil, err
	}

	var protocols []configparser.ConfigRoutingProtocol
	var routes []configparser.ConfigRoute

	for _, cpp := range doc.Routing.ControlPlaneProtocols.ControlPlaneProtocol {
		protoType := infixRoutingProtoType(cpp.Type)

		if protoType == configparser.RoutingProtoStatic {
			static := cpp.StaticRoutes
			if static == nil {
				static = cpp.StaticRoutesNS
			}
			routes = append(routes, infixStaticRouteList(static)...)
			continue
		}
		if protoType == "" {
			continue // a protocol we do not model; not an error
		}

		proto := configparser.ConfigRoutingProtocol{
			Type:     protoType,
			Enabled:  true,
			Instance: cpp.Name,
		}

		if cpp.OSPF != nil {
			proto.RouterID = cpp.OSPF.RouterID
			for _, area := range cpp.OSPF.Areas.Area {
				a := configparser.ConfigOSPFArea{
					ID:   area.AreaID,
					Type: infixOSPFAreaType(area.AreaType),
				}
				// Infix binds OSPF per interface, never per network — which is
				// exactly the case ConfigOSPFArea.Interfaces exists for. Filling
				// Networks here instead would be a lie about what the device says.
				for _, ifc := range area.Interfaces.Interface {
					if ifc.Name == "" || (ifc.Enabled != nil && !*ifc.Enabled) {
						continue
					}
					a.Interfaces = append(a.Interfaces, ifc.Name)
					proto.Interfaces = append(proto.Interfaces, ifc.Name)
				}
				proto.Areas = append(proto.Areas, a)
			}
		}

		protocols = append(protocols, proto)
	}

	return protocols, routes, nil
}

// infixStaticRouteList flattens the ipv4 and ipv6 static route lists.
func infixStaticRouteList(static *infixStaticRoutes) []configparser.ConfigRoute {
	if static == nil {
		return nil
	}
	var routes []configparser.ConfigRoute
	for _, af := range []*infixStaticAF{static.IPv4, static.IPv6} {
		if af == nil {
			continue
		}
		for _, r := range af.Route {
			if r.DestinationPrefix == "" {
				continue
			}
			routes = append(routes, configparser.ConfigRoute{
				Network:     r.DestinationPrefix,
				Gateway:     r.NextHop.NextHopAddress,
				Interface:   r.NextHop.OutgoingInterface,
				Protocol:    configparser.RoutingProtoStatic,
				Description: r.Description,
			})
		}
	}
	return routes
}

// infixRoutingProtoType maps a control-plane-protocol identityref onto a
// RoutingProto* constant. Both the Infix and the IETF spellings appear —
// "infix-routing:ospfv2" and "ietf-ospf:ospfv2" identify the same protocol — and
// stripping the prefix collapses them. An unmodelled protocol returns "".
func infixRoutingProtoType(identity string) string {
	switch stripYANGPrefix(identity) {
	case "ospfv2", "ospf":
		return configparser.RoutingProtoOSPF
	case "ospfv3":
		return configparser.RoutingProtoOSPFv3
	case "static":
		return configparser.RoutingProtoStatic
	case "bgp":
		return configparser.RoutingProtoBGP
	case "rip", "ripv2", "ripng":
		return configparser.RoutingProtoRIP
	case "isis":
		return configparser.RoutingProtoISIS
	}
	return ""
}

// infixOSPFAreaType maps an ietf-ospf area-type identityref onto the
// ConfigOSPFArea.Type vocabulary. A normal area reports "standard"; an absent
// leaf stays empty, meaning the device did not say.
func infixOSPFAreaType(identity string) string {
	switch stripYANGPrefix(identity) {
	case "normal-area", "normal":
		return "standard"
	case "stub-area", "stub", "stub-nssa-area":
		return "stub"
	case "nssa-area", "nssa":
		return "nssa"
	}
	return ""
}

// ---------------------------------------------------------------------------
// ietf-system
// ---------------------------------------------------------------------------

type infixSystemDoc struct {
	System struct {
		Hostname string `json:"hostname"`
	} `json:"ietf-system:system"`
}

// parseInfixHostname reads the hostname out of ietf-system. A missing or
// unreadable module yields "", and ParseConfig falls back to SNMP's sysName.
func parseInfixHostname(block string) string {
	var doc infixSystemDoc
	if err := decodeInfixJSON(block, &doc); err != nil {
		return ""
	}
	return doc.System.Hostname
}
