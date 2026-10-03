// SPDX-License-Identifier: AGPL-3.0-or-later
/*
Copyright © 2025 Talleyrand-34 (t34@t34.dev)

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
package vlans

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	q "nsl-graph/internal/repository/application"
)

// RegisterRoutes registers all VLAN-related routes
func RegisterRoutes(r *mux.Router, service q.NetServiceInt) {
	// VLANs
	r.HandleFunc("/vlans", AddVlanHandler(service)).Methods("POST")
	r.HandleFunc("/vlans", GetVlansHandler(service)).Methods("GET")
	r.HandleFunc("/vlans", UpdateVlanHandler(service)).Methods("PUT")
	r.HandleFunc("/vlans", DeleteVlanHandler(service)).Methods("DELETE")
}

// Placeholder handlers - these will be moved from the original files
func AddVlanHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed", "message": "Only POST method is allowed"})
			return
		}

		var req struct {
			VlanID    string `json:"vlan_id"`
			VlanName  string `json:"vlan_name"`
			IPSegment string `json:"ip_segment"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.VlanID == "" || req.VlanName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_fields", "message": "vlan_id and vlan_name are required"})
			return
		}

		err := service.AddVlan(req.VlanID, req.VlanName, req.IPSegment)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "creation_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "VLAN created successfully", "vlan_id": req.VlanID})
	}
}

func GetVlansHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		vlans, err := service.GetVlans()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vlans)
	}
}

func UpdateVlanHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "PUT" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed", "message": "Only PUT method is allowed"})
			return
		}

		var req struct {
			VlanInternalID string `json:"vlan_internal_id"`
			NewVlanID      string `json:"new_vlan_id"`
			NewVlanName    string `json:"new_vlan_name"`
			IPSegment      string `json:"ip_segment"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.VlanInternalID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_vlan_internal_id", "message": "vlan_internal_id is required"})
			return
		}

		// Update basic VLAN info if provided
		if req.NewVlanID != "" || req.NewVlanName != "" {
			err := service.UpdateVlan(req.VlanInternalID, req.NewVlanID, req.NewVlanName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "update_failed", "message": err.Error()})
				return
			}
		}

		// Update IP segment if provided
		if req.IPSegment != "" {
			err := service.UpdateVlanIPSegment(req.VlanInternalID, req.IPSegment)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "ip_segment_update_failed", "message": err.Error()})
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "VLAN updated successfully", "vlan_internal_id": req.VlanInternalID})
	}
}

func DeleteVlanHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "DELETE" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed", "message": "Only DELETE method is allowed"})
			return
		}

		var req struct {
			VlanInternalID string `json:"vlan_internal_id"`
			Cascade        bool   `json:"cascade"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.VlanInternalID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_vlan_internal_id", "message": "vlan_internal_id is required"})
			return
		}

		var err error
		if req.Cascade {
			err = service.DeleteVlanCascade(req.VlanInternalID)
		} else {
			err = service.DeleteVlan(req.VlanInternalID)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "VLAN deleted successfully", "vlan_internal_id": req.VlanInternalID})
	}
}
