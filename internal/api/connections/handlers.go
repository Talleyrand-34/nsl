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
package connections

import (
	"net/http"

	"github.com/gorilla/mux"
	q "nsl-graph/internal/repository/application"
)

// RegisterRoutes registers all connection-related routes
func RegisterRoutes(r *mux.Router, service q.NetServiceInt) {
	// Connections
	r.HandleFunc("/connections", AddConnectionHandler(service)).Methods("POST")
	r.HandleFunc("/connections", GetConnectionsHandler(service)).Methods("GET")
	r.HandleFunc("/connections", UpdateConnectionHandler(service)).Methods("PUT")
	r.HandleFunc("/connections", DeleteConnectionHandler(service)).Methods("DELETE")

	// Connection Types
	r.HandleFunc("/connectiontypes", AddConnectionTypeHandler(service)).Methods("POST")
	r.HandleFunc("/connectiontypes", GetConnectionTypesHandler(service)).Methods("GET")
	r.HandleFunc("/connectiontypes", UpdateConnectionTypeHandler(service)).Methods("PUT")
	r.HandleFunc("/connectiontypes", DeleteConnectionTypeHandler(service)).Methods("DELETE")
}

// Placeholder handlers - these will be moved from the original files
func AddConnectionHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetConnectionsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func UpdateConnectionHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteConnectionHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func AddConnectionTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetConnectionTypesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func UpdateConnectionTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteConnectionTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}