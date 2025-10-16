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

		// Get IP segments
		if fromIPSegment, ok := doc.Get("from_ip_segment").(string); ok {
			connection.FromIPSegment = fromIPSegment
		}
		if toIPSegment, ok := doc.Get("to_ip_segment").(string); ok {
			connection.ToIPSegment = toIPSegment
		}

		result = append(result, connection)
	}

	return result, nil
}

// AddConnection creates a new connection between two device ports
func (r BasicOpsCloverRepository) AddConnection(
	fromDevice string,
	fromModelPort string,
	fromIPSegment string,
	toDevice string,
	toModelPort string,
	toIPSegment string,
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

	doc := d.NewDocument()
	doc.Set("from_device_id", fromDevice)
	doc.Set("from_model_port_id", fromModelPort)
	doc.Set("to_device_id", toDevice)
	doc.Set("to_model_port_id", toModelPort)

	if fromIPSegment != "" {
		doc.Set("from_ip_segment", fromIPSegment)
	}
	if toIPSegment != "" {
		doc.Set("to_ip_segment", toIPSegment)
	}

	_, err = r.db.InsertOne(connectionsCollection, doc)
	if err != nil {
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
func (r BasicOpsCloverRepository) UpdateConnection(
	id string,
	from_device string,
	from_port string,
	from_ip_segment string,
	to_device string,
	to_port string,
	to_ip_segment string,
) error {
	updates := make(map[string]interface{})
	updates["from_device_id"] = from_device
	updates["from_model_port_id"] = from_port
	updates["to_device_id"] = to_device
	updates["to_model_port_id"] = to_port

	if from_ip_segment != "" {
		updates["from_ip_segment"] = from_ip_segment
	}
	if to_ip_segment != "" {
		updates["to_ip_segment"] = to_ip_segment
	}

	err := r.db.Update(q.NewQuery(connectionsCollection).Where(q.Field("_id").Eq(id)), updates)
	if err != nil {
		return fmt.Errorf("UpdateConnection failed: %w", err)
	}
	return nil
}
