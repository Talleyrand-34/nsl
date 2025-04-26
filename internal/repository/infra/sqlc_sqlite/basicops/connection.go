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
