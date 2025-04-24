package cmd_utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	c "nsl-graph/cmd"
	q "nsl-graph/internal/repository/application"
	infra "nsl-graph/internal/repository/infra/sqlc_sqlite/base"
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
func ServiceConnection() (q.NetServiceInt, error) {
	db, err := OpenDatabaseConnection()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	repository, err := infra.NewSQLiteRepositoryFromDB(db)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	service := q.NewNetService(repository)
	return service, err
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
