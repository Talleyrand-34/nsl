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
package scanning

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	fmtd2 "nsl-graph/internal/format"
	e "nsl-graph/internal/repository/entities"
	"nsl-graph/internal/topology"
)

// discoveredZone is the single synthetic zone all discovered nodes are placed in
// (the discovered topology isn't tied to the DB's zone hierarchy).
const discoveredZone = "Discovered"

// splitEndpoint splits a "device:port" scan label into device and port. An
// "unknown(...)" label has no device.
func splitEndpoint(label string) (device, port string) {
	if label == "" || strings.HasPrefix(label, "unknown(") {
		return "", ""
	}
	if i := strings.LastIndexByte(label, ':'); i >= 0 {
		return label[:i], label[i+1:]
	}
	return label, ""
}

// scanResultToDeviceConnections converts a discovered topology (scan edges +
// detected intermediaries) into the synthetic devices and connections the
// existing D2 generator consumes — so the diagram code is reused unchanged.
func scanResultToDeviceConnections(res topology.ConnectionScanResult) ([]e.Device, []e.Connection) {
	devSet := map[string]bool{}
	addDev := func(label string) {
		if label != "" {
			devSet[label] = true
		}
	}

	var conns []e.Connection
	addConn := func(fromDev, fromPort, toDev, toPort string) {
		if fromDev == "" || toDev == "" {
			return
		}
		addDev(fromDev)
		addDev(toDev)
		conns = append(conns, e.Connection{
			FromDevice: fromDev, FromModelPort: fromPort, FromZoneID: discoveredZone,
			ToDevice: toDev, ToModelPort: toPort, ToZoneID: discoveredZone,
		})
	}

	for _, ed := range res.Edges {
		fd, fp := splitEndpoint(ed.FromLabel)
		td, tp := splitEndpoint(ed.ToLabel)
		addConn(fd, fp, td, tp)
	}

	// Each detected intermediary becomes a synthetic node every observing host
	// links to, mirroring the ASCII tree's "via <vendor> switch" rendering.
	for _, im := range res.Intermediaries {
		vendor := strings.TrimSpace(im.Vendor)
		if vendor == "" {
			vendor = "unknown"
		}
		name := fmt.Sprintf("%s switch (%s)", vendor, im.MAC)
		p := 0
		for _, sb := range im.SeenBy {
			d, port := splitEndpoint(sb)
			if d == "" {
				continue
			}
			p++
			addConn(d, port, name, fmt.Sprintf("p%d", p))
		}
	}

	devices := make([]e.Device, 0, len(devSet))
	for label := range devSet {
		devices = append(devices, e.Device{Label: label, ZoneID: discoveredZone, ZoneName: discoveredZone})
	}
	return devices, conns
}

// ConnectionsDiagramHandler renders the posted discovered topology (a
// ConnectionScanResult) as a D2 SVG, reusing the standard diagram generator.
func ConnectionsDiagramHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var res topology.ConnectionScanResult
		if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid scan result: " + err.Error()))
			return
		}

		devices, conns := scanResultToDeviceConnections(res)
		if len(conns) == 0 {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Diagram empty: no resolvable links in the discovered topology"))
			return
		}

		dJSON, _ := json.Marshal(devices)
		cJSON, _ := json.Marshal(conns)
		d2 := fmtd2.GenerateD2FromJSON(dJSON, cJSON)
		svg, err := fmtd2.GenerateDiagramSVG(d2)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to generate SVG: " + err.Error()))
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.WriteHeader(http.StatusOK)
		w.Write(svg)
	}
}
