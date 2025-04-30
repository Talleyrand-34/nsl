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
package sqlcbase

import (
	e "nsl-graph/internal/repository/entities"
)

func (r SQLiteRepository) GetBrands() ([]e.Brand, error) {
	return r.basicops.GetBrands()
}

func (r SQLiteRepository) GetDeviceClasses() ([]e.DevClass, error) {
	return r.basicops.GetDeviceClasses()
}

func (r SQLiteRepository) GetZonetypes() ([]e.ZoneType, error) {
	return r.basicops.GetZonetypes()
}

func (r SQLiteRepository) GetProperties() ([]e.Proprietary, error) {
	return r.basicops.GetProperties()
}

func (r SQLiteRepository) GetZones() ([]e.Zone, error) {
	return r.basicops.GetZones()
}

func (r SQLiteRepository) GetModels() ([]e.ModelDevice, error) {
	return r.basicops.GetModels()
}

func (r SQLiteRepository) GetDevices() ([]e.Device, error) {
	return r.basicops.GetDevices()
}

func (r SQLiteRepository) GetModelPorts() ([]e.ModelPort, error) {
	return r.basicops.GetModelPorts()
}

func (r SQLiteRepository) GetDevicePorts() ([]e.DevicePort, error) {
	return r.basicops.GetDevicePorts()
}

func (r SQLiteRepository) GetConnections() ([]e.Connection, error) {
	return r.basicops.GetConnections()
}

func (r SQLiteRepository) GetAllPortsAll() ([]e.DevicePort, error) {
	return r.specops.GetAllPortsAll()
}

func (r SQLiteRepository) GetAllPortsDevice(deviceid string) ([]e.DevicePort, error) {
	return r.specops.GetAllPortsDevice(deviceid)
}

func (r SQLiteRepository) ExportAllStructs() (e.All, error) {
	return r.specops.ExportAllStructs()
}
