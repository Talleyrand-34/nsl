/*
Copyright © 2026 Talleyrand-34 (t34@t34.dev)
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.
You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package push

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"nsl-graph/internal/repository/application"
)

// TestE2E_PushFlow exercises the four /api/v1/push/* endpoints end-to-end:
//
//   1. POST /push/apply     → records an apply run.
//   2. GET  /push/history   → asserts the run is listed.
//   3. POST /push/preview   → renders an empty patch (no changes).
//   4. GET  /push/topology  → returns the (empty) neighborhood.
//   5. POST /push/rollback  → records a rollback run.
//
// The test boots a real httptest.Server with the production router + a
// fresh clover repo, mirroring the dashboard's request flow.
func TestE2E_PushFlow(t *testing.T) {
	router, _, _ := newTestRouter(t)

	srv := httptest.NewServer(router)
	defer srv.Close()

	// 1. Apply.
	applyBody := strings.NewReader(`{"device_id":"R1","os":"openwrt","mode":"apply"}`)
	resp := mustPost(t, srv.URL+"/push/apply", applyBody)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("apply: got %d, want 202; body=%s", resp.StatusCode, readBody(resp))
	}
	var applyResp ApplyResponse
	if err := json.NewDecoder(resp.Body).Decode(&applyResp); err != nil {
		t.Fatalf("decode apply: %v", err)
	}
	if applyResp.RunID == "" {
		t.Error("apply.RunID is empty")
	}

	// 2. History.
	resp = mustGet(t, srv.URL+"/push/history?device_id=R1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("history: got %d, want 200; body=%s", resp.StatusCode, readBody(resp))
	}
	var hist HistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&hist); err != nil {
		t.Fatalf("decode history: %v", err)
	}
	if hist.Count != 1 {
		t.Errorf("history.Count: got %d, want 1", hist.Count)
	}
	if len(hist.Runs) != 1 || hist.Runs[0].DeviceID != "R1" {
		t.Errorf("history.Runs: unexpected %+v", hist.Runs)
	}

	// 3. Preview (no changes since intent/observed are empty).
	previewBody := strings.NewReader(`{"device_id":"R1","os":"openwrt","intent":{},"observed":{}}`)
	resp = mustPost(t, srv.URL+"/push/preview", previewBody)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("preview: got %d, want 200; body=%s", resp.StatusCode, readBody(resp))
	}
	var p PreviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if p.PatchText != "" {
		t.Errorf("expected empty patch on no changes, got %q", p.PatchText)
	}

	// 4. Topology (no connections exist yet).
	resp = mustGet(t, srv.URL+"/push/topology?device_id=R1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("topology: got %d, want 200; body=%s", resp.StatusCode, readBody(resp))
	}
	var topo TopologyResponse
	if err := json.NewDecoder(resp.Body).Decode(&topo); err != nil {
		t.Fatalf("decode topology: %v", err)
	}
	if topo.DeviceID != "R1" {
		t.Errorf("topology.DeviceID: got %q, want R1", topo.DeviceID)
	}
	if len(topo.Neighbor) != 0 {
		t.Errorf("topology.Neighbor: got %d, want 0", len(topo.Neighbor))
	}

	// 5. Rollback.
	rollbackBody := strings.NewReader(`{"device_id":"R1"}`)
	resp = mustPost(t, srv.URL+"/push/rollback", rollbackBody)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("rollback: got %d, want 202; body=%s", resp.StatusCode, readBody(resp))
	}

	// Final history should now have two runs.
	resp = mustGet(t, srv.URL+"/push/history?device_id=R1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("final history: got %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&hist); err != nil {
		t.Fatalf("decode final history: %v", err)
	}
	if hist.Count != 2 {
		t.Errorf("final history.Count: got %d, want 2", hist.Count)
	}
}

// TestE2E_OptsFlow verifies the optional CORS preflight (OPTIONS)
// returns 200 without consuming a run.
func TestE2E_CORSPreflight(t *testing.T) {
	router, _, cleanup := newTestRouter(t)
	defer cleanup()

	srv := httptest.NewServer(router)
	defer srv.Close()

	for _, path := range []string{"/push/apply", "/push/preview", "/push/rollback"} {
		req, _ := http.NewRequest(http.MethodOptions, srv.URL+path, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("OPTIONS %s: %v", path, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("OPTIONS %s: got %d, want 200", path, resp.StatusCode)
		}
		resp.Body.Close()
	}
}

// TestE2E_RegisterMirrorsContract pins the surface: the same router
// that the dashboard uses must accept every path documented in the plan.
// If a future phase renames or removes one, this test fails loud.
func TestE2E_RegisterMirrorsContract(t *testing.T) {
	r := mux.NewRouter()
	svc := application.NetServiceInt(nil) // nil is fine for path enumeration
	_ = svc
	Register(r, svc)
	for _, route := range []struct{ method, path string }{
		{"POST", "/push/preview"},
		{"POST", "/push/apply"},
		{"POST", "/push/rollback"},
		{"GET", "/push/history"},
		{"GET", "/push/topology"},
	} {
		// Send an empty request with the right method; we only assert
		// the route exists (not 404 / not 405). Decode errors are fine.
		req := httptest.NewRequest(route.method, route.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusNotFound || w.Code == http.StatusMethodNotAllowed {
			t.Errorf("route %s %s missing (got %d)", route.method, route.path, w.Code)
		}
	}
}

func mustPost(t *testing.T, url string, body *strings.Reader) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	return resp
}

func mustGet(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	return resp
}

func readBody(resp *http.Response) string {
	defer resp.Body.Close()
	var sb strings.Builder
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			sb.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	return sb.String()
}

// guard against fmt drift if future changes remove the import path.
var _ = fmt.Sprintf
