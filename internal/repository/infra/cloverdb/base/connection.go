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
package basicops

import (
	"fmt"

	d "github.com/ostafen/clover/v2/document"
	q "github.com/ostafen/clover/v2/query"

	e "nsl-graph/internal/repository/entities"
)

// GetConnections returns all connections
func (r BasicOpsCloverRepository) GetConnections() ([]e.Connection, error) {
	docs, err := r.db.FindAll(q.NewQuery(connectionsCollection))
	if err != nil {
		return []e.Connection{}, err
	}

	result := make([]e.Connection, 0, len(docs))
	for _, doc := range docs {
		connection := e.Connection{
			ID: doc.ObjectId(),
		}

		// Get from device info
		if fromDeviceID, ok := doc.Get("from_device_id").(string); ok && fromDeviceID != "" {
			deviceDoc, err := r.db.FindById(devicesCollection, fromDeviceID)
			if err == nil && deviceDoc != nil {
				connection.FromDevice = deviceDoc.Get("label").(string)

				// Get from zone info
				if zoneID, ok := deviceDoc.Get("zone_id").(string); ok && zoneID != "" {
					zoneDoc, err := r.db.FindById(zonesCollection, zoneID)
					if err == nil && zoneDoc != nil {
						connection.FromZoneName = zoneDoc.Get("name").(string)
						connection.FromZoneID = zoneID
					}
				}
			}
		}

		// Get from model port info
		if fromModelPortID, ok := doc.Get("from_model_port_id").(string); ok && fromModelPortID != "" {
			modelPortDoc, err := r.db.FindById(modelportsCollection, fromModelPortID)
			if err == nil && modelPortDoc != nil {
				connection.FromModelPort = modelPortDoc.Get("name").(string)
			}
		}

		// Get to device info
		if toDeviceID, ok := doc.Get("to_device_id").(string); ok && toDeviceID != "" {
			deviceDoc, err := r.db.FindById(devicesCollection, toDeviceID)
			if err == nil && deviceDoc != nil {
				connection.ToDevice = deviceDoc.Get("label").(string)

				// Get to zone info
				if zoneID, ok := deviceDoc.Get("zone_id").(string); ok && zoneID != "" {
					zoneDoc, err := r.db.FindById(zonesCollection, zoneID)
					if err == nil && zoneDoc != nil {
						connection.ToZoneName = zoneDoc.Get("name").(string)
						connection.ToZoneID = zoneID
					}
				}
			}
		}

		// Get to model port info
		if toModelPortID, ok := doc.Get("to_model_port_id").(string); ok && toModelPortID != "" {
			modelPortDoc, err := r.db.FindById(modelportsCollection, toModelPortID)
			if err == nil && modelPortDoc != nil {
				connection.ToModelPort = modelPortDoc.Get("name").(string)
			}
		}

		result = append(result, connection)
	}

	return result, nil
}

