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

// AddVlan adds a new VLAN to the database
func (r BasicOpsCloverRepository) AddVlan(vlanID string, vlanName string) error {
	// Check if VLAN ID already exists
	exists, err := r.db.Exists(q.NewQuery(vlansCollection).Where(q.Field("vlan_id").Eq(vlanID)))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("VLAN ID %q already exists", vlanID)
	}

	doc := d.NewDocument()
	doc.Set("vlan_id", vlanID)
	doc.Set("vlan_name", vlanName)

	_, err = r.db.InsertOne(vlansCollection, doc)
	return err
}

// GetVlans gets all the VLANs available
func (r BasicOpsCloverRepository) GetVlans() ([]e.Vlan, error) {
	docs, err := r.db.FindAll(q.NewQuery(vlansCollection))
	if err != nil {
		return []e.Vlan{}, err
	}

	result := make([]e.Vlan, 0, len(docs))
	for _, doc := range docs {
		vlanName := ""
		if name, ok := doc.Get("vlan_name").(string); ok {
			vlanName = name
		}
		vlan := e.Vlan{
			ID:       doc.ObjectId(),
			VlanID:   doc.Get("vlan_id").(string),
			VlanName: vlanName,
		}
		result = append(result, vlan)
	}

	return result, nil
}

// UpdateVlan updates a VLAN in the database by its ID
func (r BasicOpsCloverRepository) UpdateVlan(vlanId string, newVlanID string, newVlanName string) error {
	updates := make(map[string]interface{})
	if newVlanID != "" {
		updates["vlan_id"] = newVlanID
	}
	if newVlanName != "" {
		updates["vlan_name"] = newVlanName
	}

	err := r.db.Update(q.NewQuery(vlansCollection).Where(q.Field("_id").Eq(vlanId)), updates)
	if err != nil {
		return err
	}

	return nil
}

// DeleteVlan deletes a VLAN from the database by its ID
func (r BasicOpsCloverRepository) DeleteVlan(vlanId string) error {
	err := r.db.Delete(q.NewQuery(vlansCollection).Where(q.Field("_id").Eq(vlanId)))
	if err != nil {
		return err
	}

	return nil
}
