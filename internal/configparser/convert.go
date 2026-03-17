package configparser

import s "nsl-graph/internal/scanner"

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
