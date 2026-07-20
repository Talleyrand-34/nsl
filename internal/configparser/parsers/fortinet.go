// fortinet.go: the FortiGate parser, written against a real FortiOS 6.0 box.
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
	"regexp"
	"strconv"
	"strings"
	"time"

	"nsl-graph/internal/configparser"
	s "nsl-graph/internal/scanner"
)

// FortinetParser handles parsing of Fortinet FortiGate CLI configuration.
//
// FortiOS is the only supported OS where the running configuration is not enough
// to describe the device: `show system interface` reports how an interface was
// *told* to get an address, not the address it *has*. A DHCP or PPPoE WAN — the
// normal shape of a branch firewall — therefore shows no address at all, and a
// parser that reads only the configuration will place the device on the map with
// no reachable IP even while it is answering on the very address it is missing.
// That is why this parser also reads operational state (`get system interface
// physical`, `diagnose hardware deviceinfo nic`) and merges it in.
type FortinetParser struct{}

// NewFortinetParser creates a new Fortinet configuration parser
func NewFortinetParser() *FortinetParser {
	return &FortinetParser{}
}

func init() { configparser.DefaultRegistry.RegisterParser(NewFortinetParser()) }

// GetOsType returns the OS type this parser handles
func (p *FortinetParser) GetOsType() string {
	return "fortinet"
}

// SupportsDevice returns true if this parser can handle the given device
func (p *FortinetParser) SupportsDevice(device s.SNMPDevice) bool {
	descr := strings.ToLower(device.SysDescr)
	return strings.Contains(descr, "fortinet") || strings.Contains(descr, "fortigate") ||
		strings.Contains(device.SysName, "FortiGate") || strings.Contains(device.SysName, "FG")
}

// The commands Fetch issues, in the order their output is written.
const (
	fortiCmdGlobal    = "show system global"
	fortiCmdInterface = "show system interface"
	fortiCmdPhysical  = "get system interface physical"
	fortiCmdStatic    = "show router static"
	fortiCmdOSPF      = "show router ospf"
	fortiCmdBGP       = "show router bgp"
	fortiCmdFull      = "show full-configuration"
	fortiCmdNICPrefix = "diagnose hardware deviceinfo nic "
)

// fortiMaxNICQueries caps how many per-interface MAC lookups one scan may cost.
// Each is a separate round trip, so a chassis with a few hundred interfaces would
// otherwise turn a single device scan into a few hundred sequential commands. The
// cap is recorded in-band when it bites so the truncation is visible rather than
// silent.
const fortiMaxNICQueries = 24

// Fetch assembles a FortiGate's state from targeted commands, tagging each block
// with "# <command>" so ParseConfig can re-split the combined output.
//
// Two things drive the command list. First, `show system interface` is preferred
// over `show full-configuration` because it is flat and small; the full dump
// remains as a fallback for accounts or builds where the targeted command is
// refused. Second, the configuration alone cannot answer "what address does this
// box have" for a DHCP interface, so operational state is fetched alongside it.
//
// A failing command is never fatal: `show router bgp` is refused on some feature
// sets, `diagnose` needs a privilege a read-only profile may lack. The error is
// recorded in-band and everything else still parses. Only losing BOTH interface
// commands is an error, because then there is nothing to report.
func (p *FortinetParser) Fetch(sess configparser.Session) (string, error) {
	var out strings.Builder

	run := func(cmd string) (string, bool) {
		output, err := sess.Execute(cmd)
		if err != nil {
			fmt.Fprintf(&out, "# Error executing %s: %v\n", cmd, err)
			return "", false
		}
		fmt.Fprintf(&out, "# %s\n%s\n", cmd, output)
		return output, true
	}

	run(fortiCmdGlobal)

	ifaceOut, ok := run(fortiCmdInterface)
	if !ok {
		full, err := sess.Execute(fortiCmdFull)
		if err != nil {
			return "", fmt.Errorf("failed to retrieve FortiGate configuration: %w", err)
		}
		fmt.Fprintf(&out, "# %s\n%s\n", fortiCmdFull, full)
		ifaceOut = full
	}

	run(fortiCmdPhysical)
	run(fortiCmdStatic)
	run(fortiCmdOSPF)
	run(fortiCmdBGP)

	// MAC addresses are one command per interface, so the interface list has to
	// exist before they can be asked for — hence the ordering above.
	names := fortiNICCandidates(sanitizeFortinetOutput(ifaceOut))
	if len(names) > fortiMaxNICQueries {
		fmt.Fprintf(&out, "# note: MAC lookup capped at %d of %d interfaces\n",
			fortiMaxNICQueries, len(names))
		names = names[:fortiMaxNICQueries]
	}
	for _, name := range names {
		run(fortiCmdNICPrefix + name)
	}

	return out.String(), nil
}

