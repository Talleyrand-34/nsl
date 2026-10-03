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
package scanning

import (
	"encoding/json"
	"net/http"

	"nsl-graph/internal/datastore"
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

// ImportConnectionsRequest carries the edges to persist.
type ImportConnectionsRequest struct {
	Edges []topology.ConnectionEdge `json:"edges"`

	// AcceptWeak records that an operator has REVIEWED the weak edges in this request
	// and accepts them despite the evidence.
	//
	// It defaults to false, and that is the point. A weak edge rests only on the
	// forwarding database: a MAC was seen on a port, which does not prove adjacency —
	// the device owning it may be several hops away behind an unmanaged switch. Without
	// this flag such edges are refused, and reported in Rejected.
	//
	// This endpoint used to import them silently, because the rule lived in the CLI and
	// the API never met it.
	AcceptWeak bool `json:"accept_weak,omitempty"`
}

// ImportConnectionsResponse reports how many connections were created, and every edge
// the commit rules refused.
type ImportConnectionsResponse struct {
	Imported int    `json:"imported"`
	Message  string `json:"message,omitempty"`

	// Rejected lists the edges that were NOT committed and why. It is not an error:
	// a scan of a real network always turns up edges that cannot be committed, and
	// refusing the whole request because of them would make discovery useless.
	Rejected []RejectedEdge `json:"rejected,omitempty"`
}

// RejectedEdge is one edge the commit rules refused.
type RejectedEdge struct {
	Edge   string `json:"edge"`   // "from <-> to"
	Rule   string `json:"rule"`   // the rule it broke
	Reason string `json:"reason"` // in words an operator can act on
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

		// Stage, then let the commit rules decide. A weak edge is committed only when
		// the caller has explicitly said it reviewed and accepted it.
		var c datastore.Candidate
		for _, edge := range req.Edges {
			c.Stage(edge, req.AcceptWeak)
		}

		n, violations, err := service.ImportConnectionEdgesChecked(c)

		resp := ImportConnectionsResponse{Imported: n}
		for _, v := range violations {
			resp.Rejected = append(resp.Rejected, RejectedEdge{
				Edge:   v.Target,
				Rule:   v.Rule,
				Reason: v.Detail,
			})
		}
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
