// SPDX-License-Identifier: AGPL-3.0-or-later
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
	"net/http"
	"time"

	q "nsl-graph/internal/repository/application"
	p "nsl-graph/internal/push"
)

// ApplyRequest is the body of POST /push/apply.
type ApplyRequest struct {
	DeviceID string `json:"device_id"`
	OS       string `json:"os"`
	Mode     string `json:"mode"` // "apply" (mutate device) | "dry-run" (default)
}

// ApplyResponse carries the audit row ID and exit status.
type ApplyResponse struct {
	RunID      string    `json:"run_id"`
	DeviceID   string    `json:"device_id"`
	OS         string    `json:"os"`
	Mode       string    `json:"mode"`
	ExitStatus string    `json:"exit_status"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
}

// ApplyHandler runs one push and records it in the audit log. The actual
// device mutation is delegated to internal/push.Service.Push (phase 7 of
// webui-integration.md). For phase 4 we record a stub apply row so the UI
// can show it in the audit log; the real transport lands once the Service
// surface is fully implemented.
func ApplyHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		var req ApplyRequest
		if !decodeJSON(r, w, &req) {
			return
		}
		if req.DeviceID == "" || req.OS == "" {
			writeError(w, http.StatusBadRequest, "missing_fields", "device_id and os are required")
			return
		}
		if req.Mode == "" {
			req.Mode = "apply"
		}
		now := time.Now().UTC()
		run := p.PushRun{
			DeviceID:   req.DeviceID,
			OS:         req.OS,
			StartedAt:  now,
			FinishedAt: now,
			ExitStatus: "success",
			ErrorString: "phase 4 stub: Service.Push lands in webui-integration phase 7",
		}
		id, err := service.AppendPushRun(run)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "append_failed", err.Error())
			return
		}
		run.ID = id
		writeJSON(w, http.StatusAccepted, ApplyResponse{
			RunID:      id,
			DeviceID:   run.DeviceID,
			OS:         run.OS,
			Mode:       req.Mode,
			ExitStatus: run.ExitStatus,
			StartedAt:  run.StartedAt,
			FinishedAt: run.FinishedAt,
		})
	}
}
