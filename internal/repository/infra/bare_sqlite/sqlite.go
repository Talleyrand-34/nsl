package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"nsl-graph/internal/repository/gen/model"

	"github.com/Masterminds/squirrel"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepositoryFromDB(db *sql.DB) (SQLiteRepository, error) {
	return SQLiteRepository{db: db}, nil
}

func NewSQLiteRepository(filePath string) (SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return SQLiteRepository{}, err
	}
	// // to-do eliminate hardcoded creation
	// _, err = db.Exec(`
	// 	CREATE TABLE IF NOT EXISTS Brand (
	// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
	// 		brand TEXT NOT NULL
	// 	);
	// `)
	// if err != nil {
	// 	db.Close() // Close the connection if table creation fails
	// 	return SQLiteRepository{}, err
	// }
	// to-do Check db structure
	// Check if the Brand table exists
	var tableName string
	err = db.QueryRow(`
		SELECT name 
		FROM sqlite_master 
		WHERE type='table' AND name='Brand';
	`).Scan(&tableName)

	if err == sql.ErrNoRows {
		db.Close()
		return SQLiteRepository{}, fmt.Errorf("table 'Brand' does not exist in the database")
	} else if err != nil {
		db.Close()
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
	// Create an INSERT statement for the brand using Squirrel
	query, args, err := squirrel.
		Insert("Brand").   // Table name
		Columns("Brand").  // Column name
		Values(brandName). // Value to insert
		ToSql()            // Generate SQL and arguments
	if err != nil {
		log.Printf("Error building SQL: %v", err)
		return err
	}

	// Execute the query
	_, execErr := r.db.Exec(query, args...)
	if execErr != nil {
		log.Printf("Error executing query: %v", execErr)
		return execErr
	}

	return nil
}

// GetBrands retrieves all brands from the database as a string array
func (r SQLiteRepository) GetBrands() []string {
	// Create a SELECT statement to get all brand names using Squirrel
	query, args, err := squirrel.
		Select("ID", "Brand"). // Select columns (ID is optional)
		From("Brand").         // Table name
		ToSql()                // Generate SQL and arguments
	if err != nil {
		log.Printf("Error building SQL: %v", err)
		return []string{}
	}

	// Define the destination structure
	var brands []model.Brand

	// Execute the query and scan results into the brands slice
	rows, queryErr := r.db.Query(query, args...)
	if queryErr != nil {
		log.Printf("Error executing query: %v", queryErr)
		return []string{}
	}
	defer rows.Close()

	for rows.Next() {
		var brand model.Brand
		if scanErr := rows.Scan(&brand.ID, &brand.Brand); scanErr != nil {
			log.Printf("Error scanning row: %v", scanErr)
			return []string{}
		}
		brands = append(brands, brand)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		log.Printf("Error iterating over rows: %v", rowsErr)
		return []string{}
	}

	// Convert the result to a string array (extracting the Brand field)
	result := make([]string, len(brands))
	for i, item := range brands {
		result[i] = item.Brand
	}

	return result
}
