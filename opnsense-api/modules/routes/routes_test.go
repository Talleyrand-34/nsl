// SPDX-License-Identifier: AGPL-3.0-or-later
// routes_test.go: TDD for the OPNsense routes wrapper.
package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/t34/opnsense-api/opnsense"
)

// newRouter builds a small lab that records paths and bodies, returning the
// canned OPNsense response the test wants. Defaults mirror what a healthy
// OPNsense box returns for each endpoint.
func newRouter(t *testing.T, cfg lab) (*httptest.Server, *Module, *router) {
	t.Helper()
	r := &router{cfg: cfg}
	srv := httptest.NewTLSServer(http.HandlerFunc(r.serve))
	c := opnsense.NewClient(srv.URL, "k", "s", opnsense.WithInsecureTLS())
	return srv, New(c), r
}

type lab struct {
	// Per-path canned responses. Path is the *Path portion only
	// (e.g. "/api/routes/routes/addroute"). Body is the JSON written back.
	Responses map[string]string
	// statusOverride, when non-empty, replaces the parsed Status field.
	statusOverride map[string]string
	// HTTPStatus is the response code; default 200.
	HTTPStatus map[string]int
	// captureBody, when true, records the request body the wrapper sent.
	captureBody bool
}

type router struct {
	cfg     lab
	bodies  map[string][]byte
	mu      chan struct{}
}

func (r *router) serve(w http.ResponseWriter, req *http.Request) {
	body, _ := readAll(req.Body)
	if r.cfg.captureBody {
		if r.bodies == nil {
			r.bodies = map[string][]byte{}
		}
		r.bodies[req.URL.Path] = body
	}
	status := http.StatusOK
	if v, ok := r.cfg.HTTPStatus[req.URL.Path]; ok {
		status = v
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if resp, ok := r.cfg.Responses[req.URL.Path]; ok {
		_, _ = w.Write([]byte(resp))
		return
	}
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func readAll(r interface{ Read(p []byte) (int, error) }) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			out = append(out, buf[:n]...)
		}
		if err != nil {
			if err.Error() == "EOF" {
				return out, nil
			}
			return out, err
		}
	}
}

func TestRouteAdd_Success(t *testing.T) {
	srv, m, _ := newRouter(t, lab{
		Responses: map[string]string{
			"/api/routes/routes/addroute": `{"status":"saved"}`,
		},
		captureBody: true,
	})
	defer srv.Close()

	if err := m.RouteAdd(context.Background(), RouteAdd{
		Network: "10.0.50.0/24",
		Gateway: "LAN_DHCP",
		Descr:   "nsl-graph push",
	}); err != nil {
		t.Fatalf("RouteAdd: %v", err)
	}
}

func TestRouteAdd_BodyEncodesNestedRoute(t *testing.T) {
	r := &router{cfg: lab{captureBody: true}}
	srv := httptest.NewTLSServer(http.HandlerFunc(r.serve))
	defer srv.Close()
	c := opnsense.NewClient(srv.URL, "k", "s", opnsense.WithInsecureTLS())
	m := New(c)

	if err := m.RouteAdd(context.Background(), RouteAdd{
		Network: "10.0.50.0/24",
		Gateway: "WAN_DHCP",
		Descr:   "test",
	}); err != nil {
		t.Fatalf("RouteAdd: %v", err)
	}
	got := r.bodies["/api/routes/routes/addroute"]
	if len(got) == 0 {
		t.Fatal("no body captured")
	}
	var env struct {
		Route RouteAdd `json:"route"`
	}
	if err := json.Unmarshal(got, &env); err != nil {
		t.Fatalf("body not nested under route: %v (raw=%s)", err, got)
	}
	if env.Route.Network != "10.0.50.0/24" || env.Route.Gateway != "WAN_DHCP" {
		t.Errorf("payload round-trip failed: %+v", env.Route)
	}
}

func TestRouteAdd_FailedStatusSurfacesError(t *testing.T) {
	srv, m, _ := newRouter(t, lab{
		Responses: map[string]string{
			"/api/routes/routes/addroute": `{"status":"failed","msg":"invalid network"}`,
		},
	})
	defer srv.Close()

	err := m.RouteAdd(context.Background(), RouteAdd{Network: "bad", Gateway: "x"})
	if err == nil {
		t.Fatal("non-success status must surface as error")
	}
	if !strings.Contains(err.Error(), "invalid network") {
		t.Errorf("error must carry API msg; got %v", err)
	}
}

func TestRouteDel_Success(t *testing.T) {
	srv, m, _ := newRouter(t, lab{
		Responses: map[string]string{
			"/api/routes/routes/delroute/abc-123": `{"status":"deleted"}`,
		},
	})
	defer srv.Close()

	if err := m.RouteDel(context.Background(), "abc-123"); err != nil {
		t.Fatalf("RouteDel: %v", err)
	}
}

func TestRouteApply_Success(t *testing.T) {
	srv, m, _ := newRouter(t, lab{
		Responses: map[string]string{
			"/api/routes/routes/reconfigure": `{"status":"ok"}`,
		},
	})
	defer srv.Close()

	if err := m.RouteApply(context.Background()); err != nil {
		t.Fatalf("RouteApply: %v", err)
	}
}

func TestSearchRoute_RawRows(t *testing.T) {
	srv, m, _ := newRouter(t, lab{
		Responses: map[string]string{
			"/api/routes/routes/searchroute": `{"rows":[{"uuid":"r1","network":"10.0.0.0/24","gateway":"WAN_DHCP"}]}`,
		},
	})
	defer srv.Close()

	resp, err := m.SearchRoute(context.Background())
	if err != nil {
		t.Fatalf("SearchRoute: %v", err)
	}
	if !strings.Contains(string(resp.RawBody), `"uuid":"r1"`) {
		t.Errorf("raw body missing uuid row: %s", resp.RawBody)
	}
}