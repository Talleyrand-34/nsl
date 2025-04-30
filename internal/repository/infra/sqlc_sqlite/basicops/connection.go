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
	"fmt"
	"strconv"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
	d "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
)

// GetModels returns all models with resolved brand/class names
func (r BasicOpsSQLiteRepository) GetConnections() ([]e.Connection, error) {
	ctx := context.Background()
	models, execErr := r.query.GetConnections(ctx)
	if execErr != nil {
		return []e.Connection{}, execErr
	}

	result := make([]e.Connection, 0, len(models))
	for _, row := range models {
		model := e.Connection{
			ID:            int(row.ID),
			FromDevice:    row.Fromdevname.String,
			FromModelPort: row.Frommodelportname.String,
			ToDevice:      row.Todevname.String,
			ToModelPort:   row.Tomodelportname.String,
			FromZoneID:    int(row.Fromzoneid.Int64),
			FromZoneName:  row.Fromzonename.String,
			ToZoneID:      int(row.Fromzoneid.Int64),
			ToZoneName:    row.Tozonename.String,
		}
		result = append(result, model)
	}
	return result, nil
}

// AddModel creates new model entry resolving brand/class names to IDs
func (r BasicOpsSQLiteRepository) AddConnection(
	fromDevice string,
	fromModelPort string,
	toDevice string,
	toModelPort string,
) error {
	ctx := context.Background()

	// Convert string arguments to integers
	sfromDevice, err := strconv.Atoi(fromDevice)
	if err != nil {
		return fmt.Errorf("invalid fromDevice: %v", err)
	}
	sfromModelPort, err := strconv.Atoi(fromModelPort)
	if err != nil {
		return fmt.Errorf("invalid fromModelPort: %v", err)
	}
	stoDevice, err := strconv.Atoi(toDevice)
	if err != nil {
		return fmt.Errorf("invalid toDevice: %v", err)
	}
	stoModelPort, err := strconv.Atoi(toModelPort)
	if err != nil {
		return fmt.Errorf("invalid toModelPort: %v", err)
	}

	err = validInputConnection(ctx, r, sfromDevice, sfromModelPort, stoDevice, stoModelPort)
	if err != nil {
		return err
	}

	// Prepare parameters for adding a new connection
	addConnParams := d.AddConnectionParams{
		FromDevicePortDeviceID:    int64(sfromDevice),
		FromDevicePortModelPortID: int64(sfromModelPort),
		ToDevicePortDeviceID:      int64(stoDevice),
		ToDevicePortModelPortID:   int64(stoModelPort),
	}

	// Add the new connection
	if err := r.query.AddConnection(ctx, addConnParams); err != nil {
		return fmt.Errorf("failed to create connection: %v", err)
	}
	return nil
}

func validInputConnection(
	ctx context.Context,
	r BasicOpsSQLiteRepository,
	sfromDevice int,
	sfromModelPort int,
	stoDevice int,
	stoModelPort int,
) error {
	// Forward direction
	checkConnParams := d.GetConnectionIdParams{
		FromDevicePortDeviceID:    int64(sfromDevice),
		FromDevicePortModelPortID: int64(sfromModelPort),
		ToDevicePortDeviceID:      int64(stoDevice),
		ToDevicePortModelPortID:   int64(stoModelPort),
	}
	// Reverse direction
	checkConnParamsReverse := d.GetConnectionIdParams{
		FromDevicePortDeviceID:    int64(stoDevice),
		FromDevicePortModelPortID: int64(stoModelPort),
		ToDevicePortDeviceID:      int64(sfromDevice),
		ToDevicePortModelPortID:   int64(sfromModelPort),
	}

	// Check if the connection already exists (forward)
	existing, err := r.query.GetConnectionId(ctx, checkConnParams)
	if err != nil {
		return fmt.Errorf("failed to check existing connection: %v", err)
	}
	if len(existing) > 0 {
		return fmt.Errorf("connection already exists")
	}
	// Check if the connection already exists (reverse)
	existing, err = r.query.GetConnectionId(ctx, checkConnParamsReverse)
	if err != nil {
		return fmt.Errorf("failed to check existing connection: %v", err)
	}
	if len(existing) > 0 {
		return fmt.Errorf("connection already exists")
	}
	return nil
}

// DeleteConnection deletes a zone from the database by its integer ID
func (r BasicOpsSQLiteRepository) DeleteConnection(id string) error {
	ctx := context.Background()

	intID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("DeleteConnection: invalid id '%s': %w", id, err)
	}

	if err := r.query.DeleteConnection(ctx, int64(intID)); err != nil {
		return fmt.Errorf("DeleteConnection failed: %w", err)
	}
	return nil
}

// DeleteConnection deletes a zone from the database by its integer ID
func (r BasicOpsSQLiteRepository) UpdateConnection(
	id string,
	from_device string,
	from_port string,
	to_device string,
	to_port string,
) error {
	ctx := context.Background()
	// Helper function to parse string to int64
	parseInt64 := func(s, field string) (int64, error) {
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid %s '%s': %w", field, s, err)
		}
		return val, nil
	}

	intID, err := parseInt64(id, "id")
	if err != nil {
		return fmt.Errorf("UpdateConnection: %w", err)
	}
	fromDeviceID, err := parseInt64(from_device, "from_device")
	if err != nil {
		return fmt.Errorf("UpdateConnection: %w", err)
	}
	fromPortID, err := parseInt64(from_port, "from_port")
	if err != nil {
		return fmt.Errorf("UpdateConnection: %w", err)
	}
	toDeviceID, err := parseInt64(to_device, "to_device")
	if err != nil {
		return fmt.Errorf("UpdateConnection: %w", err)
	}
	toPortID, err := parseInt64(to_port, "to_port")
	if err != nil {
		return fmt.Errorf("UpdateConnection: %w", err)
	}

	construct := d.UpdateConnectionParams{
		ID:                        intID,
		FromDevicePortDeviceID:    fromDeviceID,
		FromDevicePortModelPortID: fromPortID,
		ToDevicePortDeviceID:      toDeviceID,
		ToDevicePortModelPortID:   toPortID,
	}

	if err := r.query.UpdateConnection(ctx, construct); err != nil {
		return fmt.Errorf("DeleteConnection failed: %w", err)
	}
	return nil
}
