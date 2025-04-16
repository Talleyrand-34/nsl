package cmd_utils

import (
	libsql "database/sql"
	"encoding/json"
	"fmt"

	"nslgraph/src/customsql" // Import lib instead of internal

	"github.com/sirupsen/logrus"
)

func prepare() (*logrus.Logger, *libsql.DB) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel) // Set logging level to DEBUG

	db, err := customsql.OpenDatabaseConnection(srcdbpath)
	if err != nil {
		logger.Error(fmt.Sprintf("Error connecting to the database: %v", err))
		return logger, db
	}
	return logger, db
}

func closeDB(db *libsql.DB) {
	defer db.Close()
}

// Abstracted function to process commands
func ProcessCmd(fn func(logger *logrus.Logger, db *libsql.DB)) {
	logger, db := prepare()
	fn(logger, db) // Call the passed function
	closeDB(db)
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