// ParseConfig parses raw Fortinet CLI output and returns structured ConfigData.
//
// Sanitising happens here rather than in Fetch so that a configuration read from
// a file (--config-source file) gets the same treatment: those files are usually
// a terminal capture, prompt echo and pager markers included.
func (p *FortinetParser) ParseConfig(rawConfig string, deviceInfo s.SNMPDevice) (*configparser.ConfigData, error) {
	clean := sanitizeFortinetOutput(rawConfig)

	// Sections are indexed from the whole sanitised text rather than per command
	// block, which is what keeps an untagged `show full-configuration` dump
	// working: every "config X … end" it contains lands under the same key it
	// would have had if it had arrived tagged.
	sections := fortiConfigSections(clean)

	configData := &configparser.ConfigData{
		OsType:        p.GetOsType(),
		DeviceModel:   fortinetModel(deviceInfo),
		Hostname:      fortiUnquote(fortiSetValue(sections["system global"], "hostname")),
		ConfigVersion: fortinetConfigVersion(clean),
		Source:        configparser.ConfigSourceSSH, // overridden by the caller for file sources
		ParsedAt:      time.Now(),
		Raw:           clean,
	}

	interfaces, vlans := p.parseInterfaces(sections["system interface"])
	p.mergeRuntimeState(interfaces, clean)
	p.mergeMACAddresses(interfaces, clean)
	configData.Interfaces = interfaces
	configData.VLANs = vlans

	configData.Routes = p.parseStaticRoutes(sections["router static"])
	configData.RoutingProtocols = p.parseRoutingProtocols(sections)
	configData.FirewallRules = p.parseFirewallPolicies(sections["firewall policy"])

	return configData, nil
}

// ValidateConfig performs basic validation on parsed Fortinet configuration
func (p *FortinetParser) ValidateConfig(config *configparser.ConfigData) []error {
	var errors []error

	if config.Hostname == "" {
		errors = append(errors, fmt.Errorf("hostname is required"))
	}

	interfaceNames := make(map[string]bool)
	for _, iface := range config.Interfaces {
		if interfaceNames[iface.Name] {
			errors = append(errors, fmt.Errorf("duplicate interface name: %s", iface.Name))
		}
		interfaceNames[iface.Name] = true
	}

	vlanIDs := make(map[string]bool)
	for _, vlan := range config.VLANs {
		if vlanIDs[vlan.ID] {
			errors = append(errors, fmt.Errorf("duplicate VLAN ID: %s", vlan.ID))
		}
		vlanIDs[vlan.ID] = true

		if vlanID, err := strconv.Atoi(vlan.ID); err == nil {
			if vlanID < 1 || vlanID > 4094 {
				errors = append(errors, fmt.Errorf("VLAN ID %s is out of valid range (1-4094)", vlan.ID))
			}
		}
	}

	return errors
}

// -----------------------------------------------------------------------------
// Sanitising: turning a terminal capture back into a configuration
// -----------------------------------------------------------------------------

// fortiPromptRe matches the prompt FortiOS glues to the first line of every
// command's output ("FGT30D3X15012871 # config system interface"), including the
// sub-mode form ("FGT30D3X15012871 (interface) # "). It is anchored so that a
// value which merely contains a hash — a comment, a password — is untouched.
var fortiPromptRe = regexp.MustCompile(`^[A-Za-z0-9._-]+(?: \([^)]*\))? # ?`)

// sanitizeFortinetOutput strips prompt echo and pager markers from captured CLI
// output.
//
// The --More-- marker deserves a word. With the default `set output more`, FortiOS
// writes the pager marker as a PREFIX on an otherwise intact line — verified on the
// device, where "--More--                  set snmp-index 4" still carries the whole
// setting. So the fix is to drop the token and keep the rest of the line; no content
// is lost. Deliberately no attempt is made to reconfigure the device to `output
// standard`: a scanner should read a box as it finds it, and tolerating the vendor
// default is the correct answer, not changing the operator's console settings.
func sanitizeFortinetOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		line = fortiPromptRe.ReplaceAllString(line, "")
		if idx := strings.Index(line, "--More--"); idx >= 0 {
			line = line[:idx] + strings.TrimLeft(line[idx+len("--More--"):], " ")
		}
		cleaned = append(cleaned, line)
	}
	return strings.Join(cleaned, "\n")
}

