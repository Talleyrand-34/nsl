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
package push

import (
	"net/http"

	q "nsl-graph/internal/repository/application"
	e "nsl-graph/internal/repository/entities"
)

// NeighborEdge is one row of the topology scope: a connection from the
// center device to a neighbor, with the ports + VLANs that bind them.
type NeighborEdge struct {
	Neighbor     string   `json:"neighbor"`
	NeighborOS   string   `json:"neighbor_os,omitempty"`
	ThisPort     string   `json:"this_port"`
	NeighborPort string   `json:"neighbor_port"`
	Vlans        []string `json:"vlans,omitempty"`
	Confidence   string   `json:"confidence,omitempty"`
}

// TopologyResponse is the body of GET /push/topology?device_id=…
type TopologyResponse struct {
	DeviceID  string          `json:"device_id"`
	DeviceOS  string          `json:"device_os,omitempty"`
	Neighbor  []NeighborEdge  `json:"neighbors"`
}

// TopologyHandler returns the 1-hop neighborhood of a device: every
// connection where the device is either endpoint. The OS for each
// neighbor is best-effort and may be empty (the lookup goes through
// Device.Model → ModelDevice.OsType, which the service exposes via
// GetDevices + GetModels).
func TopologyHandler(service q.NetServiceInt) http.HandlerFunc {
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
		if deviceID == "" {
			writeError(w, http.StatusBadRequest, "missing_device_id", "device_id query param is required")
			return
		}
		devices, err := service.GetDevices()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "devices_failed", err.Error())
			return
		}
		models, err := service.GetModels()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "models_failed", err.Error())
			return
		}
		connections, err := service.GetConnections()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "connections_failed", err.Error())
			return
		}
		deviceOS := lookupDeviceOS(deviceID, devices, models)
		out := TopologyResponse{DeviceID: deviceID, DeviceOS: deviceOS, Neighbor: []NeighborEdge{}}
		seen := map[string]bool{}
		for _, c := range connections {
			from, to := c.FromDevice, c.ToDevice
			if from == "" || to == "" {
				continue
			}
			if from != deviceID && to != deviceID {
				continue
			}
			var neighbor, thisPort, nbrPort string
			if from == deviceID {
				neighbor, thisPort, nbrPort = to, c.FromModelPort, c.ToModelPort
			} else {
				neighbor, thisPort, nbrPort = from, c.ToModelPort, c.FromModelPort
			}
			if seen[neighbor+"|"+thisPort] {
				continue
			}
			seen[neighbor+"|"+thisPort] = true
			nbrOS := lookupDeviceOS(neighbor, devices, models)
			vlans := make([]string, 0, len(c.Vlans))
			for _, v := range c.Vlans {
				vlans = append(vlans, v.VLANID)
			}
			out.Neighbor = append(out.Neighbor, NeighborEdge{
				Neighbor:     neighbor,
				NeighborOS:   nbrOS,
				ThisPort:     thisPort,
				NeighborPort: nbrPort,
				Vlans:        vlans,
				Confidence:   c.Confidence,
			})
		}
		writeJSON(w, http.StatusOK, out)
	}
}

// lookupDeviceOS resolves a device's OS via its model. Returns "" when
func lookupDeviceOS(deviceID string, devices []e.Device, models []e.ModelDevice) string {
	var dev e.Device
	found := false
	for _, d := range devices {
		if d.ID == deviceID || d.Label == deviceID {
			dev, found = d, true
			break
		}
	}
	if !found || dev.Model == "" {
		return ""
	}
	for _, m := range models {
		if m.Model == dev.Model {
			return m.OsType
		}
	}
	return ""
}
