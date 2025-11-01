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

// AddZoneType adds a new zone type to the database
func (r BasicOpsCloverRepository) AddZoneType(zoneName string) error {
	// Check if zone type already exists
	exists, err := r.db.Exists(q.NewQuery(zonetypesCollection).Where(q.Field("location_type").Eq(zoneName)))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("zone type %q already exists", zoneName)
	}

	doc := d.NewDocument()
	doc.Set("location_type", zoneName)

	_, err = r.db.InsertOne(zonetypesCollection, doc)
	return err
}

// GetZonetypes gets all the zone types available
func (r BasicOpsCloverRepository) GetZonetypes() ([]e.ZoneType, error) {
	docs, err := r.db.FindAll(q.NewQuery(zonetypesCollection))
	if err != nil {
		return []e.ZoneType{}, err
	}

	result := make([]e.ZoneType, 0, len(docs))
	for _, doc := range docs {
		zonetype := e.ZoneType{
			ID:   doc.ObjectId(),
			Name: doc.Get("location_type").(string),
		}
		result = append(result, zonetype)
	}

	return result, nil
}

// UpdateZoneType updates a zone type in the database by its ID
func (r BasicOpsCloverRepository) UpdateZoneType(zoneTypeId string, newZoneTypeName string) error {
	updates := make(map[string]interface{})
	updates["location_type"] = newZoneTypeName

	err := r.db.Update(q.NewQuery(zonetypesCollection).Where(q.Field("_id").Eq(zoneTypeId)), updates)
	if err != nil {
		return err
	}

	return nil
}

// DeleteZoneType deletes a zone type from the database by its name
func (r BasicOpsCloverRepository) DeleteZoneType(zonetype string) error {
	err := r.db.Delete(q.NewQuery(zonetypesCollection).Where(q.Field("location_type").Eq(zonetype)))
	if err != nil {
		return fmt.Errorf("DeleteZoneType failed: %w", err)
	}

	return nil
}
