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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"nsl-graph/internal/repository/application"
	infra "nsl-graph/internal/repository/infra/cloverdb/base"
)

// newTestRouter wires Register against an in-memory clover repo. The
// service is the only thing the handlers need; the engine call inside
// preview uses an empty (intent, observed) pair which always renders
// "" for an empty diff.
func newTestRouter(t *testing.T) (*mux.Router, application.NetServiceInt, func()) {
	t.Helper()
	dir := t.TempDir()
	repo, err := infra.NewCloverRepository(dir)
	if err != nil {
		t.Fatalf("NewCloverRepository: %v", err)
	}
	cleanup := func() { _ = repo } // tempdir auto-cleaned by testing
	svc := application.NewNetService(repo)
	r := mux.NewRouter()
	Register(r, svc)
	return r, svc, cleanup
}

func TestPreviewHandler_NoChanges(t *testing.T) {
	r, _, cleanup := newTestRouter(t)
	defer cleanup()

	body := strings.NewReader(`{"device_id":"R1","os":"openwrt","intent":{},"observed":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/push/preview", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body=%s", w.Code, w.Body.String())
	}
	var resp PreviewResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.DeviceID != "R1" {
		t.Errorf("device_id: got %q, want R1", resp.DeviceID)
	}
	if resp.PatchText != "" {
		t.Errorf("expected empty patch text, got %q", resp.PatchText)
	}
	if resp.ChangeNum != 0 {
		t.Errorf("expected 0 changes, got %d", resp.ChangeNum)
	}
}

func TestApplyHandler_AppendsRun(t *testing.T) {
	r, svc, cleanup := newTestRouter(t)
	defer cleanup()

	body := strings.NewReader(`{"device_id":"R1","os":"openwrt","mode":"apply"}`)
	req := httptest.NewRequest(http.MethodPost, "/push/apply", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("status: got %d, want 202; body=%s", w.Code, w.Body.String())
	}
	runs, err := svc.AllPushRuns()
	if err != nil {
		t.Fatalf("AllPushRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].DeviceID != "R1" {
		t.Errorf("DeviceID: got %q, want R1", runs[0].DeviceID)
	}
}

func TestRollbackHandler_AppendsRun(t *testing.T) {
	r, svc, cleanup := newTestRouter(t)
	defer cleanup()

	body := strings.NewReader(`{"device_id":"R1"}`)
	req := httptest.NewRequest(http.MethodPost, "/push/rollback", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("status: got %d, want 202; body=%s", w.Code, w.Body.String())
	}
	runs, _ := svc.AllPushRuns()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
}

func TestHistoryHandler_FiltersByDevice(t *testing.T) {
	r, _, cleanup := newTestRouter(t)
	defer cleanup()

	for _, dev := range []string{"R1", "R2", "R1"} {
		body := strings.NewReader(`{"device_id":"` + dev + `","os":"openwrt"}`)
		req := httptest.NewRequest(http.MethodPost, "/push/apply", body)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
	req := httptest.NewRequest(http.MethodGet, "/push/history?device_id=R1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body=%s", w.Code, w.Body.String())
	}
	var resp HistoryResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("expected 2 R1 runs, got %d", resp.Count)
	}
}

func TestTopologyHandler_EmptyNeighborhood(t *testing.T) {
	r, _, cleanup := newTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/push/topology?device_id=R1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body=%s", w.Code, w.Body.String())
	}
	var resp TopologyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.DeviceID != "R1" {
		t.Errorf("DeviceID: got %q, want R1", resp.DeviceID)
	}
	if len(resp.Neighbor) != 0 {
		t.Errorf("expected empty neighbors, got %d", len(resp.Neighbor))
	}
}

func TestPreviewHandler_RejectsMissingDeviceID(t *testing.T) {
	r, _, cleanup := newTestRouter(t)
	defer cleanup()

	body := strings.NewReader(`{"os":"openwrt"}`)
	req := httptest.NewRequest(http.MethodPost, "/push/preview", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400; body=%s", w.Code, w.Body.String())
	}
}