// AddConnection creates a new connection between two device ports
// Validates that both ports have matching VLANs (strict mode by default)
// Set allowVLANUnion to true to allow connection if VLANs have any overlap instead of requiring exact match
func (r BasicOpsCloverRepository) AddConnection(
	fromDevice string,
	fromModelPort string,
	toDevice string,
	toModelPort string,
	allowVLANUnion bool,
) error {
	// Validate that the ports are not already in use
	// Check if from port is already used
	fromQuery := q.NewQuery(connectionsCollection).Where(
		q.Field("from_device_id").Eq(fromDevice),
	).Where(
		q.Field("from_model_port_id").Eq(fromModelPort),
	)
	fromExists, err := r.db.Exists(fromQuery)
	if err != nil {
		return fmt.Errorf("failed to check port usage: %v", err)
	}

	// Check if to port is already used
	toQuery := q.NewQuery(connectionsCollection).Where(
		q.Field("to_device_id").Eq(toDevice),
	).Where(
		q.Field("to_model_port_id").Eq(toModelPort),
	)
	toExists, err := r.db.Exists(toQuery)
	if err != nil {
		return fmt.Errorf("failed to check port usage: %v", err)
	}

	if fromExists || toExists {
		return fmt.Errorf("one or both ports are already in use")
	}

	// Get VLAN configs from both device ports
	fromPort, err := r.GetDevicePortByIDs(fromDevice, fromModelPort)
	if err != nil {
		return fmt.Errorf("failed to get from device port: %v", err)
	}

	toPort, err := r.GetDevicePortByIDs(toDevice, toModelPort)
	if err != nil {
		return fmt.Errorf("failed to get to device port: %v", err)
	}

	// Extract VLAN numbers from configs for validation
	fromVlans := extractVLANNumbers(fromPort.VlanConfigs)
	toVlans := extractVLANNumbers(toPort.VlanConfigs)

	// Validate VLAN compatibility
	if len(fromVlans) > 0 || len(toVlans) > 0 {
		if allowVLANUnion {
			// Union mode: Check if there's any VLAN overlap
			if !hasVLANOverlap(fromVlans, toVlans) {
				return fmt.Errorf("VLAN validation failed: no common VLANs between ports (from: %v, to: %v)",
					fromVlans, toVlans)
			}
		} else {
			// Strict mode: VLANs must match exactly
			if !vlanListsEqual(fromVlans, toVlans) {
				return fmt.Errorf("VLAN validation failed: VLANs must match exactly (from: %v, to: %v). Use allowVLANUnion flag to allow overlapping VLANs",
					fromVlans, toVlans)
			}
		}
	}

	doc := d.NewDocument()
	doc.Set("from_device_id", fromDevice)
	doc.Set("from_model_port_id", fromModelPort)
	doc.Set("to_device_id", toDevice)
	doc.Set("to_model_port_id", toModelPort)

	_, err = r.db.InsertOne(connectionsCollection, doc)
	if err != nil {
		return fmt.Errorf("failed to create connection: %v", err)
	}
	return nil
}

// hasVLANOverlap checks if two VLAN lists have any common elements
func hasVLANOverlap(vlans1, vlans2 []string) bool {
	// If both lists are empty, consider it valid (no VLAN restriction)
	if len(vlans1) == 0 && len(vlans2) == 0 {
		return true
	}
	// If one list is empty and the other isn't, no overlap
	if len(vlans1) == 0 || len(vlans2) == 0 {
		return false
	}

	vlanSet := make(map[string]bool)
	for _, vlan := range vlans1 {
		vlanSet[vlan] = true
	}
	for _, vlan := range vlans2 {
		if vlanSet[vlan] {
			return true
		}
	}
	return false
}

// vlanListsEqual checks if two VLAN lists contain the same elements (order doesn't matter)
func vlanListsEqual(vlans1, vlans2 []string) bool {
	if len(vlans1) != len(vlans2) {
		return false
	}
	// If both are empty, they're equal
	if len(vlans1) == 0 {
		return true
	}

	vlanSet := make(map[string]int)
	for _, vlan := range vlans1 {
		vlanSet[vlan]++
	}
	for _, vlan := range vlans2 {
		if count, ok := vlanSet[vlan]; !ok || count == 0 {
			return false
		}
		vlanSet[vlan]--
	}
	return true
}

// extractVLANNumbers extracts VLAN numbers from VlanConfigs
func extractVLANNumbers(configs []e.PortVlanConfig) []string {
	vlans := make([]string, len(configs))
	for i, config := range configs {
		vlans[i] = config.VlanNumber
	}
	return vlans
}

// DeleteConnection deletes a connection from the database by its ID
func (r BasicOpsCloverRepository) DeleteConnection(id string) error {
	err := r.db.Delete(q.NewQuery(connectionsCollection).Where(q.Field("_id").Eq(id)))
	if err != nil {
		return fmt.Errorf("DeleteConnection failed: %w", err)
	}
	return nil
}

