package application

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
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
