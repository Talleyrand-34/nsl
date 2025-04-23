package sqlc_sqlite

import (
	"context"
	"fmt"
	"strconv"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
	d "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
)

// GetModels returns all models with resolved brand/class names
func (r SQLiteRepository) GetModels() []e.ModelDevice {
	ctx := context.Background()
	models, execErr := r.query.GetModels(ctx)
	if execErr != nil {
		return []e.ModelDevice{}
	}

	result := make([]e.ModelDevice, 0, len(models))
	for _, row := range models {
		model := e.ModelDevice{
			Id:    strconv.FormatInt(row.ID, 10),
			Model: row.Model,
			Brand: row.Brand,
			Class: row.ClassName,
		}
		result = append(result, model)
	}
	return result
}

// AddModel creates new model entry resolving brand/class names to IDs
func (r SQLiteRepository) AddModel(
	modelName string,
	brandName string,
	className string,
) error {
	ctx := context.Background()

	// Get brand ID from name (required)
	brandID, err := r.query.GetBrandId(ctx, brandName)
	if err != nil {
		return fmt.Errorf("brand not found: %s", brandName)
	}

	// Get class ID from name (required)
	classID, err := r.query.GetClassId(ctx, className)
	if err != nil {
		return fmt.Errorf("class not found: %s", className)
	}

	modelStruct := d.AddModelParams{
		Model:   modelName,
		Brand:   brandID,
		ClassID: classID,
	}

	if err := r.query.AddModel(ctx, modelStruct); err != nil {
		return fmt.Errorf("failed to create model")
	}
	return nil
}
