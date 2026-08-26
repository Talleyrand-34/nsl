// SPDX-License-Identifier: MIT
// service_test.go: TDD for the canonical Service.
//
// Phase 1 lands just the Service skeleton (methods panic). These tests
// pin the wiring — they exercise NewService + deviceLock + MemoryPushRepository,
// which are the parts that exist today. The Preview / Push / History /
// Rollback tests live alongside their respective phase implementations.
package push

import (
	"context"
	"sync"
	"testing"

	"nsl-graph/internal/configparser"
)

// fakeCredentialResolver returns fixed creds for any device. Phase 2
// will use it to drive LiveConfig; today it just proves the wiring.
type fakeCredentialResolver struct {
	ssh  configparser.SSHCredentials
	api  string
	fail error
}

func (f *fakeCredentialResolver) SSH(deviceID string) (configparser.SSHCredentials, error) {
	if f.fail != nil {
		return configparser.SSHCredentials{}, f.fail
	}
	return f.ssh, nil
}

func (f *fakeCredentialResolver) API(deviceID, os string) (string, string, string, error) {
	if f.fail != nil {
		return "", "", "", f.fail
	}
	return f.api, "k", "s", nil
}

func fakeDefaultOS(deviceID string) (string, error) {
	return "openwrt", nil
}

func TestNewService_WiresFields(t *testing.T) {
	repo := NewMemoryPushRepository()
	creds := &fakeCredentialResolver{api: "https://opnsense.local"}
	svc := NewService(repo, NewDefaultEngine(), creds, fakeDefaultOS)

	if svc.repo != repo {
		t.Error("repo not wired")
	}
	if svc.engine == nil {
		t.Error("engine not wired")
	}
	if svc.creds != creds {
		t.Error("creds not wired")
	}
	if svc.defaultOS == nil {
		t.Error("defaultOS not wired")
	}
	if svc.rendererName != "typed" {
		t.Errorf("rendererName = %q; want \"typed\"", svc.rendererName)
	}
}

func TestService_DeviceLock_Serialises(t *testing.T) {
	svc := NewService(NewMemoryPushRepository(), NewDefaultEngine(), &fakeCredentialResolver{}, fakeDefaultOS)

	mu1 := svc.deviceLock("R1")
	mu2 := svc.deviceLock("R2")
	if mu1 == mu2 {
		t.Fatal("different deviceIDs must yield different mutex instances")
	}
	mu1b := svc.deviceLock("R1")
	if mu1 != mu1b {
		t.Fatal("same deviceID must return the same mutex instance")
	}
}

func TestService_DeviceLock_Concurrent(t *testing.T) {
	svc := NewService(NewMemoryPushRepository(), NewDefaultEngine(), &fakeCredentialResolver{}, fakeDefaultOS)

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			mu := svc.deviceLock("R1")
			mu.Lock()
			defer mu.Unlock()
		}()
	}
	wg.Wait()
}

func TestMemoryPushRepository_RoundTrip(t *testing.T) {
	repo := NewMemoryPushRepository()
	run := PushRun{ID: "r1", DeviceID: "R1", OS: "openwrt"}
	if err := repo.AppendRun(run); err != nil {
		t.Fatalf("AppendRun: %v", err)
	}
	got, err := repo.ListRuns("R1", 10)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(got) != 1 || got[0].ID != "r1" {
		t.Errorf("round-trip failed: %+v", got)
	}
	if _, err := repo.LatestRun("R1"); err != nil {
		t.Errorf("LatestRun: %v", err)
	}
	if _, err := repo.LatestRun("missing"); err == nil {
		t.Error("LatestRun(missing) must return an error")
	}
}

func TestService_Preview_PanicsUntilPhase2(t *testing.T) {
	// Phase 1 contract: Preview is a stub. Calling it panics with a
	// pointer to the plan. Phase 2 replaces the body and removes this
	// test.
	svc := NewService(NewMemoryPushRepository(), NewDefaultEngine(), &fakeCredentialResolver{}, fakeDefaultOS)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Preview must panic in phase 1")
		}
	}()
	_, _ = svc.Preview(context.Background(), "R1", "openwrt")
}