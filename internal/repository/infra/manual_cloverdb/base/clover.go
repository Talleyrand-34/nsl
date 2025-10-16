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
	c "github.com/ostafen/clover/v2"
)

const brandsCollection = "brands"

type BasicOpsCloverRepository struct {
	db *c.DB
}

// NewCloverRepositoryFromDB creates a repository from an existing CloverDB instance
func NewCloverRepositoryFromDB(db *c.DB) (BasicOpsCloverRepository, error) {
	// Create the brands collection if it doesn't exist
	if exists, err := db.HasCollection(brandsCollection); err != nil {
		return BasicOpsCloverRepository{}, err
	} else if !exists {
		if err := db.CreateCollection(brandsCollection); err != nil {
			return BasicOpsCloverRepository{}, err
		}
	}

	return BasicOpsCloverRepository{db: db}, nil
}

// NewCloverRepository creates a repository with a new CloverDB instance at the specified path
func NewCloverRepository(dirPath string) (BasicOpsCloverRepository, error) {
	db, err := c.Open(dirPath)
	if err != nil {
		return BasicOpsCloverRepository{}, err
	}

	return NewCloverRepositoryFromDB(db)
}

// Close closes the database connection
func (r BasicOpsCloverRepository) Close() error {
	return r.db.Close()
}
