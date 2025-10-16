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
)

func (r *SpecOpsSQLiteRepository) ExportAllStructs() (e.All, error) {
	ctx := context.Background()
	var result e.All

	// Brands
	brands, err := r.query.BasicGetBrands(ctx)
	if err != nil {
		return result, err
	}
	for _, b := range brands {

		result.Brands = append(result.Brands, e.BasicBrand{
			ID:   strconv.FormatInt(b.ID, 10),
			Name: b.Brand,
		})
	}

	// ConnectionTypes
	connectionTypes, err := r.query.BasicGetConnectionTypes(ctx)
	if err != nil {
		return result, err
	}
	for _, ct := range connectionTypes {
		result.ConnectionTypes = append(result.ConnectionTypes, e.BasicConnectiontype{
			ID:             strconv.FormatInt(ct.ID, 10),
			ConnectionType: ct.ConnectionType,
		})
	}

	// Connections
	connections, err := r.query.BasicGetConnections(ctx)
	if err != nil {
		return result, err
	}
	for _, c := range connections {
		result.Connections = append(result.Connections, e.BasicConnection{
			ID:                        strconv.FormatInt(c.ID, 10),
			FromDevicePortModelPortID: strconv.FormatInt(c.FromDevicePortModelPortID, 10),
			FromDevicePortDeviceID:    strconv.FormatInt(c.FromDevicePortDeviceID, 10),
			FromIPSegment:             nullStringToString(c.FromIpSegment),
			ToDevicePortModelPortID:   strconv.FormatInt(c.ToDevicePortModelPortID, 10),
			ToDevicePortDeviceID:      strconv.FormatInt(c.ToDevicePortDeviceID, 10),
			ToIPSegment:               nullStringToString(c.ToIpSegment),
			ConnectionType:            nullInt64ToInt64(c.ConnectionType),
		})
	}

	// DeviceClasses
	deviceClasses, err := r.query.BasicGetDeviceClasses(ctx)
	if err != nil {
		return result, err
	}
	for _, dc := range deviceClasses {
		result.DeviceClasses = append(result.DeviceClasses, e.BasicDeviceclass{
			ID:   strconv.FormatInt(dc.ID, 10),
			Name: dc.Name,
		})
	}

	// DevicePorts
	devicePorts, err := r.query.BasicGetDevicePorts(ctx)
	if err != nil {
		return result, err
	}
	for _, dp := range devicePorts {
		result.DevicePorts = append(result.DevicePorts, e.BasicDeviceport{
			ModelPortID: strconv.FormatInt(dp.ModelPortID, 10),
			DeviceID:    strconv.FormatInt(dp.DeviceID, 10),
		})
	}

	// Devices
	devices, err := r.query.BasicGetDevices(ctx)
	if err != nil {
		return result, err
	}
	for _, d := range devices {
		result.Devices = append(result.Devices, e.BasicDevice{
			ID:          strconv.FormatInt(d.ID, 10),
			Label:       d.Label,
			ModelID:     strconv.FormatInt(d.ModelID, 10),
			ZoneID:      strconv.FormatInt(nullInt64ToInt64(d.ZoneID), 10),
			Proprietary: nullInt64ToInt64(d.Proprietary),
		})
	}

	// ModelDevices
	modelDevices, err := r.query.BasicGetModelDevices(ctx)
	if err != nil {
		return result, err
	}
	for _, md := range modelDevices {
		result.ModelDevices = append(result.ModelDevices, e.BasicModeldevice{
			ID:      strconv.FormatInt(md.ID, 10),
			Model:   md.Model,
			Brand:   md.Brand,
			ClassID: strconv.FormatInt(md.ClassID, 10),
		})
	}

	// ModelPorts
	modelPorts, err := r.query.BasicGetModelPorts(ctx)
	if err != nil {
		return result, err
	}
	for _, mp := range modelPorts {
		result.ModelPorts = append(result.ModelPorts, e.BasicModelport{
			ID:        strconv.FormatInt(mp.ID, 10),
			Name:      mp.Name,
			Positionx: mp.Positionx,
			Positiony: mp.Positiony,
			ModelID:   strconv.FormatInt(mp.ModelID, 10),
		})
	}

	// Policies
	policies, err := r.query.BasicGetPolicies(ctx)
	if err != nil {
		return result, err
	}
	for _, p := range policies {
		result.Policies = append(result.Policies, e.BasicPolicy{
			ID:                   strconv.FormatInt(p.ID, 10),
			Name:                 p.Name,
			Description:          p.Description,
			AssociatedConnection: nullInt64ToInt64(p.AssociatedConnection),
			TODO:                 nullStringToString(p.Todo),
		})
	}

	// Proprietaries
	proprietaries, err := r.query.BasicGetProprietaries(ctx)
	if err != nil {
		return result, err
	}
	for _, pr := range proprietaries {
		result.Proprietaries = append(result.Proprietaries, e.BasicProprietary{
			ID:          strconv.FormatInt(pr.ID, 10),
			Proprietary: pr.Proprietary,
		})
	}

	// ZoneTypes
	zoneTypes, err := r.query.BasicGetZoneTypes(ctx)
	if err != nil {
		return result, err
	}
	for _, zt := range zoneTypes {
		result.ZoneTypes = append(result.ZoneTypes, e.BasicZonetype{
			ID:           strconv.FormatInt(zt.ID, 10),
			LocationType: zt.LocationType,
		})
	}

	// Zones
	zones, err := r.query.BasicGetZones(ctx)
	if err != nil {
		return result, err
	}
	for _, z := range zones {
		result.Zones = append(result.Zones, e.BasicZone{
			ID:           strconv.FormatInt(z.ID, 10),
			Name:         z.Name,
			Father:       nullInt64ToInt64(z.Father),
			Granularity:  nullInt64ToInt64(z.Granularity),
			Proprietary:  nullInt64ToInt64(z.Proprietary),
			LocationType: nullInt64ToInt64(z.LocationType),
		})
	}

	return result, nil
}

// Helper functions for null handling
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullInt64ToInt64(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return -1
}
