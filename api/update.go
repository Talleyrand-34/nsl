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
	"net/http"

	q "nsl-graph/internal/repository/application"
)

// genericUpdateHandler abstracts common PUT handler logic
func genericUpdateHandler[T any](
	service q.NetServiceInt,
	requiredFields []string,
	updateFunc func(service q.NetServiceInt, req *T) error,
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
		if err := updateFunc(service, &req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func updateConnectionHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID            string `json:"id"`
			FromDevice    string `json:"from_device"`
			FromPort      string `json:"from_port"`
			FromIPSegment string `json:"from_ip_segment"`
			ToDevice      string `json:"to_device"`
			ToPort        string `json:"to_port"`
			ToIPSegment   string `json:"to_ip_segment"`
			VlanId        string `json:"vlan_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.ID == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := service.UpdateConnection(req.ID, req.FromDevice, req.FromPort, req.FromIPSegment, req.ToDevice, req.ToPort, req.ToIPSegment, req.VlanId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// Brand Update
type UpdateBrandRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func updateBrandHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericUpdateHandler[UpdateBrandRequest](
		service,
		[]string{"ID", "Name"},
		func(service q.NetServiceInt, req *UpdateBrandRequest) error {
			return service.UpdateBrand(req.ID, req.Name)
		},
	)
}

// DeviceClass Update
type UpdateDeviceClassRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func updateDeviceClassHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericUpdateHandler[UpdateDeviceClassRequest](
		service,
		[]string{"ID", "Name"},
		func(service q.NetServiceInt, req *UpdateDeviceClassRequest) error {
			return service.UpdateDeviceClass(req.ID, req.Name)
		},
	)
}

// ZoneType Update
type UpdateZoneTypeRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func updateZoneTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericUpdateHandler[UpdateZoneTypeRequest](
		service,
		[]string{"ID", "Name"},
		func(service q.NetServiceInt, req *UpdateZoneTypeRequest) error {
			return service.UpdateZoneType(req.ID, req.Name)
		},
	)
}

// Proprietary Update
type UpdateProprietaryRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func updateProprietaryHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericUpdateHandler[UpdateProprietaryRequest](
		service,
		[]string{"ID", "Name"},
		func(service q.NetServiceInt, req *UpdateProprietaryRequest) error {
			return service.UpdateProprietary(req.ID, req.Name)
		},
	)
}

// Zone Update
type UpdateZoneRequest struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	FatherZoneID   string `json:"father_zone_id"`
	ZoneTypeID     string `json:"zone_type_id"`
	ProprietaryID  string `json:"proprietary_id"`
}

func updateZoneHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericUpdateHandler[UpdateZoneRequest](
		service,
		[]string{"ID", "Name", "ZoneTypeID", "ProprietaryID"},
		func(service q.NetServiceInt, req *UpdateZoneRequest) error {
			return service.UpdateZone(req.ID, req.Name, req.FatherZoneID, req.ZoneTypeID, req.ProprietaryID)
		},
	)
}

// Model Update
type UpdateModelRequest struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	BrandID         string `json:"brand_id"`
	DeviceClassID   string `json:"device_class_id"`
}

func updateModelHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericUpdateHandler[UpdateModelRequest](
		service,
		[]string{"ID", "Name", "BrandID", "DeviceClassID"},
		func(service q.NetServiceInt, req *UpdateModelRequest) error {
			return service.UpdateModel(req.ID, req.Name, req.BrandID, req.DeviceClassID)
		},
	)
}

// Device Update
type UpdateDeviceRequest struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	ModelID       string `json:"model_id"`
	ZoneID        string `json:"zone_id"`
	ProprietaryID string `json:"proprietary_id"`
}

func updateDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericUpdateHandler[UpdateDeviceRequest](
		service,
		[]string{"ID", "Label", "ModelID", "ZoneID", "ProprietaryID"},
		func(service q.NetServiceInt, req *UpdateDeviceRequest) error {
			return service.UpdateDevice(req.ID, req.Label, req.ModelID, req.ZoneID, req.ProprietaryID)
		},
	)
}

// ModelPort Update
type UpdateModelPortRequest struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PositionX string `json:"position_x"`
	PositionY string `json:"position_y"`
	ModelID   string `json:"model_id"`
}

func updateModelPortHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericUpdateHandler[UpdateModelPortRequest](
		service,
		[]string{"ID", "Name", "PositionX", "PositionY", "ModelID"},
		func(service q.NetServiceInt, req *UpdateModelPortRequest) error {
			return service.UpdateModelPort(req.ID, req.Name, req.PositionX, req.PositionY, req.ModelID)
		},
	)
}

// ConnectionType Update
type UpdateConnectionTypeRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func updateConnectionTypeHandler(service q.NetServiceInt) http.HandlerFunc {
	return genericUpdateHandler[UpdateConnectionTypeRequest](
		service,
		[]string{"ID", "Name"},
		func(service q.NetServiceInt, req *UpdateConnectionTypeRequest) error {
			return service.UpdateConnectionType(req.ID, req.Name)
		},
	)
}