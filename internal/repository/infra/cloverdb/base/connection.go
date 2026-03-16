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

	fromQuery := q.NewQuery(connectionsCollection).Where(
		q.Field("from_device_id").Eq(fromDevice).And(
			q.Field("from_model_port_id").Eq(fromModelPort),
		),
	)
	fromExists, err := r.db.Exists(fromQuery)
	if err != nil {
		return fmt.Errorf("error checking source port (device: %s, port: %s): %w", fromDevice, fromModelPort, err)
	}

	toQuery := q.NewQuery(connectionsCollection).Where(
		q.Field("to_device_id").Eq(toDevice).And(
			q.Field("to_model_port_id").Eq(toModelPort),
		),
	)
	toExists, err := r.db.Exists(toQuery)
	if err != nil {
		return fmt.Errorf("error checking target port (device: %s, port: %s): %w", toDevice, toModelPort, err)
	}
	if fromExists && toExists {
		return fmt.Errorf("both ports are already connected (source: %s/%s, target: %s/%s)", fromDevice, fromModelPort, toDevice, toModelPort)
	}
	if fromExists {
		return fmt.Errorf("source port already in use (device: %s, port: %s)", fromDevice, fromModelPort)
	}
	if toExists {
		return fmt.Errorf("target port already in use (device: %s, port: %s)", toDevice, toModelPort)
	}

	// Note: VLAN validation is now done through DeviceInterface, not DevicePort.
	// Physical ports no longer carry VLAN configs directly.

	doc := d.NewDocument()
	doc.Set("from_device_id", fromDevice)
	doc.Set("from_model_port_id", fromModelPort)
	doc.Set("to_device_id", toDevice)
	doc.Set("to_model_port_id", toModelPort)

	if _, err := r.db.InsertOne(connectionsCollection, doc); err != nil {
		return fmt.Errorf("failed to create connection: %v", err)
	}
	return nil
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
	// Note: VLAN validation is now done through DeviceInterface, not DevicePort.
	// Physical ports no longer carry VLAN configs directly.

	updates := make(map[string]interface{})
	updates["from_device_id"] = from_device
	updates["from_model_port_id"] = from_port
	updates["to_device_id"] = to_device
	updates["to_model_port_id"] = to_port

	if err := r.db.Update(q.NewQuery(connectionsCollection).Where(q.Field("_id").Eq(id)), updates); err != nil {
		return fmt.Errorf("UpdateConnection failed: %w", err)
	}
	return nil
}

// AddConnectionSimple creates a connection without port usage validation
func (r BasicOpsCloverRepository) AddConnectionSimple(
	fromDevice string,
	fromModelPort string,
	toDevice string,
	toModelPort string,
	allowVLANUnion bool,
) error {
	// Note: VLAN validation is now done through DeviceInterface, not DevicePort.
	// Physical ports no longer carry VLAN configs directly.

	doc := d.NewDocument()
	doc.Set("from_device_id", fromDevice)
	doc.Set("from_model_port_id", fromModelPort)
	doc.Set("to_device_id", toDevice)
	doc.Set("to_model_port_id", toModelPort)

	if _, err := r.db.InsertOne(connectionsCollection, doc); err != nil {
		return fmt.Errorf("failed to create connection: %v", err)
	}
	return nil
}
