/*
Copyright © 2026 Talleyrand-34 (t34@t34.dev)
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/
package basicops

import (
	"reflect"
	"testing"

	e "nsl-graph/internal/repository/entities"
)

func TestProfileDevice_AddGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer cleanup()

	p := e.ScanProfile{Name: "lab", Kind: "device", Host: "10.0.0.245", SSHUser: "root"}
	if err := repo.AddScanProfile(p); err != nil {
		t.Fatalf("AddScanProfile: %v", err)
	}

	d1 := e.ProfileDevice{ProfileName: "lab", Host: "10.0.0.10", SSHProfileName: "creds-pool"}
	d2 := e.ProfileDevice{ProfileName: "lab", Host: "10.0.0.11", SSHConfigText: "Host *\n  User admin"}
	if err := repo.AddProfileDevice(d1); err != nil {
		t.Fatalf("AddProfileDevice d1: %v", err)
	}
	if err := repo.AddProfileDevice(d2); err != nil {
		t.Fatalf("AddProfileDevice d2: %v", err)
	}

	got, err := repo.GetProfileDevices("lab")
	if err != nil {
		t.Fatalf("GetProfileDevices: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 devices, got %d: %#v", len(got), got)
	}
	byHost := map[string]e.ProfileDevice{}
	for _, d := range got {
		byHost[d.Host] = d
	}
	if byHost["10.0.0.10"].SSHProfileName != "creds-pool" {
		t.Errorf("10.0.0.10 ssh_profile_name lost: %#v", byHost["10.0.0.10"])
	}
	if byHost["10.0.0.11"].SSHConfigText == "" {
		t.Errorf("10.0.0.11 ssh_config_text lost: %#v", byHost["10.0.0.11"])
	}
}

func TestProfileDevice_DuplicateHostInProfileRejected(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer cleanup()

	if err := repo.AddScanProfile(e.ScanProfile{Name: "lab", Kind: "device", Host: "10.0.0.1", SSHUser: "root"}); err != nil {
		t.Fatalf("AddScanProfile: %v", err)
	}
	d := e.ProfileDevice{ProfileName: "lab", Host: "10.0.0.10"}
	if err := repo.AddProfileDevice(d); err != nil {
		t.Fatalf("first AddProfileDevice: %v", err)
	}
	err = repo.AddProfileDevice(d)
	if err == nil {
		t.Fatal("expected duplicate-host error, got nil")
	}
}

func TestProfileDevice_SameHostInDifferentProfileAllowed(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer cleanup()

	for _, name := range []string{"lab-prod", "lab-staging"} {
		if err := repo.AddScanProfile(e.ScanProfile{Name: name, Kind: "device", Host: "10.0.0.1", SSHUser: "root"}); err != nil {
			t.Fatalf("AddScanProfile %s: %v", name, err)
		}
	}
	d1 := e.ProfileDevice{ProfileName: "lab-prod", Host: "10.0.0.10"}
	d2 := e.ProfileDevice{ProfileName: "lab-staging", Host: "10.0.0.10"}
	if err := repo.AddProfileDevice(d1); err != nil {
		t.Fatalf("AddProfileDevice d1: %v", err)
	}
	if err := repo.AddProfileDevice(d2); err != nil {
		t.Fatalf("AddProfileDevice d2 (same host, different profile): %v", err)
	}
}

func TestProfileDevice_Delete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer cleanup()

	if err := repo.AddScanProfile(e.ScanProfile{Name: "lab", Kind: "device", Host: "10.0.0.1"}); err != nil {
		t.Fatalf("AddScanProfile: %v", err)
	}
	if err := repo.AddProfileDevice(e.ProfileDevice{ProfileName: "lab", Host: "10.0.0.10"}); err != nil {
		t.Fatalf("AddProfileDevice: %v", err)
	}
	if err := repo.AddProfileDevice(e.ProfileDevice{ProfileName: "lab", Host: "10.0.0.11"}); err != nil {
		t.Fatalf("AddProfileDevice: %v", err)
	}
	if err := repo.DeleteProfileDevice("lab", "10.0.0.10"); err != nil {
		t.Fatalf("DeleteProfileDevice: %v", err)
	}
	got, err := repo.GetProfileDevices("lab")
	if err != nil {
		t.Fatalf("GetProfileDevices: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 device left, got %d: %#v", len(got), got)
	}
	if got[0].Host != "10.0.0.11" {
		t.Fatalf("wrong survivor: %#v", got[0])
	}
}

func TestProfileDevice_DeleteAll(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer cleanup()

	if err := repo.AddScanProfile(e.ScanProfile{Name: "lab", Kind: "device", Host: "10.0.0.1"}); err != nil {
		t.Fatalf("AddScanProfile: %v", err)
	}
	for _, h := range []string{"10.0.0.10", "10.0.0.11", "10.0.0.12"} {
		if err := repo.AddProfileDevice(e.ProfileDevice{ProfileName: "lab", Host: h}); err != nil {
			t.Fatalf("AddProfileDevice %s: %v", h, err)
		}
	}
	if err := repo.DeleteAllProfileDevices("lab"); err != nil {
		t.Fatalf("DeleteAllProfileDevices: %v", err)
	}
	got, err := repo.GetProfileDevices("lab")
	if err != nil {
		t.Fatalf("GetProfileDevices: %v", err)
	}
	if !reflect.DeepEqual(got, []e.ProfileDevice{}) {
		t.Fatalf("expected empty slice, got %#v", got)
	}
}

func TestProfileDevice_MissingFields(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer cleanup()

	if err := repo.AddProfileDevice(e.ProfileDevice{ProfileName: "", Host: "10.0.0.10"}); err == nil {
		t.Errorf("empty profile_name should error")
	}
	if err := repo.AddProfileDevice(e.ProfileDevice{ProfileName: "lab", Host: ""}); err == nil {
		t.Errorf("empty host should error")
	}
}

// Inline-custom rows carry their own encrypted credentials (no SSHProfileName
// reference). The storage layer must persist and return them verbatim.
func TestProfileDevice_InlineCredsPersisted(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer cleanup()
	if err := repo.AddScanProfile(e.ScanProfile{Name: "lab", Kind: "device", Host: "10.0.0.1"}); err != nil {
		t.Fatalf("AddScanProfile: %v", err)
	}
	d := e.ProfileDevice{
		ProfileName:    "lab",
		Host:           "10.0.0.10",
		SSHUser:        "root",
		SSHPassword:    "blob:encrypted-pw",
		SSHKey:         "blob:encrypted-key",
		SSHKeyFilename: "id_ed25519",
		SSHConfigText:  "Host *\n  User admin\n  Port 2222",
	}
	if err := repo.AddProfileDevice(d); err != nil {
		t.Fatalf("AddProfileDevice: %v", err)
	}
	got, err := repo.GetProfileDevices("lab")
	if err != nil {
		t.Fatalf("GetProfileDevices: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 row, got %d", len(got))
	}
	if got[0].SSHUser != "root" || got[0].SSHPassword != "blob:encrypted-pw" {
		t.Errorf("inline creds lost: user=%q pw=%q", got[0].SSHUser, got[0].SSHPassword)
	}
	if got[0].SSHKey != "blob:encrypted-key" || got[0].SSHKeyFilename != "id_ed25519" {
		t.Errorf("inline key lost: key=%q file=%q", got[0].SSHKey, got[0].SSHKeyFilename)
	}
	if got[0].SSHConfigText != "Host *\n  User admin\n  Port 2222" {
		t.Errorf("inline config_text lost: %q", got[0].SSHConfigText)
	}
}

func TestProfileDevice_GetEmptyProfile(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer cleanup()

	got, err := repo.GetProfileDevices("nonexistent")
	if err != nil {
		t.Fatalf("GetProfileDevices: %v", err)
	}
	if !reflect.DeepEqual(got, []e.ProfileDevice{}) {
		t.Fatalf("expected empty slice for unknown profile, got %#v", got)
	}
}