// fortiCommandBlock returns the output of one tagged command.
//
// It defers to the shared extractCommandBlock, then trims at the first following
// FortiOS tag: the shared helper only recognises a handful of command verbs as
// tags, and FortiOS blocks are introduced by `get` and `diagnose` too, which it
// would otherwise swallow into the preceding block.
func fortiCommandBlock(raw, command string) string {
	block := extractCommandBlock(raw, command)
	lines := strings.Split(block, "\n")
	for i, line := range lines {
		if fortiIsCommandTag(strings.TrimSpace(line)) {
			return strings.Join(lines[:i], "\n")
		}
	}
	return block
}

// fortiIsCommandTag reports whether a line is one of Fetch's own "# <command>" tags.
func fortiIsCommandTag(line string) bool {
	rest := strings.TrimSpace(strings.TrimPrefix(line, "#"))
	if rest == line || rest == "" {
		return false
	}
	for _, verb := range []string{"show ", "get ", "diagnose ", "config ", "Error executing ", "note: "} {
		if strings.HasPrefix(rest, verb) {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// The FortiOS config grammar: config/end, edit/next, set
// -----------------------------------------------------------------------------

// fortiBlock is one "config <header> … end" or "edit <header> … next" block.
type fortiBlock struct {
	Header string
	Body   string
}

// fortiConfigSections indexes every top-level "config X … end" block by X.
//
// The first occurrence wins. A capture can legitimately contain the same section
// twice — Fetch's fallback path emits `show full-configuration` alongside whatever
// targeted commands did succeed — and concatenating them would invent duplicate
// interfaces that ValidateConfig would then rightly complain about.
func fortiConfigSections(clean string) map[string]string {
	sections := make(map[string]string)
	for _, block := range fortiConfigBlocks(clean) {
		if _, seen := sections[block.Header]; !seen && strings.TrimSpace(block.Body) != "" {
			sections[block.Header] = block.Body
		}
	}
	return sections
}

// fortiConfigBlocks splits text into the "config … end" blocks at its outermost
// level, keeping nested blocks intact inside each body.
func fortiConfigBlocks(text string) []fortiBlock {
	var blocks []fortiBlock
	var body strings.Builder
	header := ""
	depth := 0

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || fortiIsCommandTag(trimmed) {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "config "):
			if depth == 0 {
				header = strings.TrimSpace(strings.TrimPrefix(trimmed, "config "))
				body.Reset()
			} else {
				body.WriteString(trimmed + "\n")
			}
			depth++

		case trimmed == "end":
			if depth == 0 {
				continue // stray `end`, e.g. a truncated capture
			}
			depth--
			if depth == 0 {
				blocks = append(blocks, fortiBlock{Header: header, Body: body.String()})
				header = ""
				body.Reset()
			} else {
				body.WriteString(trimmed + "\n")
			}

		case depth > 0:
			body.WriteString(trimmed + "\n")
		}
	}

	return blocks
}

// fortiEditBlocks splits a section body into its "edit <id> … next" entries.
//
// Nesting is tracked because a real interface entry can contain a whole nested
// table (`config secondaryip` has its own edit/next pairs); splitting naively on
// every `edit` would tear one interface into several.
func fortiEditBlocks(body string) []fortiBlock {
	var blocks []fortiBlock
	var current strings.Builder
	header := ""
	inEdit := false
	depth := 0

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "config "):
			if inEdit {
				current.WriteString(trimmed + "\n")
			}
			depth++
		case trimmed == "end":
			if depth > 0 {
				depth--
				if inEdit {
					current.WriteString(trimmed + "\n")
				}
			}
		case depth == 0 && strings.HasPrefix(trimmed, "edit "):
			if inEdit {
				blocks = append(blocks, fortiBlock{Header: header, Body: current.String()})
			}
			header = fortiUnquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "edit ")))
			current.Reset()
			inEdit = true
		case depth == 0 && trimmed == "next":
			if inEdit {
				blocks = append(blocks, fortiBlock{Header: header, Body: current.String()})
				current.Reset()
				inEdit = false
			}
		case inEdit:
			current.WriteString(trimmed + "\n")
		}
	}

	if inEdit {
		blocks = append(blocks, fortiBlock{Header: header, Body: current.String()})
	}

	return blocks
}

// fortiNestedBlocks returns the "config … end" blocks one level inside a body.
func fortiNestedBlocks(body string) []fortiBlock {
	return fortiConfigBlocks(body)
}

