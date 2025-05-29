// update.go contains modification operations
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

func (r SQLiteRepository) UpdateConnection(
	id string,
	fromDevice string,
	fromPort string,
	toDevice string,
	toPort string,
) error {
	return r.basicops.UpdateConnection(id, fromDevice, fromPort, toDevice, toPort)
}

func (r SQLiteRepository) UpdateBrand(brandId string, newBrandName string) error {
	return r.basicops.UpdateBrand(brandId, newBrandName)
}

func (r SQLiteRepository) UpdateDeviceClass(deviceClassId string, newDeviceClassName string) error {
	return r.basicops.UpdateDeviceClass(deviceClassId, newDeviceClassName)
}

func (r SQLiteRepository) UpdateZoneType(zoneTypeId string, newZoneTypeName string) error {
	return r.basicops.UpdateZoneType(zoneTypeId, newZoneTypeName)
}

func (r SQLiteRepository) UpdateProprietary(proprietaryId string, newProprietaryName string) error {
	return r.basicops.UpdateProprietary(proprietaryId, newProprietaryName)
}

func (r SQLiteRepository) UpdateZone(
	zoneId string,
	newZoneName string,
	newFatherZoneId string,
	newZoneTypeId string,
	newProprietaryId string,
) error {
	return r.basicops.UpdateZone(zoneId, newZoneName, newFatherZoneId, newZoneTypeId, newProprietaryId)
}

func (r SQLiteRepository) UpdateModel(
	modelId string,
	newModelName string,
	newBrandId string,
	newDeviceClassId string,
) error {
	return r.basicops.UpdateModel(modelId, newModelName, newBrandId, newDeviceClassId)
}

func (r SQLiteRepository) UpdateDevice(
	deviceId string,
	newDeviceLabel string,
	newModelId string,
	newZoneId string,
	newProprietaryId string,
) error {
	return r.basicops.UpdateDevice(deviceId, newDeviceLabel, newModelId, newZoneId, newProprietaryId)
}

func (r SQLiteRepository) UpdateModelPort(
	modelPortId string,
	newPortName string,
	newPositionX string,
	newPositionY string,
	newModelId string,
) error {
	return r.basicops.UpdateModelPort(modelPortId, newPortName, newPositionX, newPositionY, newModelId)
}

func (r SQLiteRepository) UpdateConnectionType(connectionTypeId string, newConnectionTypeName string) error {
	return r.basicops.UpdateConnectionType(connectionTypeId, newConnectionTypeName)
}
