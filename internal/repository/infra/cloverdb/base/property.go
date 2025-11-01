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

// AddProprietary adds a new proprietary to the database
func (r BasicOpsCloverRepository) AddProprietary(proprietary string) error {
	// Check if proprietary already exists
	exists, err := r.db.Exists(q.NewQuery(proprietariesCollection).Where(q.Field("proprietary").Eq(proprietary)))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("proprietary %q already exists", proprietary)
	}

	doc := d.NewDocument()
	doc.Set("proprietary", proprietary)

	_, err = r.db.InsertOne(proprietariesCollection, doc)
	return err
}

// GetProperties gets all the proprietaries available
func (r BasicOpsCloverRepository) GetProperties() ([]e.Proprietary, error) {
	docs, err := r.db.FindAll(q.NewQuery(proprietariesCollection))
	if err != nil {
		return []e.Proprietary{}, err
	}

	result := make([]e.Proprietary, 0, len(docs))
	for _, doc := range docs {
		proprietary := e.Proprietary{
			ID:   doc.ObjectId(),
			Name: doc.Get("proprietary").(string),
		}
		result = append(result, proprietary)
	}

	return result, nil
}

// UpdateProprietary updates a proprietary entry in the database by its ID
func (r BasicOpsCloverRepository) UpdateProprietary(proprietaryId string, newProprietaryName string) error {
	updates := make(map[string]interface{})
	updates["proprietary"] = newProprietaryName

	err := r.db.Update(q.NewQuery(proprietariesCollection).Where(q.Field("_id").Eq(proprietaryId)), updates)
	if err != nil {
		return err
	}

	return nil
}

// DeleteProprietary deletes a proprietary entry from the database by its name
func (r BasicOpsCloverRepository) DeleteProprietary(proprietary string) error {
	err := r.db.Delete(q.NewQuery(proprietariesCollection).Where(q.Field("proprietary").Eq(proprietary)))
	if err != nil {
		return fmt.Errorf("DeleteProprietary failed: %w", err)
	}

	return nil
}