// fortiSetFields returns the arguments of "set <key>", matching the key exactly.
//
// Exact key matching rather than a prefix regex is what keeps `set ip` from
// answering with `set ip6-address`, and `set interface` from answering with
// `set interface-select-method`.
func fortiSetFields(body, key string) []string {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		fields := strings.Fields(trimmed)
		if len(fields) >= 3 && fields[0] == "set" && fields[1] == key {
			rest := strings.TrimSpace(strings.TrimSpace(trimmed[len("set"):])[len(key):])
			return fortiSplitArgs(rest)
		}
	}
	return nil
}

// fortiSplitArgs splits a `set` argument list, keeping quoted groups whole.
//
// FortiOS quotes any value that can contain a space and lists several of them on
// one line (`set srcaddr "all" "web servers"`), so splitting on whitespace alone
// turns one alias into two and truncates every description at its first word.
func fortiSplitArgs(v string) []string {
	var out []string
	var current strings.Builder
	quoted, started := false, false

	flush := func() {
		if started {
			out = append(out, current.String())
			current.Reset()
			started = false
		}
	}

	for _, r := range v {
		switch {
		case r == '"':
			quoted = !quoted
			started = true
		case (r == ' ' || r == '\t') && !quoted:
			flush()
		default:
			current.WriteRune(r)
			started = true
		}
	}
	flush()

	return out
}

// fortiSetValue returns the first argument of "set <key>", unquoted.
func fortiSetValue(body, key string) string {
	fields := fortiSetFields(body, key)
	if len(fields) == 0 {
		return ""
	}
	return fortiUnquote(fields[0])
}

// fortiHasSet reports whether "set <key> <value>" appears verbatim.
func fortiHasSet(body, key, value string) bool {
	for _, field := range fortiSetFields(body, key) {
		if fortiUnquote(field) == value {
			return true
		}
	}
	return false
}

func fortiUnquote(v string) string {
	return strings.Trim(v, `"`)
}

// fortiDottedToCIDR converts FortiOS's "ADDRESS NETMASK" pair into CIDR.
//
// FortiOS states masks in dotted-decimal everywhere — interface addresses, static
// route destinations, OSPF networks — and never as a prefix length, so this is the
// single conversion the whole parser funnels through. An unusable or absent mask
// degrades to the bare address rather than inventing a prefix.
func fortiDottedToCIDR(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	addr := fortiUnquote(fields[0])
	if addr == "" || (addr == "0.0.0.0" && len(fields) < 2) {
		return ""
	}
	if strings.Contains(addr, "/") {
		return addr
	}
	if len(fields) < 2 {
		return addr
	}
	mask := fortiUnquote(fields[1])
	if prefix := netmaskToPrefix(mask); prefix >= 0 {
		return fmt.Sprintf("%s/%d", addr, prefix)
	}
	return addr
}

// -----------------------------------------------------------------------------
// Interfaces
// -----------------------------------------------------------------------------

// parseInterfaces reads `config system interface`, returning the interfaces and
// the VLANs they declare.
func (p *FortinetParser) parseInterfaces(section string) ([]configparser.ConfigInterface, []configparser.ConfigVLAN) {
	var interfaces []configparser.ConfigInterface
	var vlans []configparser.ConfigVLAN

	for _, block := range fortiEditBlocks(section) {
		if block.Header == "" {
			continue
		}

		iface := configparser.ConfigInterface{
			Name:        block.Header,
			Description: fortinetDescription(block.Body),
			// FortiOS interfaces are administratively up unless told otherwise;
			// `set status down` is only emitted when someone disabled one.
			Enabled: !fortiHasSet(block.Body, "status", "down"),
			Type:    fortinetInterfaceType(block.Body),
		}

		if addr := fortiDottedToCIDR(fortiSetFields(block.Body, "ip")); addr != "" && !strings.HasPrefix(addr, "0.0.0.0") {
			iface.IPAddresses = append(iface.IPAddresses, addr)
		}

		if mtu := fortiSetValue(block.Body, "mtu"); mtu != "" {
			if v, err := strconv.Atoi(mtu); err == nil {
				iface.MTU = v
			}
		}

		if vlanID := fortiSetValue(block.Body, "vlanid"); vlanID != "" {
			iface.Type = "vlan"
			iface.Parent = fortiSetValue(block.Body, "interface")
			vlan := configparser.ConfigVLAN{
				ID:          vlanID,
				Name:        iface.Name,
				Description: iface.Description,
				Enabled:     iface.Enabled,
				Tagged:      true,
			}
			if len(iface.IPAddresses) > 0 {
				vlan.IPRange = iface.IPAddresses[0]
			}
			iface.VLANs = append(iface.VLANs, vlan)
			vlans = append(vlans, vlan)
		} else if parent := fortiSetValue(block.Body, "interface"); parent != "" {
			iface.Parent = parent
		}

		interfaces = append(interfaces, iface)
	}

	return interfaces, vlans
}

