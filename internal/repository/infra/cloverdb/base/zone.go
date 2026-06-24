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

// AddZone adds a new zone to the database
func (r BasicOpsCloverRepository) AddZone(
	name string,
	fatherid string,
	father string,
	proprietary string,
	zonename string,
) error {
	// Get father ID (handles all cases)
	sfatherid := r.getFatherID(fatherid, father)

	// Get proprietary ID. Strict: a named proprietary must exist (never silently
	// drop an unresolved dependency).
	spropid := r.getProprietaryID(proprietary)
	if proprietary != "" && spropid == "" {
		return fmt.Errorf("proprietary %q does not exist", proprietary)
	}

	// Get zone type ID. Strict: a named zone type must exist.
	szonetypeid := r.getZoneTypeID(zonename)
	if zonename != "" && szonetypeid == "" {
		return fmt.Errorf("zone type %q does not exist", zonename)
	}

	doc := d.NewDocument()
	doc.Set("name", name)
	if sfatherid != "" {
		doc.Set("father", sfatherid)
	}
	if szonetypeid != "" {
		doc.Set("location_type", szonetypeid)
	}
	if spropid != "" {
		doc.Set("proprietary", spropid)
	}

	_, err := r.db.InsertOne(zonesCollection, doc)
	if err != nil {
		return fmt.Errorf("AddZone failed: %w", err)
	}
	return nil
}

// GetZones gets all the zones available
func (r BasicOpsCloverRepository) GetZones() ([]e.Zone, error) {
	docs, err := r.db.FindAll(q.NewQuery(zonesCollection))
	if err != nil {
		return []e.Zone{}, err
	}

	result := make([]e.Zone, 0, len(docs))
	for _, doc := range docs {
		zone := e.Zone{
			ID:   doc.ObjectId(),
			Name: doc.Get("name").(string),
		}

		// Get father name if father ID exists
		if fatherID, ok := doc.Get("father").(string); ok && fatherID != "" {
			fatherDoc, err := r.db.FindById(zonesCollection, fatherID)
			if err == nil && fatherDoc != nil {
				zone.Father = fatherDoc.Get("name").(string)
				zone.FatherID = fatherID
			}
		}

		// Get location type name if location type ID exists
		if locationTypeID, ok := doc.Get("location_type").(string); ok && locationTypeID != "" {
			locationTypeDoc, err := r.db.FindById(zonetypesCollection, locationTypeID)
			if err == nil && locationTypeDoc != nil {
				zone.LocationType = locationTypeDoc.Get("location_type").(string)
			}
		}

		// Get proprietary name if proprietary ID exists
		if proprietaryID, ok := doc.Get("proprietary").(string); ok && proprietaryID != "" {
			proprietaryDoc, err := r.db.FindById(proprietariesCollection, proprietaryID)
			if err == nil && proprietaryDoc != nil {
				zone.Proprietary = proprietaryDoc.Get("proprietary").(string)
			}
		}

		result = append(result, zone)
	}

	return result, nil
}

// UpdateZone updates a zone in the database by its ID
func (r BasicOpsCloverRepository) UpdateZone(
	zoneId string,
	newZoneName string,
	newFatherZoneId string,
	newZoneTypeId string,
	newProprietaryId string,
) error {
	updates := make(map[string]interface{})
	updates["name"] = newZoneName

	if newFatherZoneId != "" && newFatherZoneId != "0" {
		updates["father"] = newFatherZoneId
	}

	if newZoneTypeId != "" && newZoneTypeId != "0" {
		updates["location_type"] = newZoneTypeId
	}

	if newProprietaryId != "" && newProprietaryId != "0" {
		updates["proprietary"] = newProprietaryId
	}

	err := r.db.Update(q.NewQuery(zonesCollection).Where(q.Field("_id").Eq(zoneId)), updates)
	if err != nil {
		return fmt.Errorf("UpdateZone failed: %w", err)
	}
	return nil
}

// DeleteZone deletes a zone from the database by its ID
func (r BasicOpsCloverRepository) DeleteZone(id string) error {
	// Check for child zones
	childZoneExists, err := r.db.Exists(q.NewQuery(zonesCollection).Where(q.Field("father").Eq(id)))
	if err != nil {
		return err
	}
	if childZoneExists {
		return fmt.Errorf("cannot delete zone: has child zones; use --cascade to delete all dependents")
	}

	// Check for devices in this zone
	deviceExists, err := r.db.Exists(q.NewQuery(devicesCollection).Where(q.Field("zone_id").Eq(id)))
	if err != nil {
		return err
	}
	if deviceExists {
		return fmt.Errorf("cannot delete zone: has devices; use --cascade to delete all dependents")
	}

	err = r.db.Delete(q.NewQuery(zonesCollection).Where(q.Field("_id").Eq(id)))
	if err != nil {
		return fmt.Errorf("DeleteZone failed: %w", err)
	}
	return nil
}
