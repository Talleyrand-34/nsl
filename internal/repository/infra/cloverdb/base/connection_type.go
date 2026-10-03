// SPDX-License-Identifier: AGPL-3.0-or-later
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

// AddConnectionType adds a new connection type to the database
func (r BasicOpsCloverRepository) AddConnectionType(connectiontypes string) error {
	// Check if connection type already exists
	exists, err := r.db.Exists(q.NewQuery(connectiontypesCollection).Where(q.Field("connection_type").Eq(connectiontypes)))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("connection type %q already exists", connectiontypes)
	}

	doc := d.NewDocument()
	doc.Set("connection_type", connectiontypes)

	_, err = r.db.InsertOne(connectiontypesCollection, doc)
	return err
}

// GetConnectionTypes gets all the connection types available
func (r BasicOpsCloverRepository) GetConnectionTypes() ([]e.ConnectionType, error) {
	docs, err := r.db.FindAll(q.NewQuery(connectiontypesCollection))
	if err != nil {
		return []e.ConnectionType{}, err
	}

	result := make([]e.ConnectionType, 0, len(docs))
	for _, doc := range docs {
		connectionType := e.ConnectionType{
			ID:   doc.ObjectId(),
			Name: doc.Get("connection_type").(string),
		}
		result = append(result, connectionType)
	}

	return result, nil
}

// UpdateConnectionType updates a connection type in the database by its ID
func (r BasicOpsCloverRepository) UpdateConnectionType(connectionTypeId string, newConnectionTypeName string) error {
	updates := make(map[string]interface{})
	updates["connection_type"] = newConnectionTypeName

	err := r.db.Update(q.NewQuery(connectiontypesCollection).Where(q.Field("_id").Eq(connectionTypeId)), updates)
	if err != nil {
		return err
	}

	return nil
}

// DeleteConnectionType deletes a connection type from the database by its name
func (r BasicOpsCloverRepository) DeleteConnectionType(connectionTypeName string) error {
	err := r.db.Delete(q.NewQuery(connectiontypesCollection).Where(q.Field("connection_type").Eq(connectionTypeName)))
	if err != nil {
		return err
	}

	return nil
}
