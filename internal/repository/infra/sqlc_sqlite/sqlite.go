package sqlc_sqlite

import (
	"context"
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
	d "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
)

// AddBrand adds a new brand to the database
func (r SQLiteRepository) AddBrand(brand string) error {
	ctx := context.Background()
	execErr := r.query.AddBrand(ctx, brand)
	if execErr != nil {
		log.Printf("Error executing query: %v", execErr)
		return execErr
	}
	return nil
}

// GetBrands gets all the brands available
func (r SQLiteRepository) GetBrands() []string {
	ctx := context.Background()
	brands, execErr := r.query.GetBrands(ctx)
	if execErr != nil {
		log.Printf("Error executing query: %v", execErr)
		return []string{}
	}
	return brands
}

// AddDeviceClass adds a new brand to the database
func (r SQLiteRepository) AddDeviceClass(devClassName string) error {
	ctx := context.Background()
	execErr := r.query.AddDeviceClass(ctx, devClassName)
	if execErr != nil {
		log.Printf("Error executing query: %v", execErr)
		return execErr
	}
	return nil
}

// GetDeviceClasses gets all the brands available
func (r SQLiteRepository) GetDeviceClasses() []string {
	ctx := context.Background()
	brands, execErr := r.query.GetDeviceClasses(ctx)
	if execErr != nil {
		log.Printf("Error executing query: %v", execErr)
		return []string{}
	}
	return brands
}

// deviceclass
func (r SQLiteRepository) AddZoneType(zoneName string) error {
	ctx := context.Background()
	execErr := r.query.AddZoneType(ctx, zoneName)
	if execErr != nil {
		log.Printf("Error executing query: %v", execErr)
		return execErr
	}
	return nil
}

// deviceclass
func (r SQLiteRepository) GetZonetypes() []string {
	ctx := context.Background()
	zonetypes, execErr := r.query.GetZoneTypes(ctx)
	if execErr != nil {
		log.Printf("Error executing query: %v", execErr)
		return []string{}
	}
	return zonetypes
}

// deviceclass
func (r SQLiteRepository) AddProprietary(proprietary string) error {
	ctx := context.Background()
	execErr := r.query.AddProprietary(ctx, proprietary)
	if execErr != nil {
		log.Printf("Error executing query: %v", execErr)
		return execErr
	}
	return nil
}

// deviceclass
func (r SQLiteRepository) GetProperties() []string {
	ctx := context.Background()
	proprietaries, execErr := r.query.GetProprietaries(ctx)
	if execErr != nil {
		log.Printf("Error executing query: %v", execErr)
		return []string{}
	}
	return proprietaries
}

// deviceclass
func (r SQLiteRepository) AddZone(
	name string,
	father string,
	proprietary string,
	zonename string,
) error {
	ctx := context.Background()
	// get father id
	var sfatherid sql.NullInt64
	if father != "" {
		fatherid, err := r.query.GetZoneId(ctx, father)
		if err != nil {
			log.Printf("Error executing GetZoneId of father on AddZone: %v", err)
			return err
		}
		sfatherid = sql.NullInt64{
			Int64: fatherid,
			Valid: true,
		}
	} else {
		sfatherid = sql.NullInt64{
			Valid: false,
		}
	}
	// get proprietary id
	propid, err := r.query.GetProprietary(ctx, proprietary)
	if err != nil {
		log.Printf("Error executing  GetProprietary on AddZonequery: %v", err)
		return err
	}
	spropid := sql.NullInt64{
		Int64: propid,
		Valid: true,
	}
	// get zonetype id
	zonetypeid, err := r.query.GetZoneType(ctx, zonename)
	if err != nil {
		log.Printf("Error executing  GetZonetype on AddZone: %v", err)
		return err
	}
	szonetypeid := sql.NullInt64{
		Int64: zonetypeid,
		Valid: true,
	}
	// insert in db
	zonestruct := d.AddZoneParams{
		Name:         name,
		Father:       sfatherid,
		LocationType: szonetypeid,
		Proprietary:  spropid,
	}
	execErr := r.query.AddZone(ctx, zonestruct)
	if execErr != nil {
		log.Printf("Error executing query: %v", execErr)
		return execErr
	}
	return nil
}

func (r SQLiteRepository) GetZones() []e.Zone {
	ctx := context.Background()
	zones, execErr := r.query.GetZones(ctx)
	if execErr != nil {
		log.Printf("Error executing query: %v", execErr)
		return []e.Zone{}
	}

	result := make([]e.Zone, 0, len(zones))
	for _, row := range zones {
		zone := e.Zone{
			Name:         row.Name,
			Father:       row.Father,
			LocationType: row.LocationType,
			Proprietary:  row.Proprietary,
		}
		result = append(result, zone)
	}
	return result
}

// deviceclass
// func (r SQLiteRepository) GetZone(name string) int {
// }
