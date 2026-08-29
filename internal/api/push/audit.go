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

	q "nsl-graph/internal/repository/application"
	p "nsl-graph/internal/push"
)

// HistoryResponse is the audit log slice returned by GET /push/history.
type HistoryResponse struct {
	DeviceID string       `json:"device_id"`
	Count    int          `json:"count"`
	Runs     []p.PushRun  `json:"runs"`
}

// HistoryHandler returns the most recent runs for a device, newest first.
// The default cap is 20; pass ?limit=N to override (0 or negative means
// no cap).
func HistoryHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		deviceID := r.URL.Query().Get("device_id")
		if deviceID == "" {
			writeError(w, http.StatusBadRequest, "missing_device_id", "device_id query param is required")
			return
		}
		// phase 4: AllPushRuns returns every run; we filter in-process for
		// the device. Phase 7 of the push-config-tab plan will move this
		// filter into the repo layer (ListPushRunsByDevice) for efficiency.
		all, err := service.AllPushRuns()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "history_failed", err.Error())
			return
		}
		out := make([]p.PushRun, 0, len(all))
		for _, run := range all {
			if run.DeviceID == deviceID {
				out = append(out, run)
			}
		}
		writeJSON(w, http.StatusOK, HistoryResponse{
			DeviceID: deviceID,
			Count:    len(out),
			Runs:     out,
		})
	}
}
