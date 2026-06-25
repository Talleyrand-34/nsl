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
package core

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	fmtd2 "nsl-graph/internal/format"
	q "nsl-graph/internal/repository/application"
)

// RegisterRoutes registers all core routes
func RegisterRoutes(r *mux.Router, service q.NetServiceInt) {
	r.HandleFunc("/", RootHandler()).Methods("GET")
	r.HandleFunc("/diagram", GetDiagramHandler(service)).Methods("GET")
	r.HandleFunc("/vault/status", VaultStatusHandler(service)).Methods("GET", "OPTIONS")
	r.HandleFunc("/vault/init", VaultInitHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/vault/unlock", VaultUnlockHandler(service)).Methods("POST", "OPTIONS")
	r.HandleFunc("/vault/lock", VaultLockHandler(service)).Methods("POST", "OPTIONS")
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

			// Network scanning
			"POST   /scan/network",
			"POST   /scan/host",
			"POST   /scan/import",
			"GET    /scan/status?scan_id=<id>",
			"GET    /scan/validate?subnet=<subnet>",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(endpoints)
	}
}

// GetDiagramHandler generates and returns network diagrams
func GetDiagramHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Get query parameters
		format := r.URL.Query().Get("format")
		if format == "" {
			format = "connections" // default
		}

		vlan := r.URL.Query().Get("vlan") == "true"
		colorports := r.URL.Query().Get("colorports") == "true"
		allports := r.URL.Query().Get("allports") == "true"

		vlanScope := r.URL.Query().Get("vlan_scope")
		if vlanScope == "" {
			vlanScope = "untagged"
		}
		colorTarget := r.URL.Query().Get("color_target")
		if colorTarget == "" {
			colorTarget = "both"
		}

		// Validate format parameter
		if format != "ports" && format != "connections" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid format parameter. Use 'ports' or 'connections'"))
			return
		}

		// Get data from service
		devices, err := service.GetDevices()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get devices: " + err.Error()))
			return
		}

		// Nothing to render — signal the client (the <img> falls back to its alt
		// text) instead of returning a blank SVG.
		if len(devices) == 0 {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Diagram empty: No devices"))
			return
		}

		connections, err := service.GetConnections()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get connections: " + err.Error()))
			return
		}

		zones, err := service.GetZones()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get zones: " + err.Error()))
			return
		}

		devicePorts, err := service.GetDevicePorts()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get device ports: " + err.Error()))
			return
		}

		allInterfaces, err := service.GetAllDeviceInterfaces()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get device interfaces: " + err.Error()))
			return
		}

		ifacePorts, err := service.GetAllInterfacePorts()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get interface ports: " + err.Error()))
			return
		}

		// Generate D2 diagram based on parameters
		var d2Script string

		if vlan {
			if format == "ports" {
				d2Script = fmtd2.GenerateD2FocusPortsWithVlans(devices, connections, zones, devicePorts, allInterfaces, ifacePorts, allports, vlanScope, colorTarget)
			} else {
				d2Script = fmtd2.GenerateD2FocusConnectionsWithVlans(devices, connections, zones, devicePorts, allInterfaces, ifacePorts, allports, vlanScope, colorTarget)
			}
		} else {
			if format == "ports" {
				d2Script = fmtd2.GenerateD2FocusPorts(devices, connections, zones, devicePorts, allInterfaces, ifacePorts, colorports, allports)
			} else {
				d2Script = fmtd2.GenerateD2FocusConnections(devices, connections, zones, devicePorts, allInterfaces, ifacePorts, colorports, allports)
			}
		}

		// Generate SVG from D2 script
		svgBytes, err := fmtd2.GenerateDiagramSVG(d2Script)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to generate SVG: " + err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(svgBytes)
	}
}
