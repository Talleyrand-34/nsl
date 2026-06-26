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

		// Get model-type name if the model-type ID exists
		if mtID, ok := doc.Get("model_type_id").(string); ok && mtID != "" {
			mtDoc, err := r.db.FindById(modeltypesCollection, mtID)
			if err == nil && mtDoc != nil {
				model.ModelType = mtDoc.Get("name").(string)
			}
		}

		// Get OS-type name if the os-type ID exists
		if osID, ok := doc.Get("os_type_id").(string); ok && osID != "" {
			osDoc, err := r.db.FindById(ostypesCollection, osID)
			if err == nil && osDoc != nil {
				model.OsType = osDoc.Get("name").(string)
			}
		}

		result = append(result, model)
	}

	return result, nil
}

// AddModel creates new model entry resolving brand/model-type/os-type names to
// IDs. An unknown os-type name is auto-registered (so scans that detect a new OS
// family populate the catalogue); os-type is optional.
func (r BasicOpsCloverRepository) AddModel(
	modelName string,
	brandName string,
	modelTypeName string,
	osTypeName string,
) error {
	// Get brand ID from name (required)
	brandID, err := r.getBrandID(brandName)
	if err != nil {
		return fmt.Errorf("brand not found: %s", brandName)
	}

	// Get model-type ID from name (required)
	classID, err := r.getModelTypeID(modelTypeName)
	if err != nil {
		return fmt.Errorf("model type not found: %s", modelTypeName)
	}

	doc := d.NewDocument()
	doc.Set("model", modelName)
	doc.Set("brand", brandID)
	doc.Set("model_type_id", classID)
	doc.Set("os_type_id", r.ensureOsTypeID(osTypeName))

	_, err = r.db.InsertOne(modelsCollection, doc)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}
	return nil
}

// ensureOsTypeID resolves an os-type name to its id, creating the os-type if it
// doesn't exist yet. Empty name -> empty id.
func (r BasicOpsCloverRepository) ensureOsTypeID(osTypeName string) string {
	if osTypeName == "" {
		return ""
	}
	if id := r.getOsTypeID(osTypeName); id != "" {
		return id
	}
	_ = r.AddOsType(osTypeName)
	return r.getOsTypeID(osTypeName)
}

// UpdateModel updates a model in the database by its ID
func (r BasicOpsCloverRepository) UpdateModel(
	modelId string,
	newModelName string,
	newBrandId string,
	newModelTypeId string,
	newOsTypeId string,
) error {
	updates := make(map[string]interface{})
	updates["model"] = newModelName
	if newBrandId != "" {
		updates["brand"] = newBrandId
	}
	if newModelTypeId != "" {
		updates["model_type_id"] = newModelTypeId
	}
	if newOsTypeId != "" {
		updates["os_type_id"] = newOsTypeId
	}

	err := r.db.Update(q.NewQuery(modelsCollection).Where(q.Field("_id").Eq(modelId)), updates)
	if err != nil {
		return fmt.Errorf("UpdateModel failed: %w", err)
	}
	return nil
}

// DeleteModel deletes a model from the database by its ID
func (r BasicOpsCloverRepository) DeleteModel(id string) error {
	// Check for dependent devices
	deviceExists, err := r.db.Exists(q.NewQuery(devicesCollection).Where(q.Field("model_id").Eq(id)))
	if err != nil {
		return err
	}
	if deviceExists {
		return fmt.Errorf("cannot delete model: referenced by devices; use --cascade to delete all dependents")
	}

	// Check for dependent model ports
	modelPortExists, err := r.db.Exists(q.NewQuery(modelportsCollection).Where(q.Field("model_id").Eq(id)))
	if err != nil {
		return err
	}
	if modelPortExists {
		return fmt.Errorf("cannot delete model: referenced by model ports; use --cascade to delete all dependents")
	}

	err = r.db.Delete(q.NewQuery(modelsCollection).Where(q.Field("_id").Eq(id)))
	if err != nil {
		return fmt.Errorf("DeleteModel failed: %w", err)
	}
	return nil
}
