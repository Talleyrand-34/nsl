package sqlite

import (
	"database/sql"
	"log"
	"nsl-graph/internal/repository/gen/model"
	"nsl-graph/internal/repository/gen/table"

	s "github.com/go-jet/jet/v2/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(filePath string) (SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return SQLiteRepository{}, err
	}

	// Create the Brand table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Brand (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			brand TEXT NOT NULL
		);
	`)
	if err != nil {
		db.Close() // Close the connection if table creation fails
		return SQLiteRepository{}, err
	}

	return SQLiteRepository{db: db}, nil
}

// Close closes the database connection
func (r SQLiteRepository) Close() error {
	return r.db.Close()
}

// AddBrand adds a new brand to the database
func (r SQLiteRepository) AddBrand(brandName string) error {
	// Create an INSERT statement for the brand
	stmt := table.Brand.INSERT(table.Brand.Brand).
		VALUES(s.String(brandName))

	// Execute the insert statement
	_, err := stmt.Exec(r.db)
	return err
}

// GetBrands retrieves all brands from the database as a string array
func (r SQLiteRepository) GetBrands() []string {
	// Create a SELECT statement to get all brand names
	stmt := s.SELECT(
		table.Brand.Brand,
	).FROM(
		table.Brand,
	)
	// Define the destination structure
	var brandStrings []struct {
		brand model.Brand
	}
	// Execute the query and store results in brandStrings
	err := stmt.Query(r.db, &brandStrings)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		// In case of error, return empty slice
		return []string{}
	}

	// Convert the result to a string array
	result := make([]string, len(brandStrings))
	for i, item := range brandStrings {
		result[i] = item.brand.Brand // Access the Brand field inside the struct
	}
	return result
}
