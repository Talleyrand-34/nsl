/*
Copyright © 2026 Talleyrand-34 (t34@t34.dev)
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

// AddOsType adds a new OS type (openwrt, opnsense, …) to the catalogue.
func (r BasicOpsCloverRepository) AddOsType(osTypeName string) error {
	exists, err := r.db.Exists(q.NewQuery(ostypesCollection).Where(q.Field("name").Eq(osTypeName)))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("OS type %q already exists", osTypeName)
	}
	doc := d.NewDocument()
	doc.Set("name", osTypeName)
	_, err = r.db.InsertOne(ostypesCollection, doc)
	return err
}

// GetOsTypes returns all OS types in the catalogue.
func (r BasicOpsCloverRepository) GetOsTypes() ([]e.OsType, error) {
	docs, err := r.db.FindAll(q.NewQuery(ostypesCollection))
	if err != nil {
		return []e.OsType{}, err
	}
	result := make([]e.OsType, 0, len(docs))
	for _, doc := range docs {
		result = append(result, e.OsType{ID: doc.ObjectId(), Name: doc.Get("name").(string)})
	}
	return result, nil
}

// UpdateOsType renames an OS type by its ID.
func (r BasicOpsCloverRepository) UpdateOsType(osTypeId string, newOsTypeName string) error {
	return r.db.Update(q.NewQuery(ostypesCollection).Where(q.Field("_id").Eq(osTypeId)),
		map[string]interface{}{"name": newOsTypeName})
}

// DeleteOsType deletes an OS type by name (refused if a model references it).
func (r BasicOpsCloverRepository) DeleteOsType(osTypeName string) error {
	doc, err := r.db.FindFirst(q.NewQuery(ostypesCollection).Where(q.Field("name").Eq(osTypeName)))
	if err != nil {
		return err
	}
	if doc != nil {
		exists, err := r.db.Exists(q.NewQuery(modelsCollection).Where(q.Field("os_type_id").Eq(doc.ObjectId())))
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("cannot delete OS type %q: referenced by models", osTypeName)
		}
	}
	return r.db.Delete(q.NewQuery(ostypesCollection).Where(q.Field("name").Eq(osTypeName)))
}

// getOsTypeID resolves an OS-type name to its document id ("" name -> "" id, ok).
func (r BasicOpsCloverRepository) getOsTypeID(osTypeName string) string {
	if osTypeName == "" {
		return ""
	}
	doc, err := r.db.FindFirst(q.NewQuery(ostypesCollection).Where(q.Field("name").Eq(osTypeName)))
	if err != nil || doc == nil {
		return ""
	}
	return doc.ObjectId()
}
