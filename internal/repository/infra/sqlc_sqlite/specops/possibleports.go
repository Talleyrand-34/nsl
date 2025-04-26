package specops

import (
	"context"
	"strconv"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
)

func (r SpecOpsSQLiteRepository) GetAllPortsAll() ([]e.DevicePort, error) {
	ctx := context.Background()
	devports, execErr := r.query.GetPossiblePortsAll(ctx)
	if execErr != nil {
		return []e.DevicePort{}, execErr
	}
	result := make([]e.DevicePort, 0, len(devports))
	for _, row := range devports {
		model := e.DevicePort{
			DeviceID: int(row.Deviceid),
			ModelID:  int(row.Modelid),
		}
		result = append(result, model)
	}
	return result, nil
}

func (r SpecOpsSQLiteRepository) GetAllPortsDevice(deviceid string) ([]e.DevicePort, error) {
	id, err := strconv.ParseInt(deviceid, 10, 64)
	if err != nil {
		return []e.DevicePort{}, err
	}
	ctx := context.Background()
	devports, execErr := r.query.GetPossiblePortsDevice(ctx, id)
	if execErr != nil {
		return []e.DevicePort{}, execErr
	}
	result := make([]e.DevicePort, 0, len(devports))
	for _, row := range devports {
		model := e.DevicePort{
			DeviceID: int(row.Deviceid),
			ModelID:  int(row.Modelid),
		}
		result = append(result, model)
	}
	return result, nil
}
