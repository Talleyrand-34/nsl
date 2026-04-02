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
func (r BasicOpsCloverRepository) AddDevicePort(deviceid string, modelportid string, macAddress string, vlanConfigs []e.PortVlanConfig) (string, error) {
	// Validate that the model port belongs to the device's model
	deviceDoc, err := r.db.FindById(devicesCollection, deviceid)
	if err != nil {
		return "", fmt.Errorf("device not found: %v", err)
	}

	modelID, ok := deviceDoc.Get("model_id").(string)
	if !ok {
		return "", fmt.Errorf("device has no model")
	}

	modelPortDoc, err := r.db.FindById(modelportsCollection, modelportid)
	if err != nil {
		return "", fmt.Errorf("model port not found: %v", err)
	}

	portModelID, ok := modelPortDoc.Get("model_id").(string)
	if !ok || portModelID != modelID {
		return "", fmt.Errorf("model port does not belong to device's model")
	}

	doc := d.NewDocument()
	doc.Set("device_id", deviceid)
	doc.Set("model_port_id", modelportid)
	if macAddress != "" {
		doc.Set("mac_address", macAddress)
	}

	if len(vlanConfigs) > 0 {
		vlansMap := make([]map[string]interface{}, len(vlanConfigs))
		for i, vc := range vlanConfigs {
			vlansMap[i] = map[string]interface{}{
				"vlan_number": vc.VlanNumber,
				"tagged":      vc.Tagged,
			}
		}
		doc.Set("vlan_configs", vlansMap)
	}

	docID, err := r.db.InsertOne(deviceportsCollection, doc)
	if err != nil {
		return "", fmt.Errorf("AddDevicePort failed: %w", err)
	}
	return docID, nil
}

// getUnmanagedDevicePortVLANs computes VLANs for an unmanaged deviceport dynamically.
// It recursively follows connections to collect all VLANs from connected deviceports.
func (r BasicOpsCloverRepository) getUnmanagedDevicePortVLANs(deviceportID string, visited map[string]struct{}) ([]e.PortVlanConfig, error) {
	// Mark as visited to prevent cycles
	visited[deviceportID] = struct{}{}

	// Find all connections where this deviceport is an endpoint
	fromConns, err := r.db.FindAll(q.NewQuery(connectionsCollection).Where(
		q.Field("from_deviceport_id").Eq(deviceportID)))
	if err != nil {
		return nil, fmt.Errorf("failed to find connections: %w", err)
	}
	toConns, err := r.db.FindAll(q.NewQuery(connectionsCollection).Where(
		q.Field("to_deviceport_id").Eq(deviceportID)))
	if err != nil {
		return nil, fmt.Errorf("failed to find connections: %w", err)
	}

	// Collect all connected deviceport IDs
	type connInfo struct {
		otherDeviceportID string
		isFrom            bool
	}
	var connectedPorts []connInfo
	for _, c := range fromConns {
		if toID, ok := c.Get("to_deviceport_id").(string); ok {
			connectedPorts = append(connectedPorts, connInfo{otherDeviceportID: toID, isFrom: false})
		}
	}
	for _, c := range toConns {
		if fromID, ok := c.Get("from_deviceport_id").(string); ok {
			connectedPorts = append(connectedPorts, connInfo{otherDeviceportID: fromID, isFrom: true})
		}
	}

	// Collect VLANs from all connected deviceports
	result := make([]e.PortVlanConfig, 0)
	seen := make(map[string]struct{})

	for _, conn := range connectedPorts {
		otherID := conn.otherDeviceportID

		// Check if already visited (cycle)
		if _, wasVisited := visited[otherID]; wasVisited {
			continue
		}

		// Get the other deviceport
		otherDoc, err := r.db.FindById(deviceportsCollection, otherID)
		if err != nil || otherDoc == nil {
			continue
		}

		// Get the parent device of the other deviceport
		otherDeviceID, _ := otherDoc.Get("device_id").(string)
		otherDeviceDoc, err := r.db.FindById(devicesCollection, otherDeviceID)
		if err != nil || otherDeviceDoc == nil {
			continue
		}

		// Check if the other device is unmanaged
		isUnmanaged, _ := otherDeviceDoc.Get("is_unmanaged").(bool)

		if isUnmanaged {
			// Recursively get VLANs from the unmanaged deviceport
			subVLANs, err := r.getUnmanagedDevicePortVLANs(otherID, visited)
			if err != nil {
				continue
			}
			for _, vc := range subVLANs {
				key := vc.VlanNumber + ":" + fmt.Sprintf("%v", vc.Tagged)
				if _, exists := seen[key]; !exists {
					seen[key] = struct{}{}
					result = append(result, vc)
				}
			}
		} else {
			// Get VLANs from the regular deviceport (stored VLANs + InterfacePorts)
			vlans := r.getStoredVLANsForDeviceport(otherDoc, otherDeviceID)
			for _, vc := range vlans {
				key := vc.VlanNumber + ":" + fmt.Sprintf("%v", vc.Tagged)
				if _, exists := seen[key]; !exists {
					seen[key] = struct{}{}
					result = append(result, vc)
				}
			}
		}
	}

	return result, nil
}

