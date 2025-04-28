
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
			ID:   b.ID,
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
			ID:             ct.ID,
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
			ID:                        c.ID,
			FromDevicePortModelPortID: c.FromDevicePortModelPortID,
			FromDevicePortDeviceID:    c.FromDevicePortDeviceID,
			FromIPSegment:             nullStringToString(c.FromIpSegment),
			ToDevicePortModelPortID:   c.ToDevicePortModelPortID,
			ToDevicePortDeviceID:      c.ToDevicePortDeviceID,
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
			ID:   dc.ID,
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
			ModelPortID: dp.ModelPortID,
			DeviceID:    dp.DeviceID,
		})
	}

	// Devices
	devices, err := r.query.BasicGetDevices(ctx)
	if err != nil {
		return result, err
	}
	for _, d := range devices {
		result.Devices = append(result.Devices, e.BasicDevice{
			ID:          d.ID,
			Label:       d.Label,
			ModelID:     d.ModelID,
			ZoneID:      nullInt64ToInt64(d.ZoneID),
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
			ID:      md.ID,
			Model:   md.Model,
			Brand:   md.Brand,
			ClassID: md.ClassID,
		})
	}

	// ModelPorts
	modelPorts, err := r.query.BasicGetModelPorts(ctx)
	if err != nil {
		return result, err
	}
	for _, mp := range modelPorts {
		result.ModelPorts = append(result.ModelPorts, e.BasicModelport{
			ID:        mp.ID,
			Name:      mp.Name,
			Positionx: mp.Positionx,
			Positiony: mp.Positiony,
			ModelID:   mp.ModelID,
		})
	}

	// Policies
	policies, err := r.query.BasicGetPolicies(ctx)
	if err != nil {
		return result, err
	}
	for _, p := range policies {
		result.Policies = append(result.Policies, e.BasicPolicy{
			ID:                   p.ID,
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
			ID:          pr.ID,
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
			ID:           zt.ID,
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
			ID:           z.ID,
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
