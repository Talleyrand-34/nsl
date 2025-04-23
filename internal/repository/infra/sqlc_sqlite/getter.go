package sqlc_sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db
)

// Subfunction to get father ID as sql.NullInt64
func (r SQLiteRepository) getFatherID(
	ctx context.Context,
	fatherid string,
	father string,
) sql.NullInt64 {
	// Prefer fatherid if provided
	if fatherid != "" {
		id, err := strconv.ParseInt(fatherid, 10, 64)
		if err != nil {
			return sql.NullInt64{Valid: false}
		}
		return sql.NullInt64{Int64: id, Valid: true}
	}

	// Otherwise, try to get from father name
	if father != "" {
		id, err := r.query.GetZoneId(ctx, father)
		if err != nil {
			return sql.NullInt64{Valid: false}
		}
		return sql.NullInt64{Int64: id, Valid: true}
	}

	// Neither provided, set as invalid
	return sql.NullInt64{Valid: false}
}

// Subfunction to get proprietary ID as sql.NullInt64
func (r SQLiteRepository) getProprietaryID(
	ctx context.Context,
	proprietary string,
) sql.NullInt64 {
	propid, err := r.query.GetProprietary(ctx, proprietary)
	if err != nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: propid, Valid: true}
}

// Subfunction to get zone type ID as sql.NullInt64
func (r SQLiteRepository) getZoneTypeID(
	ctx context.Context,
	zonename string,
) sql.NullInt64 {
	zonetypeid, err := r.query.GetZoneType(ctx, zonename)
	if err != nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: zonetypeid, Valid: true}
}

// Subfunction to get zone type ID as sql.NullInt64
func (r SQLiteRepository) getZoneID(
	ctx context.Context,
	zoneid string,
	zonename string,
) sql.NullInt64 {
	// Prefer fatherid if provided
	if zoneid != "" {
		id, err := strconv.ParseInt(zoneid, 10, 64)
		if err != nil {
			return sql.NullInt64{Valid: false}
		}
		return sql.NullInt64{Int64: id, Valid: true}
	}
	id, err := r.query.GetZoneId(ctx, zonename)
	if err != nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: id, Valid: true}
}

// Subfunction to get zone type ID as sql.NullInt64
func (r SQLiteRepository) getModelID(
	ctx context.Context,
	modelname string,
) (int64, error) {
	szoneid, err := r.query.GetModelId(ctx, modelname)
	if err != nil {
		return 0, fmt.Errorf("error getting modelid from model name: %v", err)
	}
	return szoneid, nil
}