// getStoredVLANsForDeviceport returns the stored VLAN configs for a deviceport document
func (r BasicOpsCloverRepository) getStoredVLANsForDeviceport(doc *d.Document, deviceID string) []e.PortVlanConfig {
	var vlans []e.PortVlanConfig
	seen := make(map[string]struct{})

	modelPortID, _ := doc.Get("model_port_id").(string)

	// First add VLANs from DevicePort's own vlan_configs field
	if raw, ok := doc.Get("vlan_configs").([]interface{}); ok {
		for _, item := range raw {
			if m, ok := item.(map[string]interface{}); ok {
				vc := e.PortVlanConfig{}
				if vn, ok := m["vlan_number"].(string); ok {
					vc.VlanNumber = vn
				}
				if t, ok := m["tagged"].(bool); ok {
					vc.Tagged = t
				}
				if vc.VlanNumber != "" {
					key := vc.VlanNumber + ":" + fmt.Sprintf("%v", vc.Tagged)
					if _, exists := seen[key]; !exists {
						seen[key] = struct{}{}
						vlans = append(vlans, vc)
					}
				}
			}
		}
	}

	// Add VLANs from InterfacePorts
	ifacePortDocs, _ := r.db.FindAll(q.NewQuery(interfacePortsCollection).
		Where(q.Field("device_id").Eq(deviceID)).
		Where(q.Field("model_port_id").Eq(modelPortID)))
	for _, ipDoc := range ifacePortDocs {
		if raw, ok := ipDoc.Get("vlan_configs").([]interface{}); ok {
			for _, item := range raw {
				if m, ok := item.(map[string]interface{}); ok {
					vc := e.PortVlanConfig{}
					if vn, ok := m["vlan_number"].(string); ok {
						vc.VlanNumber = vn
					}
					if t, ok := m["tagged"].(bool); ok {
						vc.Tagged = t
					}
					if vc.VlanNumber != "" {
						key := vc.VlanNumber + ":" + fmt.Sprintf("%v", vc.Tagged)
						if _, exists := seen[key]; !exists {
							seen[key] = struct{}{}
							vlans = append(vlans, vc)
						}
					}
				}
			}
		}
	}

	return vlans
}

