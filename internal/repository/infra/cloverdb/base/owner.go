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

// AddOwner adds a new owner to the database
func (r BasicOpsCloverRepository) AddOwner(owner string) error {
	// Check if owner already exists
	exists, err := r.db.Exists(q.NewQuery(ownersCollection).Where(q.Field("owner").Eq(owner)))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("owner %q already exists", owner)
	}

	doc := d.NewDocument()
	doc.Set("owner", owner)

	_, err = r.db.InsertOne(ownersCollection, doc)
	return err
}

// GetOwners gets all the owners available
func (r BasicOpsCloverRepository) GetOwners() ([]e.Owner, error) {
	docs, err := r.db.FindAll(q.NewQuery(ownersCollection))
	if err != nil {
		return []e.Owner{}, err
	}

	result := make([]e.Owner, 0, len(docs))
	for _, doc := range docs {
		owner := e.Owner{
			ID:   doc.ObjectId(),
			Name: doc.Get("owner").(string),
		}
		result = append(result, owner)
	}

	return result, nil
}

// UpdateOwner updates a owner entry in the database by its ID
func (r BasicOpsCloverRepository) UpdateOwner(ownerId string, newOwnerName string) error {
	updates := make(map[string]interface{})
	updates["owner"] = newOwnerName

	err := r.db.Update(q.NewQuery(ownersCollection).Where(q.Field("_id").Eq(ownerId)), updates)
	if err != nil {
		return err
	}

	return nil
}

// DeleteOwner deletes a owner entry from the database by its name.
// Before deleting, it sets the owner field to empty on all referencing zones and devices.
func (r BasicOpsCloverRepository) DeleteOwner(owner string) error {
	// Find the owner ID to clear references
	ownerDoc, err := r.db.FindFirst(q.NewQuery(ownersCollection).Where(q.Field("owner").Eq(owner)))
	if err != nil {
		return fmt.Errorf("DeleteOwner failed: %w", err)
	}
	if ownerDoc != nil {
		ownerID := ownerDoc.ObjectId()
		// Set null on zones referencing this owner
		if err := r.db.Update(
			q.NewQuery(zonesCollection).Where(q.Field("owner").Eq(ownerID)),
			map[string]interface{}{"owner": ""},
		); err != nil {
			return fmt.Errorf("DeleteOwner: failed to clear owner from zones: %w", err)
		}
		// Set null on devices referencing this owner
		if err := r.db.Update(
			q.NewQuery(devicesCollection).Where(q.Field("owner").Eq(ownerID)),
			map[string]interface{}{"owner": ""},
		); err != nil {
			return fmt.Errorf("DeleteOwner: failed to clear owner from devices: %w", err)
		}
	}

	err = r.db.Delete(q.NewQuery(ownersCollection).Where(q.Field("owner").Eq(owner)))
	if err != nil {
		return fmt.Errorf("DeleteOwner failed: %w", err)
	}

	return nil
}
