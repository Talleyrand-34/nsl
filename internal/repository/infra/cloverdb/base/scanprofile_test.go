package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

func TestScanProfile_CRUD(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer cleanup()

	p := e.ScanProfile{
		Name:          "opnsense",
		Host:          "10.0.0.1",
		SNMPCommunity: "public",
		SNMPVersion:   "v2c",
		SNMPPort:      161,
		ScanSource:    "ssh",
		OsType:        "opnsense",
		SSHUser:       "admin",
		SSHPassword:   "ENCBLOB",
		SSHPort:       22,
		VLANAccuracy:  2,
	}
	if err := repo.AddScanProfile(p); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := repo.AddScanProfile(p); err == nil {
		t.Fatal("expected duplicate-name error")
	}

	got, err := repo.GetScanProfileByName("opnsense")
	if err != nil || got == nil {
		t.Fatalf("getByName: %v / %v", err, got)
	}
	if got.Host != "10.0.0.1" || got.SSHUser != "admin" || got.SSHPassword != "ENCBLOB" ||
		got.SNMPPort != 161 || got.VLANAccuracy != 2 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}

	byHost, err := repo.GetScanProfileByHost("10.0.0.1")
	if err != nil || byHost == nil || byHost.Name != "opnsense" {
		t.Fatalf("getByHost: %v / %v", err, byHost)
	}

	p.SNMPCommunity = "private"
	if err := repo.UpdateScanProfile(p); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ = repo.GetScanProfileByName("opnsense")
	if got.SNMPCommunity != "private" {
		t.Fatalf("update not applied: %q", got.SNMPCommunity)
	}

	if err := repo.DeleteScanProfile("opnsense"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, _ = repo.GetScanProfileByName("opnsense")
	if got != nil {
		t.Fatal("expected nil after delete")
	}
}
