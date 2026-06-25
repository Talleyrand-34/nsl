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

// AddDevice adds a new device to the database
func (r BasicOpsCloverRepository) AddDevice(
	label string,
	model string,
	zoneId string,
	zoneName string,
	owner string,
	isUnmanaged bool,
	isInvisible bool,
) error {
	// Check for name collision with existing devices
	existingDevices, err := r.GetDevices()
	if err != nil {
		return fmt.Errorf("failed to check existing devices: %w", err)
	}
	for _, d := range existingDevices {
		if d.Name == label {
			return fmt.Errorf("device with name %q already exists", label)
		}
	}

	// Get owner ID
	spropid := r.getOwnerID(owner)

	// Get zone ID
	szoneid := r.getZoneID(zoneId, zoneName)

	// Get model ID
	smodelid, err := r.getModelID(model)
	if err != nil {
		return err
	}

	doc := d.NewDocument()
	doc.Set("label", label)
	doc.Set("model_id", smodelid)
	doc.Set("is_unmanaged", isUnmanaged)
	doc.Set("is_invisible", isInvisible)
	if spropid != "" {
		doc.Set("owner", spropid)
	}
	if szoneid != "" {
		doc.Set("zone_id", szoneid)
	}

	_, err = r.db.InsertOne(devicesCollection, doc)
	if err != nil {
		return fmt.Errorf("AddDevice failed: %w", err)
	}
	return nil
}

// GetDevices gets all the devices available
func (r BasicOpsCloverRepository) GetDevices() ([]e.Device, error) {
	docs, err := r.db.FindAll(q.NewQuery(devicesCollection))
	if err != nil {
		return []e.Device{}, err
	}

	result := make([]e.Device, 0, len(docs))
	for _, doc := range docs {
		device := e.Device{
			ID:   doc.ObjectId(),
			Name: doc.Get("label").(string),
		}

		// IP addresses (stored as a string array on the device document).
		switch ips := doc.Get("ips").(type) {
		case []string:
			device.Ips = ips
		case []interface{}:
			for _, v := range ips {
				if s, ok := v.(string); ok {
					device.Ips = append(device.Ips, s)
				}
			}
		}

		// Associated scan-profile name (device or generic), if any.
		if profile, ok := doc.Get("profile").(string); ok {
			device.Profile = profile
		}

		// Get unmanaged and invisible flags
		if isUnmanaged, ok := doc.Get("is_unmanaged").(bool); ok {
			device.IsUnmanaged = isUnmanaged
		}
		if isInvisible, ok := doc.Get("is_invisible").(bool); ok {
			device.IsInvisible = isInvisible
		}

		// Get model name and brand if model ID exists
		if modelID, ok := doc.Get("model_id").(string); ok && modelID != "" {
			modelDoc, err := r.db.FindById(modelsCollection, modelID)
			if err == nil && modelDoc != nil {
				device.Model = modelDoc.Get("model").(string)

				// Get brand name from model's brand ID
				if brandID, ok := modelDoc.Get("brand").(string); ok && brandID != "" {
					brandDoc, err := r.db.FindById(brandsCollection, brandID)
					if err == nil && brandDoc != nil {
						device.Brand = brandDoc.Get("brand").(string)
					}
				}
			}
		}

		// Get zone name and father if zone ID exists
		if zoneID, ok := doc.Get("zone_id").(string); ok && zoneID != "" {
			zoneDoc, err := r.db.FindById(zonesCollection, zoneID)
			if err == nil && zoneDoc != nil {
				device.ZoneName = zoneDoc.Get("name").(string)
				device.ZoneID = zoneID

				// Get zone father name if it exists
				if fatherID, ok := zoneDoc.Get("father").(string); ok && fatherID != "" {
					fatherDoc, err := r.db.FindById(zonesCollection, fatherID)
					if err == nil && fatherDoc != nil {
						device.ZoneFather = fatherDoc.Get("name").(string)
					}
				}
			}
		}

		// Get owner name if owner ID exists
		if ownerID, ok := doc.Get("owner").(string); ok && ownerID != "" {
			ownerDoc, err := r.db.FindById(ownersCollection, ownerID)
			if err == nil && ownerDoc != nil {
				device.Owner = ownerDoc.Get("owner").(string)
			}
		}

		result = append(result, device)
	}

	return result, nil
}

