package configparser

import s "nsl-graph/internal/scanner"

// ConfigDataToDiscoveredDeviceInfo converts parsed SSH/file configuration into
// DiscoveredDeviceInfo, preserving all data including SwitchPorts and properly
// separating physical ports from logical interfaces.
func ConfigDataToDiscoveredDeviceInfo(cd *ConfigData, ip string) *s.DiscoveredDeviceInfo {
	name := ip
	if cd.Hostname != "" {
		name = cd.Hostname
	}

	info := &s.DiscoveredDeviceInfo{
		IP:            ip,
		SysName:       name,
		SysDescr:      cd.DeviceType,
		PhysicalPorts: make([]s.PhysicalPortInfo, 0, len(cd.SwitchPorts)),
		Interfaces:    make([]s.DeviceInterface, 0),
		Source:        "ssh",
	}

	// Convert SwitchPorts to PhysicalPorts
	for _, sp := range cd.SwitchPorts {
		info.PhysicalPorts = append(info.PhysicalPorts, s.PhysicalPortInfo{
			Name:       sp.PortName,
			PortNumber: sp.PortNumber,
			Role:       sp.Role,
			LinkStatus: sp.LinkStatus,
			PVID:       sp.PVID,
			MAC:        "",
		})
	}

	// Convert ConfigInterfaces to DeviceInterfaces, separating physical from logical
	for i, ci := range cd.Interfaces {
		oper := 1
		if !ci.Enabled {
			oper = 2
		}
		di := s.DeviceInterface{
			Index:       i + 1,
			Name:        ci.Name,
			MAC:         ci.MACAddress,
			IPAddresses: ci.IPAddresses,
			IPNetmasks:  make(map[string]string),
			OperStatus:  oper,
			AdminStatus: oper,
			Parent:      ci.Parent,
		}
		switch ci.Type {
		case "wifi-radio":
			di.IfType = s.IfTypeIEEE80211
			di.WifiBand = ci.WifiBand
		case "wifi-iface":
			di.IfType = s.IfTypePropVirtual
			di.WifiSSID = ci.WifiSSID
			di.WifiSecurity = ci.WifiSecurity
			di.Parent = ci.WifiRadio
		case "bridge":
			di.IfType = s.IfTypePropVirtual
			di.IsBridge = true
		case "physical", "switch-port":
			// Skip - these are physical ports, not logical interfaces
			continue
		default: // "logical", "vlan", ""
			di.IfType = s.IfTypePropVirtual
		}
		for _, cv := range ci.VLANs {
			di.VLANs = append(di.VLANs, s.VLANMembership{
				VLANNumber: cv.ID,
				Tagged:     cv.Tagged,
			})
		}
		info.Interfaces = append(info.Interfaces, di)
	}
	return info
}

// ConfigDataToSNMPDevice converts parsed SSH/file configuration into the standard
// SNMPDevice used by the rest of the pipeline. Both SNMP and SSH sources produce
// this same struct so downstream code is source-agnostic.
func ConfigDataToSNMPDevice(cd *ConfigData, ip string) *s.SNMPDevice {
	name := ip
	if cd.Hostname != "" {
		name = cd.Hostname
	}
	device := &s.SNMPDevice{IP: ip, SysName: name, Reachable: true}
	if cd.DeviceType != "" {
		device.SysDescr = cd.DeviceType
	}

	// Convert SwitchPorts from ConfigData
	for _, sp := range cd.SwitchPorts {
		device.SwitchPorts = append(device.SwitchPorts, s.PhysicalPortInfo{
			Name:       sp.PortName,
			PortNumber: sp.PortNumber,
			Role:       sp.Role,
			LinkStatus: sp.LinkStatus,
			PVID:       sp.PVID,
			MAC:        "",
		})
	}

	for i, ci := range cd.Interfaces {
		oper := 1
		if !ci.Enabled {
			oper = 2
		}
		di := s.DeviceInterface{
			Index:       i + 1,
			Name:        ci.Name,
			MAC:         ci.MACAddress,
			IPAddresses: ci.IPAddresses,
			IPNetmasks:  make(map[string]string),
			OperStatus:  oper,
			AdminStatus: oper,
			Parent:      ci.Parent,
		}
		switch ci.Type {
		case "wifi-radio":
			di.IfType = s.IfTypeIEEE80211
			di.WifiBand = ci.WifiBand
		case "wifi-iface":
			di.IfType = s.IfTypePropVirtual
			di.WifiSSID = ci.WifiSSID
			di.WifiSecurity = ci.WifiSecurity
			di.Parent = ci.WifiRadio
		case "physical":
			di.IfType = s.IfTypeEthernetCsmacd
		case "bridge":
			di.IfType = s.IfTypePropVirtual
			di.IsBridge = true
		case "switch-port":
			// Switch ports from board.json are in SwitchPorts, not Interfaces
			continue
		default: // "logical", "vlan", ""
			di.IfType = s.IfTypePropVirtual
		}
		for _, cv := range ci.VLANs {
			di.VLANs = append(di.VLANs, s.VLANMembership{
				VLANNumber: cv.ID,
				Tagged:     cv.Tagged,
			})
		}
		device.Interfaces = append(device.Interfaces, di)
	}
	return device
}
