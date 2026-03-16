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

// AddDevice adds a new device to the database
func (r BasicOpsCloverRepository) AddDevice(
	label string,
	model string,
	zoneId string,
	zoneName string,
	proprietary string,
	ips []string,
) error {
	// Get proprietary ID
	spropid := r.getProprietaryID(proprietary)

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
	if spropid != "" {
		doc.Set("proprietary", spropid)
	}
	if szoneid != "" {
		doc.Set("zone_id", szoneid)
	}
	if len(ips) > 0 {
		doc.Set("ips", ips)
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

		// Get proprietary name if proprietary ID exists
		if proprietaryID, ok := doc.Get("proprietary").(string); ok && proprietaryID != "" {
			proprietaryDoc, err := r.db.FindById(proprietariesCollection, proprietaryID)
			if err == nil && proprietaryDoc != nil {
				device.Proprietary = proprietaryDoc.Get("proprietary").(string)
			}
		}

		// Get IPs if they exist
		device.IPs = make([]string, 0)
		if ips, ok := doc.Get("ips").([]interface{}); ok && len(ips) > 0 {
			for _, ipInterface := range ips {
				if ip, ok := ipInterface.(string); ok && ip != "" {
					device.IPs = append(device.IPs, ip)
				}
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
	newProprietaryId string,
) error {
	updates := make(map[string]interface{})
	updates["label"] = newDeviceLabel

	if newModelId != "" && newModelId != "0" {
		updates["model_id"] = newModelId
	}

	if newZoneId != "" && newZoneId != "0" {
		updates["zone_id"] = newZoneId
	}

	if newProprietaryId != "" && newProprietaryId != "0" {
		updates["proprietary"] = newProprietaryId
	}

	err := r.db.Update(q.NewQuery(devicesCollection).Where(q.Field("_id").Eq(deviceId)), updates)
	if err != nil {
		return fmt.Errorf("UpdateDevice failed: %w", err)
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
