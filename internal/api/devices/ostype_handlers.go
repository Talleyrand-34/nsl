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
package devices

import (
	"encoding/json"
	"net/http"

	q "nsl-graph/internal/repository/application"
)

// AddOsTypeHandler creates an OS type (openwrt, opnsense, …).
func AddOsTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		osTypeHeaders(w, "POST")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed"})
			return
		}
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}
		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_name", "message": "name is required"})
			return
		}
		if err := service.AddOsType(req.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "add_failed", "message": err.Error()})
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "OS type created successfully", "name": req.Name})
	}
}

// GetOsTypesHandler lists OS types.
func GetOsTypesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		osTypeHeaders(w, "GET")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		osTypes, err := service.GetOsTypes()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(osTypes)
	}
}

// UpdateOsTypeHandler renames an OS type by id.
func UpdateOsTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		osTypeHeaders(w, "PUT")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "PUT" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed"})
			return
		}
		var req struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}
		if req.ID == "" || req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_field", "message": "id and name are required"})
			return
		}
		if err := service.UpdateOsType(req.ID, req.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "update_failed", "message": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "OS type updated successfully", "id": req.ID, "name": req.Name})
	}
}

// DeleteOsTypeHandler deletes an OS type by name.
func DeleteOsTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		osTypeHeaders(w, "DELETE")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "DELETE" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed"})
			return
		}
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}
		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_name", "message": "name is required"})
			return
		}
		if err := service.DeleteOsType(req.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "OS type deleted successfully", "name": req.Name})
	}
}

func osTypeHeaders(w http.ResponseWriter, methods string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", methods+", OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
