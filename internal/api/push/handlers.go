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

// Package push implements the HTTP surface for the Push config tab.
//
// Five endpoints:
//
//   POST /api/v1/push/preview     Engine.Preview — read-only patch text.
//   POST /api/v1/push/apply       SafetyFloor.Record(Apply) — mutates device.
//   POST /api/v1/push/rollback    Reads latest apply row, restores snapshot.
//   GET  /api/v1/push/history     PushRun audit log, newest first.
//   GET  /api/v1/push/topology    1-hop neighborhood of a device.
//
// Engine + SafetyFloor are taken from internal/push. The audit log is
// persisted via NetService.AppendPushRun/AllPushRuns (phase 3).
package push

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	q "nsl-graph/internal/repository/application"
)

// ErrorResponse is the package-wide error envelope.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Register mounts every push endpoint on r. Mirrors the shape used by
// internal/api/{devices,scanning,vlans}.
func Register(r *mux.Router, service q.NetServiceInt) {
	r.HandleFunc("/push/preview", PreviewHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/push/apply", ApplyHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/push/rollback", RollbackHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/push/history", HistoryHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/push/topology", TopologyHandler(service)).Methods("GET", "OPTIONS")
}

// writeJSON writes v with the standard CORS headers + content type.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError is the standard error envelope.
func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, ErrorResponse{Error: code, Message: msg})
}

// decodeJSON decodes r.Body into v, writing 400 on parse failure.
func decodeJSON(r *http.Request, w http.ResponseWriter, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return false
	}
	return true
}