// fortinetDescription prefers `set alias`, which is what the GUI shows and what
// operators actually fill in; `set description` exists too and is the fallback.
func fortinetDescription(body string) string {
	if alias := fortiSetValue(body, "alias"); alias != "" {
		return alias
	}
	return fortiSetValue(body, "description")
}

// fortinetInterfaceType maps FortiOS's `set type` onto the vocabulary ConfigInterface
// documents ("physical", "vlan", "bridge", "tunnel").
//
// hard-switch and switch are FortiGate's internal LAN switch: several front panel
// ports behind one logical interface, which is a bridge in every sense that matters
// to a topology. Aggregates are reported as bridges for the same reason — the field
// has no "lag" value, and no other parser in this package emits one.
func fortinetInterfaceType(body string) string {
	switch fortiSetValue(body, "type") {
	case "physical":
		return "physical"
	case "tunnel":
		return "tunnel"
	case "hard-switch", "switch", "software-switch", "aggregate", "redundant":
		return "bridge"
	case "vlan":
		return "vlan"
	case "loopback":
		return "loopback"
	}
	// An entry with no `set type` and a vlanid is a VLAN sub-interface; anything
	// else without a type is a plain port.
	if fortiSetValue(body, "vlanid") != "" {
		return "vlan"
	}
	return "physical"
}

// fortiNICCandidates lists the interfaces worth a `diagnose … nic` query.
//
// Only interfaces backed by a NIC have a MAC of their own: tunnels have none, and
// a VLAN sub-interface merely borrows its parent's, so asking costs a round trip
// to learn nothing.
func fortiNICCandidates(cleanInterfaceOutput string) []string {
	var names []string
	for _, block := range fortiEditBlocks(fortiConfigSections(cleanInterfaceOutput)["system interface"]) {
		if block.Header == "" {
			continue
		}
		switch fortinetInterfaceType(block.Body) {
		case "tunnel", "vlan", "loopback":
			continue
		}
		names = append(names, block.Header)
	}
	return names
}

// fortiRuntimeIface is the operational state of one interface, as reported by
// `get system interface physical`.
type fortiRuntimeIface struct {
	Address string // CIDR, empty when unaddressed
	Status  string // "up" / "down"
}

// fortiPhysIfaceRe matches the interface header of `get system interface physical`
// ("==[wan]"). The enclosing hardware group header is spelled with a space
// ("== [onboard]") and deliberately does not match: it names a chassis section, not
// an interface.
var fortiPhysIfaceRe = regexp.MustCompile(`^==\[([^\]]+)\]$`)

// mergeRuntimeState folds `get system interface physical` onto the configured
// interfaces.
//
// This is the whole reason the command is fetched. A DHCP or PPPoE interface has
// no address in the configuration — the FortiGate this parser was written against
// answers on 10.0.50.50 while `show system interface` says only `set mode dhcp` —
// so without this merge the device's own management address is missing from the
// map.
//
// Only interfaces the configuration already named are touched. The physical view
// also lists hardware that is not an interface, and inventing entries from it
// would put ghosts on the map.
func (p *FortinetParser) mergeRuntimeState(interfaces []configparser.ConfigInterface, clean string) {
	runtime := parseFortiPhysical(fortiCommandBlock(clean, fortiCmdPhysical))
	if len(runtime) == 0 {
		return
	}

	for i := range interfaces {
		state, ok := runtime[interfaces[i].Name]
		if !ok {
			continue
		}

		if state.Address != "" && !fortiHasAddress(interfaces[i].IPAddresses, state.Address) {
			interfaces[i].IPAddresses = append(interfaces[i].IPAddresses, state.Address)
		}

		// Operational state stands in for administrative state, which FortiOS only
		// states when an interface has been disabled. `set status down` therefore
		// wins: a link that is down because nothing is plugged in is a different
		// fact from one an operator switched off, and only the latter is config.
		if state.Status != "" && interfaces[i].Enabled {
			interfaces[i].Enabled = state.Status == "up"
		}
	}
}

// fortiHasAddress reports whether the list already carries this address, comparing
// on the address only. A statically configured `set ip` and the runtime address of
// the same interface are the same fact stated twice, and must not both be listed
// merely because one of them arrived with a different mask.
func fortiHasAddress(existing []string, addr string) bool {
	host := strings.SplitN(addr, "/", 2)[0]
	for _, e := range existing {
		if strings.SplitN(e, "/", 2)[0] == host {
			return true
		}
	}
	return false
}

