// observed.go: project a scan result onto the canonical tree — what the network IS,
// as opposed to what the specification says it should be.
package mapping

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
	"sort"
	"strconv"
	"strings"

	e "nsl-graph/internal/repository/entities"
	s "nsl-graph/internal/scanner"
	"nsl-graph/internal/yang/canon"
)

// FromScan projects scanned devices onto the canonical tree, so the result can be
// diffed against the specification produced by FromEntities.
//
// # The identity problem, which is the whole difficulty here
//
// The specification keys a device by UUID. A scan does not know that UUID -- it knows an
// IP address and whatever the device calls itself. Diffing the two trees is meaningless
// unless the same physical device carries the same node-id in both, so this resolves
// each scanned device against `spec`:
//
//   - by IP, against the addresses the specification records for each device;
//   - failing that, by name (sysName against the device label).
//
// A device that resolves adopts its specification node-id, and its ports adopt their
// specification tp-ids (matched by interface name). It is then comparable, port for port.
//
// A device that resolves to nothing is minted an address-derived node-id, so it surfaces
// in the diff as UNEXPECTED rather than being silently dropped. That is the correct
// outcome: something is on the network that nobody wrote down.
func FromScan(devices []s.SNMPDevice, spec Source) (*canon.Root, []string) {
	m := &mapper{src: spec}
	m.index()

	r := &resolver{mapper: m}
	r.indexSpec()

	nodes := make([]canon.Node, 0, len(devices))
	for _, dev := range devices {
		nodes = append(nodes, r.node(dev))
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].NodeID < nodes[j].NodeID })

	root := &canon.Root{
		Networks: &canon.Networks{
			Network: []canon.Network{{
				NetworkID: canon.NetworkID,
				NetworkTypes: canon.NetworkTypes{
					L2Topology:  &struct{}{},
					NSLTopology: &struct{}{},
				},
				Nodes: nodes,
				// Links are deliberately absent. A device scan sees interfaces, not
				// cabling: adjacency comes from correlating LLDP/CDP/FDB across several
				// hosts, which is what `scan connections` does. Emitting no links means
				// the diff reports no link changes, which is honest -- rather than
				// reporting every cable as missing because this scan could not see one.
			}},
		},
	}
	return root, m.warnings
}

// resolver maps scanned devices and interfaces back onto the identities the
// specification already assigned them.
type resolver struct {
	*mapper

	deviceByIP   map[string]e.Device
	deviceByName map[string]e.Device

	// portsByDevice maps a spec device ID to its ports, keyed by port name.
	portsByName map[string]map[string]e.DevicePort
}

func (r *resolver) indexSpec() {
	r.deviceByIP = make(map[string]e.Device)
	r.deviceByName = make(map[string]e.Device)

	for _, d := range r.src.Devices {
		for _, ip := range d.Ips {
			if ip = usableAddr(ip); ip != "" {
				r.deviceByIP[ip] = d
			}
		}
		if d.Label != "" {
			r.deviceByName[strings.ToLower(d.Label)] = d
		}
	}

	r.portsByName = make(map[string]map[string]e.DevicePort)
	for _, dp := range r.src.DevicePorts {
		byName, ok := r.portsByName[dp.DeviceID]
		if !ok {
			byName = make(map[string]e.DevicePort)
			r.portsByName[dp.DeviceID] = byName
		}
		byName[dp.PortName] = dp
	}
}

// usableAddr strips a CIDR suffix and rejects the addresses that identify nothing.
func usableAddr(raw string) string {
	ip := raw
	if i := strings.IndexByte(ip, '/'); i >= 0 {
		ip = ip[:i]
	}
	ip = strings.TrimSpace(ip)
	if ip == "" || ip == "0.0.0.0" || strings.HasPrefix(ip, "127.") {
		return ""
	}
	return ip
}

// resolve finds the specification device this scanned device IS, if any.
func (r *resolver) resolve(dev s.SNMPDevice) (e.Device, bool) {
	if ip := usableAddr(dev.IP); ip != "" {
		if d, ok := r.deviceByIP[ip]; ok {
			return d, true
		}
	}
	// The device may have been re-addressed since the specification was written, so fall
	// back to what it calls itself.
	if dev.SysName != "" {
		if d, ok := r.deviceByName[strings.ToLower(dev.SysName)]; ok {
			return d, true
		}
	}
	return e.Device{}, false
}

