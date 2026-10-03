// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

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

func TestEnumerateCIDR_Multi(t *testing.T) {
	// Two /30s (all 4 addresses each, network+broadcast included) plus a bare IP.
	got := enumerateCIDR("10.0.0.0/30, 10.0.1.0/30, 10.0.2.5")
	want := []string{
		"10.0.0.0", "10.0.0.1", "10.0.0.2", "10.0.0.3",
		"10.0.1.0", "10.0.1.1", "10.0.1.2", "10.0.1.3",
		"10.0.2.5",
	}
	if len(got) != len(want) {
		t.Fatalf("enumerateCIDR multi = %v (len %d), want len %d", got, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("enumerateCIDR multi = %v, want %v", got, want)
		}
	}

	// Overlapping ranges must not produce duplicates.
	dup := enumerateCIDR("10.0.0.0/30,10.0.0.0/29")
	seen := map[string]bool{}
	for _, ip := range dup {
		if seen[ip] {
			t.Fatalf("duplicate address %s in %v", ip, dup)
		}
		seen[ip] = true
	}
}

func TestProfileSSHCreds_GenericFallback(t *testing.T) {
	var store string
	v := secret.NewVault(
		func() (string, error) { return store, nil },
		func(s string) error { store = s; return nil },
		0,
	)
	if err := v.Init("master"); err != nil {
		t.Fatalf("vault init: %v", err)
	}

	// No SSH user -> no usable credentials.
	if c := profileSSHCreds(&e.ScanProfile{Kind: "generic"}, v); c != nil {
		t.Errorf("expected nil creds without an SSH user, got %+v", c)
	}

	// Key file: used verbatim, no decryption needed.
	c := profileSSHCreds(&e.ScanProfile{Kind: "generic", SSHUser: "root", SSHKeyFile: "/k/id", SSHPort: 2222}, v)
	if c == nil || c.Username != "root" || c.KeyFile != "/k/id" || c.Port != 2222 {
		t.Fatalf("keyfile creds wrong: %+v", c)
	}

	// Encrypted in-memory private key: decrypted via the unlocked vault.
	blob, err := v.Encrypt("PEMDATA")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	c = profileSSHCreds(&e.ScanProfile{Kind: "generic", SSHUser: "root", SSHKey: blob}, v)
	if c == nil || c.PrivateKey != "PEMDATA" {
		t.Fatalf("expected decrypted private key, got %+v", c)
	}
	// Locked vault -> nil (can't decrypt).
	v.Lock()
	if c := profileSSHCreds(&e.ScanProfile{Kind: "generic", SSHUser: "root", SSHKey: blob}, v); c != nil {
		t.Errorf("expected nil creds when vault is locked, got %+v", c)
	}
}

// A from-db connections scan must exclude DB devices that have no scan profile
// and surface a "device-no-profile" soft-warning naming them (rather than
// silently scanning them). With the only device lacking a profile, no targets
// remain, so this needs no network access.
func TestDiscoverConnections_DeviceNoProfileWarning(t *testing.T) {
	service, cleanup := setupTestScanningService(t)
	defer cleanup()

	if err := service.EnsureModelExists("M1", "B1", "B1", ""); err != nil {
		t.Fatalf("EnsureModelExists: %v", err)
	}
	if err := service.AddDevice("dev-noprof", "M1", "", "Generic", "Discovered", false, false); err != nil {
		t.Fatalf("AddDevice: %v", err)
	}
	devs, err := service.GetDevices()
	if err != nil {
		t.Fatalf("GetDevices: %v", err)
	}
	var devID string
	for _, d := range devs {
		if d.Label == "dev-noprof" {
			devID = d.ID
		}
	}
	if devID == "" {
		t.Fatal("created device not found")
	}
	// Give it a management IP so it becomes a from-db target candidate.
	if err := service.AddDeviceInterface(devID, "eth0", "", "", nil, []string{"10.9.9.9"}, "", ""); err != nil {
		t.Fatalf("AddDeviceInterface: %v", err)
	}

	result, err := service.DiscoverConnectionsByMode(ConnectionScanOptions{FromDB: true}, nil)
	if err != nil {
		t.Fatalf("DiscoverConnectionsByMode: %v", err)
	}
	var found *topology.Discrepancy
	for i := range result.Discrepancies {
		if result.Discrepancies[i].Kind == "device-no-profile" {
			found = &result.Discrepancies[i]
		}
	}
	if found == nil {
		t.Fatalf("expected a device-no-profile discrepancy, got %+v", result.Discrepancies)
	}
	if !strings.Contains(strings.Join(found.Provenance, ","), "dev-noprof") {
		t.Errorf("expected the device label in provenance, got %v", found.Provenance)
	}
}

// Once a device is tied to a profile, its profile association round-trips
// through the repository (GetDevices exposes it).
func TestUpdateDeviceProfile_RoundTrip(t *testing.T) {
	service, cleanup := setupTestScanningService(t)
	defer cleanup()

	if err := service.EnsureModelExists("M1", "B1", "B1", ""); err != nil {
		t.Fatalf("EnsureModelExists: %v", err)
	}
	if err := service.AddDevice("dev-p", "M1", "", "Generic", "Discovered", false, false); err != nil {
		t.Fatalf("AddDevice: %v", err)
	}
	devs, _ := service.GetDevices()
	var devID string
	for _, d := range devs {
		if d.Label == "dev-p" {
			devID = d.ID
		}
	}
	if err := service.UpdateDeviceProfile(devID, "generic-ssh"); err != nil {
		t.Fatalf("UpdateDeviceProfile: %v", err)
	}
	devs, _ = service.GetDevices()
	for _, d := range devs {
		if d.ID == devID && d.Profile != "generic-ssh" {
			t.Fatalf("expected profile generic-ssh, got %q", d.Profile)
		}
	}
}