// GetDevicePorts gets all the device ports available
func (r BasicOpsCloverRepository) GetDevicePorts() ([]e.DevicePort, error) {
	docs, err := r.db.FindAll(q.NewQuery(deviceportsCollection))
	if err != nil {
		return []e.DevicePort{}, err
	}

	// Pre-load interface ports with their VLAN configs
	// key = "deviceID:modelPortID" → []e.PortVlanConfig (aggregated from all InterfacePorts on that port)
	ifacePortDocs, _ := r.db.FindAll(q.NewQuery(interfacePortsCollection))
	portToVlanConfigs := make(map[string][]e.PortVlanConfig)
	for _, ipDoc := range ifacePortDocs {
		devID, _ := ipDoc.Get("device_id").(string)
		mpID, _ := ipDoc.Get("model_port_id").(string)
		if devID != "" && mpID != "" {
			key := devID + ":" + mpID
			// Parse VLAN configs from InterfacePort
			if raw, ok := ipDoc.Get("vlan_configs").([]interface{}); ok {
				for _, item := range raw {
					if m, ok := item.(map[string]interface{}); ok {
						vc := e.PortVlanConfig{}
						if vn, ok := m["vlan_number"].(string); ok {
							vc.VlanNumber = vn
						}
						if t, ok := m["tagged"].(bool); ok {
							vc.Tagged = t
						}
						if vc.VlanNumber != "" {
							portToVlanConfigs[key] = append(portToVlanConfigs[key], vc)
						}
					}
				}
			}
		}
	}

	result := make([]e.DevicePort, 0, len(docs))
	for _, doc := range docs {
		devicePort := e.DevicePort{
			ID: doc.ObjectId(),
		}

		// Get device ID and check if unmanaged
		var isUnmanaged bool
		if deviceID, ok := doc.Get("device_id").(string); ok {
			devicePort.DeviceID = deviceID

			// Get device label
			deviceDoc, err := r.db.FindById(devicesCollection, deviceID)
			if err == nil && deviceDoc != nil {
				devicePort.DevLabel = deviceDoc.Get("label").(string)
				isUnmanaged, _ = deviceDoc.Get("is_unmanaged").(bool)
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

			// For unmanaged devices, compute VLANs dynamically
			if isUnmanaged {
				devicePort.VlanConfigs, err = r.getUnmanagedDevicePortVLANs(devicePort.ID, make(map[string]struct{}))
				if err != nil {
					devicePort.VlanConfigs = []e.PortVlanConfig{}
				}
			} else {
				// Collect VLAN configs from all InterfacePorts linked to this port AND from DevicePort's own vlan_configs
				seen := make(map[string]struct{})
				// First add VLANs from DevicePort's own vlan_configs field
				if raw, ok := doc.Get("vlan_configs").([]interface{}); ok {
					for _, item := range raw {
						if m, ok := item.(map[string]interface{}); ok {
							vc := e.PortVlanConfig{}
							if vn, ok := m["vlan_number"].(string); ok {
								vc.VlanNumber = vn
							}
							if t, ok := m["tagged"].(bool); ok {
								vc.Tagged = t
							}
							if vc.VlanNumber != "" {
								key := vc.VlanNumber + ":" + fmt.Sprintf("%v", vc.Tagged)
								if _, exists := seen[key]; !exists {
									seen[key] = struct{}{}
									devicePort.VlanConfigs = append(devicePort.VlanConfigs, vc)
								}
							}
						}
					}
				}
				// Then add VLANs from InterfacePorts
				for _, vc := range portToVlanConfigs[devicePort.DeviceID+":"+modelPortID] {
					key := vc.VlanNumber + ":" + fmt.Sprintf("%v", vc.Tagged)
					if _, exists := seen[key]; !exists {
						seen[key] = struct{}{}
						devicePort.VlanConfigs = append(devicePort.VlanConfigs, vc)
					}
				}
				if devicePort.VlanConfigs == nil {
					devicePort.VlanConfigs = []e.PortVlanConfig{}
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
	// First, find the deviceport document to get its ID
	query := q.NewQuery(deviceportsCollection).
		Where(q.Field("device_id").Eq(deviceid)).
		Where(q.Field("model_port_id").Eq(modelportid))

	doc, err := r.db.FindFirst(query)
	if err != nil {
		return fmt.Errorf("DeleteDevicePort: failed to find deviceport: %w", err)
	}
	if doc == nil {
		return fmt.Errorf("DeleteDevicePort: device port not found")
	}

	deviceportID := doc.ObjectId()

	// Check for dependent connections (from side)
	fromExists, err := r.db.Exists(q.NewQuery(connectionsCollection).Where(
		q.Field("from_deviceport_id").Eq(deviceportID)))
	if err != nil {
		return err
	}
	// Check for dependent connections (to side)
	toExists, err := r.db.Exists(q.NewQuery(connectionsCollection).Where(
		q.Field("to_deviceport_id").Eq(deviceportID)))
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
		ID:       doc.ObjectId(),
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
