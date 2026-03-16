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
	"log"

	q "github.com/ostafen/clover/v2/query"

	e "nsl-graph/internal/repository/entities"
)

// DeleteBrandCascade deletes a brand and all dependent models and devices
func (r BasicOpsCloverRepository) DeleteBrandCascade(brandName string) error {
	log.Printf("Cascade deleting brand: %s", brandName)

	// Get all models with this brand
	models, err := r.GetModels()
	if err != nil {
		return fmt.Errorf("failed to get models: %w", err)
	}

	for _, model := range models {
		if model.Brand == brandName {
			// Cascade delete each model (which will delete devices and their dependencies)
			if err := r.DeleteModelCascade(model.ID); err != nil {
				return fmt.Errorf("failed to cascade delete model %s: %w", model.ID, err)
			}
		}
	}

	// Finally delete the brand itself
	return r.DeleteBrand(brandName)
}

// DeleteDeviceClassCascade deletes a device class and all dependent models and devices
func (r BasicOpsCloverRepository) DeleteDeviceClassCascade(deviceClassName string) error {
	log.Printf("Cascade deleting device class: %s", deviceClassName)

	// Get all models with this device class
	models, err := r.GetModels()
	if err != nil {
		return fmt.Errorf("failed to get models: %w", err)
	}

	for _, model := range models {
		if model.Class == deviceClassName {
			// Cascade delete each model
			if err := r.DeleteModelCascade(model.ID); err != nil {
				return fmt.Errorf("failed to cascade delete model %s: %w", model.ID, err)
			}
		}
	}

	// Finally delete the device class itself
	return r.DeleteDeviceClass(deviceClassName)
}

// DeleteZoneTypeCascade deletes a zone type and all dependent zones and devices
func (r BasicOpsCloverRepository) DeleteZoneTypeCascade(zoneTypeName string) error {
	log.Printf("Cascade deleting zone type: %s", zoneTypeName)

	// Get all zones with this zone type
	zones, err := r.GetZones()
	if err != nil {
		return fmt.Errorf("failed to get zones: %w", err)
	}

	for _, zone := range zones {
		if zone.LocationType == zoneTypeName {
			// Cascade delete each zone
			if err := r.DeleteZoneCascade(zone.ID); err != nil {
				return fmt.Errorf("failed to cascade delete zone %s: %w", zone.ID, err)
			}
		}
	}

	// Finally delete the zone type itself
	return r.DeleteZoneType(zoneTypeName)
}

// DeleteProprietaryCascade deletes a proprietary, setting the proprietary field to empty
// on all referencing zones and devices (set-null, not cascade-delete).
func (r BasicOpsCloverRepository) DeleteProprietaryCascade(proprietaryName string) error {
	log.Printf("Cascade deleting proprietary: %s", proprietaryName)
	// DeleteProprietary already handles set-null on zones and devices before deleting
	return r.DeleteProprietary(proprietaryName)
}

// DeleteZoneCascade deletes a zone and all dependent devices
func (r BasicOpsCloverRepository) DeleteZoneCascade(zoneId string) error {
	log.Printf("Cascade deleting zone: %s", zoneId)

	// Get all devices in this zone
	devices, err := r.GetDevices()
	if err != nil {
		return fmt.Errorf("failed to get devices: %w", err)
	}

	for _, device := range devices {
		if device.ZoneID == zoneId {
			// Cascade delete each device
			if err := r.DeleteDeviceCascade(device.ID); err != nil {
				return fmt.Errorf("failed to cascade delete device %s: %w", device.ID, err)
			}
		}
	}

	// Finally delete the zone itself
	return r.DeleteZone(zoneId)
}

// DeleteModelCascade deletes a model and all dependent devices and model ports
func (r BasicOpsCloverRepository) DeleteModelCascade(modelId string) error {
	log.Printf("Cascade deleting model: %s", modelId)

	// Get all devices using this model
	devices, err := r.GetDevices()
	if err != nil {
		return fmt.Errorf("failed to get devices: %w", err)
	}

	for _, device := range devices {
		if device.Model == modelId {
			// Cascade delete each device
			if err := r.DeleteDeviceCascade(device.ID); err != nil {
				return fmt.Errorf("failed to cascade delete device %s: %w", device.ID, err)
			}
		}
	}

	// Get all model ports for this model
	modelPorts, err := r.GetModelPorts()
	if err != nil {
		return fmt.Errorf("failed to get model ports: %w", err)
	}

	for _, modelPort := range modelPorts {
		if modelPort.Model == modelId {
			// Cascade delete each model port
			if err := r.DeleteModelPortCascade(modelPort.ID); err != nil {
				return fmt.Errorf("failed to cascade delete model port %s: %w", modelPort.ID, err)
			}
		}
	}

	// Finally delete the model itself
	return r.DeleteModel(modelId)
}

