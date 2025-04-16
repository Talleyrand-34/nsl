package cmd_utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"

	c "nsl-graph/cmd"
	q "nsl-graph/internal/repository/application"
	infra "nsl-graph/internal/repository/infra/bare_sqlite"
)

// openDatabaseConnection establishes and returns a database connection.
func OpenDatabaseConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", c.Srcdbpath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}
	return db, nil
}

// Return repositoryDB connection
func RepositoryConnection() (q.NetServiceInt, error) {
	db, err := OpenDatabaseConnection()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	repository, err := infra.NewSQLiteRepositoryFromDB(db)
	return repository, err
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
