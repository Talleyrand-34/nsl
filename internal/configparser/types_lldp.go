// SPDX-License-Identifier: MIT
// types_lldp.go: LLDP-transmit settings model.
package configparser

// ConfigLLDPSettings controls LLDP advertisement behaviour.
type ConfigLLDPSettings struct {
	Enabled          bool   `json:"enabled"`                    // tx on/off
	ChassisID       string `json:"chassis_id,omitempty"`      // override MAC-derived chassis ID
	SystemName      string `json:"system_name,omitempty"`    // LLDP System-Name TLV
	SystemDesc      string `json:"system_desc,omitempty"`    // LLDP System-Description TLV
	MgmtAddress     string `json:"mgmt_address,omitempty"`   // Management Address TLV
	InterfacePattern string `json:"interface_pattern,omitempty"` // OpenWrt: glob pattern for which ifaces to tx on
}
