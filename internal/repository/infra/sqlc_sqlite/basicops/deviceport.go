package basicops

import (
	"context"
	"strconv"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
	d "nsl-graph/internal/repository/infra/sqlc_sqlite/basicops/internal_sqlc_sqlite"
)

// AddBrand adds a new brand to the database
func (r SQLiteRepository) AddDevicePort(deviceid string, modelportid string) error {
	ctx := context.Background()
	sdeviceid, err := strconv.Atoi(deviceid)
	if err != nil {
		return nil
	}
	smodelportid, err := strconv.Atoi(modelportid)
	if err != nil {
		return nil
	}
	devportstruct := d.AddDevicePortParams{
		DeviceID:    int64(sdeviceid),
		ModelPortID: int64(smodelportid),
	}
	execErr := r.query.AddDevicePort(ctx, devportstruct)
	if execErr != nil {
		return execErr
	}
	return nil
}

// GetBrands gets all the brands available
func (r SQLiteRepository) GetDevicePorts() []e.DevicePort {
	ctx := context.Background()
	devports, execErr := r.query.GetDevicePorts(ctx)
	if execErr != nil {
		return []e.DevicePort{}
	}
	result := make([]e.DevicePort, 0, len(devports))
	for _, row := range devports {
		model := e.DevicePort{
			DeviceID:  int(row.DeviceID),
			ModelID:   int(row.ModelPortID),
			DevLabel:  row.Label.String,
			PortName:  row.Name.String,
			Positionx: int(row.Positionx.Int64),
			Positiony: int(row.Positiony.Int64),
		}
		result = append(result, model)
	}
	return result
}
