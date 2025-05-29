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
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	q "nsl-graph/internal/repository/application"
)

// validateRequiredFields checks that all required fields in v are non-empty.
// v must be a pointer to a struct.
func validateRequiredFields(v interface{}, requiredFields []string) error {
	rv := reflect.ValueOf(v).Elem()
	for _, field := range requiredFields {
		f := rv.FieldByName(field)
		if !f.IsValid() {
			return fmt.Errorf("field %s not found", field)
		}
		// Only check string fields for emptiness
		if f.Kind() == reflect.String && strings.TrimSpace(f.String()) == "" {
			return fmt.Errorf("field %s cannot be empty", field)
		}
	}
	return nil
}

// genericAddHandler abstracts common POST handler logic
func genericAddHandler[T any](
	service q.NetServiceInt,
	requiredFields []string,
	addFunc func(service q.NetServiceInt, req *T) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req T
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := validateRequiredFields(&req, requiredFields); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := addFunc(service, &req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

// Brand
type AddBrandRequest struct {
	Brand string `json:"brand"`
}

func addBrandHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericAddHandler[AddBrandRequest](
		service,
		[]string{"Brand"},
		func(service q.NetServiceInt, req *AddBrandRequest) error {
			return service.AddBrand(req.Brand)
		},
	)
}

// DeviceClass
type AddDeviceClassRequest struct {
	DevClass string `json:"name"`
}

func addDeviceClassHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericAddHandler[AddDeviceClassRequest](
		service,
		[]string{"DevClass"},
		func(service q.NetServiceInt, req *AddDeviceClassRequest) error {
			return service.AddDeviceClass(req.DevClass)
		},
	)
}

// ZoneType
type AddZoneTypeRequest struct {
	Name string `json:"name"`
}

func addZoneTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericAddHandler[AddZoneTypeRequest](
		service,
		[]string{"Name"},
		func(service q.NetServiceInt, req *AddZoneTypeRequest) error {
			return service.AddZoneType(req.Name)
		},
	)
}

// Proprietary
type AddProprietaryRequest struct {
	Name string `json:"name"`
}

func addProprietaryHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericAddHandler[AddProprietaryRequest](
		service,
		[]string{"Name"},
		func(service q.NetServiceInt, req *AddProprietaryRequest) error {
			return service.AddProprietary(req.Name)
		},
	)
}

// Zone
type AddZoneRequest struct {
	Name          string `json:"name"`
	Father        string `json:"father"`
	FatherID      string `json:"fatherid"`
	Proprietary   string `json:"proprietary"`
	Location_type string `json:"location_type"`
}

func addZoneHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericAddHandler[AddZoneRequest](
		service,
		[]string{"Name", "Proprietary"}, // Only these are required, adjust as needed
		func(service q.NetServiceInt, req *AddZoneRequest) error {
			return service.AddZone(
				req.Name,
				req.FatherID,
				req.Father,
				req.Proprietary,
				req.Location_type,
			)
		},
	)
}

// Model
type AddModelRequest struct {
	ModelName string `json:"model"`
	BrandName string `json:"brand"`
	ClassName string `json:"class"`
}

func addModelHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericAddHandler[AddModelRequest](
		service,
		[]string{"ModelName", "BrandName", "ClassName"},
		func(service q.NetServiceInt, req *AddModelRequest) error {
			return service.AddModel(req.ModelName, req.BrandName, req.ClassName)
		},
	)
}

// Device
type AddDeviceRequest struct {
	Label       string `json:"label"`
	Model       string `json:"model"`
	ZoneId      string `json:"zoneId"`
	ZoneName    string `json:"zoneName"`
	Proprietary string `json:"proprietary"`
}

func addDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericAddHandler[AddDeviceRequest](
		service,
		[]string{"Label", "Model", "Proprietary"}, // Adjust required fields as needed
		func(service q.NetServiceInt, req *AddDeviceRequest) error {
			return service.AddDevice(
				req.Label,
				req.Model,
				req.ZoneId,
				req.ZoneName,
				req.Proprietary,
			)
		},
	)
}

// ModelPort
type AddModelPortRequest struct {
	Name      string `json:"name"`
	PosX      string `json:"posx"`
	PosY      string `json:"posy"`
	ModelName string `json:"modelName"`
}

func addModelPortHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericAddHandler[AddModelPortRequest](
		service,
		[]string{"Name", "ModelName"}, // Adjust required fields as needed
		func(service q.NetServiceInt, req *AddModelPortRequest) error {
			return service.AddModelPort(req.Name, req.PosX, req.PosY, req.ModelName)
		},
	)
}

// DevicePort
type AddDevicePortRequest struct {
	DeviceID    string `json:"deviceid"`
	ModelPortID string `json:"modelportid"`
	MacAddress  string `json:"mac_address"`
}

func addDevicePortHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericAddHandler[AddDevicePortRequest](
		service,
		[]string{"DeviceID", "ModelPortID"},
		func(service q.NetServiceInt, req *AddDevicePortRequest) error {
			return service.AddDevicePort(req.DeviceID, req.ModelPortID, req.MacAddress)
		},
	)
}

// Connection
type AddConnectionRequest struct {
	FromDevice    string `json:"fromDevice"`
	FromModelPort string `json:"fromModelPort"`
	FromIPSegment string `json:"fromIPSegment"`
	ToDevice      string `json:"toDevice"`
	ToModelPort   string `json:"toModelPort"`
	ToIPSegment   string `json:"toIPSegment"`
}

func addConnectionHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericAddHandler[AddConnectionRequest](
		service,
		[]string{"FromDevice", "FromModelPort", "ToDevice", "ToModelPort"},
		func(service q.NetServiceInt, req *AddConnectionRequest) error {
			return service.AddConnection(
				req.FromDevice,
				req.FromModelPort,
				req.FromIPSegment,
				req.ToDevice,
				req.ToModelPort,
				req.ToIPSegment,
			)
		},
	)
}

// ConnectionType
type AddConnectionTypeRequest struct {
	Name string `json:"name"`
}

func addConnectionTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericAddHandler[AddConnectionTypeRequest](
		service,
		[]string{"Name"},
		func(service q.NetServiceInt, req *AddConnectionTypeRequest) error {
			return service.AddConnectionType(req.Name)
		},
	)
}