func (r *resolver) node(dev s.SNMPDevice) canon.Node {
	spec, known := r.resolve(dev)

	name := dev.SysName
	if name == "" {
		name = dev.IP
	}

	n := canon.Node{}
	if known {
		n.NodeID = deviceURIPrefix + spec.ID
		// Prefer the specification's label so the two trees name the same device the
		// same way; the diff is read by a human, not a parser.
		if spec.Label != "" {
			name = spec.Label
		}
	} else {
		// Nothing in the specification matches. Mint an address-derived id so it shows
		// up as unexpected rather than vanishing.
		n.NodeID = deviceURIPrefix + "unknown:" + dev.IP
		r.warnf("scanned device %s (%s) matches nothing in the specification; "+
			"it will be reported as unexpected", dev.IP, dev.SysName)
	}

	n.L2 = &canon.L2NodeAttributes{Name: name}
	if ip := usableAddr(dev.IP); ip != "" {
		n.L2.ManagementAddress = []string{ip}
	}

	n.TerminationPoints = r.terminationPoints(dev, spec, known)
	return n
}

func (r *resolver) terminationPoints(dev s.SNMPDevice, spec e.Device, known bool) []canon.TerminationPoint {
	specPorts := r.portsByName[spec.ID]

	tps := make([]canon.TerminationPoint, 0, len(dev.Interfaces))
	for i := range dev.Interfaces {
		iface := dev.Interfaces[i]

		// Only physical ports are comparable with the specification's model ports. A
		// bridge or a VLAN sub-interface is a logical construct that the spec models as
		// a DeviceInterface, not as a port, so comparing them here would report
		// differences that are not differences.
		if !iface.IsPhysicalPort() {
			continue
		}

		tp := canon.TerminationPoint{
			L2: &canon.L2TPAttributes{
				InterfaceName: iface.Name,
				MacAddress:    strings.ToLower(iface.MAC),
			},
			VlanMembership: r.vlans(iface, dev, known),
		}

		if known {
			if sp, ok := specPorts[iface.Name]; ok {
				tp.TpID = tpURIPrefix + sp.ID
			}
		}
		if tp.TpID == "" {
			// A port the specification does not know about, on a device it does. It
			// belongs in the diff as unexpected.
			tp.TpID = tpURIPrefix + "unknown:" + dev.IP + ":" + iface.Name
		}

		tps = append(tps, tp)
	}

	sort.Slice(tps, func(i, j int) bool { return tps[i].TpID < tps[j].TpID })
	return tps
}

// vlans converts observed VLAN memberships, applying exactly the same typing rules as
// the specification side. A device that reports VLAN 4999 is reporting nonsense, and it
// must be dropped here too -- otherwise the diff would show it as an "unexpected VLAN"
// and send an operator hunting for a VLAN that cannot exist.
func (r *resolver) vlans(iface s.DeviceInterface, dev s.SNMPDevice, known bool) []canon.VlanMembership {
	if len(iface.VLANs) == 0 {
		return nil
	}

	seen := make(map[uint16]bool, len(iface.VLANs))
	out := make([]canon.VlanMembership, 0, len(iface.VLANs))

	where := fmt.Sprintf("scanned device %s port %q", dev.IP, iface.Name)

	for _, v := range iface.VLANs {
		n, err := strconv.ParseUint(strings.TrimSpace(v.VLANNumber), 10, 16)
		if err != nil {
			r.warnf("%s: reported VLAN %q is not a number; ignored", where, v.VLANNumber)
			continue
		}
		if n < 1 || n > 4094 {
			r.warnf("%s: reported VLAN %d is out of range; ignored (802.1Q VLAN IDs are 1..4094)", where, n)
			continue
		}
		vid := uint16(n)

		if tagged, dup := seen[vid]; dup {
			if tagged != v.Tagged {
				r.warnf("%s: VLAN %d is reported as BOTH tagged and untagged; keeping tagged=%t",
					where, vid, tagged)
			}
			continue
		}
		seen[vid] = v.Tagged

		out = append(out, canon.VlanMembership{VlanID: vid, Tagged: v.Tagged})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].VlanID < out[j].VlanID })
	return out
}