// UpdateConnection updates a connection in the database by its ID
// Validates VLAN compatibility between the new ports
func (r BasicOpsCloverRepository) UpdateConnection(
	id string,
	from_device string,
	from_port string,
	to_device string,
	to_port string,
	allowVLANUnion bool,
) error {
	// Get VLAN numbers from both device ports
	fromPortData, err := r.GetDevicePortByIDs(from_device, from_port)
	if err != nil {
		return fmt.Errorf("failed to get from device port: %v", err)
	}

	toPortData, err := r.GetDevicePortByIDs(to_device, to_port)
	if err != nil {
		return fmt.Errorf("failed to get to device port: %v", err)
	}

	// Extract VLAN numbers from configs for validation
	fromVlans := extractVLANNumbers(fromPortData.VlanConfigs)
	toVlans := extractVLANNumbers(toPortData.VlanConfigs)

	// Validate VLAN compatibility
	if len(fromVlans) > 0 || len(toVlans) > 0 {
		if allowVLANUnion {
			// Union mode: Check if there's any VLAN overlap
			if !hasVLANOverlap(fromVlans, toVlans) {
				return fmt.Errorf("VLAN validation failed: no common VLANs between ports (from: %v, to: %v)",
					fromVlans, toVlans)
			}
		} else {
			// Strict mode: VLANs must match exactly
			if !vlanListsEqual(fromVlans, toVlans) {
				return fmt.Errorf("VLAN validation failed: VLANs must match exactly (from: %v, to: %v). Use allowVLANUnion flag to allow overlapping VLANs",
					fromVlans, toVlans)
			}
		}
	}

	updates := make(map[string]interface{})
	updates["from_device_id"] = from_device
	updates["from_model_port_id"] = from_port
	updates["to_device_id"] = to_device
	updates["to_model_port_id"] = to_port

	err = r.db.Update(q.NewQuery(connectionsCollection).Where(q.Field("_id").Eq(id)), updates)
	if err != nil {
		return fmt.Errorf("UpdateConnection failed: %w", err)
	}
	return nil
}
// AddConnectionSimple creates a connection without port usage validation
// Still validates VLAN compatibility
func (r BasicOpsCloverRepository) AddConnectionSimple(
	fromDevice string,
	fromModelPort string,
	toDevice string,
	toModelPort string,
	allowVLANUnion bool,
) error {
	// Get VLAN configs from both device ports
	fromPort, err := r.GetDevicePortByIDs(fromDevice, fromModelPort)
	if err != nil {
		return fmt.Errorf("failed to get from device port: %v", err)
	}

	toPort, err := r.GetDevicePortByIDs(toDevice, toModelPort)
	if err != nil {
		return fmt.Errorf("failed to get to device port: %v", err)
	}

	// Extract VLAN numbers from configs for validation
	fromVlans := extractVLANNumbers(fromPort.VlanConfigs)
	toVlans := extractVLANNumbers(toPort.VlanConfigs)

	// Validate VLAN compatibility
	if len(fromVlans) > 0 || len(toVlans) > 0 {
		if allowVLANUnion {
			// Union mode: Check if there's any VLAN overlap
			if !hasVLANOverlap(fromVlans, toVlans) {
				return fmt.Errorf("VLAN validation failed: no common VLANs between ports (from: %v, to: %v)",
					fromVlans, toVlans)
			}
		} else {
			// Strict mode: VLANs must match exactly
			if !vlanListsEqual(fromVlans, toVlans) {
				return fmt.Errorf("VLAN validation failed: VLANs must match exactly (from: %v, to: %v). Use allowVLANUnion flag to allow overlapping VLANs",
					fromVlans, toVlans)
			}
		}
	}

	doc := d.NewDocument()
	doc.Set("from_device_id", fromDevice)
	doc.Set("from_model_port_id", fromModelPort)
	doc.Set("to_device_id", toDevice)
	doc.Set("to_model_port_id", toModelPort)

	_, err = r.db.InsertOne(connectionsCollection, doc)
	if err != nil {
		return fmt.Errorf("failed to create connection: %v", err)
	}
	return nil
}
