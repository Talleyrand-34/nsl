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

// DevicePortExists checks if a device port already exists
func (r BasicOpsCloverRepository) DevicePortExists(deviceid string, modelportid string) (bool, error) {
	query := q.NewQuery(deviceportsCollection).
		Where(q.Field("device_id").Eq(deviceid)).
		Where(q.Field("model_port_id").Eq(modelportid))

	exists, err := r.db.Exists(query)
	if err != nil {
		return false, fmt.Errorf("error checking device port existence: %w", err)
	}
	return exists, nil
}

// AddDevicePort adds a new device port to the database.
// vlanConfigs is accepted but ignored — VLAN data is now stored on DeviceInterface.
func (r BasicOpsCloverRepository) AddDevicePort(deviceid string, modelportid string, macAddress string, vlanConfigs []e.PortVlanConfig) error {
	// Validate that the model port belongs to the device's model
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
	// Check for dependent connections (from side)
	fromExists, err := r.db.Exists(q.NewQuery(connectionsCollection).Where(
		q.Field("from_device_id").Eq(deviceid).And(q.Field("from_model_port_id").Eq(modelportid))))
	if err != nil {
		return err
	}
	// Check for dependent connections (to side)
	toExists, err := r.db.Exists(q.NewQuery(connectionsCollection).Where(
		q.Field("to_device_id").Eq(deviceid).And(q.Field("to_model_port_id").Eq(modelportid))))
	if err != nil {
		return err
	}
	if fromExists || toExists {
		return fmt.Errorf("cannot delete device port: referenced by connections; use --cascade to delete all dependents")
	}

	// Cascade-delete InterfacePort entries for this physical port
	if err := r.db.Delete(q.NewQuery(interfacePortsCollection).
		Where(q.Field("device_id").Eq(deviceid)).
		Where(q.Field("model_port_id").Eq(modelportid))); err != nil {
		return fmt.Errorf("DeleteDevicePort: failed to delete interface ports: %w", err)
	}

	query := q.NewQuery(deviceportsCollection).
		Where(q.Field("device_id").Eq(deviceid)).
		Where(q.Field("model_port_id").Eq(modelportid))

	err = r.db.Delete(query)
	if err != nil {
		return fmt.Errorf("DeleteDevicePort failed: %w", err)
	}
	return nil
}

// UpdateDevicePort updates a device port's MAC address.
// vlanConfigs is accepted for interface compatibility but ignored — VLAN data
// is now managed via DeviceInterface.
func (r BasicOpsCloverRepository) UpdateDevicePort(deviceid string, modelportid string, macAddress string, vlanConfigs []e.PortVlanConfig) error {
	query := q.NewQuery(deviceportsCollection).Where(
		q.Field("device_id").Eq(deviceid).And(
			q.Field("model_port_id").Eq(modelportid),
		),
	)

	updates := map[string]interface{}{
		"mac_address": macAddress,
	}

	err := r.db.Update(query, updates)
	if err != nil {
		return fmt.Errorf("UpdateDevicePort failed: %w", err)
	}
	return nil
}

// GetDevicePortByIDs retrieves a device port by device and model port IDs
func (r BasicOpsCloverRepository) GetDevicePortByIDs(deviceid string, modelportid string) (*e.DevicePort, error) {
	query := q.NewQuery(deviceportsCollection).
		Where(q.Field("device_id").Eq(deviceid)).
		Where(q.Field("model_port_id").Eq(modelportid))

	doc, err := r.db.FindFirst(query)
	if err != nil {
		return nil, fmt.Errorf("GetDevicePortByIDs failed: %w", err)
	}
	if doc == nil {
		return nil, fmt.Errorf("device port not found")
	}

	devicePort := &e.DevicePort{
		DeviceID: deviceid,
		ModelID:  modelportid,
	}

	// Get MAC address if it exists
	if macAddr, ok := doc.Get("mac_address").(string); ok {
		devicePort.MacAddress = macAddr
	}

	// Get device label
	deviceDoc, err := r.db.FindById(devicesCollection, deviceid)
	if err == nil && deviceDoc != nil {
		devicePort.DevLabel = deviceDoc.Get("label").(string)
	}

	// Get model port details
	modelPortDoc, err := r.db.FindById(modelportsCollection, modelportid)
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

	return devicePort, nil
}
