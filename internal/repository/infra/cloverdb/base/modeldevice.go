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

// GetModels returns all models with resolved brand/class names
func (r BasicOpsCloverRepository) GetModels() ([]e.ModelDevice, error) {
	docs, err := r.db.FindAll(q.NewQuery(modelsCollection))
	if err != nil {
		return []e.ModelDevice{}, err
	}

	result := make([]e.ModelDevice, 0, len(docs))
	for _, doc := range docs {
		model := e.ModelDevice{
			ID:    doc.ObjectId(),
			Model: doc.Get("model").(string),
		}

		// Get brand name if brand ID exists
		if brandID, ok := doc.Get("brand").(string); ok && brandID != "" {
			brandDoc, err := r.db.FindById(brandsCollection, brandID)
			if err == nil && brandDoc != nil {
				model.Brand = brandDoc.Get("brand").(string)
			}
		}

		// Get class name if class ID exists
		if classID, ok := doc.Get("class_id").(string); ok && classID != "" {
			classDoc, err := r.db.FindById(devclassesCollection, classID)
			if err == nil && classDoc != nil {
				model.Class = classDoc.Get("name").(string)
			}
		}

		result = append(result, model)
	}

	return result, nil
}

// AddModel creates new model entry resolving brand/class names to IDs
func (r BasicOpsCloverRepository) AddModel(
	modelName string,
	brandName string,
	className string,
) error {
	// Get brand ID from name (required)
	brandID, err := r.getBrandID(brandName)
	if err != nil {
		return fmt.Errorf("brand not found: %s", brandName)
	}

	// Get class ID from name (required)
	classID, err := r.getClassID(className)
	if err != nil {
		return fmt.Errorf("class not found: %s", className)
	}

	doc := d.NewDocument()
	doc.Set("model", modelName)
	doc.Set("brand", brandID)
	doc.Set("class_id", classID)

	_, err = r.db.InsertOne(modelsCollection, doc)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}
	return nil
}

// UpdateModel updates a model in the database by its ID
func (r BasicOpsCloverRepository) UpdateModel(
	modelId string,
	newModelName string,
	newBrandId string,
	newDeviceClassId string,
) error {
	updates := make(map[string]interface{})
	updates["model"] = newModelName
	if newBrandId != "" {
		updates["brand"] = newBrandId
	}
	if newDeviceClassId != "" {
		updates["class_id"] = newDeviceClassId
	}

	err := r.db.Update(q.NewQuery(modelsCollection).Where(q.Field("_id").Eq(modelId)), updates)
	if err != nil {
		return fmt.Errorf("UpdateModel failed: %w", err)
	}
	return nil
}

// DeleteModel deletes a model from the database by its ID
func (r BasicOpsCloverRepository) DeleteModel(id string) error {
	err := r.db.Delete(q.NewQuery(modelsCollection).Where(q.Field("_id").Eq(id)))
	if err != nil {
		return fmt.Errorf("DeleteModel failed: %w", err)
	}
	return nil
}
