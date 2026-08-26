// SPDX-License-Identifier: MIT
// opnsense_e2e_test.go: end-to-end tests for the OPNsense push path.
package cmd_push_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/configparser/parsers"
	"nsl-graph/internal/push"
)

// opnsenseLabHandler is a small in-process OPNsense REST simulator. It only
// implements the endpoints the push path touches; everything else 404s so
// accidental new calls surface as test failures.
type opnsenseLabHandler struct {
	mu       sync.Mutex
	paths    []string
	commitOK bool
	rows     string
}

func newOpnsenseLab() *opnsenseLabHandler {
	return &opnsenseLabHandler{
		commitOK: true,
		rows:     `{"rows":[{"device":"vtnet0","identifier":"wan"}]}`,
	}
}

func (h *opnsenseLabHandler) record(p string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.paths = append(h.paths, p)
}

func (h *opnsenseLabHandler) snapshot() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]string, len(h.paths))
	copy(out, h.paths)
	return out
}

func (h *opnsenseLabHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.record(r.URL.Path)
	switch {
	case r.URL.Path == "/api/interfaces/overview/list" && r.Method == http.MethodGet:
		w.WriteHeader(200)
		_, _ = w.Write([]byte(h.rows))
	case r.URL.Path == "/api/routes/routes/searchroute" && r.Method == http.MethodGet:
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"rows":[]}`))
	case r.URL.Path == "/api/routing/settings/searchGateway" && r.Method == http.MethodGet:
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"rows":[]}`))
	case r.URL.Path == "/api/interfaces/overview/commit" && r.Method == http.MethodPost:
		if !h.commitOK {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(`{"status":"failed","msg":"commit rejected"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	case r.URL.Path == "/api/interfaces/vlan/add" && r.Method == http.MethodPost:
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"ok","uuid":"v-new"}`))
	default:
		w.WriteHeader(404)
	}
}

// engineFor builds a NewDefaultEngine wired to a typed renderer pointing at
// srv.URL. The WithOpnsenseFactory hook is the only place we touch the
// engine from outside; the rest of the renderers are the registry's
// defaults, so the test exercises the real wire-up.
func engineFor(srv *httptest.Server) *push.Engine {
	return push.NewDefaultEngine(push.WithOpnsenseFactory(func(_, _, _ string) configparser.ConfigRenderer {
		return parsers.NewOpnsenseRendererForURL(srv.URL, "k", "s")
	}))
}

