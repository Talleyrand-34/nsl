/*
Copyright © 2025 Talleyrand-34 (t34@t34.dev)
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
	"fmt"
	"log"
	"os"

	c "github.com/ostafen/clover/v2"
)

// Collection constants
const (
	brandsCollection          = "brands"
	devclassesCollection      = "devclasses"
	proprietariesCollection   = "proprietaries"
	zonetypesCollection       = "zonetypes"
	zonesCollection           = "zones"
	modelsCollection          = "models"
	modelportsCollection      = "modelports"
	devicesCollection         = "devices"
	deviceportsCollection     = "deviceports"
	connectionsCollection     = "connections"
	connectiontypesCollection = "connectiontypes"
	vlansCollection           = "vlans"
	scanProfilesCollection    = "scanprofiles"
)

type BasicOpsCloverRepository struct {
	db *c.DB
}

// NewCloverRepositoryFromDB creates a repository from an existing CloverDB instance
func NewCloverRepositoryFromDB(db *c.DB) (BasicOpsCloverRepository, error) {
	// List of all collections to create
	collections := []string{
		brandsCollection,
		devclassesCollection,
		proprietariesCollection,
		zonetypesCollection,
		zonesCollection,
		modelsCollection,
		modelportsCollection,
		devicesCollection,
		deviceportsCollection,
		connectionsCollection,
		connectiontypesCollection,
		vlansCollection,
		deviceInterfacesCollection,
		interfacePortsCollection,
		scanProfilesCollection,
	}

	// Create each collection if it doesn't exist
	for _, collectionName := range collections {
		if exists, err := db.HasCollection(collectionName); err != nil {
			return BasicOpsCloverRepository{}, err
		} else if !exists {
			if err := db.CreateCollection(collectionName); err != nil {
				return BasicOpsCloverRepository{}, err
			}
		}
	}

	return BasicOpsCloverRepository{db: db}, nil
}

// NewCloverRepository creates a repository with a new CloverDB instance at the
// specified directory path. The directory is created if it does not exist;
// either way the outcome is reported via a log message.
func NewCloverRepository(dirPath string) (BasicOpsCloverRepository, error) {
	if info, err := os.Stat(dirPath); err != nil {
		if !os.IsNotExist(err) {
			return BasicOpsCloverRepository{}, fmt.Errorf("failed to access database directory %q: %w", dirPath, err)
		}
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			return BasicOpsCloverRepository{}, fmt.Errorf("failed to create database directory %q: %w", dirPath, err)
		}
		log.Printf("Created new database directory %q", dirPath)
	} else if !info.IsDir() {
		return BasicOpsCloverRepository{}, fmt.Errorf("database path %q exists but is not a directory", dirPath)
	} else {
		log.Printf("Using existing database directory %q", dirPath)
	}

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
