// Package entities contains the structs needed for the processing of sql queries
//
// datastruct.go: Contains core data structures for SQL query processing.
package entities

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

// Zone This struct containd the info about a zone
type Zone struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	FatherID     string `json:"fatherid"`
	Father       string `json:"father"`
	LocationType string `json:"location_type"`
	Proprietary  string `json:"proprietary"`
}

// ModelDevice This struct contains the info about a model
type ModelDevice struct {
	ID    string `json:"id"`
	Model string `json:"model"`
	Brand string `json:"brand"`
	Class string `json:"class"`
}

// ModelPort This struct contains the info about a port from the model perspective
type ModelPort struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Positionx int    `json:"positionx"`
	Positiony int    `json:"positiony"`
	Model     string `json:"model"`
	Brand     string `json:"brand"`
}

// Device This struct contains the info about a device
type Device struct {
	ID          string `json:"id"`
	Name        string `json:"label"`
	Model       string `json:"model"`
	Brand       string `json:"brand"`
	ZoneID      string `json:"zoneid"`
	ZoneName    string `json:"zonename"`
	ZoneFather  string `json:"zonefathername"`
	Proprietary string `json:"proprietary"`
}

// DevicePort This struct contains the info about ports asociated to a device since the information is contained in the model
type DevicePort struct {
	DeviceID   string `json:"devid"`
	ModelID    string `json:"modelid"`
	MacAddress string `json:"mac_address"`
	PortName   string `json:"portname"`
	DevLabel   string `json:"devname"`
	Positionx  int    `json:"positionx"`
	Positiony  int    `json:"positiony"`
}

// Connection This struct contains the info about a connection
type Connection struct {
	ID            string   `json:"id"`
	FromDevice    string   `json:"fromdevice"`    // Name of the FromDevice
	FromModelPort string   `json:"frommodel"`     // Name of the port on the model
	FromIPSegment string   `json:"fromipsegment"` // IP segment of the from port
	FromZoneName  string   `json:"fromzonename"`
	FromZoneID    string   `json:"fromzoneid"`
	ToDevice      string   `json:"todevice"`    // Name of the ToDevice
	ToModelPort   string   `json:"tomodel"`     // Name of the port on the model
	ToIPSegment   string   `json:"toipsegment"` // IP segment of the to port
	ToZoneName    string   `json:"tozonename"`
	ToZoneID      string   `json:"tozoneid"`
	VlanCon       []string `json:"vlanids"` // List of VLAN IDs
}

type Vlan struct {
	ID       string `json:"id"`
	VlanID   string `json:"vlanid"`
	VlanName string `json:"vlanname"`
}

// Brand represents a brand
type Brand struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DevClass represents the class of a device
type DevClass struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Proprietary represents a proprietary of a device o zone
type Proprietary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ZoneType represents the type of zone a zone can be
type ZoneType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ConnectionType represents the type of connection a connection can be
type ConnectionType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
