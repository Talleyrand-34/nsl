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
	"context"
	"strconv"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
	d "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
)

// AddDeviceClass adds a new brand to the database
func (r BasicOpsSQLiteRepository) AddDeviceClass(devClassName string) error {
	ctx := context.Background()
	execErr := r.query.AddDeviceClass(ctx, devClassName)
	if execErr != nil {
		return execErr
	}
	return nil
}

// GetDeviceClasses gets all the device classes available
func (r BasicOpsSQLiteRepository) GetDeviceClasses() ([]e.DevClass, error) {
	ctx := context.Background()
	devclasses, execErr := r.query.GetDeviceClasses(ctx)
	if execErr != nil {
		return []e.DevClass{}, execErr
	}
	result := make([]e.DevClass, 0, len(devclasses))
	for _, row := range devclasses {
		singleDevClass := e.DevClass{
			ID:   row.ID,
			Name: row.Name,
		}
		result = append(result, singleDevClass)
	}
	return result, nil
}

// UpdateDeviceClass updates a device class in the database by its ID
func (r BasicOpsSQLiteRepository) UpdateDeviceClass(deviceClassId string, newDeviceClassName string) error {
	ctx := context.Background()

	// Convert string ID to int64
	id, err := strconv.ParseInt(deviceClassId, 10, 64)
	if err != nil {
		return err
	}

	execErr := r.query.UpdateDeviceClass(ctx, d.UpdateDeviceClassParams{
		Name: newDeviceClassName,
		ID:   id,
	})
	if execErr != nil {
		return execErr
	}
	return nil
}

// DeleteDeviceClass deletes a device class from the database
func (r BasicOpsSQLiteRepository) DeleteDeviceClass(devClassName string) error {
	ctx := context.Background()
	execErr := r.query.DeleteDeviceClass(ctx, devClassName)
	if execErr != nil {
		return execErr
	}
	return nil
}
