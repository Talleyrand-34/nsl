// SPDX-License-Identifier: MIT
// service_test.go: TDD for the canonical Service.
package push

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"nsl-graph/internal/configparser"
)

// fakeCredentialResolver returns fixed creds for any device.
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

// ---------------------------------------------------------------------------
// Phase 1 wiring
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Phase 2: Preview
// ---------------------------------------------------------------------------

func TestService_Preview_DelegatesToEngine(t *testing.T) {
	svc := NewService(
		NewMemoryPushRepository(),
		NewDefaultEngine(),
		&fakeCredentialResolver{},
		fakeDefaultOS,
	)
	intent := &configparser.ConfigData{
		Hostname: "openwrt-1",
		Interfaces: []configparser.ConfigInterface{
			{Name: "br-lan", Type: "bridge", Enabled: true,
				VLANs: []configparser.ConfigVLAN{{ID: "30", Tagged: true}}},
		},
	}
	observed := &configparser.ConfigData{
		Hostname: "openwrt-1",
		Interfaces: []configparser.ConfigInterface{
			{Name: "br-lan", Type: "bridge", Enabled: true},
		},
	}
	got, err := svc.PreviewDiff(context.Background(), "R1", "openwrt", intent, observed)
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}
	if !strings.Contains(got, "br-lan VLAN 30") {
		t.Errorf("Preview output missing expected VLAN path; got: %q", got)
	}
	if !strings.Contains(got, "openwrt") {
		t.Errorf("Preview output missing OS header; got: %q", got)
	}
}

func TestService_Preview_NoDiff_ReturnsEmpty(t *testing.T) {
	svc := NewService(
		NewMemoryPushRepository(),
		NewDefaultEngine(),
		&fakeCredentialResolver{},
		fakeDefaultOS,
	)
	empty := &configparser.ConfigData{Hostname: "R1"}
	got, err := svc.PreviewDiff(context.Background(), "R1", "openwrt", empty, empty)
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}
	if got != "" {
		t.Errorf("Preview with no diff must return empty string; got: %q", got)
	}
}

func TestService_Preview_UnknownOS_ReturnsErrUnsupported(t *testing.T) {
	svc := NewService(
		NewMemoryPushRepository(),
		NewDefaultEngine(),
		&fakeCredentialResolver{},
		fakeDefaultOS,
	)
	_, err := svc.PreviewDiff(context.Background(), "R1", "netscaler", &configparser.ConfigData{}, &configparser.ConfigData{})
	if err == nil {
		t.Fatal("Preview on unknown OS must error")
	}
	var unsup configparser.ErrUnsupported
	if !errors.As(err, &unsup) {
		t.Errorf("Preview on unknown OS must return ErrUnsupported; got %v", err)
	}
}