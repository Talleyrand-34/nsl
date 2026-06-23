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

// Package topology collects layer-2/layer-1 adjacency evidence from hosts using
// multiple sources (LLDP via SNMP or SSH, CDP, bridge forwarding databases) and
// keeps the full gather. It is deliberately DB-free: it observes and records,
// it never writes to the repository or configures the targets. The correlation
// of this evidence into connection edges lives in the application layer.
package topology

import (
	"nsl-graph/internal/configparser"
	"nsl-graph/internal/scanner"
)

// Source identifiers for collected evidence (used for provenance and --source).
const (
	SourceSNMPLLDP  = "snmp-lldp"
	SourceSNMPCDP   = "snmp-cdp"
	SourceSNMPFDB   = "snmp-fdb"
	SourceSSHLLDP   = "ssh-lldp"
	SourceSSHFDB    = "ssh-fdb"
	SourceLocalLLDP = "local-lldp"
)

// AllSources lists every collector identifier, in preference order.
var AllSources = []string{SourceSNMPLLDP, SourceSNMPCDP, SourceSNMPFDB, SourceSSHLLDP, SourceSSHFDB, SourceLocalLLDP}

// Target describes one host to scan and how to reach it. Credentials are already
// resolved (decrypted) by the caller; this package never touches scan profiles
// or passphrases.
type Target struct {
	Host        string                       // management IP / hostname
	DeviceLabel string                       // resolved DB device label (may be empty)
	Profile     string                       // scan profile name used (for reporting)
	SNMP        *scanner.ScanOptions         // non-nil enables snmp-lldp/snmp-cdp/snmp-fdb
	SSH         *configparser.SSHCredentials // non-nil enables ssh-lldp
	Local       bool                         // collect local-lldp from the machine running the tool
}

// NeighborEvidence is one observation that "remote" sits at the other end of
// LocalPort on ObservedHost, as reported by Source. Every field carries its
// origin so discrepancies between sources can be surfaced.
type NeighborEvidence struct {
	Source           string `json:"source"`
	ObservedHost     string `json:"observed_host"`
	ObservedDevice   string `json:"observed_device,omitempty"`
	LocalPort        string `json:"local_port"`
	RemoteChassisMAC string `json:"remote_chassis_mac,omitempty"`
	RemoteSysName    string `json:"remote_sys_name,omitempty"`
	RemotePort       string `json:"remote_port,omitempty"`
	RemoteIP         string `json:"remote_ip,omitempty"`
}

// FdbEvidence is one learned MAC from a bridge forwarding database: MAC was seen
// on local bridge port Port of ObservedHost (weak adjacency — the MAC may be
// several transparent hops away).
type FdbEvidence struct {
	Source         string `json:"source"`
	ObservedHost   string `json:"observed_host"`
	ObservedDevice string `json:"observed_device,omitempty"`
	MAC            string `json:"mac"`
	Port           string `json:"port"`
	VLAN           string `json:"vlan,omitempty"`
}

// HostScan is the complete gather from one target: the raw SNMP device, all
// neighbour evidence, the FDB, and any non-fatal collector errors.
type HostScan struct {
	Host            string              `json:"host"`
	DeviceLabel     string              `json:"device_label,omitempty"`
	Profile         string              `json:"profile,omitempty"`
	LocalChassisMAC string              `json:"local_chassis_mac,omitempty"` // this host's own LLDP chassis MAC
	Device          *scanner.SNMPDevice `json:"device,omitempty"`
	Evidence        []NeighborEvidence  `json:"evidence,omitempty"`
	FDB             []FdbEvidence       `json:"fdb,omitempty"`
	Errors          []string            `json:"errors,omitempty"`
}

// Confidence levels for a derived edge.
const (
	ConfidenceConfirmed = "confirmed" // observed from both endpoints (bidirectional)
	ConfidenceCandidate = "candidate" // a single direct LLDP/CDP observation
	ConfidenceWeak      = "weak"      // FDB-only corroboration
)

// ConnectionEdge is a derived link between two device ports. RemoteResolved is
// false when one endpoint could not be matched to a device port in the DB (an
// unknown intermediary, e.g. an unmanaged switch) — such edges are reported but
// never imported.
type ConnectionEdge struct {
	FromDevicePortID string   `json:"from_deviceport_id,omitempty"`
	ToDevicePortID   string   `json:"to_deviceport_id,omitempty"`
	FromLabel        string   `json:"from"`
	ToLabel          string   `json:"to"`
	Confidence       string   `json:"confidence"`
	Provenance       []string `json:"provenance"`
	RemoteResolved   bool     `json:"remote_resolved"`
}

// Discrepancy is something the user should review before committing.
type Discrepancy struct {
	Kind       string   `json:"kind"`
	Detail     string   `json:"detail"`
	Provenance []string `json:"provenance,omitempty"`
}

// Intermediary is a device detected "in the middle" — a MAC learned in the
// forwarding tables of multiple hosts that never speaks LLDP and isn't in the DB
// (e.g. an unmanaged / Netgear "Plus" switch that bridges hosts transparently).
type Intermediary struct {
	MAC    string   `json:"mac"`
	Vendor string   `json:"vendor,omitempty"`
	SeenBy []string `json:"seen_by"` // "device:port" observations
}

// ConnectionScanResult is the full output of a scan: every host's gather, the
// derived edges, the detected intermediary devices, and the discrepancies to
// review. Only resolved, confirmed/candidate edges are eligible to be imported.
type ConnectionScanResult struct {
	Hosts         []HostScan       `json:"hosts"`
	Edges         []ConnectionEdge `json:"edges"`
	Intermediaries []Intermediary  `json:"intermediaries,omitempty"`
	Discrepancies []Discrepancy    `json:"discrepancies,omitempty"`
}
