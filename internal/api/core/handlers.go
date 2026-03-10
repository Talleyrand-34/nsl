/*
Copyright © 2025 Tecdesoft (rodrigo-gonzalez@tecdesoft.es, t34@t34.dev)

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
package core

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	q "nsl-graph/internal/repository/application"
)

// RegisterRoutes registers all core routes
func RegisterRoutes(r *mux.Router, service q.NetServiceInt) {
	r.HandleFunc("/", RootHandler()).Methods("GET")
	r.HandleFunc("/diagram", GetDiagramHandler(service)).Methods("GET")
}

// RootHandler returns a list of all available endpoints
func RootHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpoints := []string{
			// /allports/all and /allports/device
			"GET    /allports/all",
			"GET    /allports/device?deviceid=<id>",

			// /brands
			"DELETE /brands",
			"GET    /brands",
			"POST   /brands",
			"PUT    /brands",

			// /connections
			"DELETE /connections",
			"GET    /connections",
			"POST   /connections",
			"PUT    /connections",

			// /connectiontypes
			"DELETE /connectiontypes",
			"GET    /connectiontypes",
			"POST   /connectiontypes",
			"PUT    /connectiontypes",

			// /deviceclasses
			"DELETE /deviceclasses",
			"GET    /deviceclasses",
			"POST   /deviceclasses",
			"PUT    /deviceclasses",

			// /deviceports
			"DELETE /deviceports",
			"GET    /deviceports",
			"POST   /deviceports",
			"PUT    /deviceports",

			// /devices
			"DELETE /devices",
			"GET    /devices",
			"POST   /devices",
			"PUT    /devices",

			// /modelports
			"DELETE /modelports",
			"GET    /modelports",
			"POST   /modelports",
			"POST   /modelports/bulk",
			"PUT    /modelports",

			// /models
			"DELETE /models",
			"GET    /models",
			"POST   /models",
			"PUT    /models",

			// /proprietaries
			"DELETE /proprietaries",
			"GET    /proprietaries",
			"POST   /proprietaries",
			"PUT    /proprietaries",

			// /zones
			"DELETE /zones",
			"GET    /zones",
			"POST   /zones",
			"PUT    /zones",

			// /zonetypes
			"DELETE /zonetypes",
			"GET    /zonetypes",
			"POST   /zonetypes",
			"PUT    /zonetypes",

			// /vlans
			"DELETE /vlans",
			"GET    /vlans",
			"POST   /vlans",
			"PUT    /vlans",

			// /diagram
			"GET    /diagram",

			// /plugins
			"GET    /plugins",
			"POST   /plugins/active",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(endpoints)
	}
}

// GetDiagramHandler generates and returns network diagrams
func GetDiagramHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Implementation to be moved from api/get.go
		w.WriteHeader(http.StatusNotImplemented)
	}
}