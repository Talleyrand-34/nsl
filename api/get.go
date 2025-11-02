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
package api

import (
	"encoding/json"
	"net/http"

	"nsl-graph/internal/format"
	q "nsl-graph/internal/repository/application"
)

// func getDiagram(service q.NetServiceInt) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		connections, err := service.GetConnections()
// 		if err != nil {
// 			http.Error(w, "Failed to get connections: "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}
//
// 		devices, err := service.GetDevices()
// 		if err != nil {
// 			http.Error(w, "Failed to get devices: "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		zones, err := service.GetZones()
// 		if err != nil {
// 			http.Error(w, "Failed to get devices: "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}
//
// 		diagramString := format.GenerateD2FromStruct2(devices, connections, zones)
//
// 		var diagram []byte
// 		diagram, err = format.GenerateDiagramSVG(diagramString)
// 		if err != nil {
// 			http.Error(
// 				w,
// 				"Failed to generate diagram: "+err.Error(),
// 				http.StatusInternalServerError,
// 			)
// 			return
// 		}
//
// 		w.Header().Set("Content-Type", "image/svg+xml")
// 		w.WriteHeader(http.StatusOK)
// 		_, _ = w.Write(diagram)
// 	}
// }

type DiagramFormat string

const (
	FormatPorts       DiagramFormat = "ports"
	FormatConnections DiagramFormat = "connections"
)

// Updated function to read format from query parameter
func getDiagram(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get format from query parameter, default to connections
		formatParam := r.URL.Query().Get("format")
		vlanParam := r.URL.Query().Get("vlan") // "true" or "false"

		var diagramFormat DiagramFormat
		includeVlans := vlanParam == "true"

		switch formatParam {
		case "ports":
			diagramFormat = FormatPorts
		case "connections", "": // Default to connections if not specified
			diagramFormat = FormatConnections
		default:
			http.Error(
				w,
				"Invalid diagram format. Use 'ports' or 'connections'",
				http.StatusBadRequest,
			)
			return
		}

		connections, err := service.GetConnections()
		if err != nil {
			http.Error(w, "Failed to get connections: "+err.Error(), http.StatusInternalServerError)
			return
		}
		devices, err := service.GetDevices()
		if err != nil {
			http.Error(w, "Failed to get devices: "+err.Error(), http.StatusInternalServerError)
			return
		}
		zones, err := service.GetZones()
		if err != nil {
			http.Error(w, "Failed to get zones: "+err.Error(), http.StatusInternalServerError)
			return
		}
		devicePorts, err := service.GetDevicePorts()
		if err != nil {
			http.Error(w, "Failed to get device ports: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var diagramString string
		switch diagramFormat {
		case FormatPorts:
			if includeVlans {
				diagramString = format.GenerateD2FocusPortsWithVlans(devices, connections, zones, devicePorts)
			} else {
				diagramString = format.GenerateD2FocusPorts(devices, connections, zones)
			}
		case FormatConnections:
			if includeVlans {
				diagramString = format.GenerateD2FocusConnectionsWithVlans(devices, connections, zones, devicePorts)
			} else {
				diagramString = format.GenerateD2FocusConnections(devices, connections, zones)
			}
		}

		diagram, err := format.GenerateDiagramSVG(diagramString)
		if err != nil {
			http.Error(
				w,
				"Failed to generate diagram: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		// Add cache headers to prevent unwanted caching during format switching
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.WriteHeader(http.StatusOK)
		w.Write(diagram)
	}
}

// --- Brand ---

func getBrandsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		brands, err := service.GetBrands()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(brands)
	}
}

// --- DeviceClass ---

func getDeviceClassesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		classes, err := service.GetDeviceClasses()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(classes)
	}
}

// --- ZoneType ---

func getZoneTypesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		types, err := service.GetZonetypes()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types)
	}
}

// --- Proprietary ---

func getProprietariesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		props, err := service.GetProperties()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(props)
	}
}

// --- Zone ---

func getZonesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zones, err := service.GetZones()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(zones)
	}
}

// --- Model ---

func getModelsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		models, err := service.GetModels()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models)
	}
}

// --- Device ---

func getDevicesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		devices, err := service.GetDevices()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(devices)
	}
}

// --- ModelPort ---

func getModelPortsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ports, err := service.GetModelPorts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ports)
	}
}

// --- DevicePort ---

func getDevicePortsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ports, err := service.GetDevicePorts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ports)
	}
}

// --- ConnectionType ---

func getConnectionsTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conns, err := service.GetConnectionTypes()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(conns)
	}
}

// --- Connection ---

func getConnectionsHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conns, err := service.GetConnections()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(conns)
	}
}

// --- VLAN ---

func getVlansHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vlans, err := service.GetVlans()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(vlans)
	}
}

// --- All Ports for Device ---
func getAllPortsDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.URL.Query().Get("deviceid")
		if deviceID == "" {
			http.Error(w, "deviceid is required", http.StatusBadRequest)
			return
		}
		ports, err := service.GetAllPortsDevice(deviceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ports)
	}
}

// --- All Ports (All Devices) ---
func getAllPortsAllHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ports, err := service.GetAllPortsAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ports)
	}
}
