// SPDX-License-Identifier: AGPL-3.0-or-later
/*
Copyright © 2025 Talleyrand-34 (t34@t34.dev)

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
package basicops

import (
	"fmt"

	"github.com/google/uuid"
	d "github.com/ostafen/clover/v2/document"
	q "github.com/ostafen/clover/v2/query"

	e "nsl-graph/internal/repository/entities"
)

const (
	deviceInterfacesCollection = "device_interfaces"
	interfacePortsCollection   = "interface_ports"
)

// AddDeviceInterface adds a new logical interface for a device
func (r BasicOpsCloverRepository) AddDeviceInterface(deviceID, name, description, parent string, vlanConfigs []e.PortVlanConfig, ips []string, wifiSSID, wifiSecurity string) error {
	id := uuid.New().String()

	doc := d.NewDocument()
	doc.Set("_id", id)
	doc.Set("device_id", deviceID)
	doc.Set("name", name)
	doc.Set("description", description)
	if parent != "" {
		doc.Set("parent", parent)
	}

	if len(vlanConfigs) > 0 {
		vlanConfigsMap := make([]map[string]interface{}, len(vlanConfigs))
		for i, vc := range vlanConfigs {
			vlanConfigsMap[i] = map[string]interface{}{
				"vlan_number": vc.VlanNumber,
				"tagged":      vc.Tagged,
			}
		}
		doc.Set("vlan_configs", vlanConfigsMap)
	}

	if len(ips) > 0 {
		doc.Set("ip_addresses", ips)
	}
	if wifiSSID != "" {
		doc.Set("wifi_ssid", wifiSSID)
	}
	if wifiSecurity != "" {
		doc.Set("wifi_security", wifiSecurity)
	}

	_, err := r.db.InsertOne(deviceInterfacesCollection, doc)
	if err != nil {
		return fmt.Errorf("AddDeviceInterface failed: %w", err)
	}
	return nil
}

// GetDeviceInterfaces returns all logical interfaces for a specific device
func (r BasicOpsCloverRepository) GetDeviceInterfaces(deviceID string) ([]e.DeviceInterface, error) {
	docs, err := r.db.FindAll(q.NewQuery(deviceInterfacesCollection).Where(q.Field("device_id").Eq(deviceID)))
	if err != nil {
		return nil, fmt.Errorf("GetDeviceInterfaces failed: %w", err)
	}
	return r.docsToDeviceInterfaces(docs), nil
}

// GetAllDeviceInterfaces returns all logical interfaces across all devices
func (r BasicOpsCloverRepository) GetAllDeviceInterfaces() ([]e.DeviceInterface, error) {
	docs, err := r.db.FindAll(q.NewQuery(deviceInterfacesCollection))
	if err != nil {
		return nil, fmt.Errorf("GetAllDeviceInterfaces failed: %w", err)
	}
	return r.docsToDeviceInterfaces(docs), nil
}

// UpdateDeviceInterface updates the VLAN configurations of a logical interface
func (r BasicOpsCloverRepository) UpdateDeviceInterface(id string, vlanConfigs []e.PortVlanConfig) error {
	updates := make(map[string]interface{})

	if len(vlanConfigs) > 0 {
		vlanConfigsMap := make([]map[string]interface{}, len(vlanConfigs))
		for i, vc := range vlanConfigs {
			vlanConfigsMap[i] = map[string]interface{}{
				"vlan_number": vc.VlanNumber,
				"tagged":      vc.Tagged,
			}
		}
		updates["vlan_configs"] = vlanConfigsMap
	} else {
		updates["vlan_configs"] = []map[string]interface{}{}
	}

	err := r.db.Update(q.NewQuery(deviceInterfacesCollection).Where(q.Field("_id").Eq(id)), updates)
	if err != nil {
		return fmt.Errorf("UpdateDeviceInterface failed: %w", err)
	}
	return nil
}

// DeleteDeviceInterface deletes a logical interface and its related InterfacePort entries
func (r BasicOpsCloverRepository) DeleteDeviceInterface(id string) error {
	// Delete all InterfacePort entries for this interface
	if err := r.db.Delete(q.NewQuery(interfacePortsCollection).Where(q.Field("interface_id").Eq(id))); err != nil {
		return fmt.Errorf("DeleteDeviceInterface: failed to delete interface ports: %w", err)
	}

	// Delete the interface itself
	if err := r.db.Delete(q.NewQuery(deviceInterfacesCollection).Where(q.Field("_id").Eq(id))); err != nil {
		return fmt.Errorf("DeleteDeviceInterface failed: %w", err)
	}
	return nil
}

// AddInterfacePort links a logical interface to a physical device port.
// VLAN configurations and IP addresses are NOT stored here; they belong to the
// DeviceInterface and are resolved via interface_id.
func (r BasicOpsCloverRepository) AddInterfacePort(interfaceID, deviceID, modelPortID string) error {
	id := uuid.New().String()

	doc := d.NewDocument()
	doc.Set("_id", id)
	doc.Set("interface_id", interfaceID)
	doc.Set("device_id", deviceID)
	doc.Set("model_port_id", modelPortID)

	_, err := r.db.InsertOne(interfacePortsCollection, doc)
	if err != nil {
		return fmt.Errorf("AddInterfacePort failed: %w", err)
	}
	return nil
}

// GetInterfacePortsByInterface returns all InterfacePort entries for a given interface ID
func (r BasicOpsCloverRepository) GetInterfacePortsByInterface(interfaceID string) ([]e.InterfacePort, error) {
	docs, err := r.db.FindAll(q.NewQuery(interfacePortsCollection).Where(q.Field("interface_id").Eq(interfaceID)))
	if err != nil {
		return nil, fmt.Errorf("GetInterfacePortsByInterface failed: %w", err)
	}
	return r.docsToInterfacePorts(docs), nil
}

// GetInterfacePortsByPort returns all InterfacePort entries for a given device+model-port combination
func (r BasicOpsCloverRepository) GetInterfacePortsByPort(deviceID, modelPortID string) ([]e.InterfacePort, error) {
	docs, err := r.db.FindAll(q.NewQuery(interfacePortsCollection).Where(
		q.Field("device_id").Eq(deviceID).And(
			q.Field("model_port_id").Eq(modelPortID),
		),
	))
	if err != nil {
		return nil, fmt.Errorf("GetInterfacePortsByPort failed: %w", err)
	}
	return r.docsToInterfacePorts(docs), nil
}

// GetAllInterfacePorts returns every InterfacePort record in the collection
func (r BasicOpsCloverRepository) GetAllInterfacePorts() ([]e.InterfacePort, error) {
	docs, err := r.db.FindAll(q.NewQuery(interfacePortsCollection))
	if err != nil {
		return nil, fmt.Errorf("GetAllInterfacePorts failed: %w", err)
	}
	return r.docsToInterfacePorts(docs), nil
}

// DeleteInterfacePort removes a single InterfacePort entry identified by
// interface ID, device ID and model port ID.
func (r BasicOpsCloverRepository) DeleteInterfacePort(interfaceID, deviceID, modelPortID string) error {
	err := r.db.Delete(q.NewQuery(interfacePortsCollection).Where(
		q.Field("interface_id").Eq(interfaceID).And(
			q.Field("device_id").Eq(deviceID).And(
				q.Field("model_port_id").Eq(modelPortID),
			),
		),
	))
	if err != nil {
		return fmt.Errorf("DeleteInterfacePort failed: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func (r BasicOpsCloverRepository) docsToDeviceInterfaces(docs []*d.Document) []e.DeviceInterface {
	result := make([]e.DeviceInterface, 0, len(docs))
	for _, doc := range docs {
		iface := e.DeviceInterface{}

		if id, ok := doc.Get("_id").(string); ok {
			iface.ID = id
		} else {
			iface.ID = doc.ObjectId()
		}
		if deviceID, ok := doc.Get("device_id").(string); ok {
			iface.DeviceID = deviceID
		}
		if name, ok := doc.Get("name").(string); ok {
			iface.Name = name
		}
		if desc, ok := doc.Get("description").(string); ok {
			iface.Description = desc
		}
		if parent, ok := doc.Get("parent").(string); ok {
			iface.Parent = parent
		}

		iface.VlanConfigs = make([]e.PortVlanConfig, 0)
		if raw, ok := doc.Get("vlan_configs").([]interface{}); ok {
			for _, item := range raw {
				if m, ok := item.(map[string]interface{}); ok {
					vc := e.PortVlanConfig{}
					if vn, ok := m["vlan_number"].(string); ok {
						vc.VlanNumber = vn
					}
					if t, ok := m["tagged"].(bool); ok {
						vc.Tagged = t
					}
					iface.VlanConfigs = append(iface.VlanConfigs, vc)
				}
			}
		}

		iface.IPAddresses = make([]string, 0)
		if raw, ok := doc.Get("ip_addresses").([]interface{}); ok {
			for _, item := range raw {
				if ip, ok := item.(string); ok && ip != "" {
					iface.IPAddresses = append(iface.IPAddresses, ip)
				}
			}
		}

		if v, ok := doc.Get("wifi_ssid").(string); ok {
			iface.WifiSSID = v
		}
		if v, ok := doc.Get("wifi_security").(string); ok {
			iface.WifiSecurity = v
		}

		result = append(result, iface)
	}
	return result
}

func (r BasicOpsCloverRepository) docsToInterfacePorts(docs []*d.Document) []e.InterfacePort {
	result := make([]e.InterfacePort, 0, len(docs))
	for _, doc := range docs {
		ip := e.InterfacePort{}

		if id, ok := doc.Get("_id").(string); ok {
			ip.ID = id
		} else {
			ip.ID = doc.ObjectId()
		}
		if v, ok := doc.Get("interface_id").(string); ok {
			ip.InterfaceID = v
		}
		if v, ok := doc.Get("device_id").(string); ok {
			ip.DeviceID = v
		}
		if v, ok := doc.Get("model_port_id").(string); ok {
			ip.ModelPortID = v
		}

		result = append(result, ip)
	}
	return result
}
