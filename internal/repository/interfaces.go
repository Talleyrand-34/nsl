package repository

// d "nsl-graph/internal/types"

type DeviceRepository interface {
	// FetchPossiblePortsDevices() ([]d.DeviceModelPortDetails, error)
	// FetchDevicesWithModelPorts() ([]d.Device, error)
	// FetchDevices() ([]d.Device, error)
	AddBrand(string) error
	GetBrands() []string
}
