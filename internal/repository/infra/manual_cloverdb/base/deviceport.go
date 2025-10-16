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

// AddDevicePort adds a new device port to the database
func (r BasicOpsCloverRepository) AddDevicePort(deviceid string, modelportid string, macAddress string) error {
	// Validate that the model port belongs to the device's model
	// This check ensures the port is valid for this device
	deviceDoc, err := r.db.FindById(devicesCollection, deviceid)
	if err != nil {
		return fmt.Errorf("device not found: %v", err)
	}

	modelID, ok := deviceDoc.Get("model_id").(string)
	if !ok {
		return fmt.Errorf("device has no model")
	}

	modelPortDoc, err := r.db.FindById(modelportsCollection, modelportid)
	if err != nil {
		return fmt.Errorf("model port not found: %v", err)
	}

	portModelID, ok := modelPortDoc.Get("model_id").(string)
	if !ok || portModelID != modelID {
		return fmt.Errorf("model port does not belong to device's model")
	}

	doc := d.NewDocument()
	doc.Set("device_id", deviceid)
	doc.Set("model_port_id", modelportid)
	if macAddress != "" {
		doc.Set("mac_address", macAddress)
	}

	_, err = r.db.InsertOne(deviceportsCollection, doc)
	if err != nil {
		return fmt.Errorf("AddDevicePort failed: %w", err)
	}
	return nil
}

// GetDevicePorts gets all the device ports available
func (r BasicOpsCloverRepository) GetDevicePorts() ([]e.DevicePort, error) {
	docs, err := r.db.FindAll(q.NewQuery(deviceportsCollection))
	if err != nil {
		return []e.DevicePort{}, err
	}

	result := make([]e.DevicePort, 0, len(docs))
	for _, doc := range docs {
		devicePort := e.DevicePort{}

		// Get device ID
		if deviceID, ok := doc.Get("device_id").(string); ok {
			devicePort.DeviceID = deviceID

			// Get device label
			deviceDoc, err := r.db.FindById(devicesCollection, deviceID)
			if err == nil && deviceDoc != nil {
				devicePort.DevLabel = deviceDoc.Get("label").(string)
			}
		}

		// Get model port ID and details
		if modelPortID, ok := doc.Get("model_port_id").(string); ok {
			devicePort.ModelID = modelPortID

			// Get model port details
			modelPortDoc, err := r.db.FindById(modelportsCollection, modelPortID)
			if err == nil && modelPortDoc != nil {
				devicePort.PortName = modelPortDoc.Get("name").(string)

				if px, ok := modelPortDoc.Get("position_x").(float64); ok {
					devicePort.Positionx = int(px)
				} else if px, ok := modelPortDoc.Get("position_x").(int); ok {
					devicePort.Positionx = px
				}

				if py, ok := modelPortDoc.Get("position_y").(float64); ok {
					devicePort.Positiony = int(py)
				} else if py, ok := modelPortDoc.Get("position_y").(int); ok {
					devicePort.Positiony = py
				}
			}
		}

		// Get MAC address if it exists
		if macAddr, ok := doc.Get("mac_address").(string); ok {
			devicePort.MacAddress = macAddr
		}

		result = append(result, devicePort)
	}

	return result, nil
}

// DeleteDevicePort deletes a device port from the database by device and model port IDs
func (r BasicOpsCloverRepository) DeleteDevicePort(deviceid string, modelportid string) error {
	query := q.NewQuery(deviceportsCollection).
		Where(q.Field("device_id").Eq(deviceid)).
		Where(q.Field("model_port_id").Eq(modelportid))

	err := r.db.Delete(query)
	if err != nil {
		return fmt.Errorf("DeleteDevicePort failed: %w", err)
	}
	return nil
}
