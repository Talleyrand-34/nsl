// exportstruct.go: Contains helper structs for exporting data.
// SPDX-License-Identifier: AGPL-3.0-or-later
package entities

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

// All is the composition of all the info in the db
type All struct {
	Brands          []BasicBrand
	ConnectionTypes []BasicConnectiontype
	Connections     []BasicConnection
	ModelTypes      []BasicModelType
	DevicePorts     []BasicDeviceport
	Devices         []BasicDevice
	ModelDevices    []BasicModeldevice
	ModelPorts      []BasicModelport
	Policies        []BasicPolicy
	Owners          []BasicOwner
	ZoneTypes       []BasicZonetype
	Zones           []BasicZone
	Vlans           []BasicVlan
}

// BasicBrand maps the info of the brand in the db
type BasicBrand struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// BasicConnectiontype maps the info of the connectiontype in the db
type BasicConnectiontype struct {
	ID             string `json:"id"`
	ConnectionType string `json:"connection_type"`
}

// BasicConnectionmaps the info of the connection in the db
type BasicConnection struct {
	ID                        string   `json:"id"`
	FromDevicePortModelPortID string   `json:"from_device_port_model_port_id"`
	FromDevicePortDeviceID    string   `json:"from_device_port_device_id"`
	FromIPSegment             string   `json:"from_ip_segment"` // "" if null
	ToDevicePortModelPortID   string   `json:"to_device_port_model_port_id"`
	ToDevicePortDeviceID      string   `json:"to_device_port_device_id"`
	ToIPSegment               string   `json:"to_ip_segment"`   // "" if null
	ConnectionType            int64    `json:"connection_type"` // -1 if null
	VlanIDs                   []string `json:"vlan_ids"`        // Array of VLAN IDs
}

type BasicModelType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type BasicDeviceport struct {
	ModelPortID string `json:"model_port_id"`
	DeviceID    string `json:"device_id"`
}

type BasicDevice struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	ModelID string `json:"model_id"`
	ZoneID  string `json:"zone_id"` // -1 if null
	Owner   int64  `json:"owner"`   // -1 if null
}

type BasicModeldevice struct {
	ID          string `json:"id"`
	Model       string `json:"model"`
	Brand       int64  `json:"brand"`
	ModelTypeID string `json:"model_type_id"`
}

type BasicModelport struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Positionx int64  `json:"positionx"`
	Positiony int64  `json:"positiony"`
	ModelID   string `json:"model_id"`
}

// Experimental/Not completed
type BasicPolicy struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	AssociatedConnection int64  `json:"associated_connection"` // -1 if null
	TODO                 string `json:"todo"`                  // "" if null
}

type BasicOwner struct {
	ID    string `json:"id"`
	Owner string `json:"owner"`
}

type BasicZonetype struct {
	ID           string `json:"id"`
	LocationType string `json:"location_type"`
}

type BasicZone struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Father       int64  `json:"father"`        // -1 if null
	Granularity  int64  `json:"granularity"`   // -1 if null
	Owner        int64  `json:"owner"`         // -1 if null
	LocationType int64  `json:"location_type"` // -1 if null
}

type BasicVlan struct {
	ID       string `json:"id"`
	VlanID   string `json:"vlan_id"`
	VlanName string `json:"vlan_name"` // "" if null
}
