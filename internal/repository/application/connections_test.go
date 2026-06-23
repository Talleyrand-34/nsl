package application

import (
	"strings"
	"testing"

	e "nsl-graph/internal/repository/entities"
	s "nsl-graph/internal/scanner"
	"nsl-graph/internal/secret"
	"nsl-graph/internal/topology"
)

// ports models a small lab: opnsense(igc1) — openwrt(br-lan).
func labPorts() []e.DevicePort {
	return []e.DevicePort{
		{ID: "p-opn", DevLabel: "opnsense", PortName: "igc1", MacAddress: "00:e0:b4:60:e9:b0"},
		{ID: "p-ow", DevLabel: "openwrt", PortName: "br-lan", MacAddress: "60:22:32:d5:84:81"},
	}
}

func edgeBetween(edges []topology.ConnectionEdge, a, b string) *topology.ConnectionEdge {
	for i := range edges {
		from, to := edges[i].FromDevicePortID, edges[i].ToDevicePortID
		if (from == a && to == b) || (from == b && to == a) {
			return &edges[i]
		}
	}
	return nil
}

func TestCorrelate_BidirectionalConfirmed(t *testing.T) {
	result := &topology.ConnectionScanResult{
		Hosts: []topology.HostScan{
			{Host: "10.0.0.1", DeviceLabel: "opnsense", Evidence: []topology.NeighborEvidence{
				{Source: topology.SourceSSHLLDP, ObservedDevice: "opnsense", LocalPort: "igc1",
					RemoteChassisMAC: "60:22:32:d5:84:81", RemoteSysName: "openwrt", RemotePort: "br-lan"},
			}},
			{Host: "10.0.0.245", DeviceLabel: "openwrt", Evidence: []topology.NeighborEvidence{
				{Source: topology.SourceSSHLLDP, ObservedDevice: "openwrt", LocalPort: "br-lan",
					RemoteChassisMAC: "00:e0:b4:60:e9:b0", RemoteSysName: "opnsense", RemotePort: "igc1"},
			}},
		},
	}

	correlateEvidence(result, labPorts(), []string{"opnsense", "openwrt"}, nil)

	edge := edgeBetween(result.Edges, "p-opn", "p-ow")
	if edge == nil {
		t.Fatalf("expected an edge between opnsense:igc1 and openwrt:br-lan; got %+v", result.Edges)
	}
	if edge.Confidence != topology.ConfidenceConfirmed {
		t.Errorf("expected confirmed (bidirectional), got %q", edge.Confidence)
	}
	if !edge.RemoteResolved {
		t.Errorf("expected resolved endpoints")
	}
	if len(edge.Provenance) != 2 {
		t.Errorf("expected provenance from both ends, got %v", edge.Provenance)
	}
}

func TestCorrelate_SingleSidedCandidate(t *testing.T) {
	result := &topology.ConnectionScanResult{
		Hosts: []topology.HostScan{
			{Host: "10.0.0.1", DeviceLabel: "opnsense", Evidence: []topology.NeighborEvidence{
				{Source: topology.SourceSSHLLDP, ObservedDevice: "opnsense", LocalPort: "igc1",
					RemoteChassisMAC: "60:22:32:d5:84:81", RemoteSysName: "openwrt", RemotePort: "br-lan"},
			}},
		},
	}
	correlateEvidence(result, labPorts(), []string{"opnsense", "openwrt"}, nil)
	edge := edgeBetween(result.Edges, "p-opn", "p-ow")
	if edge == nil {
		t.Fatalf("expected an edge, got %+v", result.Edges)
	}
	if edge.Confidence != topology.ConfidenceCandidate {
		t.Errorf("expected candidate (one-sided), got %q", edge.Confidence)
	}
}

func TestCorrelate_UnknownRemoteIsDiscrepancyNotImported(t *testing.T) {
	result := &topology.ConnectionScanResult{
		Hosts: []topology.HostScan{
			{Host: "10.0.0.1", DeviceLabel: "opnsense", Evidence: []topology.NeighborEvidence{
				{Source: topology.SourceSSHLLDP, ObservedDevice: "opnsense", LocalPort: "igc1",
					RemoteChassisMAC: "aa:bb:cc:dd:ee:ff", RemoteSysName: "mystery-switch", RemotePort: "1"},
			}},
		},
	}
	correlateEvidence(result, labPorts(), []string{"opnsense", "openwrt"}, nil)

	if len(result.Discrepancies) == 0 || result.Discrepancies[0].Kind != "unknown-remote" {
		t.Fatalf("expected an unknown-remote discrepancy, got %+v", result.Discrepancies)
	}
	for _, edge := range result.Edges {
		if edge.RemoteResolved {
			t.Errorf("no edge should be resolved/importable, got %+v", edge)
		}
	}
}