// DeleteDeviceCascade deletes a device and all dependent device ports and connections
func (r BasicOpsCloverRepository) DeleteDeviceCascade(deviceId string) error {
	log.Printf("Cascade deleting device: %s", deviceId)

	// Get all device ports for this device
	devicePorts, err := r.GetDevicePorts()
	if err != nil {
		return fmt.Errorf("failed to get device ports: %w", err)
	}

	for _, devicePort := range devicePorts {
		if devicePort.DeviceID == deviceId {
			// Cascade delete each device port
			if err := r.DeleteDevicePortCascade(devicePort.DeviceID, devicePort.ModelID); err != nil {
				return fmt.Errorf("failed to cascade delete device port %s:%s: %w", devicePort.DeviceID, devicePort.ModelID, err)
			}
		}
	}

	// Get all connections involving this device
	connections, err := r.GetConnections()
	if err != nil {
		return fmt.Errorf("failed to get connections: %w", err)
	}

	for _, connection := range connections {
		if connection.FromDevice == deviceId || connection.ToDevice == deviceId {
			// Delete connections (no cascade needed for connections)
			if err := r.DeleteConnection(connection.ID); err != nil {
				return fmt.Errorf("failed to delete connection %s: %w", connection.ID, err)
			}
		}
	}

	// Finally delete the device itself
	return r.DeleteDevice(deviceId)
}

// DeleteModelPortCascade deletes a model port and all dependent device ports and connections
func (r BasicOpsCloverRepository) DeleteModelPortCascade(modelPortId string) error {
	log.Printf("Cascade deleting model port: %s", modelPortId)

	// Get all device ports using this model port
	devicePorts, err := r.GetDevicePorts()
	if err != nil {
		return fmt.Errorf("failed to get device ports: %w", err)
	}

	for _, devicePort := range devicePorts {
		if devicePort.ModelID == modelPortId {
			// Cascade delete each device port
			if err := r.DeleteDevicePortCascade(devicePort.DeviceID, devicePort.ModelID); err != nil {
				return fmt.Errorf("failed to cascade delete device port %s:%s: %w", devicePort.DeviceID, devicePort.ModelID, err)
			}
		}
	}

	// Finally delete the model port itself
	return r.DeleteModelPort(modelPortId)
}

// DeleteDevicePortCascade deletes a device port and all dependent connections
func (r BasicOpsCloverRepository) DeleteDevicePortCascade(deviceId string, modelPortId string) error {
	log.Printf("Cascade deleting device port: %s:%s", deviceId, modelPortId)

	// Get all connections using this device port
	connections, err := r.GetConnections()
	if err != nil {
		return fmt.Errorf("failed to get connections: %w", err)
	}

	for _, connection := range connections {
		// Check if this connection involves the device port being deleted
		if (connection.FromDevice == deviceId && connection.FromModelPort == modelPortId) ||
			(connection.ToDevice == deviceId && connection.ToModelPort == modelPortId) {
			// Delete connection (no cascade needed for connections)
			if err := r.DeleteConnection(connection.ID); err != nil {
				return fmt.Errorf("failed to delete connection %s: %w", connection.ID, err)
			}
		}
	}

	// Finally delete the device port itself
	return r.DeleteDevicePort(deviceId, modelPortId)
}

// DeleteConnectionCascade deletes a connection (connections have no dependencies)
func (r BasicOpsCloverRepository) DeleteConnectionCascade(connectionId string) error {
	log.Printf("Cascade deleting connection: %s", connectionId)
	// Connections have no dependencies, so just delete normally
	return r.DeleteConnection(connectionId)
}

// DeleteVlanCascade deletes a VLAN, its local VLAN entries, and removes it from all device port configurations
func (r BasicOpsCloverRepository) DeleteVlanCascade(vlanId string) error {
	log.Printf("Cascade deleting VLAN: %s", vlanId)

	// Get the VLAN to find its VLAN number
	vlans, err := r.GetVlans()
	if err != nil {
		return fmt.Errorf("failed to get VLANs: %w", err)
	}

	var vlanNumber string
	for _, vlan := range vlans {
		if vlan.ID == vlanId {
			vlanNumber = vlan.VlanID
			break
		}
	}

	if vlanNumber == "" {
		return fmt.Errorf("VLAN with ID %s not found", vlanId)
	}

	// Delete all local VLAN entries referencing this VLAN number
	if err := r.db.Delete(q.NewQuery(localvlansCollection).Where(q.Field("vlan_id").Eq(vlanNumber))); err != nil {
		return fmt.Errorf("failed to delete local VLANs for vlan %s: %w", vlanNumber, err)
	}

	// Get all device ports and remove VLAN configurations
	devicePorts, err := r.GetDevicePorts()
	if err != nil {
		return fmt.Errorf("failed to get device ports: %w", err)
	}

	for _, devicePort := range devicePorts {
		// Check if this device port has the VLAN configured
		var updatedVlanConfigs []e.PortVlanConfig
		modified := false

		for _, vlanConfig := range devicePort.VlanConfigs {
			if vlanConfig.VlanNumber != vlanNumber {
				updatedVlanConfigs = append(updatedVlanConfigs, vlanConfig)
			} else {
				modified = true
			}
		}

		// Update the device port if VLAN configs were modified
		if modified {
			if err := r.UpdateDevicePort(devicePort.DeviceID, devicePort.ModelID, devicePort.MacAddress, updatedVlanConfigs); err != nil {
				return fmt.Errorf("failed to update device port %s:%s: %w", devicePort.DeviceID, devicePort.ModelID, err)
			}
		}
	}

	// Finally delete the VLAN itself (local vlans already deleted, so check passes)
	return r.DeleteVlan(vlanId)
}