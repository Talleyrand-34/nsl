
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
package specops

import (
	"context"
	"fmt"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db
)

// --- ID Resolution Helpers ---
// These return -1 if not found, or the found ID.

func (r *SpecOpsSQLiteRepository) getBrandID(ctx context.Context, name string) (int64, error) {
	id, err := r.query.GetBrandId(ctx, name)
	if err != nil {
		return -1, fmt.Errorf("brand '%s' not found: %w", name, err)
	}
	return id, nil
}

func (r *SpecOpsSQLiteRepository) getProprietaryID(
	ctx context.Context,
	name string,
) (int64, error) {
	id, err := r.query.GetProprietary(ctx, name)
	if err != nil {
		return -1, fmt.Errorf("proprietary '%s' not found: %w", name, err)
	}
	return id, nil
}

func (r *SpecOpsSQLiteRepository) getDeviceClassID(
	ctx context.Context,
	name string,
) (int64, error) {
	id, err := r.query.GetClassId(ctx, name)
	if err != nil {
		return -1, fmt.Errorf("device class '%s' not found: %w", name, err)
	}
	return id, nil
}

func (r *SpecOpsSQLiteRepository) getZoneTypeID(
	ctx context.Context,
	locationType string,
) (int64, error) {
	id, err := r.query.GetZoneType(ctx, locationType)
	if err != nil {
		return -1, fmt.Errorf("zone type '%s' not found: %w", locationType, err)
	}
	return id, nil
}

// func (r *SpecOpsSQLiteRepository) getZoneID(ctx context.Context, name string) (int64, error) {
// 	id, err := r.query.GetZoneId(ctx, name)
// 	if err != nil {
// 		return -1, fmt.Errorf("zone '%s' not found: %w", name, err)
// 	}
// 	return id, nil
// }
//
// func (r *SpecOpsSQLiteRepository) getModelID(ctx context.Context, model string) (int64, error) {
// 	id, err := r.query.GetModelId(ctx, model)
// 	if err != nil {
// 		return -1, fmt.Errorf("model '%s' not found: %w", model, err)
// 	}
// 	return id, nil
// }
