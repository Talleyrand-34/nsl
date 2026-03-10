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
package devices

import (
	"net/http"

	"github.com/gorilla/mux"
	q "nsl-graph/internal/repository/application"
)

// RegisterRoutes registers all device-related routes
func RegisterRoutes(r *mux.Router, service q.NetServiceInt) {
	// Brands
	r.HandleFunc("/brands", AddBrandHandler(service)).Methods("POST")
	r.HandleFunc("/brands", GetBrandsHandler(service)).Methods("GET")
	r.HandleFunc("/brands", UpdateBrandHandler(service)).Methods("PUT")
	r.HandleFunc("/brands", DeleteBrandHandler(service)).Methods("DELETE")

	// Device Classes
	r.HandleFunc("/deviceclasses", AddDeviceClassHandler(service)).Methods("POST")
	r.HandleFunc("/deviceclasses", GetDeviceClassesHandler(service)).Methods("GET")
	r.HandleFunc("/deviceclasses", UpdateDeviceClassHandler(service)).Methods("PUT")
	r.HandleFunc("/deviceclasses", DeleteDeviceClassHandler(service)).Methods("DELETE")

	// Zone Types
	r.HandleFunc("/zonetypes", AddZoneTypeHandler(service)).Methods("POST")
	r.HandleFunc("/zonetypes", GetZoneTypesHandler(service)).Methods("GET")
	r.HandleFunc("/zonetypes", UpdateZoneTypeHandler(service)).Methods("PUT")
	r.HandleFunc("/zonetypes", DeleteZoneTypeHandler(service)).Methods("DELETE")

	// Proprietaries
	r.HandleFunc("/proprietaries", AddProprietaryHandler(service)).Methods("POST")
	r.HandleFunc("/proprietaries", GetProprietariesHandler(service)).Methods("GET")
	r.HandleFunc("/proprietaries", UpdateProprietaryHandler(service)).Methods("PUT")
	r.HandleFunc("/proprietaries", DeleteProprietaryHandler(service)).Methods("DELETE")

	// Zones
	r.HandleFunc("/zones", AddZoneHandler(service)).Methods("POST")
	r.HandleFunc("/zones", GetZonesHandler(service)).Methods("GET")
	r.HandleFunc("/zones", UpdateZoneHandler(service)).Methods("PUT")
	r.HandleFunc("/zones", DeleteZoneHandler(service)).Methods("DELETE")

	// Models
	r.HandleFunc("/models", AddModelHandler(service)).Methods("POST")
	r.HandleFunc("/models", GetModelsHandler(service)).Methods("GET")
	r.HandleFunc("/models", UpdateModelHandler(service)).Methods("PUT")
	r.HandleFunc("/models", DeleteModelHandler(service)).Methods("DELETE")

	// Devices
	r.HandleFunc("/devices", AddDeviceHandler(service)).Methods("POST")
	r.HandleFunc("/devices", GetDevicesHandler(service)).Methods("GET")
	r.HandleFunc("/devices", UpdateDeviceHandler(service)).Methods("PUT")
	r.HandleFunc("/devices", DeleteDeviceHandler(service)).Methods("DELETE")

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
}

// Placeholder handlers - these will be moved from the original files
func AddBrandHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Implementation to be moved from api/add.go
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetBrandsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Implementation to be moved from api/get.go
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func UpdateBrandHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Implementation to be moved from api/update.go
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteBrandHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Implementation to be moved from api/delete.go
		w.WriteHeader(http.StatusNotImplemented)
	}
}

// Add placeholder handlers for other device entities...
// (DeviceClass, ZoneType, Proprietary, Zone, Model, Device, ModelPort, DevicePort handlers)

func AddDeviceClassHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetDeviceClassesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func UpdateDeviceClassHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteDeviceClassHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func AddZoneTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetZoneTypesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func UpdateZoneTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteZoneTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func AddProprietaryHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetProprietariesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func UpdateProprietaryHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteProprietaryHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func AddZoneHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetZonesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
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
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetModelsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func UpdateModelHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteModelHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func AddDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetDevicesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func UpdateDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func AddModelPortHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func AddBulkModelPortHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetModelPortsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func UpdateModelPortHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteModelPortHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func AddDevicePortHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetDevicePortsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func UpdateDevicePortHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteDevicePortHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetAllPortsDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetAllPortsAllHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}