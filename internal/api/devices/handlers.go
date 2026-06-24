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
package devices

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	q "nsl-graph/internal/repository/application"
	e "nsl-graph/internal/repository/entities"
)

// RegisterRoutes registers all device-related routes
func RegisterRoutes(r *mux.Router, service q.NetServiceInt) {
	// Brands
	r.HandleFunc("/brands", AddBrandHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/brands", GetBrandsHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/brands", UpdateBrandHandler(service)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/brands", DeleteBrandHandler(service)).Methods("DELETE", "OPTIONS")

	// Device Classes
	r.HandleFunc("/deviceclasses", AddDeviceClassHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/deviceclasses", GetDeviceClassesHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/deviceclasses", UpdateDeviceClassHandler(service)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/deviceclasses", DeleteDeviceClassHandler(service)).Methods("DELETE", "OPTIONS")

	// Zone Types
	r.HandleFunc("/zonetypes", AddZoneTypeHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/zonetypes", GetZoneTypesHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/zonetypes", UpdateZoneTypeHandler(service)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/zonetypes", DeleteZoneTypeHandler(service)).Methods("DELETE", "OPTIONS")

	// Proprietaries
	r.HandleFunc("/proprietaries", AddProprietaryHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/proprietaries", GetProprietariesHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/proprietaries", UpdateProprietaryHandler(service)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/proprietaries", DeleteProprietaryHandler(service)).Methods("DELETE", "OPTIONS")

	// Zones
	r.HandleFunc("/zones", AddZoneHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/zones", GetZonesHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/zones", UpdateZoneHandler(service)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/zones", DeleteZoneHandler(service)).Methods("DELETE", "OPTIONS")

	// Models
	r.HandleFunc("/models", AddModelHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/models", GetModelsHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/models", UpdateModelHandler(service)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/models", DeleteModelHandler(service)).Methods("DELETE", "OPTIONS")

	// Devices
	r.HandleFunc("/devices", AddDeviceHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/devices", GetDevicesHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/devices", UpdateDeviceHandler(service)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/devices", DeleteDeviceHandler(service)).Methods("DELETE", "OPTIONS")

	// Model Ports
	r.HandleFunc("/modelports", AddModelPortHandler(service)).Methods("POST")
	r.HandleFunc("/modelports/bulk", AddBulkModelPortHandler(service)).Methods("POST")
	r.HandleFunc("/modelports", GetModelPortsHandler(service)).Methods("GET")
	r.HandleFunc("/modelports", UpdateModelPortHandler(service)).Methods("PUT")
	r.HandleFunc("/modelports", DeleteModelPortHandler(service)).Methods("DELETE")

	// Device Ports
	r.HandleFunc("/deviceports", AddDevicePortHandler(service)).Methods("POST")
	r.HandleFunc("/deviceports", GetDevicePortsHandler(service)).Methods("GET")
	r.HandleFunc("/deviceports", UpdateDevicePortHandler(service)).Methods("PUT")
	r.HandleFunc("/deviceports", DeleteDevicePortHandler(service)).Methods("DELETE")

	// All Ports
	r.HandleFunc("/allports/device", GetAllPortsDeviceHandler(service)).Methods("GET")
	r.HandleFunc("/allports/all", GetAllPortsAllHandler(service)).Methods("GET")

	// Device Interfaces
	r.HandleFunc("/deviceinterfaces", AddDeviceInterfaceHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/deviceinterfaces", GetDeviceInterfacesHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/deviceinterfaces", DeleteDeviceInterfaceHandler(service)).Methods("DELETE", "OPTIONS")

	// Interface Ports
	r.HandleFunc("/interfaceports", AddInterfacePortHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/interfaceports", GetInterfacePortsHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/interfaceports", DeleteInterfacePortHandler(service)).Methods("DELETE", "OPTIONS")
}

// Placeholder handlers - these will be moved from the original files
func AddBrandHandler(service q.NetServiceInt) http.HandlerFunc {
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

		err := service.AddBrand(req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "add_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Brand created successfully", "name": req.Name})
	}
}

func GetBrandsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		brands, err := service.GetBrands()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(brands)
	}
}

func UpdateBrandHandler(service q.NetServiceInt) http.HandlerFunc {
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
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_id", "message": "id is required"})
			return
		}

		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_name", "message": "name is required"})
			return
		}

		err := service.UpdateBrand(req.ID, req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "update_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Brand updated successfully", "id": req.ID, "name": req.Name})
	}
}

