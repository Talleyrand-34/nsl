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

// AddLocalVlan adds a new local VLAN mapping to the database
func (r BasicOpsCloverRepository) AddLocalVlan(vlanID string, deviceID string, vlanName string) error {
	// Check if local VLAN mapping already exists
	exists, err := r.db.Exists(q.NewQuery(localvlansCollection).Where(
		q.Field("vlan_id").Eq(vlanID).And(q.Field("device_id").Eq(deviceID))))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("local VLAN mapping for VLAN ID %q on device %q already exists", vlanID, deviceID)
	}

	doc := d.NewDocument()
	doc.Set("vlan_id", vlanID)
	doc.Set("device_id", deviceID)
	doc.Set("vlan_name", vlanName)

	_, err = r.db.InsertOne(localvlansCollection, doc)
	return err
}

// GetLocalVlans gets all local VLAN mappings
func (r BasicOpsCloverRepository) GetLocalVlans() ([]e.LocalVlan, error) {
	docs, err := r.db.FindAll(q.NewQuery(localvlansCollection))
	if err != nil {
		return []e.LocalVlan{}, err
	}

	result := make([]e.LocalVlan, 0, len(docs))
	for _, doc := range docs {
		vlanName := ""
		if name, ok := doc.Get("vlan_name").(string); ok {
			vlanName = name
		}
		localVlan := e.LocalVlan{
			ID:       doc.ObjectId(),
			VlanID:   doc.Get("vlan_id").(string),
			DeviceID: doc.Get("device_id").(string),
			VlanName: vlanName,
		}
		result = append(result, localVlan)
	}

	return result, nil
}

// GetLocalVlansByDevice gets all local VLAN mappings for a specific device
func (r BasicOpsCloverRepository) GetLocalVlansByDevice(deviceID string) ([]e.LocalVlan, error) {
	docs, err := r.db.FindAll(q.NewQuery(localvlansCollection).Where(q.Field("device_id").Eq(deviceID)))
	if err != nil {
		return []e.LocalVlan{}, err
	}

	result := make([]e.LocalVlan, 0, len(docs))
	for _, doc := range docs {
		vlanName := ""
		if name, ok := doc.Get("vlan_name").(string); ok {
			vlanName = name
		}
		localVlan := e.LocalVlan{
			ID:       doc.ObjectId(),
			VlanID:   doc.Get("vlan_id").(string),
			DeviceID: doc.Get("device_id").(string),
			VlanName: vlanName,
		}
		result = append(result, localVlan)
	}

	return result, nil
}

// GetLocalVlansByVlanID gets all local VLAN mappings for a specific VLAN ID
func (r BasicOpsCloverRepository) GetLocalVlansByVlanID(vlanID string) ([]e.LocalVlan, error) {
	docs, err := r.db.FindAll(q.NewQuery(localvlansCollection).Where(q.Field("vlan_id").Eq(vlanID)))
	if err != nil {
		return []e.LocalVlan{}, err
	}

	result := make([]e.LocalVlan, 0, len(docs))
	for _, doc := range docs {
		vlanName := ""
		if name, ok := doc.Get("vlan_name").(string); ok {
			vlanName = name
		}
		localVlan := e.LocalVlan{
			ID:       doc.ObjectId(),
			VlanID:   doc.Get("vlan_id").(string),
			DeviceID: doc.Get("device_id").(string),
			VlanName: vlanName,
		}
		result = append(result, localVlan)
	}

	return result, nil
}

// UpdateLocalVlan updates a local VLAN mapping in the database by its ID
func (r BasicOpsCloverRepository) UpdateLocalVlan(localVlanId string, newVlanID string, newDeviceID string, newVlanName string) error {
	updates := make(map[string]interface{})
	if newVlanID != "" {
		updates["vlan_id"] = newVlanID
	}
	if newDeviceID != "" {
		updates["device_id"] = newDeviceID
	}
	if newVlanName != "" {
		updates["vlan_name"] = newVlanName
	}

	err := r.db.Update(q.NewQuery(localvlansCollection).Where(q.Field("_id").Eq(localVlanId)), updates)
	if err != nil {
		return err
	}

	return nil
}

// UpdateLocalVlanByMapping updates a local VLAN mapping by VLAN ID and device ID
func (r BasicOpsCloverRepository) UpdateLocalVlanByMapping(vlanID string, deviceID string, newVlanName string) error {
	updates := map[string]interface{}{
		"vlan_name": newVlanName,
	}

	err := r.db.Update(q.NewQuery(localvlansCollection).Where(
		q.Field("vlan_id").Eq(vlanID).And(q.Field("device_id").Eq(deviceID))), updates)
	if err != nil {
		return err
	}

	return nil
}

// DeleteLocalVlan deletes a local VLAN mapping from the database by its ID
func (r BasicOpsCloverRepository) DeleteLocalVlan(localVlanId string) error {
	err := r.db.Delete(q.NewQuery(localvlansCollection).Where(q.Field("_id").Eq(localVlanId)))
	if err != nil {
		return err
	}

	return nil
}

// DeleteLocalVlansByDevice deletes all local VLAN mappings for a specific device
func (r BasicOpsCloverRepository) DeleteLocalVlansByDevice(deviceID string) error {
	err := r.db.Delete(q.NewQuery(localvlansCollection).Where(q.Field("device_id").Eq(deviceID)))
	if err != nil {
		return err
	}

	return nil
}

// DeleteLocalVlansByVlanID deletes all local VLAN mappings for a specific VLAN ID
func (r BasicOpsCloverRepository) DeleteLocalVlansByVlanID(vlanID string) error {
	err := r.db.Delete(q.NewQuery(localvlansCollection).Where(q.Field("vlan_id").Eq(vlanID)))
	if err != nil {
		return err
	}

	return nil
}