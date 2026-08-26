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

// ---------------------------------------------------------------------------
// Phase 3: LiveConfig
// ---------------------------------------------------------------------------

// fakeFetcher is a stand-in for the per-OS Fetcher strategy. It avoids the
// SSH/REST plumbing in the Service test.
type fakeFetcher struct {
	got *configparser.ConfigData
	err error
}

func (f *fakeFetcher) Fetch(ctx context.Context, deviceID string) (*configparser.ConfigData, error) {
	return f.got, f.err
}

func TestService_LiveConfig_PicksFetcherByOS(t *testing.T) {
	openwrtFetcher := &fakeFetcher{got: &configparser.ConfigData{OsType: "openwrt"}}
	opnFetcher := &fakeFetcher{got: &configparser.ConfigData{OsType: "opnsense"}}

	svc := NewService(
		NewMemoryPushRepository(),
		NewDefaultEngine(),
		&fakeCredentialResolver{},
		fakeDefaultOS,
	)
	// Override the fetcher lookup with a stub keyed by OS.
	svc.fetchers = map[string]Fetcher{
		"openwrt": openwrtFetcher,
		"opnsense": opnFetcher,
	}

	got, err := svc.LiveConfig(context.Background(), "R1", "openwrt")
	if err != nil {
		t.Fatalf("LiveConfig openwrt: %v", err)
	}
	if got == nil || got.OsType != "openwrt" {
		t.Errorf("openwrt fetcher not called; got %+v", got)
	}

	got, err = svc.LiveConfig(context.Background(), "R2", "opnsense")
	if err != nil {
		t.Fatalf("LiveConfig opnsense: %v", err)
	}
	if got == nil || got.OsType != "opnsense" {
		t.Errorf("opnsense fetcher not called; got %+v", got)
	}
}

func TestService_LiveConfig_UnknownOS_ReturnsErrUnsupported(t *testing.T) {
	svc := NewService(
		NewMemoryPushRepository(),
		NewDefaultEngine(),
		&fakeCredentialResolver{},
		fakeDefaultOS,
	)
	svc.fetchers = map[string]Fetcher{}
	_, err := svc.LiveConfig(context.Background(), "R1", "vyos")
	if err == nil {
		t.Fatal("LiveConfig on unregistered OS must error")
	}
	var unsup configparser.ErrUnsupported
	if !errors.As(err, &unsup) {
		t.Errorf("must return ErrUnsupported; got %v", err)
	}
}

func TestService_LiveConfig_PropagatesError(t *testing.T) {
	svc := NewService(
		NewMemoryPushRepository(),
		NewDefaultEngine(),
		&fakeCredentialResolver{},
		fakeDefaultOS,
	)
	svc.fetchers = map[string]Fetcher{
		"openwrt": &fakeFetcher{err: errors.New("ssh unreachable")},
	}
	_, err := svc.LiveConfig(context.Background(), "R1", "openwrt")
	if err == nil {
		t.Fatal("LiveConfig must propagate fetcher errors")
	}
	if !strings.Contains(err.Error(), "ssh unreachable") {
		t.Errorf("error must wrap fetcher error; got %v", err)
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

// ---------------------------------------------------------------------------
// Phase 4: FetchAndSnapshot
// ---------------------------------------------------------------------------

// stubFetcherAndRaw is a stand-in for an OS that supports both Fetcher
// and RawFetcher (e.g. OpenWrt).
type stubFetcherAndRaw struct {
	cd     *configparser.ConfigData
	raw    []byte
	ext    string
	err    error
	rawErr error
}

func (s *stubFetcherAndRaw) Fetch(ctx context.Context, deviceID string) (*configparser.ConfigData, error) {
	return s.cd, s.err
}

func (s *stubFetcherAndRaw) FetchRaw(ctx context.Context, deviceID string) ([]byte, string, error) {
	return s.raw, s.ext, s.rawErr
}

func TestService_FetchAndSnapshot_WritesParsedAndBackup(t *testing.T) {
	repo := NewMemoryPushRepository()
	svc := NewService(repo, NewDefaultEngine(), &fakeCredentialResolver{}, fakeDefaultOS)
	fetcher := &stubFetcherAndRaw{
		cd:  &configparser.ConfigData{Hostname: "openwrt-1", OsType: "openwrt"},
		raw: []byte("config system 'foo'\n\toption bar 'baz'\n"),
		ext: "conf",
	}
	svc.SetFetcher("openwrt", fetcher)

	id, err := svc.FetchAndSnapshot(context.Background(), "R1", "openwrt", "run-1")
	if err != nil {
		t.Fatalf("FetchAndSnapshot: %v", err)
	}
	if id == "" {
		t.Fatal("FetchAndSnapshot returned empty snapshot ID")
	}

	got, err := repo.GetSnapshot(id)
	if err != nil {
		t.Fatalf("GetSnapshot(%q): %v", id, err)
	}
	if got.OS != "openwrt" {
		t.Errorf("snapshot.OsType = %q; want openwrt", got.OS)
	}
	if got.CapturedByRun != "run-1" {
		t.Errorf("snapshot.CapturedByRun = %q; want run-1", got.CapturedByRun)
	}
	if got.Config == nil {
		t.Error("snapshot.Config is nil; expected parsed *ConfigData")
	}
	if got.BackupRelpath == "" {
		t.Error("snapshot.BackupRelpath is empty; expected a backup file path")
	}

	r, err := repo.OpenBackup("R1", "run-1")
	if err != nil {
		t.Fatalf("OpenBackup: %v", err)
	}
	defer r.Close()
	buf := make([]byte, 256)
	n, _ := r.Read(buf)
	if string(buf[:n]) != string(fetcher.raw) {
		t.Errorf("backup contents = %q; want %q", buf[:n], fetcher.raw)
	}
}

func TestService_FetchAndSnapshot_NoRawFetcher_StillWritesSnapshot(t *testing.T) {
	repo := NewMemoryPushRepository()
	svc := NewService(repo, NewDefaultEngine(), &fakeCredentialResolver{}, fakeDefaultOS)
	svc.SetFetcher("opnsense", &fakeFetcher{got: &configparser.ConfigData{OsType: "opnsense"}})

	id, err := svc.FetchAndSnapshot(context.Background(), "R1", "opnsense", "run-2")
	if err != nil {
		t.Fatalf("FetchAndSnapshot: %v", err)
	}
	got, err := repo.GetSnapshot(id)
	if err != nil {
		t.Fatalf("GetSnapshot: %v", err)
	}
	if got.BackupRelpath != "" {
		t.Errorf("BackupRelpath = %q; want empty when no RawFetcher", got.BackupRelpath)
	}
	if got.Config == nil {
		t.Error("Config is nil; parsed snapshot should still be saved")
	}
}

func TestService_FetchAndSnapshot_PropagatesFetchError(t *testing.T) {
	repo := NewMemoryPushRepository()
	svc := NewService(repo, NewDefaultEngine(), &fakeCredentialResolver{}, fakeDefaultOS)
	svc.SetFetcher("openwrt", &fakeFetcher{err: errors.New("ssh down")})

	_, err := svc.FetchAndSnapshot(context.Background(), "R1", "openwrt", "run-3")
	if err == nil {
		t.Fatal("Fetch error must propagate")
	}
	if !strings.Contains(err.Error(), "ssh down") {
		t.Errorf("error must wrap fetcher error; got %v", err)
	}
	if all, _ := repo.ListSnapshots("R1", 10); len(all) != 0 {
		t.Errorf("snapshot must not be written on fetcher error; got %d", len(all))
	}
}