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

// AddDeviceClass adds a new device class to the database
func (r BasicOpsCloverRepository) AddDeviceClass(devClassName string) error {
	// Check if device class already exists
	exists, err := r.db.Exists(q.NewQuery(devclassesCollection).Where(q.Field("name").Eq(devClassName)))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("device class %q already exists", devClassName)
	}

	doc := d.NewDocument()
	doc.Set("name", devClassName)

	_, err = r.db.InsertOne(devclassesCollection, doc)
	return err
}

// GetDeviceClasses gets all the device classes available
func (r BasicOpsCloverRepository) GetDeviceClasses() ([]e.DevClass, error) {
	docs, err := r.db.FindAll(q.NewQuery(devclassesCollection))
	if err != nil {
		return []e.DevClass{}, err
	}

	result := make([]e.DevClass, 0, len(docs))
	for _, doc := range docs {
		devClass := e.DevClass{
			ID:   doc.ObjectId(),
			Name: doc.Get("name").(string),
		}
		result = append(result, devClass)
	}

	return result, nil
}

// UpdateDeviceClass updates a device class in the database by its ID
func (r BasicOpsCloverRepository) UpdateDeviceClass(deviceClassId string, newDeviceClassName string) error {
	updates := make(map[string]interface{})
	updates["name"] = newDeviceClassName

	err := r.db.Update(q.NewQuery(devclassesCollection).Where(q.Field("_id").Eq(deviceClassId)), updates)
	if err != nil {
		return err
	}

	return nil
}

// DeleteDeviceClass deletes a device class from the database by its name
func (r BasicOpsCloverRepository) DeleteDeviceClass(devClassName string) error {
	err := r.db.Delete(q.NewQuery(devclassesCollection).Where(q.Field("name").Eq(devClassName)))
	if err != nil {
		return err
	}

	return nil
}
