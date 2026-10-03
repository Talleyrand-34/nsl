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

// RollbackRequest is the body of POST /push/rollback. Either DeviceID
// (latest apply for the device) or RunID (specific run) is required.
type RollbackRequest struct {
	DeviceID string `json:"device_id,omitempty"`
	RunID    string `json:"run_id,omitempty"`
}

// RollbackResponse mirrors ApplyResponse.
type RollbackResponse struct {
	RunID      string    `json:"run_id"`
	DeviceID   string    `json:"device_id"`
	SnapshotID string    `json:"snapshot_id,omitempty"`
	ExitStatus string    `json:"exit_status"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
}

// RollbackHandler restores the snapshot referenced by the most recent
// apply run for a device. For phase 4 we record a stub rollback row that
// mirrors Apply; the real snapshot fetch + restore lands once
// Service.Rollback is implemented (webui-integration phase 8).
func RollbackHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		var req RollbackRequest
		if !decodeJSON(r, w, &req) {
			return
		}
		if req.DeviceID == "" && req.RunID == "" {
			writeError(w, http.StatusBadRequest, "missing_target", "device_id or run_id is required")
			return
		}
		now := time.Now().UTC()
		run := p.PushRun{
			DeviceID:    req.DeviceID,
			OS:          "rollback",
			StartedAt:   now,
			FinishedAt:  now,
			ExitStatus:  "success",
			ErrorString: "phase 4 stub: Service.Rollback lands in webui-integration phase 8",
		}
		id, err := service.AppendPushRun(run)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "append_failed", err.Error())
			return
		}
		run.ID = id
		writeJSON(w, http.StatusAccepted, RollbackResponse{
			RunID:      id,
			DeviceID:   run.DeviceID,
			SnapshotID: req.RunID,
			ExitStatus: run.ExitStatus,
			StartedAt:  run.StartedAt,
			FinishedAt: run.FinishedAt,
		})
	}
}
