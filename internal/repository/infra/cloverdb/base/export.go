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
	"strconv"

	"github.com/ostafen/clover/v2/document"
	q "github.com/ostafen/clover/v2/query"

	e "nsl-graph/internal/repository/entities"
)

// GetAllPortsAll returns all possible device ports (combination of devices and their model ports)
func (r BasicOpsCloverRepository) GetAllPortsAll() ([]e.DevicePort, error) {
	devices, err := r.GetDevices()
	if err != nil {
		return []e.DevicePort{}, err
	}

	result := make([]e.DevicePort, 0)
	for _, device := range devices {
		// Get the model for this device
		modelDoc, err := r.db.FindById(modelsCollection, device.Model)
		if err != nil {
			continue
		}

		modelID, ok := modelDoc.Get("_id").(string)
		if !ok {
			continue
		}

		// Get all ports for this model
		modelPortDocs, err := r.db.FindAll(q.NewQuery(modelportsCollection).Where(q.Field("model_id").Eq(modelID)))
		if err != nil {
			continue
		}

		// Create a device port entry for each model port
		for _, portDoc := range modelPortDocs {
			result = append(result, e.DevicePort{
				DeviceID: device.ID,
				ModelID:  portDoc.ObjectId(),
			})
		}
	}

	return result, nil
}

// GetAllPortsDevice returns all possible ports for a specific device
func (r BasicOpsCloverRepository) GetAllPortsDevice(deviceid string) ([]e.DevicePort, error) {
	// Get the device
	deviceDoc, err := r.db.FindById(devicesCollection, deviceid)
	if err != nil {
		return []e.DevicePort{}, err
	}

	modelID, ok := deviceDoc.Get("model_id").(string)
	if !ok {
		return []e.DevicePort{}, nil
	}

	// Get all ports for this model
	modelPortDocs, err := r.db.FindAll(q.NewQuery(modelportsCollection).Where(q.Field("model_id").Eq(modelID)))
	if err != nil {
		return []e.DevicePort{}, err
	}

	result := make([]e.DevicePort, 0, len(modelPortDocs))
	for _, portDoc := range modelPortDocs {
		result = append(result, e.DevicePort{
			DeviceID: deviceid,
			ModelID:  portDoc.ObjectId(),
		})
	}

	return result, nil
}

