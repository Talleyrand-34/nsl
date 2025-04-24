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
func (r BasicOpsSQLiteRepository) GetConnections() []e.Connection {
	ctx := context.Background()
	models, execErr := r.query.GetConnections(ctx)
	if execErr != nil {
		return []e.Connection{}
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
	return result
}

// AddModel creates new model entry resolving brand/class names to IDs
func (r BasicOpsSQLiteRepository) AddConnection(
	fromDevice string,
	fromModelPort string,
	toDevice string,
	toModelPort string,
) error {
	ctx := context.Background()

	sfromDevice, err := strconv.Atoi(fromDevice)
	if err != nil {
		return nil
	}
	sfromModelPort, err := strconv.Atoi(fromModelPort)
	if err != nil {
		return nil
	}
	stoDevice, err := strconv.Atoi(toDevice)
	if err != nil {
		return nil
	}
	stoModelPort, err := strconv.Atoi(toModelPort)
	if err != nil {
		return nil
	}

	connStruct := d.AddConnectionParams{
		FromDevicePortDeviceID:    int64(sfromDevice),
		FromDevicePortModelPortID: int64(sfromModelPort),
		ToDevicePortDeviceID:      int64(stoDevice),
		ToDevicePortModelPortID:   int64(stoModelPort),
	}

	if err := r.query.AddConnection(ctx, connStruct); err != nil {
		return fmt.Errorf("failed to create connection: %v", err)
	}
	return nil
}