// UpdateDevice updates a device in the database by its ID
func (r BasicOpsCloverRepository) UpdateDevice(
	deviceId string,
	newDeviceLabel string,
	newModelId string,
	newZoneId string,
	newOwnerId string,
	isUnmanaged *bool,
) error {
	updates := make(map[string]interface{})

	if newDeviceLabel != "" {
		updates["label"] = newDeviceLabel
	}

	if newModelId != "" && newModelId != "0" {
		updates["model_id"] = newModelId
	}

	if newZoneId != "" && newZoneId != "0" {
		updates["zone_id"] = newZoneId
	}

	if newOwnerId != "" && newOwnerId != "0" {
		updates["owner"] = newOwnerId
	}

	if isUnmanaged != nil {
		updates["is_unmanaged"] = *isUnmanaged
	}

	if len(updates) == 0 {
		return nil
	}

	err := r.db.Update(q.NewQuery(devicesCollection).Where(q.Field("_id").Eq(deviceId)), updates)
	if err != nil {
		return fmt.Errorf("UpdateDevice failed: %w", err)
	}
	return nil
}

// MigrateDeviceModel re-points a device to newModelID and remaps each of its
// device ports to a target model port of the new model per portMap
// (devicePortID -> newModelPortID). The device is the source of truth: its port
// data (MAC, VLAN configs, connections and interface links) is preserved — only
// the model-port pointers change, so no device information is destroyed.
func (r BasicOpsCloverRepository) MigrateDeviceModel(deviceID, newModelID string, portMap map[string]string) error {
	devDoc, err := r.db.FindById(devicesCollection, deviceID)
	if err != nil || devDoc == nil {
		return fmt.Errorf("device %q not found", deviceID)
	}
	modelDoc, err := r.db.FindById(modelsCollection, newModelID)
	if err != nil || modelDoc == nil {
		return fmt.Errorf("model %q not found", newModelID)
	}

	for devPortID, newMPID := range portMap {
		if newMPID == "" {
			return fmt.Errorf("device port %q has no target model port selected", devPortID)
		}
		dpDoc, err := r.db.FindById(deviceportsCollection, devPortID)
		if err != nil || dpDoc == nil {
			return fmt.Errorf("device port %q not found", devPortID)
		}
		if did, _ := dpDoc.Get("device_id").(string); did != deviceID {
			return fmt.Errorf("device port %q does not belong to device %q", devPortID, deviceID)
		}
		mpDoc, err := r.db.FindById(modelportsCollection, newMPID)
		if err != nil || mpDoc == nil {
			return fmt.Errorf("target model port %q not found", newMPID)
		}
		if mid, _ := mpDoc.Get("model_id").(string); mid != newModelID {
			return fmt.Errorf("target model port %q does not belong to the target model", newMPID)
		}
		oldMPID, _ := dpDoc.Get("model_port_id").(string)

		// Re-point the device port to the target model port.
		if err := r.db.Update(
			q.NewQuery(deviceportsCollection).Where(q.Field("_id").Eq(devPortID)),
			map[string]interface{}{"model_port_id": newMPID},
		); err != nil {
			return fmt.Errorf("update device port %q: %w", devPortID, err)
		}

		// Re-point any interface-port links for this device that referenced the old port.
		if oldMPID != "" && oldMPID != newMPID {
			if err := r.db.Update(
				q.NewQuery(interfacePortsCollection).Where(
					q.Field("device_id").Eq(deviceID).And(q.Field("model_port_id").Eq(oldMPID)),
				),
				map[string]interface{}{"model_port_id": newMPID},
			); err != nil {
				return fmt.Errorf("update interface ports for %q: %w", devPortID, err)
			}
		}
	}

	// Finally re-point the device to the new model.
	if err := r.db.Update(
		q.NewQuery(devicesCollection).Where(q.Field("_id").Eq(deviceID)),
		map[string]interface{}{"model_id": newModelID},
	); err != nil {
		return fmt.Errorf("update device model: %w", err)
	}
	return nil
}

