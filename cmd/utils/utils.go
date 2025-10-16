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
package cmd_utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	c "nsl-graph/cmd"
	q "nsl-graph/internal/repository/application"
	cloverInfra "nsl-graph/internal/repository/infra/manual_cloverdb/base"
	sqliteInfra "nsl-graph/internal/repository/infra/sqlc_sqlite/base"
	// sqliteInfra "nsl-graph/internal/repository/infra/sqlc_sqlite/base"
)

// openDatabaseConnection establishes and returns a database connection.
func OpenDatabaseConnectionSqlite() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", c.Srcdbpath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}
	return db, nil
}
func OpenDatabaseConnectionClover() {
	os.Mkdir(c.Srcdbpath, 0755)
}

// Return repositoryDB connection
func ServiceConnection() (q.NetServiceInt, error) {
	backend := strings.ToLower(c.Backend)

	switch backend {
	case "cloverdb", "clover":
		// Use CloverDB backend
		OpenDatabaseConnectionClover()
		repository, err := cloverInfra.NewCloverRepository(c.Srcdbpath)
		if err != nil {
			log.Fatalf("Error connecting to CloverDB: %v", err)
		}
		service := q.NewNetService(repository)
		return service, nil

	case "sqlite", "":
		// Use SQLite backend (default)
		db, err := OpenDatabaseConnectionSqlite()
		if err != nil {
			log.Fatalf("Error connecting to SQLite database: %v", err)
		}
		repository, err := sqliteInfra.NewSQLiteRepositoryFromDB(db)
		if err != nil {
			log.Fatalf("Error initializing SQLite repository: %v", err)
		}
		service := q.NewNetService(repository)
		return service, nil

	default:
		return nil, fmt.Errorf("unsupported backend type: %s (supported: sqlite, cloverdb)", backend)
	}
}

// PrintPrettyJSON prints the JSON data in a human-readable format.
func PrintPrettyJSON(jsonData string) error {
	var prettyData map[string]interface{}

	// Unmarshal the JSON string into a map for formatting
	err := json.Unmarshal([]byte(jsonData), &prettyData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Marshal the map back into a pretty-printed JSON string
	prettyJSON, err := json.MarshalIndent(prettyData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(prettyJSON))
	return nil
}

func PrintStringArrayPrettyJson(arr []string) {
	b, _ := json.MarshalIndent(arr, "", "  ")
	fmt.Println(string(b))
}

// flagproc gets the values of the given flag names from the command and returns them as a slice of strings.
func Flagproc(cmd *cobra.Command, flagNames []string) []string {
	values := make([]string, len(flagNames))
	for i, flag := range flagNames {
		val, err := cmd.Flags().GetString(flag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag '%v': %v\n", flag, err)
			os.Exit(1)
		}
		values[i] = val
	}
	return values
}