// intentWithVLAN30 is the canonical "I want VLAN 30 on vtnet0" intent.
func intentWithVLAN30() *configparser.ConfigData {
	return &configparser.ConfigData{
		Hostname: "opnsense-1",
		Interfaces: []configparser.ConfigInterface{
			{
				Name: "vtnet0", Type: "physical", Enabled: true,
				VLANs: []configparser.ConfigVLAN{{ID: "30", Tagged: true}},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Engine wire-up
// ---------------------------------------------------------------------------

func TestOPNsenseE2E_EngineResolvesTypedRenderer(t *testing.T) {
	srv := httptest.NewTLSServer(newOpnsenseLab())
	defer srv.Close()

	engine := engineFor(srv)
	r, ok := engine.Renderers()["opnsense"]
	if !ok {
		t.Fatal("opnsense renderer must be registered in default engine")
	}
	if r.GetOsType() != "opnsense" {
		t.Errorf("renderer GetOsType = %q; want opnsense", r.GetOsType())
	}
	if err := r.Render(configparser.SafetyDryRun,
		&configparser.ConfigData{Hostname: "opnsense-1"},
		nil, configparser.SSHCredentials{}); err != nil {
		t.Fatalf("typed renderer DryRun must succeed; got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Render end-to-end
// ---------------------------------------------------------------------------

func TestOPNsenseE2E_Apply_ReachesVLANAddAndCommit(t *testing.T) {
	lab := newOpnsenseLab()
	srv := httptest.NewTLSServer(lab)
	defer srv.Close()

	r := engineFor(srv).Renderers()["opnsense"]

	if err := r.Render(configparser.SafetyApply, intentWithVLAN30(), nil, configparser.SSHCredentials{}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	var sawList, sawAdd, sawCommit bool
	for _, p := range lab.snapshot() {
		switch {
		case strings.HasSuffix(p, "/interfaces/overview/list"):
			sawList = true
		case strings.HasSuffix(p, "/interfaces/vlan/add"):
			sawAdd = true
		case strings.HasSuffix(p, "/interfaces/overview/commit"):
			sawCommit = true
		}
	}
	if !sawList {
		t.Errorf("Apply must fetch observed via /interfaces/overview/list; paths: %v", lab.snapshot())
	}
	if !sawAdd {
		t.Errorf("Apply must POST /interfaces/vlan/add for the missing VLAN; paths: %v", lab.snapshot())
	}
	if !sawCommit {
		t.Errorf("Apply must POST /interfaces/overview/commit when diff is non-empty; paths: %v", lab.snapshot())
	}
}

func TestOPNsenseE2E_Apply_CommitFailureSurfacesError(t *testing.T) {
	lab := newOpnsenseLab()
	lab.commitOK = false
	srv := httptest.NewTLSServer(lab)
	defer srv.Close()

	r := engineFor(srv).Renderers()["opnsense"]

	err := r.Render(configparser.SafetyApply, intentWithVLAN30(), nil, configparser.SSHCredentials{})
	if err == nil {
		t.Fatal("commit failure must surface as a Render error")
	}
	var unsup configparser.ErrUnsupported
	if errors.As(err, &unsup) {
		t.Errorf("server-side 500 must NOT be classified as ErrUnsupported; got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Audit-trail integration
// ---------------------------------------------------------------------------

func TestOPNsenseE2E_AuditRow_Success(t *testing.T) {
	srv := httptest.NewTLSServer(newOpnsenseLab())
	defer srv.Close()

	store := push.NewMemoryRunStore()
	floor := push.NewSafetyFloor(store, nil)
	r := engineFor(srv).Renderers()["opnsense"]
	intent := intentWithVLAN30()

	floor.Record(push.RunParams{
		DeviceID: "opnsense-test",
		OS:       "opnsense",
		Safety:   configparser.SafetyApply,
		Renderer: "opnsense-typed",
		ApplyFn: func() error {
			return r.Render(configparser.SafetyApply, intent, nil, configparser.SSHCredentials{})
		},
	})

	runs := store.All()
	if len(runs) != 1 {
		t.Fatalf("expected 1 audit row; got %d", len(runs))
	}
	got := runs[0]
	if got.ExitStatus != "success" {
		t.Errorf("ExitStatus = %q; want success; ErrorString=%q", got.ExitStatus, got.ErrorString)
	}
	if got.OS != "opnsense" {
		t.Errorf("OS = %q; want opnsense", got.OS)
	}
	if got.Safety != configparser.SafetyApply {
		t.Errorf("Safety = %d; want SafetyApply(%d)", got.Safety, configparser.SafetyApply)
	}
	if got.DeviceID != "opnsense-test" {
		t.Errorf("DeviceID = %q; want opnsense-test", got.DeviceID)
	}
}

func TestOPNsenseE2E_AuditRow_CommitFailure(t *testing.T) {
	lab := newOpnsenseLab()
	lab.commitOK = false
	srv := httptest.NewTLSServer(lab)
	defer srv.Close()

	store := push.NewMemoryRunStore()
	floor := push.NewSafetyFloor(store, nil)
	r := engineFor(srv).Renderers()["opnsense"]
	intent := intentWithVLAN30()

	floor.Record(push.RunParams{
		DeviceID: "opnsense-test",
		OS:       "opnsense",
		Safety:   configparser.SafetyApply,
		Renderer: "opnsense-typed",
		ApplyFn: func() error {
			return r.Render(configparser.SafetyApply, intent, nil, configparser.SSHCredentials{})
		},
	})

	runs := store.All()
	if len(runs) != 1 {
		t.Fatalf("expected 1 audit row; got %d", len(runs))
	}
	got := runs[0]
	if got.ExitStatus != "failed" {
		t.Errorf("ExitStatus = %q; want failed", got.ExitStatus)
	}
	if got.ErrorString == "" {
		t.Error("failed audit row must carry ErrorString")
	}
}