// DeleteDevice deletes a device from the database by its ID
func (r BasicOpsCloverRepository) DeleteDevice(id string) error {
	// Check for dependent device ports
	devicePortExists, err := r.db.Exists(q.NewQuery(deviceportsCollection).Where(q.Field("device_id").Eq(id)))
	if err != nil {
		return err
	}
	if devicePortExists {
		return fmt.Errorf("cannot delete device: referenced by device ports; use --cascade to delete all dependents")
	}

	// Check for dependent connections
	fromConnExists, err := r.db.Exists(q.NewQuery(connectionsCollection).Where(q.Field("from_device_id").Eq(id)))
	if err != nil {
		return err
	}
	toConnExists, err := r.db.Exists(q.NewQuery(connectionsCollection).Where(q.Field("to_device_id").Eq(id)))
	if err != nil {
		return err
	}
	if fromConnExists || toConnExists {
		return fmt.Errorf("cannot delete device: referenced by connections; use --cascade to delete all dependents")
	}

	// Cascade-delete DeviceInterfaces (which in turn delete their InterfacePorts)
	ifaceDocs, err := r.db.FindAll(q.NewQuery(deviceInterfacesCollection).Where(q.Field("device_id").Eq(id)))
	if err != nil {
		return fmt.Errorf("DeleteDevice: failed to query device interfaces: %w", err)
	}
	for _, ifaceDoc := range ifaceDocs {
		ifaceID := ifaceDoc.ObjectId()
		if v, ok := ifaceDoc.Get("_id").(string); ok {
			ifaceID = v
		}
		if err := r.DeleteDeviceInterface(ifaceID); err != nil {
			return fmt.Errorf("DeleteDevice: failed to delete interface %s: %w", ifaceID, err)
		}
	}

	// Also delete any orphaned InterfacePort entries referencing this device
	if err := r.db.Delete(q.NewQuery(interfacePortsCollection).Where(q.Field("device_id").Eq(id))); err != nil {
		return fmt.Errorf("DeleteDevice: failed to delete interface ports for device %s: %w", id, err)
	}

	err = r.db.Delete(q.NewQuery(devicesCollection).Where(q.Field("_id").Eq(id)))
	if err != nil {
		return fmt.Errorf("DeleteDevice failed: %w", err)
	}
	return nil
}

// UpdateDeviceIPs updates the IP addresses for a device
func (r BasicOpsCloverRepository) UpdateDeviceIPs(deviceId string, ips []string) error {
	updates := make(map[string]interface{})
	if len(ips) > 0 {
		updates["ips"] = ips
	} else {
		// If empty, remove the field
		updates["ips"] = []string{}
	}

	err := r.db.Update(q.NewQuery(devicesCollection).Where(q.Field("_id").Eq(deviceId)), updates)
	if err != nil {
		return fmt.Errorf("UpdateDeviceIPs failed: %w", err)
	}
	return nil
}

// UpdateDeviceProfile sets (or clears, when profile == "") the scan-profile name
// associated with a device.
func (r BasicOpsCloverRepository) UpdateDeviceProfile(deviceId string, profile string) error {
	updates := map[string]interface{}{"profile": profile}
	if err := r.db.Update(q.NewQuery(devicesCollection).Where(q.Field("_id").Eq(deviceId)), updates); err != nil {
		return fmt.Errorf("UpdateDeviceProfile failed: %w", err)
	}
	return nil
}
