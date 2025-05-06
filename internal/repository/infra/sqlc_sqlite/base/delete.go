// add.go contains deletion operations
package sqlcbase

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

func (r SQLiteRepository) DeleteBrand(brand string) error {
	return r.basicops.DeleteBrand(brand)
}

func (r SQLiteRepository) DeleteDeviceClass(name string) error {
	return r.basicops.DeleteDeviceClass(name)
}

func (r SQLiteRepository) DeleteZoneType(locationType string) error {
	return r.basicops.DeleteZoneType(locationType)
}

func (r SQLiteRepository) DeleteProprietary(proprietary string) error {
	return r.basicops.DeleteProprietary(proprietary)
}

func (r SQLiteRepository) DeleteZone(id string) error {
	return r.basicops.DeleteZone(id)
}

func (r SQLiteRepository) DeleteModel(id string) error {
	return r.basicops.DeleteModel(id)
}

func (r SQLiteRepository) DeleteDevice(id string) error {
	return r.basicops.DeleteDevice(id)
}

func (r SQLiteRepository) DeleteModelPort(id string) error {
	return r.basicops.DeleteModelPort(id)
}

func (r SQLiteRepository) DeleteDevicePort(deviceID, modelPortID string) error {
	return r.basicops.DeleteDevicePort(deviceID, modelPortID)
}

func (r SQLiteRepository) DeleteConnection(id string) error {
	return r.basicops.DeleteConnection(id)
}
