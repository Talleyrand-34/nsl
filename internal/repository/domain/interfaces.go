// Comment
package repository

import e "nsl-graph/internal/repository/entities"

// Comment
type NetRepository interface {
	// Brand
	AddBrand(brand string) error
	// Brand
	GetBrands() []string
	// deviceclass
	AddDeviceClass(devclass string) error
	// deviceclass
	GetDeviceClasses() []string
	// deviceclass
	AddZoneType(name string) error
	// deviceclass
	GetZonetypes() []string
	// deviceclass
	AddProprietary(name string) error
	// deviceclass
	GetProperties() []string
	// deviceclass
	AddZone(
		name string,
		father string,
		proprietary string,
		zonename string,
	) error
	// deviceclass
	GetZones() []e.Zone
	// GetZone(name string) int
}