func DeleteBrandHandler(service q.NetServiceInt) http.HandlerFunc {
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

		err := service.DeleteBrand(req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Brand deleted successfully", "name": req.Name})
	}
}

// Add placeholder handlers for other device entities...
// (DeviceClass, ZoneType, Proprietary, Zone, Model, Device, ModelPort, DevicePort handlers)

func AddDeviceClassHandler(service q.NetServiceInt) http.HandlerFunc {
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

		err := service.AddDeviceClass(req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "add_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Device class created successfully", "name": req.Name})
	}
}

func GetDeviceClassesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		deviceClasses, err := service.GetDeviceClasses()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(deviceClasses)
	}
}

func UpdateDeviceClassHandler(service q.NetServiceInt) http.HandlerFunc {
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
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_id", "message": "id is required"})
			return
		}

		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_name", "message": "name is required"})
			return
		}

		err := service.UpdateDeviceClass(req.ID, req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "update_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Device class updated successfully", "id": req.ID, "name": req.Name})
	}
}

func DeleteDeviceClassHandler(service q.NetServiceInt) http.HandlerFunc {
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

		err := service.DeleteDeviceClass(req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Device class deleted successfully", "name": req.Name})
	}
}

func AddZoneTypeHandler(service q.NetServiceInt) http.HandlerFunc {
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

		err := service.AddZoneType(req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "add_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Zone type created successfully", "name": req.Name})
	}
}

func GetZoneTypesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		zoneTypes, err := service.GetZonetypes()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(zoneTypes)
	}
}

func UpdateZoneTypeHandler(service q.NetServiceInt) http.HandlerFunc {
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
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_id", "message": "id is required"})
			return
		}

		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_name", "message": "name is required"})
			return
		}

		err := service.UpdateZoneType(req.ID, req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "update_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Zone type updated successfully", "id": req.ID, "name": req.Name})
	}
}

func DeleteZoneTypeHandler(service q.NetServiceInt) http.HandlerFunc {
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

		err := service.DeleteZoneType(req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Zone type deleted successfully", "name": req.Name})
	}
}

func AddProprietaryHandler(service q.NetServiceInt) http.HandlerFunc {
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

		err := service.AddProprietary(req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "add_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Proprietary created successfully", "name": req.Name})
	}
}

func GetProprietariesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		proprietaries, err := service.GetProperties()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(proprietaries)
	}
}

func UpdateProprietaryHandler(service q.NetServiceInt) http.HandlerFunc {
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
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_id", "message": "id is required"})
			return
		}

		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_name", "message": "name is required"})
			return
		}

		err := service.UpdateProprietary(req.ID, req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "update_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Proprietary updated successfully", "id": req.ID, "name": req.Name})
	}
}

func DeleteProprietaryHandler(service q.NetServiceInt) http.HandlerFunc {
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

		err := service.DeleteProprietary(req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Proprietary deleted successfully", "name": req.Name})
	}
}

func AddZoneHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetZonesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		zones, err := service.GetZones()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(zones)
	}
}

func UpdateZoneHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteZoneHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func AddModelHandler(service q.NetServiceInt) http.HandlerFunc {
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
			ModelName       string `json:"model_name"`
			BrandName       string `json:"brand_name"`
			DeviceClassName string `json:"device_class_name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.ModelName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_model_name", "message": "model_name is required"})
			return
		}

		if req.BrandName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_brand_name", "message": "brand_name is required"})
			return
		}

		if req.DeviceClassName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_device_class_name", "message": "device_class_name is required"})
			return
		}

		err := service.AddModel(req.ModelName, req.BrandName, req.DeviceClassName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "add_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Model created successfully", "model_name": req.ModelName})
	}
}

func GetModelsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		models, err := service.GetModels()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models)
	}
}

func UpdateModelHandler(service q.NetServiceInt) http.HandlerFunc {
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
			ModelID       string `json:"model_id"`
			ModelName     string `json:"model_name"`
			BrandID       string `json:"brand_id"`
			DeviceClassID string `json:"device_class_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.ModelID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_model_id", "message": "model_id is required"})
			return
		}

		err := service.UpdateModel(req.ModelID, req.ModelName, req.BrandID, req.DeviceClassID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "update_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Model updated successfully", "model_id": req.ModelID})
	}
}

func DeleteModelHandler(service q.NetServiceInt) http.HandlerFunc {
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
			ModelID string `json:"model_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.ModelID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_model_id", "message": "model_id is required"})
			return
		}

		err := service.DeleteModel(req.ModelID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Model deleted successfully", "model_id": req.ModelID})
	}
}

func AddDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
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
			Label       string `json:"label"`
			ModelName   string `json:"model_name"`
			ZoneID      string `json:"zone_id"`
			ZoneName    string `json:"zone_name"`
			Proprietary string `json:"proprietary"`
			IsUnmanaged bool   `json:"is_unmanaged"`
			IsInvisible bool   `json:"is_invisible"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.Label == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_label", "message": "label is required"})
			return
		}

		if req.ModelName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_model", "message": "model_name is required"})
			return
		}

		err := service.AddDevice(req.Label, req.ModelName, req.ZoneID, req.ZoneName, req.Proprietary, req.IsUnmanaged, req.IsInvisible)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "creation_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Device created successfully", "label": req.Label})
	}
}

func GetDevicesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		devices, err := service.GetDevices()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(devices)
	}
}

func UpdateDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
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
			ID          string `json:"id"`
			Label       string `json:"label"`
			ModelID     string `json:"model_id"`
			ZoneID      string `json:"zone_id"`
			Proprietary string `json:"proprietary_id"`
			IsUnmanaged *bool  `json:"is_unmanaged"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_id", "message": "id is required"})
			return
		}

		err := service.UpdateDevice(req.ID, req.Label, req.ModelID, req.ZoneID, req.Proprietary, req.IsUnmanaged)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "update_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Device updated successfully", "id": req.ID})
	}
}

func DeleteDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func AddModelPortHandler(service q.NetServiceInt) http.HandlerFunc {
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
			PortName                 string `json:"port_name"`
			PositionX                string `json:"position_x"`
			PositionY                string `json:"position_y"`
			ModelName                string `json:"model_name"`
			AllowMultipleConnections bool   `json:"allow_multiple_connections"`
			PortType                 string `json:"port_type"`
			Band                     string `json:"band"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.PortName == "" || req.ModelName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_fields", "message": "port_name and model_name are required"})
			return
		}

		err := service.AddModelPort(req.PortName, req.PositionX, req.PositionY, req.ModelName, req.AllowMultipleConnections, req.PortType, req.Band)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "creation_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Model port created successfully", "port_name": req.PortName, "model_name": req.ModelName})
	}
}

