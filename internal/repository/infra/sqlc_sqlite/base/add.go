// package sqlcbase contains the struct and public functions to interact with the satabase
//
// add.go contains addition operations
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

func (r SQLiteRepository) AddBrand(brand string) error {
	return r.basicops.AddBrand(brand)
}

func (r SQLiteRepository) AddDeviceClass(devclass string) error {
	return r.basicops.AddDeviceClass(devclass)
}

func (r SQLiteRepository) AddZoneType(name string) error {
	return r.basicops.AddZoneType(name)
}

func (r SQLiteRepository) AddProprietary(name string) error {
	return r.basicops.AddProprietary(name)
}

func (r SQLiteRepository) AddZone(
	name string,
	fatherid string,
	father string,
	proprietary string,
	zonename string,
) error {
	return r.basicops.AddZone(name, fatherid, father, proprietary, zonename)
}

func (r SQLiteRepository) AddModel(
	modelName string,
	brandName string,
	className string,
) error {
	return r.basicops.AddModel(modelName, brandName, className)
}

func (r SQLiteRepository) AddDevice(
	label string,
	model string,
	zoneId string,
	zoneName string,
	proprietary string,
) error {
	return r.basicops.AddDevice(label, model, zoneId, zoneName, proprietary)
}

func (r SQLiteRepository) AddModelPort(
	name string,
	posx string,
	posy string,
	modelName string,
) error {
	return r.basicops.AddModelPort(name, posx, posy, modelName)
}

func (r SQLiteRepository) AddDevicePort(deviceid string, modelportid string, macAddress string) error {
	return r.basicops.AddDevicePort(deviceid, modelportid, macAddress)
}

func (r SQLiteRepository) AddConnectionType(connectionTypeName string) error {
	return r.basicops.AddConnectionType(connectionTypeName)
}

func (r SQLiteRepository) AddConnection(
	fromDevice string,
	fromModelPort string,
	toDevice string,
	toModelPort string,
) error {
	return r.basicops.AddConnection(fromDevice, fromModelPort, toDevice, toModelPort)
}
