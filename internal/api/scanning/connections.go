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
package scanning

import (
	"encoding/json"
	"net/http"

	"nsl-graph/internal/observ"
	q "nsl-graph/internal/repository/application"
	"nsl-graph/internal/topology"
)

func setConnHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// ScanConnectionsHandler runs multi-source L2/L1 link discovery and returns the
// full result (hosts, edges, intermediaries, discrepancies). It never writes to
// the DB; importing is a separate, explicit step.
func ScanConnectionsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setConnHeaders(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method_not_allowed", Message: "Only POST is allowed"})
			return
		}

		var opts q.ConnectionScanOptions
		if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_json", Message: err.Error()})
			return
		}

		title := opts.Subnet
		if title == "" {
			if opts.FromDB {
				title = "from-db"
			} else if opts.Profiles {
				title = "profiles"
			}
		}
		startAsyncScan(w, "connections", title, func(run *observ.Run) (any, error) {
			return service.DiscoverConnectionsByMode(opts, run)
		})
	}
}

// ImportConnectionsRequest carries the (reviewed) edges to persist.
type ImportConnectionsRequest struct {
	Edges []topology.ConnectionEdge `json:"edges"`
}

// ImportConnectionsResponse reports how many connections were created.
type ImportConnectionsResponse struct {
	Imported int    `json:"imported"`
	Message  string `json:"message,omitempty"`
}

// ImportConnectionsHandler persists the selected discovered edges as connections.
func ImportConnectionsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setConnHeaders(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method_not_allowed", Message: "Only POST is allowed"})
			return
		}

		var req ImportConnectionsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_json", Message: err.Error()})
			return
		}

		n, err := service.ImportConnectionEdges(req.Edges)
		resp := ImportConnectionsResponse{Imported: n}
		if err != nil {
			resp.Message = err.Error()
		}
		json.NewEncoder(w).Encode(resp)
	}
}

// ImportPlaceholdersRequest carries the (reviewed) intermediaries to materialize
// as a shared placeholder unmanaged device.
type ImportPlaceholdersRequest struct {
	Intermediaries []topology.Intermediary `json:"intermediaries"`
}

// ImportPlaceholdersHandler creates a placeholder zone + shared unmanaged device
// for the selected detected intermediaries and wires the observing endpoints to it.
func ImportPlaceholdersHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setConnHeaders(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method_not_allowed", Message: "Only POST is allowed"})
			return
		}

		var req ImportPlaceholdersRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_json", Message: err.Error()})
			return
		}

		res, err := service.CreatePlaceholderForIntermediaries(req.Intermediaries)
		out := map[string]any{
			"zone":        res.Zone,
			"device":      res.Device,
			"connections": res.Connections,
		}
		if err != nil {
			out["message"] = err.Error()
		}
		json.NewEncoder(w).Encode(out)
	}
}
