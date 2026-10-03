// SPDX-License-Identifier: AGPL-3.0-or-later
// opnsense_renderer_test.go: TDD for the OPNsense renderer end-to-end.
// Covers the fetch-observed → diff → applyChange → commit sequence for both
// VLANs (existing) and routes (new).
package parsers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"nsl-graph/internal/configparser"
)

// opnsenseLab is an in-process OPNsense REST simulator. Only the endpoints
// the renderer touches are implemented; everything else 404s so a stray
// new call surfaces as a test failure rather than passing silently.
type opnsenseLab struct {
	mu           sync.Mutex
	paths        []string
	interfacesRows string
	routesRows     string
	gatewaysRows   string
	commitOK     bool
}

func (l *opnsenseLab) record(p string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.paths = append(l.paths, p)
}

func (l *opnsenseLab) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]string, len(l.paths))
	copy(out, l.paths)
	return out
}

func (l *opnsenseLab) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.record(r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.URL.Path == "/api/interfaces/overview/list" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(l.interfacesRows))
	case r.URL.Path == "/api/interfaces/overview/commit" && r.Method == http.MethodPost:
		if !l.commitOK {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(`{"status":"failed","msg":"commit rejected"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	case r.URL.Path == "/api/routes/routes/searchroute" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(l.routesRows))
	case r.URL.Path == "/api/routes/routes/addroute" && r.Method == http.MethodPost:
		var env struct {
			Route map[string]any `json:"route"`
		}
		_ = json.NewDecoder(r.Body).Decode(&env)
		_, _ = w.Write([]byte(`{"status":"saved"}`))
	case r.URL.Path == "/api/routes/routes/delroute" && r.Method == http.MethodPost:
		_, _ = w.Write([]byte(`{"status":"deleted"}`))
	case r.URL.Path == "/api/routes/routes/reconfigure" && r.Method == http.MethodPost:
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	case r.URL.Path == "/api/routing/settings/searchGateway" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(l.gatewaysRows))
	default:
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"status":"failed","msg":"unhandled"}`))
	}
}

func newLab(interfacesRows, routesRows string) *opnsenseLab {
	return &opnsenseLab{
		interfacesRows: interfacesRows,
		routesRows:     routesRows,
		gatewaysRows:   `{"rows":[{"uuid":"gw1","name":"WAN_DHCP","address":"192.168.1.1","gateway_type":"ipv4"}]}`,
		commitOK:       true,
	}
}

func newOpnsenseRendererForTest(t *testing.T, lab *opnsenseLab) configparser.ConfigRenderer {
	t.Helper()
	srv := httptest.NewTLSServer(lab)
	t.Cleanup(srv.Close)
	return NewOpnsenseRendererForURL(srv.URL, "k", "s")
}

// ---------------------------------------------------------------------------
// route push — red-green coverage of phase 1
// ---------------------------------------------------------------------------

func TestOPNsenseRenderer_Apply_RouteAdd(t *testing.T) {
	lab := newLab(
		`{"rows":[{"device":"vtnet0","identifier":"wan"}]}`, // observed: no VLANs, no routes
		`{"rows":[]}`,                                       // observed routes: empty
	)
	r := newOpnsenseRendererForTest(t, lab)

	intent := &configparser.ConfigData{
		Hostname: "opnsense-1",
		Routes: []configparser.ConfigRoute{
			{Network: "10.0.50.0/24", Gateway: "WAN_DHCP", Interface: "vtnet0"},
		},
	}
	if err := r.Render(configparser.SafetyApply, intent, nil, configparser.SSHCredentials{}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	paths := lab.snapshot()
	var sawList, sawSearch, sawAdd, sawReconfigure bool
	for _, p := range paths {
		switch {
		case strings.HasSuffix(p, "/interfaces/overview/list"):
			sawList = true
		case strings.HasSuffix(p, "/routes/routes/searchroute"):
			sawSearch = true
		case strings.HasSuffix(p, "/routes/routes/addroute"):
			sawAdd = true
		case strings.HasSuffix(p, "/routes/routes/reconfigure"):
			sawReconfigure = true
		}
	}
	if !sawList {
		t.Errorf("must fetch interfaces/observed; paths: %v", paths)
	}
	if !sawSearch {
		t.Errorf("must fetch routes/observed; paths: %v", paths)
	}
	if !sawAdd {
		t.Errorf("must POST /api/routes/routes/addroute for the missing route; paths: %v", paths)
	}
	if !sawReconfigure {
		t.Errorf("must POST /api/routes/routes/reconfigure to commit route changes; paths: %v", paths)
	}
}

func TestOPNsenseRenderer_Apply_RouteAlreadyPresent_NoCommit(t *testing.T) {
	lab := newLab(
		`{"rows":[{"device":"vtnet0","identifier":"wan"}]}`,
		`{"rows":[{"uuid":"r1","network":"10.0.50.0/24","gateway":"WAN_DHCP","descr":"existing"}]}`,
	)
	r := newOpnsenseRendererForTest(t, lab)

	intent := &configparser.ConfigData{
		Hostname: "opnsense-1",
		Routes: []configparser.ConfigRoute{
			{Network: "10.0.50.0/24", Gateway: "WAN_DHCP", Interface: "vtnet0"},
		},
	}
	if err := r.Render(configparser.SafetyApply, intent, nil, configparser.SSHCredentials{}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	for _, p := range lab.snapshot() {
		if strings.HasSuffix(p, "/routes/routes/addroute") {
			t.Errorf("must NOT POST addroute when route already exists; paths: %v", lab.snapshot())
		}
		if strings.HasSuffix(p, "/interfaces/overview/commit") {
			t.Errorf("must NOT commit when diff is empty; paths: %v", lab.snapshot())
		}
	}
}

func TestOPNsenseRenderer_DryRun_NoNetworkForRoutes(t *testing.T) {
	lab := newLab(`{"rows":[]}`, `{"rows":[]}`)
	r := newOpnsenseRendererForTest(t, lab)

	intent := &configparser.ConfigData{
		Hostname: "opnsense-1",
		Routes:   []configparser.ConfigRoute{{Network: "10.0.50.0/24", Gateway: "WAN_DHCP"}},
	}
	if err := r.Render(configparser.SafetyDryRun, intent, nil, configparser.SSHCredentials{}); err != nil {
		t.Fatalf("DryRun: %v", err)
	}
	for _, p := range lab.snapshot() {
		t.Errorf("DryRun must not call the network; saw %q", p)
	}
}

func TestOPNsenseRenderer_FetchObserved_ParsesRoutes(t *testing.T) {
	lab := newLab(
		`{"rows":[{"device":"vtnet0","identifier":"wan"}]}`,
		`{"rows":[{"uuid":"r1","network":"10.0.50.0/24","gateway":"WAN_DHCP","descr":"a"}]}`,
	)
	r := newOpnsenseRendererForTest(t, lab).(*opnsenseRenderer)

	obs, err := r.fetchObserved(t.Context())
	if err != nil {
		t.Fatalf("fetchObserved: %v", err)
	}
	if len(obs.Routes) != 1 {
		t.Fatalf("want 1 observed route, got %d", len(obs.Routes))
	}
	if obs.Routes[0].Network != "10.0.50.0/24" || obs.Routes[0].Gateway != "WAN_DHCP" {
		t.Errorf("route round-trip failed: %+v", obs.Routes[0])
	}
}