func TestCorrelate_WarnsHostNotInDB(t *testing.T) {
	result := &topology.ConnectionScanResult{
		Hosts: []topology.HostScan{
			// responded to SNMP but isn't a DB device -> should warn
			{Host: "10.0.0.245", Device: &s.SNMPDevice{SysName: "HeartOfGold", Reachable: true}},
			// in the DB (matched by IP -> DeviceLabel set) -> no warning
			{Host: "10.0.0.1", DeviceLabel: "opnsense"},
		},
	}
	correlateEvidence(result, labPorts(), []string{"opnsense", "openwrt"}, nil)

	warned := false
	for _, d := range result.Discrepancies {
		if d.Kind == "host-not-in-db" {
			if !strings.Contains(d.Detail, "HeartOfGold") || !strings.Contains(d.Detail, "10.0.0.245") {
				t.Errorf("unexpected warning detail: %q", d.Detail)
			}
			warned = true
		}
	}
	if !warned {
		t.Fatalf("expected a host-not-in-db warning for HeartOfGold; got %+v", result.Discrepancies)
	}
}

func TestCorrelate_AgnosticEdgesNoDB(t *testing.T) {
	result := &topology.ConnectionScanResult{
		Hosts: []topology.HostScan{
			{Host: "10.0.2.246", LocalSysName: "OpenWrt", LocalChassisMAC: "74:83:c2:f5:ea:12",
				Evidence: []topology.NeighborEvidence{
					{Source: topology.SourceSSHLLDP, ObservedHost: "10.0.2.246", LocalPort: "eth1",
						RemoteChassisMAC: "6c:cd:d6:d1:9c:3e", RemoteSysName: "Netgear-PBaja", RemotePort: "eth0"},
				}},
			{Host: "10.0.2.242", LocalSysName: "Netgear-PBaja", LocalChassisMAC: "6c:cd:d6:d1:9c:3e",
				Evidence: []topology.NeighborEvidence{
					{Source: topology.SourceSSHLLDP, ObservedHost: "10.0.2.242", LocalPort: "eth0",
						RemoteChassisMAC: "74:83:c2:f5:ea:12", RemoteSysName: "OpenWrt", RemotePort: "eth1"},
				}},
		},
	}
	// Empty DB: no ports, no device names — edges must still be derived.
	correlateEvidence(result, nil, nil, nil)

	var found *topology.ConnectionEdge
	for i := range result.Edges {
		e := result.Edges[i]
		if (e.FromLabel == "OpenWrt:eth1" && e.ToLabel == "Netgear-PBaja:eth0") ||
			(e.FromLabel == "Netgear-PBaja:eth0" && e.ToLabel == "OpenWrt:eth1") {
			found = &result.Edges[i]
		}
	}
	if found == nil {
		t.Fatalf("expected agnostic edge OpenWrt:eth1 <-> Netgear-PBaja:eth0; got %+v", result.Edges)
	}
	if found.Confidence != topology.ConfidenceConfirmed {
		t.Errorf("expected confirmed (both ends), got %q", found.Confidence)
	}
}

func TestProfileSSHCreds_GenericFallback(t *testing.T) {
	// No SSH user -> no usable credentials.
	if c := profileSSHCreds(&e.ScanProfile{Kind: "generic"}, ""); c != nil {
		t.Errorf("expected nil creds without an SSH user, got %+v", c)
	}

	// Key file: used verbatim, no passphrase needed.
	c := profileSSHCreds(&e.ScanProfile{Kind: "generic", SSHUser: "root", SSHKeyFile: "/k/id", SSHPort: 2222}, "")
	if c == nil || c.Username != "root" || c.KeyFile != "/k/id" || c.Port != 2222 {
		t.Fatalf("keyfile creds wrong: %+v", c)
	}

	// Encrypted in-memory private key: decrypted with the passphrase into PrivateKey.
	blob, err := secret.Encrypt("PEMDATA", "pw")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	c = profileSSHCreds(&e.ScanProfile{Kind: "generic", SSHUser: "root", SSHKey: blob}, "pw")
	if c == nil || c.PrivateKey != "PEMDATA" {
		t.Fatalf("expected decrypted private key, got %+v", c)
	}
	// Wrong passphrase -> nil (can't unlock).
	if c := profileSSHCreds(&e.ScanProfile{Kind: "generic", SSHUser: "root", SSHKey: blob}, "wrong"); c != nil {
		t.Errorf("expected nil creds on bad passphrase, got %+v", c)
	}
}