// parseFortiPhysical reads the indented `get system interface physical` report.
func parseFortiPhysical(block string) map[string]fortiRuntimeIface {
	out := make(map[string]fortiRuntimeIface)
	name := ""
	var current fortiRuntimeIface
	populated := false

	flush := func() {
		// An entry with no fields at all is a header we misread; dropping it is
		// cheaper than guessing.
		if name != "" && populated {
			out[name] = current
		}
		name, current, populated = "", fortiRuntimeIface{}, false
	}

	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if m := fortiPhysIfaceRe.FindStringSubmatch(trimmed); m != nil {
			flush()
			name = m[1]
			continue
		}
		if strings.HasPrefix(trimmed, "==") {
			flush() // a hardware group header ends the previous interface
			continue
		}
		if name == "" {
			continue
		}

		key, value, found := strings.Cut(trimmed, ":")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		fields := strings.Fields(value)

		switch key {
		case "ip":
			populated = true
			// 0.0.0.0 0.0.0.0 is FortiOS for "no address" — a down PPPoE link
			// reports it, and recording it would put every unconfigured port on
			// the same imaginary network.
			if len(fields) >= 1 && fields[0] != "0.0.0.0" {
				current.Address = fortiDottedToCIDR(fields)
			}
		case "status":
			populated = true
			if len(fields) >= 1 {
				current.Status = fields[0]
			}
		case "mode", "speed", "link", "mtu":
			populated = true
		}
	}
	flush()

	return out
}

// mergeMACAddresses folds in the per-interface `diagnose hardware deviceinfo nic`
// output, which is the only place FortiOS states a MAC.
func (p *FortinetParser) mergeMACAddresses(interfaces []configparser.ConfigInterface, clean string) {
	for i := range interfaces {
		block := fortiCommandBlock(clean, fortiCmdNICPrefix+interfaces[i].Name)
		if strings.TrimSpace(block) == "" {
			continue
		}
		if mac := fortiNICMAC(block); mac != "" {
			interfaces[i].MACAddress = mac
		}
	}
}

// fortiNICMAC pulls the MAC out of a nic dump, preferring Current_HWaddr:
// Permanent_HWaddr is the burnt-in address, which differs from the one actually on
// the wire whenever the interface has been overridden or is an HA member.
func fortiNICMAC(block string) string {
	permanent := ""
	for _, line := range strings.Split(block, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "Current_HWaddr":
			return fields[1]
		case "Permanent_HWaddr":
			permanent = fields[1]
		}
	}
	return permanent
}

// -----------------------------------------------------------------------------
// Static routes
// -----------------------------------------------------------------------------

// parseStaticRoutes reads `config router static`.
func (p *FortinetParser) parseStaticRoutes(section string) []configparser.ConfigRoute {
	var routes []configparser.ConfigRoute

	for _, block := range fortiEditBlocks(section) {
		gateway := fortiSetValue(block.Body, "gateway")
		device := fortiSetValue(block.Body, "device")
		dst := fortiDottedToCIDR(fortiSetFields(block.Body, "dst"))

		if gateway == "" && device == "" && dst == "" {
			continue
		}

		// FortiOS omits `set dst` on a default route rather than spelling out
		// 0.0.0.0 0.0.0.0, so an entry with a gateway and no destination is THE
		// default route — the single most important route on a branch firewall.
		if dst == "" {
			dst = "0.0.0.0/0"
		}

		route := configparser.ConfigRoute{
			Network:   dst,
			Gateway:   gateway,
			Interface: device,
			Protocol:  configparser.RoutingProtoStatic,
			Metric:    fortiRouteMetric(block.Body),
		}
		if comment := fortiSetValue(block.Body, "comment"); comment != "" {
			route.Description = comment
		}

		routes = append(routes, route)
	}

	return routes
}

// fortiRouteMetric prefers `set distance` — FortiOS's administrative distance is
// what selects between competing routes — and falls back to `set priority`, which
// only breaks ties among equals.
func fortiRouteMetric(body string) int {
	for _, key := range []string{"distance", "priority"} {
		if v := fortiSetValue(body, key); v != "" {
			if m, err := strconv.Atoi(v); err == nil {
				return m
			}
		}
	}
	return 0
}

// -----------------------------------------------------------------------------
// Dynamic routing
// -----------------------------------------------------------------------------

