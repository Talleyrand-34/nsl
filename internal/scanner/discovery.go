package scanner

import (
	"fmt"
	"net"
	"strings"
)

// DeviceDiscoverer classifies devices based on their SNMP sysDescr.
type DeviceDiscoverer struct{}

func NewDeviceDiscoverer() *DeviceDiscoverer {
	return &DeviceDiscoverer{}
}

// ClassifyDevice returns (brand, model, deviceClass) based on sysDescr.
func (dd *DeviceDiscoverer) ClassifyDevice(device SNMPDevice) (brand, model, deviceClass string) {
	descr := strings.ToLower(device.SysDescr)

	switch {
	case strings.Contains(descr, "cisco ios xe"):
		return "Cisco", "IOS XE Device", "Router"
	case strings.Contains(descr, "cisco ios xr"):
		return "Cisco", "IOS XR Device", "Router"
	case strings.Contains(descr, "cisco nx-os"):
		return "Cisco", "NX-OS Device", "Switch"
	case strings.Contains(descr, "cisco ios"):
		return "Cisco", "IOS Device", "Router"
	case strings.Contains(descr, "cisco"):
		return "Cisco", "Cisco Device", "Switch"
	case strings.Contains(descr, "juniper"):
		return "Juniper", "Juniper Device", "Router"
	case strings.Contains(descr, "aruba"):
		return "Aruba", "Aruba Device", "Access Point"
	case strings.Contains(descr, "ubiquiti") || strings.Contains(descr, "unifi"):
		return "Ubiquiti", "UniFi Device", "Access Point"
	case strings.Contains(descr, "fortinet") || strings.Contains(descr, "fortigate"):
		return "Fortinet", "FortiGate", "Firewall"
	case strings.Contains(descr, "palo alto"):
		return "Palo Alto", "PAN Device", "Firewall"
	case strings.Contains(descr, "mikrotik") || strings.Contains(descr, "routeros"):
		return "MikroTik", "RouterOS Device", "Router"
	case strings.Contains(descr, "hp procurve") || strings.Contains(descr, "hpe aruba"):
		return "HP", "ProCurve Switch", "Switch"
	case strings.Contains(descr, "openwrt"):
		return "OpenWrt", "OpenWrt Router", "Router"
	case strings.Contains(descr, "linux"):
		return "Linux", "Linux Server", "Server"
	case strings.Contains(descr, "windows"):
		return "Microsoft", "Windows Server", "Server"
	case strings.Contains(descr, "freebsd") || strings.Contains(descr, "openbsd"):
		return "BSD", "BSD Server", "Server"
	case strings.Contains(descr, "printer"):
		return "Generic", "Network Printer", "Printer"
	default:
		return "Generic", "Network Device", "Generic"
	}
}

// GenerateDeviceName returns the device name to use for import.
// Prefers sysName if set; otherwise generates a name from deviceClass and the last IP octet.
func (dd *DeviceDiscoverer) GenerateDeviceName(device SNMPDevice, deviceClass string) string {
	if device.SysName != "" {
		return device.SysName
	}

	prefix := classPrefix(deviceClass)
	parts := strings.Split(device.IP, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s-%s", prefix, parts[3])
	}
	return fmt.Sprintf("%s-%s", prefix, strings.ReplaceAll(device.IP, ".", "-"))
}

// SuggestZone suggests a zone name based on the device IP.
func (dd *DeviceDiscoverer) SuggestZone(device SNMPDevice) string {
	ip := net.ParseIP(device.IP)
	if ip == nil {
		return "Discovered"
	}
	if ip.IsPrivate() {
		switch {
		case strings.HasPrefix(device.IP, "192.168."):
			return "LAN"
		case strings.HasPrefix(device.IP, "10."):
			return "Internal"
		case strings.HasPrefix(device.IP, "172."):
			return "Private"
		}
	}
	return "External"
}

func classPrefix(deviceClass string) string {
	switch strings.ToLower(deviceClass) {
	case "switch":
		return "SW"
	case "router":
		return "RTR"
	case "access point":
		return "AP"
	case "firewall":
		return "FW"
	case "server":
		return "SRV"
	case "workstation":
		return "PC"
	case "printer":
		return "PRT"
	default:
		return "DEV"
	}
}
