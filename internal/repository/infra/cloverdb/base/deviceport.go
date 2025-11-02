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

// AddDevicePort adds a new device port to the database
func (r BasicOpsCloverRepository) AddDevicePort(deviceid string, modelportid string, macAddress string, vlanConfigs []e.PortVlanConfig, allowMultipleUntagged bool) error {
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

	// Validate untagged VLAN restriction (soft restriction)
	if !allowMultipleUntagged && len(vlanConfigs) > 0 {
		untaggedCount := 0
		for _, vc := range vlanConfigs {
			if !vc.Tagged {
				untaggedCount++
			}
		}
		if untaggedCount > 1 {
			return fmt.Errorf("multiple untagged VLANs are not allowed per port (found %d). Use allowMultipleUntagged flag to override", untaggedCount)
		}
	}

	doc := d.NewDocument()
	doc.Set("device_id", deviceid)
	doc.Set("model_port_id", modelportid)
	if macAddress != "" {
		doc.Set("mac_address", macAddress)
	}
	if len(vlanConfigs) > 0 {
		// Convert vlanConfigs to a format suitable for storage
		vlanConfigsMap := make([]map[string]interface{}, len(vlanConfigs))
		for i, vc := range vlanConfigs {
			vlanConfigsMap[i] = map[string]interface{}{
				"vlan_number": vc.VlanNumber,
				"tagged":      vc.Tagged,
			}
		}
		doc.Set("vlan_configs", vlanConfigsMap)
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

		// Get VLAN configs if they exist
		devicePort.VlanConfigs = make([]e.PortVlanConfig, 0)
		if vlanConfigs, ok := doc.Get("vlan_configs").([]interface{}); ok && len(vlanConfigs) > 0 {
			for _, vcInterface := range vlanConfigs {
				if vcMap, ok := vcInterface.(map[string]interface{}); ok {
					vlanConfig := e.PortVlanConfig{}
					if vlanNum, ok := vcMap["vlan_number"].(string); ok {
						vlanConfig.VlanNumber = vlanNum
					}
					if tagged, ok := vcMap["tagged"].(bool); ok {
						vlanConfig.Tagged = tagged
					}
					devicePort.VlanConfigs = append(devicePort.VlanConfigs, vlanConfig)
				}
			}
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

// UpdateDevicePortVLANs updates the VLAN configurations for a device port
func (r BasicOpsCloverRepository) UpdateDevicePortVLANs(deviceid string, modelportid string, vlanConfigs []e.PortVlanConfig, allowMultipleUntagged bool) error {
	// Validate untagged VLAN restriction (soft restriction)
	if !allowMultipleUntagged && len(vlanConfigs) > 0 {
		untaggedCount := 0
		for _, vc := range vlanConfigs {
			if !vc.Tagged {
				untaggedCount++
			}
		}
		if untaggedCount > 1 {
			return fmt.Errorf("multiple untagged VLANs are not allowed per port (found %d). Use allowMultipleUntagged flag to override", untaggedCount)
		}
	}

	query := q.NewQuery(deviceportsCollection).
		Where(q.Field("device_id").Eq(deviceid)).
		Where(q.Field("model_port_id").Eq(modelportid))

	updates := make(map[string]interface{})
	if len(vlanConfigs) > 0 {
		// Convert vlanConfigs to a format suitable for storage
		vlanConfigsMap := make([]map[string]interface{}, len(vlanConfigs))
		for i, vc := range vlanConfigs {
			vlanConfigsMap[i] = map[string]interface{}{
				"vlan_number": vc.VlanNumber,
				"tagged":      vc.Tagged,
			}
		}
		updates["vlan_configs"] = vlanConfigsMap
	} else {
		// If empty, remove the field
		updates["vlan_configs"] = []map[string]interface{}{}
	}

	err := r.db.Update(query, updates)
	if err != nil {
		return fmt.Errorf("UpdateDevicePortVLANs failed: %w", err)
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

	// Get VLAN configs if they exist
	devicePort.VlanConfigs = make([]e.PortVlanConfig, 0)
	if vlanConfigs, ok := doc.Get("vlan_configs").([]interface{}); ok && len(vlanConfigs) > 0 {
		for _, vcInterface := range vlanConfigs {
			if vcMap, ok := vcInterface.(map[string]interface{}); ok {
				vlanConfig := e.PortVlanConfig{}
				if vlanNum, ok := vcMap["vlan_number"].(string); ok {
					vlanConfig.VlanNumber = vlanNum
				}
				if tagged, ok := vcMap["tagged"].(bool); ok {
					vlanConfig.Tagged = tagged
				}
				devicePort.VlanConfigs = append(devicePort.VlanConfigs, vlanConfig)
			}
		}
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
