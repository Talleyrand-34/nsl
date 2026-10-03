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

	q "nsl-graph/internal/repository/application"
)

// PreviewRequest is the body of POST /push/preview.
//
// The intent + observed fields are *configparser.ConfigData, but we keep
// them as raw JSON here so the handler does not need to import the
// configparser package for type-only purposes. The handler decodes them
// into the right shape before calling the engine.
type PreviewRequest struct {
	DeviceID string         `json:"device_id"`
	OS       string         `json:"os"`
	Intent   map[string]any `json:"intent,omitempty"`
	Observed map[string]any `json:"observed,omitempty"`
}

// PreviewResponse is what the handler writes back.
type PreviewResponse struct {
	DeviceID  string `json:"device_id"`
	OS        string `json:"os"`
	PatchText string `json:"patch_text"`
	ChangeNum int    `json:"change_count"`
}

// PreviewHandler is a thin adapter over internal/push.Engine.Preview.
// It does not snapshot or apply — it just renders the diff.
func PreviewHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		var req PreviewRequest
		if !decodeJSON(r, w, &req) {
			return
		}
		if req.DeviceID == "" {
			writeError(w, http.StatusBadRequest, "missing_device_id", "device_id is required")
			return
		}
		if req.OS == "" {
			writeError(w, http.StatusBadRequest, "missing_os", "os is required")
			return
		}
		// Phase 4 wires the real engine call. For now we delegate to the
		// existing internal/push.Engine (which supports NewDefaultEngine).
		// The handler composes (intent, observed) into ConfigData and
		// returns the rendered patch text.
		text, count, err := previewPatch(req)
		if err != nil {
			writeError(w, http.StatusBadRequest, "preview_failed", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, PreviewResponse{
			DeviceID:  req.DeviceID,
			OS:        req.OS,
			PatchText: text,
			ChangeNum: count,
		})
	}
}
