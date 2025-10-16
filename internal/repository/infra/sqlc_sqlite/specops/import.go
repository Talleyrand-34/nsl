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
package specops

import (
	"context"
	"database/sql"
	"strconv"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
	d "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
)

func (r *SpecOpsSQLiteRepository) ImportAllStructs(ctx context.Context, all e.All) error {
	// 1. Brands
	for _, b := range all.Brands {
		if err := r.query.AddBrand(ctx, b.Name); err != nil {
			return err
		}
	}

	// 2. ConnectionTypes
	// for _, ct := range all.ConnectionTypes {
	// 	if err := r.query.AddConnectionType(ctx, ct.ConnectionType); err != nil {
	// 		return err
	// 	}
	// }

	// 3. Proprietaries
	for _, p := range all.Proprietaries {
		if err := r.query.AddProprietary(ctx, p.Proprietary); err != nil {
			return err
		}
	}

	// 4. DeviceClasses
	for _, dc := range all.DeviceClasses {
		if err := r.query.AddDeviceClass(ctx, dc.Name); err != nil {
			return err
		}
	}

	// 5. ZoneTypes
	for _, zt := range all.ZoneTypes {
		if err := r.query.AddZoneType(ctx, zt.LocationType); err != nil {
			return err
		}
	}

	// 6. Zones
	for _, z := range all.Zones {
		// Assume you have helpers: getProprietaryID, getZoneTypeID
		propID, _ := r.getProprietaryID(ctx, strconv.FormatInt(z.Proprietary, 10))
		zoneTypeID, _ := r.getZoneTypeID(ctx, strconv.FormatInt(z.LocationType, 10))
		arg := d.AddZoneParams{
			Name:         z.Name,
			Father:       int64ToNull(z.Father),
			LocationType: int64ToNull(zoneTypeID),
			Proprietary:  int64ToNull(propID),
		}
		if err := r.query.AddZone(ctx, arg); err != nil {
			return err
		}
	}

	// 7. ModelDevices
	for _, md := range all.ModelDevices {
		// Assume you have helpers: getBrandID, getDeviceClassID
		brandID, _ := r.getBrandID(ctx, strconv.FormatInt(md.Brand, 10))
		classID, _ := r.getDeviceClassID(ctx, md.ClassID)
		arg := d.AddModelParams{
			Model:   md.Model,
			Brand:   brandID,
			ClassID: classID,
		}
		if err := r.query.AddModel(ctx, arg); err != nil {
			return err
		}
	}

	// 8. ModelPorts
	for _, mp := range all.ModelPorts {
		modelid, err := strconv.Atoi(mp.ModelID)
		if err != nil {
			modelid = 0

		}
		arg := d.AddModelPortParams{
			Name:      mp.Name,
			Positionx: mp.Positionx,
			Positiony: mp.Positiony,
			ModelID:   int64(modelid),
		}
		if err := r.query.AddModelPort(ctx, arg); err != nil {
			return err
		}
	}

	// 9. Devices
	for _, dv := range all.Devices {
		modelid, err := strconv.Atoi(dv.ModelID)
		if err != nil {
			modelid = 0

		}
		zoneid, err := strconv.Atoi(dv.ZoneID)
		if err != nil {
			modelid = 0

		}
		arg := d.AddDeviceParams{
			Label:       dv.Label,
			ModelID:     int64(modelid),
			ZoneID:      int64ToNull(int64(zoneid)),
			Proprietary: int64ToNull(dv.Proprietary),
		}
		if err := r.query.AddDevice(ctx, arg); err != nil {
			return err
		}
	}

	// 10. DevicePorts
	for _, dp := range all.DevicePorts {
		deviceid, err := strconv.Atoi(dp.DeviceID)
		if err != nil {
			deviceid = 0

		}
		modelportid, err := strconv.Atoi(dp.ModelPortID)
		if err != nil {
			modelportid = 0

		}
		arg := d.AddDevicePortParams{
			DeviceID:    int64(deviceid),
			ModelPortID: int64(modelportid),
		}
		if err := r.query.AddDevicePort(ctx, arg); err != nil {
			return err
		}
	}

	// 11. Connections

	for _, c := range all.Connections {
		fromDevicePortDeviceID, err1 := strconv.ParseInt(c.FromDevicePortDeviceID, 10, 64)
		fromDevicePortModelPortID, err2 := strconv.ParseInt(c.FromDevicePortModelPortID, 10, 64)
		toDevicePortDeviceID, err3 := strconv.ParseInt(c.ToDevicePortDeviceID, 10, 64)
		toDevicePortModelPortID, err4 := strconv.ParseInt(c.ToDevicePortModelPortID, 10, 64)

		// Handle errors gracefully
		if err1 != nil {
			fromDevicePortDeviceID = 0
		}
		if err2 != nil {
			fromDevicePortModelPortID = 0
		}
		if err3 != nil {
			toDevicePortDeviceID = 0
		}
		if err4 != nil {
			toDevicePortModelPortID = 0
		}

		arg := d.AddConnectionParams{
			FromDevicePortDeviceID:    fromDevicePortDeviceID,
			FromDevicePortModelPortID: fromDevicePortModelPortID,
			ToDevicePortDeviceID:      toDevicePortDeviceID,
			ToDevicePortModelPortID:   toDevicePortModelPortID,
			ConnectionType:            int64ToNull(c.ConnectionType),
		}
		if err := r.query.AddConnection(ctx, arg); err != nil {
			return err
		}
	}
	// 12. Policies
	// for _, p := range all.Policies {
	// 	// You may need to adapt this if AddPolicy uses a struct or separate args
	// 	if err := r.query.AddPolicy(ctx, p.Name, p.Description, int64ToNull(p.AssociatedConnection), stringToNull(p.TODO)); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

// Helper: convert -1 to sql.NullInt64
func int64ToNull(val int64) sql.NullInt64 {
	if val == -1 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: val, Valid: true}
}

// // Helper: convert "" to sql.NullString
// func stringToNull(val string) sql.NullString {
// 	if val == "" {
// 		return sql.NullString{}
// 	}
// 	return sql.NullString{String: val, Valid: true}
// }
