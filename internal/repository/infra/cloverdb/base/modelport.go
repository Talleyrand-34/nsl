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
	"strconv"

	d "github.com/ostafen/clover/v2/document"
	q "github.com/ostafen/clover/v2/query"

	e "nsl-graph/internal/repository/entities"
)

// GetModelPorts returns all model ports
func (r BasicOpsCloverRepository) GetModelPorts() ([]e.ModelPort, error) {
	docs, err := r.db.FindAll(q.NewQuery(modelportsCollection))
	if err != nil {
		return []e.ModelPort{}, err
	}

	result := make([]e.ModelPort, 0, len(docs))
	for _, doc := range docs {
		posX := 0
		if px, ok := doc.Get("position_x").(float64); ok {
			posX = int(px)
		} else if px, ok := doc.Get("position_x").(int); ok {
			posX = px
		}

		posY := 0
		if py, ok := doc.Get("position_y").(float64); ok {
			posY = int(py)
		} else if py, ok := doc.Get("position_y").(int); ok {
			posY = py
		}

		// Get allow_multiple_connections field, default to false
		allowMultiple := false
		if val, ok := doc.Get("allow_multiple_connections").(bool); ok {
			allowMultiple = val
		}

		modelPort := e.ModelPort{
			ID:                       doc.ObjectId(),
			Name:                     doc.Get("name").(string),
			Positionx:                posX,
			Positiony:                posY,
			AllowMultipleConnections: allowMultiple,
		}

		if pt, ok := doc.Get("port_type").(string); ok {
			modelPort.PortType = pt
		}
		if band, ok := doc.Get("band").(string); ok {
			modelPort.Band = band
		}

		// Get model name and brand if model ID exists
		if modelID, ok := doc.Get("model_id").(string); ok && modelID != "" {
			modelDoc, err := r.db.FindById(modelsCollection, modelID)
			if err == nil && modelDoc != nil {
				modelPort.Model = modelDoc.Get("model").(string)

				// Get brand name from model's brand ID
				if brandID, ok := modelDoc.Get("brand").(string); ok && brandID != "" {
					brandDoc, err := r.db.FindById(brandsCollection, brandID)
					if err == nil && brandDoc != nil {
						modelPort.Brand = brandDoc.Get("brand").(string)
					}
				}
			}
		}

		result = append(result, modelPort)
	}

	return result, nil
}

// AddModelPort creates new model port entry
func (r BasicOpsCloverRepository) AddModelPort(
	name string,
	posx string,
	posy string,
	modelName string,
	allowMultipleConnections bool,
	portType string,
	band string,
) error {
	// Get model ID from name (required)
	modelID, err := r.getModelID(modelName)
	if err != nil {
		return fmt.Errorf("model not found: %s", modelName)
	}

	iposx, err := strconv.Atoi(posx)
	if err != nil {
		return fmt.Errorf("invalid position x: %v", err)
	}

	iposy, err := strconv.Atoi(posy)
	if err != nil {
		return fmt.Errorf("invalid position y: %v", err)
	}

	doc := d.NewDocument()
	doc.Set("name", name)
	doc.Set("position_x", iposx)
	doc.Set("position_y", iposy)
	doc.Set("model_id", modelID)
	doc.Set("allow_multiple_connections", allowMultipleConnections)
	if portType != "" {
		doc.Set("port_type", portType)
	}
	if band != "" {
		doc.Set("band", band)
	}

	_, err = r.db.InsertOne(modelportsCollection, doc)
	if err != nil {
		return fmt.Errorf("failed to create modelport: %v", err)
	}
	return nil
}

// UpdateModelPort updates a model port in the database by its ID
func (r BasicOpsCloverRepository) UpdateModelPort(
	modelPortId string,
	newPortName string,
	newPositionX string,
	newPositionY string,
	newModelId string,
	newAllowMultipleConnections bool,
) error {
	posX, err := strconv.Atoi(newPositionX)
	if err != nil {
		return fmt.Errorf("invalid position X '%s': %w", newPositionX, err)
	}

	posY, err := strconv.Atoi(newPositionY)
	if err != nil {
		return fmt.Errorf("invalid position Y '%s': %w", newPositionY, err)
	}

	updates := make(map[string]interface{})
	updates["name"] = newPortName
	updates["position_x"] = posX
	updates["position_y"] = posY
	updates["allow_multiple_connections"] = newAllowMultipleConnections
	if newModelId != "" {
		updates["model_id"] = newModelId
	}

	err = r.db.Update(q.NewQuery(modelportsCollection).Where(q.Field("_id").Eq(modelPortId)), updates)
	if err != nil {
		return fmt.Errorf("UpdateModelPort failed: %w", err)
	}
	return nil
}

// DeleteModelPort deletes a model port from the database by its ID
func (r BasicOpsCloverRepository) DeleteModelPort(id string) error {
	// Check for dependent device ports
	exists, err := r.db.Exists(q.NewQuery(deviceportsCollection).Where(q.Field("model_port_id").Eq(id)))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("cannot delete model port: referenced by device ports; use --cascade to delete all dependents")
	}

	err = r.db.Delete(q.NewQuery(modelportsCollection).Where(q.Field("_id").Eq(id)))
	if err != nil {
		return fmt.Errorf("DeleteModelPort failed: %w", err)
	}
	return nil
}