// ExportAllStructs exports all data from the database in a structured format
func (r BasicOpsCloverRepository) ExportAllStructs() (e.All, error) {
	var result e.All

	// Brands
	brandDocs, err := r.db.FindAll(q.NewQuery(brandsCollection))
	if err != nil {
		return result, err
	}
	for _, doc := range brandDocs {
		result.Brands = append(result.Brands, e.BasicBrand{
			ID:   doc.ObjectId(),
			Name: doc.Get("brand").(string),
		})
	}

	// ConnectionTypes
	connTypeDocs, err := r.db.FindAll(q.NewQuery(connectiontypesCollection))
	if err != nil {
		return result, err
	}
	for _, doc := range connTypeDocs {
		result.ConnectionTypes = append(result.ConnectionTypes, e.BasicConnectiontype{
			ID:             doc.ObjectId(),
			ConnectionType: doc.Get("connection_type").(string),
		})
	}

	// Connections
	connDocs, err := r.db.FindAll(q.NewQuery(connectionsCollection))
	if err != nil {
		return result, err
	}
	for _, doc := range connDocs {
		fromIPSegment := ""
		if ip, ok := doc.Get("from_ip_segment").(string); ok {
			fromIPSegment = ip
		}
		toIPSegment := ""
		if ip, ok := doc.Get("to_ip_segment").(string); ok {
			toIPSegment = ip
		}

		// Get VLAN IDs - array of VLAN database IDs
		vlanIDs := make([]string, 0)
		if vlanIDsInterface, ok := doc.Get("vlan_ids").([]interface{}); ok && len(vlanIDsInterface) > 0 {
			for _, vlanIDInterface := range vlanIDsInterface {
				if vlanDbID, ok := vlanIDInterface.(string); ok && vlanDbID != "" {
					// Look up the VLAN document by its database ID
					vlanDoc, err := r.db.FindById(vlansCollection, vlanDbID)
					if err == nil && vlanDoc != nil {
						// Get the actual VLAN ID (e.g., "100", "200")
						if vlanIDStr, ok := vlanDoc.Get("vlan_id").(string); ok && vlanIDStr != "" {
							vlanIDs = append(vlanIDs, vlanIDStr)
						}
					}
				}
			}
		}

		result.Connections = append(result.Connections, e.BasicConnection{
			ID:                        doc.ObjectId(),
			FromDevicePortDeviceID:    getStringField(doc, "from_device_id"),
			FromDevicePortModelPortID: getStringField(doc, "from_model_port_id"),
			FromIPSegment:             fromIPSegment,
			ToDevicePortDeviceID:      getStringField(doc, "to_device_id"),
			ToDevicePortModelPortID:   getStringField(doc, "to_model_port_id"),
			ToIPSegment:               toIPSegment,
			ConnectionType:            -1, // CloverDB doesn't track this
			VlanIDs:                   vlanIDs,
		})
	}

	// ModelTypes
	modelTypeDocs, err := r.db.FindAll(q.NewQuery(modeltypesCollection))
	if err != nil {
		return result, err
	}
	for _, doc := range modelTypeDocs {
		result.ModelTypes = append(result.ModelTypes, e.BasicModelType{
			ID:   doc.ObjectId(),
			Name: doc.Get("name").(string),
		})
	}

	// DevicePorts
	devPortDocs, err := r.db.FindAll(q.NewQuery(deviceportsCollection))
	if err != nil {
		return result, err
	}
	for _, doc := range devPortDocs {
		result.DevicePorts = append(result.DevicePorts, e.BasicDeviceport{
			DeviceID:    getStringField(doc, "device_id"),
			ModelPortID: getStringField(doc, "model_port_id"),
		})
	}

	// Devices
	deviceDocs, err := r.db.FindAll(q.NewQuery(devicesCollection))
	if err != nil {
		return result, err
	}
	for _, doc := range deviceDocs {
		ownerID := int64(-1)
		if propID, ok := doc.Get("owner").(string); ok && propID != "" {
			// Try to convert to int64 if needed, otherwise use -1
			if val, err := strconv.ParseInt(propID, 10, 64); err == nil {
				ownerID = val
			}
		}

		result.Devices = append(result.Devices, e.BasicDevice{
			ID:      doc.ObjectId(),
			Label:   doc.Get("label").(string),
			ModelID: getStringField(doc, "model_id"),
			ZoneID:  getStringField(doc, "zone_id"),
			Owner:   ownerID,
		})
	}

	// ModelDevices
	modelDocs, err := r.db.FindAll(q.NewQuery(modelsCollection))
	if err != nil {
		return result, err
	}
	for _, doc := range modelDocs {
		brandID := int64(-1)
		if bID, ok := doc.Get("brand").(string); ok && bID != "" {
			if val, err := strconv.ParseInt(bID, 10, 64); err == nil {
				brandID = val
			}
		}

		result.ModelDevices = append(result.ModelDevices, e.BasicModeldevice{
			ID:      doc.ObjectId(),
			Model:   doc.Get("model").(string),
			Brand:   brandID,
			ModelTypeID: getStringField(doc, "model_type_id"),
		})
	}

	// ModelPorts
	modelPortDocs, err := r.db.FindAll(q.NewQuery(modelportsCollection))
	if err != nil {
		return result, err
	}
	for _, doc := range modelPortDocs {
		posX := int64(0)
		posY := int64(0)
		if px, ok := doc.Get("position_x").(float64); ok {
			posX = int64(px)
		} else if px, ok := doc.Get("position_x").(int); ok {
			posX = int64(px)
		}
		if py, ok := doc.Get("position_y").(float64); ok {
			posY = int64(py)
		} else if py, ok := doc.Get("position_y").(int); ok {
			posY = int64(py)
		}

		result.ModelPorts = append(result.ModelPorts, e.BasicModelport{
			ID:        doc.ObjectId(),
			Name:      doc.Get("name").(string),
			Positionx: posX,
			Positiony: posY,
			ModelID:   getStringField(doc, "model_id"),
		})
	}

	// Policies (not implemented in CloverDB, return empty)
	result.Policies = []e.BasicPolicy{}

	// Owners
	propDocs, err := r.db.FindAll(q.NewQuery(ownersCollection))
	if err != nil {
		return result, err
	}
	for _, doc := range propDocs {
		result.Owners = append(result.Owners, e.BasicOwner{
			ID:    doc.ObjectId(),
			Owner: doc.Get("owner").(string),
		})
	}

	// ZoneTypes
	zoneTypeDocs, err := r.db.FindAll(q.NewQuery(zonetypesCollection))
	if err != nil {
		return result, err
	}
	for _, doc := range zoneTypeDocs {
		result.ZoneTypes = append(result.ZoneTypes, e.BasicZonetype{
			ID:           doc.ObjectId(),
			LocationType: doc.Get("location_type").(string),
		})
	}

	// Zones
	zoneDocs, err := r.db.FindAll(q.NewQuery(zonesCollection))
	if err != nil {
		return result, err
	}
	for _, doc := range zoneDocs {
		fatherID := int64(-1)
		if fID, ok := doc.Get("father").(string); ok && fID != "" {
			if val, err := strconv.ParseInt(fID, 10, 64); err == nil {
				fatherID = val
			}
		}

		ownerID := int64(-1)
		if propID, ok := doc.Get("owner").(string); ok && propID != "" {
			if val, err := strconv.ParseInt(propID, 10, 64); err == nil {
				ownerID = val
			}
		}

		locationTypeID := int64(-1)
		if ltID, ok := doc.Get("location_type").(string); ok && ltID != "" {
			if val, err := strconv.ParseInt(ltID, 10, 64); err == nil {
				locationTypeID = val
			}
		}

		result.Zones = append(result.Zones, e.BasicZone{
			ID:           doc.ObjectId(),
			Name:         doc.Get("name").(string),
			Father:       fatherID,
			Granularity:  -1, // Not tracked in CloverDB
			Owner:        ownerID,
			LocationType: locationTypeID,
		})
	}

	// VLANs
	vlanDocs, err := r.db.FindAll(q.NewQuery(vlansCollection))
	if err != nil {
		return result, err
	}
	for _, doc := range vlanDocs {
		vlanName := ""
		if name, ok := doc.Get("vlan_name").(string); ok {
			vlanName = name
		}
		result.Vlans = append(result.Vlans, e.BasicVlan{
			ID:       doc.ObjectId(),
			VlanID:   doc.Get("vlan_id").(string),
			VlanName: vlanName,
		})
	}

	return result, nil
}

// Helper function to safely get string fields
func getStringField(doc *document.Document, field string) string {
	if val, ok := doc.Get(field).(string); ok {
		return val
	}
	return ""
}