// parseRoutingProtocols reports the dynamic protocols this device actually runs.
func (p *FortinetParser) parseRoutingProtocols(sections map[string]string) []configparser.ConfigRoutingProtocol {
	var protocols []configparser.ConfigRoutingProtocol

	if ospf := parseFortiOSPF(sections["router ospf"]); ospf != nil {
		protocols = append(protocols, *ospf)
	}
	if bgp := parseFortiBGP(sections["router bgp"]); bgp != nil {
		protocols = append(protocols, *bgp)
	}

	return protocols
}

// The empty-skeleton trap.
//
// FortiOS emits a `config router ospf` and a `config router bgp` block on EVERY
// device, configured or not. On the FortiGate this parser was written against —
// which has never had a dynamic protocol touched — `show router ospf` still
// returns the block, containing nothing but a row of empty `config redistribute
// "x" / end` pairs that FortiOS pre-creates as placeholders.
//
// So the presence of the block means nothing at all, and a parser that infers OSPF
// from it would report OSPF on every FortiGate ever scanned. That is a silent,
// plausible-looking wrong answer, which is the worst kind: the topology would show
// a routed control plane where there is a single default route. Substance must be
// positively demonstrated instead — a router-id, an area, a network, or (BGP) a
// local AS or a neighbor. Empty redistribute placeholders never count.
func fortiHasRoutingSubstance(p configparser.ConfigRoutingProtocol) bool {
	return p.RouterID != "" || len(p.Areas) > 0 || len(p.Networks) > 0 ||
		len(p.Neighbors) > 0 || (p.LocalAS != "" && p.LocalAS != "0")
}

// parseFortiOSPF reads `config router ospf`, returning nil when the block is a
// bare skeleton.
func parseFortiOSPF(section string) *configparser.ConfigRoutingProtocol {
	if strings.TrimSpace(section) == "" {
		return nil
	}

	proto := configparser.ConfigRoutingProtocol{
		Type:         configparser.RoutingProtoOSPF,
		Enabled:      true,
		RouterID:     fortiSetValue(section, "router-id"),
		Redistribute: fortiRedistribute(section),
	}

	// Areas come from `config area`; the networks that populate them are declared
	// separately in `config network`, each naming its area, so the two are read
	// independently and joined afterwards.
	areaIndex := make(map[string]int)
	for _, nested := range fortiNestedBlocks(section) {
		if nested.Header != "area" {
			continue
		}
		for _, entry := range fortiEditBlocks(nested.Body) {
			if entry.Header == "" {
				continue
			}
			areaIndex[entry.Header] = len(proto.Areas)
			proto.Areas = append(proto.Areas, configparser.ConfigOSPFArea{
				ID:   entry.Header,
				Type: fortiSetValue(entry.Body, "type"),
			})
		}
	}

	for _, nested := range fortiNestedBlocks(section) {
		if nested.Header != "network" {
			continue
		}
		for _, entry := range fortiEditBlocks(nested.Body) {
			prefix := fortiDottedToCIDR(fortiSetFields(entry.Body, "prefix"))
			if prefix == "" {
				continue
			}
			area := fortiSetValue(entry.Body, "area")
			idx, ok := areaIndex[area]
			if !ok {
				// A network naming an area with no `config area` entry of its own
				// is still participating; keep the area rather than dropping the
				// network into a bucket where nothing describes it.
				if area == "" {
					proto.Networks = append(proto.Networks, prefix)
					continue
				}
				areaIndex[area] = len(proto.Areas)
				proto.Areas = append(proto.Areas, configparser.ConfigOSPFArea{ID: area})
				idx = areaIndex[area]
			}
			proto.Areas[idx].Networks = append(proto.Areas[idx].Networks, prefix)
		}
	}

	for _, nested := range fortiNestedBlocks(section) {
		if nested.Header != "ospf-interface" {
			continue
		}
		for _, entry := range fortiEditBlocks(nested.Body) {
			if iface := fortiSetValue(entry.Body, "interface"); iface != "" {
				proto.Interfaces = append(proto.Interfaces, iface)
			}
		}
	}

	if !fortiHasRoutingSubstance(proto) {
		return nil
	}
	return &proto
}

