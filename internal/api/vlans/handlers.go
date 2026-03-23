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

	// Local VLANs
	r.HandleFunc("/localvlans", AddLocalVlanHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/localvlans", GetLocalVlansHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/localvlans/device/{deviceId}", GetLocalVlansByDeviceHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/localvlans/vlan/{vlanId}", GetLocalVlansByVlanIdHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/localvlans", UpdateLocalVlanHandler(service)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/localvlans", DeleteLocalVlanHandler(service)).Methods("DELETE", "OPTIONS")
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
			VlanID      string `json:"vlan_id"`
			VlanName    string `json:"vlan_name"`
			IPSegment   string `json:"ip_segment"`
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

		err := service.DeleteVlan(req.VlanInternalID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "VLAN deleted successfully", "vlan_internal_id": req.VlanInternalID})
	}
}

// Local VLAN Handlers

func AddLocalVlanHandler(service q.NetServiceInt) http.HandlerFunc {
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
			VlanID   string `json:"vlan_id"`
			DeviceID string `json:"device_id"`
			VlanName string `json:"vlan_name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.VlanID == "" || req.DeviceID == "" || req.VlanName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_fields", "message": "vlan_id, device_id, and vlan_name are required"})
			return
		}

		err := service.AddLocalVlan(req.VlanID, req.DeviceID, req.VlanName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "creation_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Local VLAN created successfully", "vlan_id": req.VlanID, "device_id": req.DeviceID})
	}
}

func GetLocalVlansHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed", "message": "Only GET method is allowed"})
			return
		}

		localVlans, err := service.GetLocalVlans()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "retrieval_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(localVlans)
	}
}

func GetLocalVlansByDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed", "message": "Only GET method is allowed"})
			return
		}

		vars := mux.Vars(r)
		deviceID := vars["deviceId"]

		if deviceID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_device_id", "message": "Device ID is required"})
			return
		}

		localVlans, err := service.GetLocalVlansByDevice(deviceID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "retrieval_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(localVlans)
	}
}

func GetLocalVlansByVlanIdHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed", "message": "Only GET method is allowed"})
			return
		}

		vars := mux.Vars(r)
		vlanID := vars["vlanId"]

		if vlanID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_vlan_id", "message": "VLAN ID is required"})
			return
		}

		localVlans, err := service.GetLocalVlansByVlanID(vlanID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "retrieval_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(localVlans)
	}
}

func UpdateLocalVlanHandler(service q.NetServiceInt) http.HandlerFunc {
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
			LocalVlanID  string `json:"local_vlan_id"`
			NewVlanID    string `json:"new_vlan_id"`
			NewDeviceID  string `json:"new_device_id"`
			NewVlanName  string `json:"new_vlan_name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.LocalVlanID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_id", "message": "Local VLAN ID is required"})
			return
		}

		err := service.UpdateLocalVlan(req.LocalVlanID, req.NewVlanID, req.NewDeviceID, req.NewVlanName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "update_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Local VLAN updated successfully", "local_vlan_id": req.LocalVlanID})
	}
}

func DeleteLocalVlanHandler(service q.NetServiceInt) http.HandlerFunc {
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
			LocalVlanID string `json:"local_vlan_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.LocalVlanID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_id", "message": "Local VLAN ID is required"})
			return
		}

		err := service.DeleteLocalVlan(req.LocalVlanID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Local VLAN deleted successfully", "local_vlan_id": req.LocalVlanID})
	}
}