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

// AddModelType adds a new device class to the database
func (r BasicOpsCloverRepository) AddModelType(modelTypeName string) error {
	// Check if device class already exists
	exists, err := r.db.Exists(q.NewQuery(modeltypesCollection).Where(q.Field("name").Eq(modelTypeName)))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("device class %q already exists", modelTypeName)
	}

	doc := d.NewDocument()
	doc.Set("name", modelTypeName)

	_, err = r.db.InsertOne(modeltypesCollection, doc)
	return err
}

// GetModelTypes gets all the device classes available
func (r BasicOpsCloverRepository) GetModelTypes() ([]e.ModelType, error) {
	docs, err := r.db.FindAll(q.NewQuery(modeltypesCollection))
	if err != nil {
		return []e.ModelType{}, err
	}

	result := make([]e.ModelType, 0, len(docs))
	for _, doc := range docs {
		modelType := e.ModelType{
			ID:   doc.ObjectId(),
			Name: doc.Get("name").(string),
		}
		result = append(result, modelType)
	}

	return result, nil
}

// UpdateModelType updates a device class in the database by its ID
func (r BasicOpsCloverRepository) UpdateModelType(modelTypeId string, newModelTypeName string) error {
	updates := make(map[string]interface{})
	updates["name"] = newModelTypeName

	err := r.db.Update(q.NewQuery(modeltypesCollection).Where(q.Field("_id").Eq(modelTypeId)), updates)
	if err != nil {
		return err
	}

	return nil
}

// DeleteModelType deletes a device class from the database by its name
func (r BasicOpsCloverRepository) DeleteModelType(modelTypeName string) error {
	// Check for dependent models before deleting
	classDoc, err := r.db.FindFirst(q.NewQuery(modeltypesCollection).Where(q.Field("name").Eq(modelTypeName)))
	if err != nil {
		return err
	}
	if classDoc != nil {
		classID := classDoc.ObjectId()
		exists, err := r.db.Exists(q.NewQuery(modelsCollection).Where(q.Field("model_type_id").Eq(classID)))
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("cannot delete device class '%s': referenced by models; use --cascade to delete all dependents", modelTypeName)
		}
	}

	err = r.db.Delete(q.NewQuery(modeltypesCollection).Where(q.Field("name").Eq(modelTypeName)))
	if err != nil {
		return err
	}

	return nil
}