func AddBulkModelPortHandler(service q.NetServiceInt) http.HandlerFunc {
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
			Ports []struct {
				Name                     string `json:"name"`
				PosX                     string `json:"posx"`
				PosY                     string `json:"posy"`
				ModelName                string `json:"modelName"`
				AllowMultipleConnections bool   `json:"allow_multiple_connections"`
				PortType                 string `json:"port_type"`
				Band                     string `json:"band"`
			} `json:"ports"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if len(req.Ports) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "empty_ports", "message": "No ports provided"})
			return
		}

		var errors []string
		successCount := 0
		failureCount := 0

		for i, port := range req.Ports {
			if port.Name == "" || port.ModelName == "" {
				errors = append(errors, fmt.Sprintf("Port %d: name and modelName are required", i+1))
				failureCount++
				continue
			}

			err := service.AddModelPort(port.Name, port.PosX, port.PosY, port.ModelName, port.AllowMultipleConnections, port.PortType, port.Band)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Port %d (%s): %s", i+1, port.Name, err.Error()))
				failureCount++
			} else {
				successCount++
			}
		}

		response := map[string]interface{}{
			"successCount": successCount,
			"failureCount": failureCount,
		}

		if len(errors) > 0 {
			response["errors"] = errors
		}

		if failureCount > 0 && successCount == 0 {
			w.WriteHeader(http.StatusBadRequest)
		} else if failureCount > 0 {
			w.WriteHeader(207) // 207 Multi-Status for partial success
		} else {
			w.WriteHeader(http.StatusCreated)
		}

		response["message"] = fmt.Sprintf("Bulk operation completed: %d succeeded, %d failed", successCount, failureCount)
		json.NewEncoder(w).Encode(response)
	}
}

func GetModelPortsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		modelPorts, err := service.GetModelPorts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(modelPorts)
	}
}

func UpdateModelPortHandler(service q.NetServiceInt) http.HandlerFunc {
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
			ModelPortID                 string `json:"model_port_id"`
			NewPortName                 string `json:"new_port_name"`
			NewPositionX                string `json:"new_position_x"`
			NewPositionY                string `json:"new_position_y"`
			NewModelID                  string `json:"new_model_id"`
			NewAllowMultipleConnections bool   `json:"new_allow_multiple_connections"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.ModelPortID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_model_port_id", "message": "model_port_id is required"})
			return
		}

		err := service.UpdateModelPort(req.ModelPortID, req.NewPortName, req.NewPositionX, req.NewPositionY, req.NewModelID, req.NewAllowMultipleConnections)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "update_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Model port updated successfully", "model_port_id": req.ModelPortID})
	}
}

func DeleteModelPortHandler(service q.NetServiceInt) http.HandlerFunc {
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
			ModelPortID string `json:"model_port_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.ModelPortID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_model_port_id", "message": "model_port_id is required"})
			return
		}

		err := service.DeleteModelPort(req.ModelPortID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Model port deleted successfully", "model_port_id": req.ModelPortID})
	}
}

func AddDevicePortHandler(service q.NetServiceInt) http.HandlerFunc {
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
			DeviceID    string             `json:"device_id"`
			ModelPortID string             `json:"model_port_id"`
			MacAddress  string             `json:"mac_address"`
			VlanConfigs []e.PortVlanConfig `json:"vlan_configs"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.DeviceID == "" || req.ModelPortID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_fields", "message": "device_id and model_port_id are required"})
			return
		}

		if _, err := service.AddDevicePort(req.DeviceID, req.ModelPortID, req.MacAddress, req.VlanConfigs); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "creation_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Device port created successfully", "device_id": req.DeviceID, "model_port_id": req.ModelPortID})
	}
}

func GetDevicePortsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		devicePorts, err := service.GetDevicePorts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(devicePorts)
	}
}

func UpdateDevicePortHandler(service q.NetServiceInt) http.HandlerFunc {
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
			DeviceID    string             `json:"device_id"`
			ModelPortID string             `json:"model_port_id"`
			MacAddress  string             `json:"mac_address"`
			VlanConfigs []e.PortVlanConfig `json:"vlan_configs"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.DeviceID == "" || req.ModelPortID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_fields", "message": "device_id and model_port_id are required"})
			return
		}

		err := service.UpdateDevicePort(req.DeviceID, req.ModelPortID, req.MacAddress, req.VlanConfigs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "update_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Device port updated successfully", "device_id": req.DeviceID, "model_port_id": req.ModelPortID})
	}
}

func DeleteDevicePortHandler(service q.NetServiceInt) http.HandlerFunc {
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
			DeviceID    string `json:"device_id"`
			ModelPortID string `json:"model_port_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.DeviceID == "" || req.ModelPortID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_fields", "message": "device_id and model_port_id are required"})
			return
		}

		err := service.DeleteDevicePort(req.DeviceID, req.ModelPortID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Device port deleted successfully", "device_id": req.DeviceID, "model_port_id": req.ModelPortID})
	}
}

func GetAllPortsDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		deviceID := r.URL.Query().Get("deviceid")
		if deviceID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_deviceid", "message": "deviceid query parameter is required"})
			return
		}

		devicePorts, err := service.GetAllPortsDevice(deviceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(devicePorts)
	}
}

func GetAllPortsAllHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		devicePorts, err := service.GetAllPortsAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(devicePorts)
	}
}

// --- Device Interface handlers ---

func AddDeviceInterfaceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var req struct {
			DeviceID     string             `json:"device_id"`
			Name         string             `json:"name"`
			Description  string             `json:"description"`
			VlanConfigs  []e.PortVlanConfig `json:"vlan_configs"`
			IPAddresses  []string           `json:"ip_addresses"`
			WifiSSID     string             `json:"wifi_ssid"`
			WifiSecurity string             `json:"wifi_security"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.DeviceID == "" || req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_fields", "message": "device_id and name are required"})
			return
		}

		err := service.AddDeviceInterface(req.DeviceID, req.Name, req.Description, "", req.VlanConfigs, req.IPAddresses, req.WifiSSID, req.WifiSecurity)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "creation_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Device interface created successfully"})
	}
}

func GetDeviceInterfacesHandler(service q.NetServiceInt) http.HandlerFunc {
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

		var result interface{}
		var err error
		if deviceID != "" {
			result, err = service.GetDeviceInterfaces(deviceID)
		} else {
			result, err = service.GetAllDeviceInterfaces()
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func DeleteDeviceInterfaceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var req struct {
			ID string `json:"id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_fields", "message": "id is required"})
			return
		}

		err := service.DeleteDeviceInterface(req.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Device interface deleted successfully", "id": req.ID})
	}
}

// --- Interface Port handlers ---

func AddInterfacePortHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var req struct {
			InterfaceID string `json:"interface_id"`
			DeviceID    string `json:"device_id"`
			ModelPortID string `json:"model_port_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.InterfaceID == "" || req.DeviceID == "" || req.ModelPortID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_fields", "message": "interface_id, device_id, and model_port_id are required"})
			return
		}

		err := service.AddInterfacePort(req.InterfaceID, req.DeviceID, req.ModelPortID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "creation_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Interface-port link created successfully"})
	}
}

func GetInterfacePortsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		interfaceID := r.URL.Query().Get("interface_id")
		deviceID := r.URL.Query().Get("device_id")
		modelPortID := r.URL.Query().Get("model_port_id")

		var result interface{}
		var err error
		switch {
		case interfaceID != "":
			result, err = service.GetPortsForInterface(interfaceID)
		case deviceID != "" && modelPortID != "":
			result, err = service.GetInterfacesForPort(deviceID, modelPortID)
		default:
			result, err = service.GetAllInterfacePorts()
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func DeleteInterfacePortHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var req struct {
			InterfaceID string `json:"interface_id"`
			DeviceID    string `json:"device_id"`
			ModelPortID string `json:"model_port_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
			return
		}

		if req.InterfaceID == "" || req.DeviceID == "" || req.ModelPortID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing_fields", "message": "interface_id, device_id, and model_port_id are required"})
			return
		}

		err := service.DeleteInterfacePort(req.InterfaceID, req.DeviceID, req.ModelPortID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "delete_failed", "message": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Interface-port link deleted successfully"})
	}
}