// parseFortiBGP reads `config router bgp`, returning nil when the block is a bare
// skeleton.
func parseFortiBGP(section string) *configparser.ConfigRoutingProtocol {
	if strings.TrimSpace(section) == "" {
		return nil
	}

	proto := configparser.ConfigRoutingProtocol{
		Type:         configparser.RoutingProtoBGP,
		Enabled:      true,
		RouterID:     fortiSetValue(section, "router-id"),
		Redistribute: fortiRedistribute(section),
	}
	// FortiOS writes `set as 0` for "no AS" in a full-configuration dump, which is
	// a skeleton value, not a configured one.
	if as := fortiSetValue(section, "as"); as != "" && as != "0" {
		proto.LocalAS = as
	}

	for _, nested := range fortiNestedBlocks(section) {
		switch nested.Header {
		case "neighbor":
			for _, entry := range fortiEditBlocks(nested.Body) {
				if entry.Header == "" {
					continue
				}
				proto.Neighbors = append(proto.Neighbors, configparser.ConfigBGPNeighbor{
					Address:      entry.Header,
					RemoteAS:     fortiSetValue(entry.Body, "remote-as"),
					Description:  fortiSetValue(entry.Body, "description"),
					UpdateSource: fortiSetValue(entry.Body, "update-source"),
					NextHopSelf:  fortiHasSet(entry.Body, "next-hop-self", "enable"),
				})
			}
		case "network":
			for _, entry := range fortiEditBlocks(nested.Body) {
				if prefix := fortiDottedToCIDR(fortiSetFields(entry.Body, "prefix")); prefix != "" {
					proto.Networks = append(proto.Networks, prefix)
				}
			}
		}
	}

	if !fortiHasRoutingSubstance(proto) {
		return nil
	}
	return &proto
}

// fortiRedistribute lists the protocols redistributed into an instance.
//
// The placeholder blocks FortiOS pre-creates are empty; only one carrying
// `set status enable` describes redistribution that is actually happening. The
// IPv6 `redistribute6` siblings are skipped: they would duplicate every entry
// under a name the rest of the pipeline has no address family to attach it to.
func fortiRedistribute(section string) []string {
	var out []string
	for _, nested := range fortiNestedBlocks(section) {
		name, arg, found := strings.Cut(nested.Header, " ")
		if !found || name != "redistribute" {
			continue
		}
		if !fortiHasSet(nested.Body, "status", "enable") {
			continue
		}
		out = append(out, fortiUnquote(strings.TrimSpace(arg)))
	}
	return out
}

// -----------------------------------------------------------------------------
// Firewall policies
// -----------------------------------------------------------------------------

// parseFirewallPolicies reads `config firewall policy`, which only a full
// configuration dump carries.
func (p *FortinetParser) parseFirewallPolicies(section string) []configparser.ConfigFirewallRule {
	var rules []configparser.ConfigFirewallRule

	for _, block := range fortiEditBlocks(section) {
		rule := configparser.ConfigFirewallRule{
			ID:          block.Header,
			Name:        fortiSetValue(block.Body, "name"),
			Enabled:     !fortiHasSet(block.Body, "status", "disable"),
			Action:      fortinetAction(fortiSetValue(block.Body, "action")),
			Direction:   "forward",
			SourceZone:  fortiSetValue(block.Body, "srcintf"),
			DestZone:    fortiSetValue(block.Body, "dstintf"),
			Source:      fortiSetList(block.Body, "srcaddr"),
			Destination: fortiSetList(block.Body, "dstaddr"),
			Ports:       fortiSetList(block.Body, "service"),
		}
		rules = append(rules, rule)
	}

	return rules
}

// fortiSetList returns every argument of "set <key>", which FortiOS uses for the
// multi-valued object references in a policy (`set srcaddr "a" "b"`).
func fortiSetList(body, key string) []string {
	return fortiSetFields(body, key)
}

func fortinetAction(action string) string {
	switch strings.ToLower(action) {
	case "accept", "allow":
		return "allow"
	case "deny", "drop":
		return "deny"
	default:
		return action
	}
}

// -----------------------------------------------------------------------------
// Device identity
// -----------------------------------------------------------------------------

// fortinetConfigVersion reads the "#config-version:" banner a full configuration
// dump opens with. Targeted commands do not carry it, so the parse date stands in.
func fortinetConfigVersion(clean string) string {
	for _, line := range strings.Split(clean, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#config-version") {
			if _, rest, found := strings.Cut(line, ":"); found {
				return strings.TrimSpace(rest)
			}
		}
	}
	return time.Now().Format("2006-01-02")
}

// fortinetModel reports the model, which FortiOS states nowhere in the
// configuration — the CLI knows it only from `get system status`, which is not
// worth a round trip when SNMP already carries it in sysDescr.
func fortinetModel(deviceInfo s.SNMPDevice) string {
	descr := strings.ToLower(deviceInfo.SysDescr)
	if strings.Contains(descr, "forti") {
		return deviceInfo.SysDescr
	}
	return "FortiGate Firewall"
}
