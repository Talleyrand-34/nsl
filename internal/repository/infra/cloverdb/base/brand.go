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

// AddBrand adds a new brand to the database
func (r BasicOpsCloverRepository) AddBrand(brand string) error {
	// Check if brand already exists
	exists, err := r.db.Exists(q.NewQuery(brandsCollection).Where(q.Field("brand").Eq(brand)))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("brand %q already exists", brand)
	}

	doc := d.NewDocument()
	doc.Set("brand", brand)

	_, err = r.db.InsertOne(brandsCollection, doc)
	return err
}

// GetBrands gets all the brands available
func (r BasicOpsCloverRepository) GetBrands() ([]e.Brand, error) {
	docs, err := r.db.FindAll(q.NewQuery(brandsCollection))
	if err != nil {
		return []e.Brand{}, err
	}

	result := make([]e.Brand, 0, len(docs))
	for _, doc := range docs {
		brand := e.Brand{
			ID:   doc.ObjectId(),
			Name: doc.Get("brand").(string),
		}
		result = append(result, brand)
	}

	return result, nil
}

// UpdateBrand updates a brand in the database by its ID
func (r BasicOpsCloverRepository) UpdateBrand(brandId string, newBrandName string) error {
	updates := make(map[string]interface{})
	updates["brand"] = newBrandName

	err := r.db.Update(q.NewQuery(brandsCollection).Where(q.Field("_id").Eq(brandId)), updates)
	if err != nil {
		return err
	}

	return nil
}

// DeleteBrand deletes a brand from the database by its name
func (r BasicOpsCloverRepository) DeleteBrand(brand string) error {
	// Check for dependent models before deleting
	brandDoc, err := r.db.FindFirst(q.NewQuery(brandsCollection).Where(q.Field("brand").Eq(brand)))
	if err != nil {
		return err
	}
	if brandDoc != nil {
		brandID := brandDoc.ObjectId()
		exists, err := r.db.Exists(q.NewQuery(modelsCollection).Where(q.Field("brand").Eq(brandID)))
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("cannot delete brand '%s': referenced by models; use --cascade to delete all dependents", brand)
		}
	}

	err = r.db.Delete(q.NewQuery(brandsCollection).Where(q.Field("brand").Eq(brand)))
	if err != nil {
		return err
	}

	return nil
}
