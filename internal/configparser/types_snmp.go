// SPDX-License-Identifier: MIT
// types_snmp.go: SNMP configuration model types.
package configparser

// ConfigSNMPCommunity is one SNMP community + access level.
type ConfigSNMPCommunity struct {
	Name        string   `json:"name"`                    // community name
	Access      string   `json:"access"`                 // "ro" or "rw"
	Network     []string `json:"network,omitempty"`    // allowed source CIDRs
	TrapTarget  string   `json:"trap_target,omitempty"` // hostname/IP for trap sends
}

// ConfigSNMPConfig is the top-level SNMP configuration for a device.
type ConfigSNMPConfig struct {
	Enabled        bool                    `json:"enabled"`
	Location       string                  `json:"location,omitempty"`    // "Building A, Rack 4"
	Contact        string                  `json:"contact,omitempty"`       // "noc@example.com"
	Communities    []ConfigSNMPCommunity   `json:"communities"`
	ListenInterface string                  `json:"listen_interface,omitempty"` // OPNsense: bind to specific interface
}
