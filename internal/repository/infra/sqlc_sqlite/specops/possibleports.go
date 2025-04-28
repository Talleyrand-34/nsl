
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
