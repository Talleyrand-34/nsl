package basicops

import (
	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
)

func (r SQLiteRepository) GetBrands() []e.DevicePort {
}
