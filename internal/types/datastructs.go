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
package datastructs

// DevicePort comment
type DevicePort struct {
	ModelPortID int    // Port ID
	DeviceID    int    // Port ID
	Name        string // Port name
	PositionX   int    // PositionX
	PositionY   int    // PositionY
}

// Connection represents a connection with detailed zone, device, and port information.
type Connection struct {
	ConnectionID int
	FromZone     []string   // Zone hierarchy as a list
	FromDevice   string     // Device label
	FromPort     DevicePort // From port information
	ToZone       []string   // Zone hierarchy as a list
	ToDevice     string     // Device label
	ToPort       DevicePort // To port information
}

// Device represents a device with its properties and ports.
type Device struct {
	ZoneHierarchy []string // Zone hierarchy as a list
	ID            int
	ModelID       int
	Label         string                // Device label
	Shape         string                // Shape based on device class
	Ports         map[string]DevicePort // Ports mapped by their IDs
}

// DeviceModelPortDetails comment
type DeviceModelPortDetails struct {
	DeviceID      int
	DeviceLabel   string
	ModelID       int
	ModelName     string
	ModelPortID   int
	ModelPortName string
}
