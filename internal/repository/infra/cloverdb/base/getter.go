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

	q "github.com/ostafen/clover/v2/query"
)

// Helper function to get father ID from zone ID or name
func (r BasicOpsCloverRepository) getFatherID(fatherid string, father string) string {
	// Prefer fatherid if provided
	if fatherid != "" {
		return fatherid
	}

	// Otherwise, try to get from father name
	if father != "" {
		doc, err := r.db.FindFirst(q.NewQuery(zonesCollection).Where(q.Field("name").Eq(father)))
		if err == nil && doc != nil {
			return doc.ObjectId()
		}
	}

	// Neither provided, return empty string
	return ""
}

// Helper function to get owner ID from owner name
func (r BasicOpsCloverRepository) getOwnerID(owner string) string {
	if owner == "" {
		return ""
	}

	doc, err := r.db.FindFirst(q.NewQuery(ownersCollection).Where(q.Field("owner").Eq(owner)))
	if err == nil && doc != nil {
		return doc.ObjectId()
	}

	return ""
}

// Helper function to get zone type ID from zone type name
func (r BasicOpsCloverRepository) getZoneTypeID(zonename string) string {
	if zonename == "" {
		return ""
	}

	doc, err := r.db.FindFirst(q.NewQuery(zonetypesCollection).Where(q.Field("location_type").Eq(zonename)))
	if err == nil && doc != nil {
		return doc.ObjectId()
	}

	return ""
}

// getConnectionTypeID returns the id of a connection type by name, or "" if it
// doesn't exist.
func (r BasicOpsCloverRepository) getConnectionTypeID(connectionType string) string {
	if connectionType == "" {
		return ""
	}

	doc, err := r.db.FindFirst(q.NewQuery(connectiontypesCollection).Where(q.Field("connection_type").Eq(connectionType)))
	if err == nil && doc != nil {
		return doc.ObjectId()
	}

	return ""
}

// Helper function to get zone ID from zone ID or name
func (r BasicOpsCloverRepository) getZoneID(zoneid string, zonename string) string {
	// Prefer zoneid if provided
	if zoneid != "" {
		return zoneid
	}

	if zonename == "" {
		return ""
	}

	doc, err := r.db.FindFirst(q.NewQuery(zonesCollection).Where(q.Field("name").Eq(zonename)))
	if err == nil && doc != nil {
		return doc.ObjectId()
	}

	return ""
}

// Helper function to get model ID from model name
func (r BasicOpsCloverRepository) getModelID(modelname string) (string, error) {
	if modelname == "" {
		return "", fmt.Errorf("model name is required")
	}

	doc, err := r.db.FindFirst(q.NewQuery(modelsCollection).Where(q.Field("model").Eq(modelname)))
	if err != nil {
		return "", fmt.Errorf("error getting model ID from model name: %v", err)
	}
	if doc == nil {
		return "", fmt.Errorf("model not found: %s", modelname)
	}

	return doc.ObjectId(), nil
}

// Helper function to get brand ID from brand name
func (r BasicOpsCloverRepository) getBrandID(brandName string) (string, error) {
	if brandName == "" {
		return "", fmt.Errorf("brand name is required")
	}

	doc, err := r.db.FindFirst(q.NewQuery(brandsCollection).Where(q.Field("brand").Eq(brandName)))
	if err != nil {
		return "", fmt.Errorf("brand not found: %s", brandName)
	}
	if doc == nil {
		return "", fmt.Errorf("brand not found: %s", brandName)
	}

	return doc.ObjectId(), nil
}

// Helper function to get device class ID from class name
func (r BasicOpsCloverRepository) getModelTypeID(modelTypeName string) (string, error) {
	if modelTypeName == "" {
		return "", fmt.Errorf("class name is required")
	}

	doc, err := r.db.FindFirst(q.NewQuery(modeltypesCollection).Where(q.Field("name").Eq(modelTypeName)))
	if err != nil {
		return "", fmt.Errorf("class not found: %s", modelTypeName)
	}
	if doc == nil {
		return "", fmt.Errorf("class not found: %s", modelTypeName)
	}

	return doc.ObjectId(), nil
}
