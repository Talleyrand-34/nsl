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

		// Resolve the connection type name from its id.
		if typeID, ok := doc.Get("connection_type").(string); ok && typeID != "" {
			if typeDoc, err := r.db.FindById(connectiontypesCollection, typeID); err == nil && typeDoc != nil {
				if name, ok := typeDoc.Get("connection_type").(string); ok {
					connection.ConnectionType = name
				}
			}
		}

		// Provenance of an auto-discovered link, if any.
		if raw, ok := doc.Get("discovered_via").([]interface{}); ok {
			for _, v := range raw {
				if s, ok := v.(string); ok {
					connection.DiscoveredVia = append(connection.DiscoveredVia, s)
				}
			}
		}

		// Get from deviceport info
		if fromDeviceportID, ok := doc.Get("from_deviceport_id").(string); ok && fromDeviceportID != "" {
			deviceportDoc, err := r.db.FindById(deviceportsCollection, fromDeviceportID)
			if err == nil && deviceportDoc != nil {
				// Get device info
				if deviceID, ok := deviceportDoc.Get("device_id").(string); ok && deviceID != "" {
					deviceDoc, err := r.db.FindById(devicesCollection, deviceID)
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
				if modelPortID, ok := deviceportDoc.Get("model_port_id").(string); ok && modelPortID != "" {
					modelPortDoc, err := r.db.FindById(modelportsCollection, modelPortID)
					if err == nil && modelPortDoc != nil {
						connection.FromModelPort = modelPortDoc.Get("name").(string)
					}
				}
			}
		}

		// Get to deviceport info
		if toDeviceportID, ok := doc.Get("to_deviceport_id").(string); ok && toDeviceportID != "" {
			deviceportDoc, err := r.db.FindById(deviceportsCollection, toDeviceportID)
			if err == nil && deviceportDoc != nil {
				// Get device info
				if deviceID, ok := deviceportDoc.Get("device_id").(string); ok && deviceID != "" {
					deviceDoc, err := r.db.FindById(devicesCollection, deviceID)
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
				if modelPortID, ok := deviceportDoc.Get("model_port_id").(string); ok && modelPortID != "" {
					modelPortDoc, err := r.db.FindById(modelportsCollection, modelPortID)
					if err == nil && modelPortDoc != nil {
						connection.ToModelPort = modelPortDoc.Get("name").(string)
					}
				}
			}
		}

		result = append(result, connection)
	}

	return result, nil
}

// AddConnection creates a new connection between two device ports. Optional
// discoveredVia strings record the provenance of an auto-discovered link.
func (r BasicOpsCloverRepository) AddConnection(
	fromDeviceportID string,
	toDeviceportID string,
	connectionType string,
	discoveredVia ...string,
) error {
	// A connection must reference an existing connection type (strict: the data
	// layer never silently drops an unresolved dependency).
	if connectionType == "" {
		return fmt.Errorf("connection type is required")
	}
	connectionTypeID := r.getConnectionTypeID(connectionType)
	if connectionTypeID == "" {
		return fmt.Errorf("connection type %q does not exist", connectionType)
	}

	fromQuery := q.NewQuery(connectionsCollection).Where(
		q.Field("from_deviceport_id").Eq(fromDeviceportID),
	)
	fromExists, err := r.db.Exists(fromQuery)
	if err != nil {
		return fmt.Errorf("error checking source port: %w", err)
	}

	toQuery := q.NewQuery(connectionsCollection).Where(
		q.Field("to_deviceport_id").Eq(toDeviceportID),
	)
	toExists, err := r.db.Exists(toQuery)
	if err != nil {
		return fmt.Errorf("error checking target port: %w", err)
	}
	if fromExists && toExists {
		return fmt.Errorf("both ports are already connected")
	}
	if fromExists {
		return fmt.Errorf("source port already in use")
	}
	if toExists {
		return fmt.Errorf("target port already in use")
	}

	doc := d.NewDocument()
	doc.Set("from_deviceport_id", fromDeviceportID)
	doc.Set("to_deviceport_id", toDeviceportID)
	doc.Set("connection_type", connectionTypeID)
	if len(discoveredVia) > 0 {
		doc.Set("discovered_via", discoveredVia)
	}

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
func (r BasicOpsCloverRepository) UpdateConnection(
	id string,
	fromDeviceportID string,
	toDeviceportID string,
) error {
	updates := make(map[string]interface{})
	updates["from_deviceport_id"] = fromDeviceportID
	updates["to_deviceport_id"] = toDeviceportID

	if err := r.db.Update(q.NewQuery(connectionsCollection).Where(q.Field("_id").Eq(id)), updates); err != nil {
		return fmt.Errorf("UpdateConnection failed: %w", err)
	}
	return nil
}

// AddConnectionSimple creates a connection without port usage validation
func (r BasicOpsCloverRepository) AddConnectionSimple(
	fromDeviceportID string,
	toDeviceportID string,
) error {
	doc := d.NewDocument()
	doc.Set("from_deviceport_id", fromDeviceportID)
	doc.Set("to_deviceport_id", toDeviceportID)

	if _, err := r.db.InsertOne(connectionsCollection, doc); err != nil {
		return fmt.Errorf("failed to create connection: %v", err)
	}
	return nil
}
