package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- Delete dependency check tests --- //

// ---- Brand ---- //

func TestDeleteBrand_BlocksWithDependentModel(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	if err := repo.DeleteBrand("Juniper"); err == nil {
		t.Errorf("expected error when deleting brand referenced by a model, got nil")
	}
}

func TestDeleteBrand_SucceedsWithoutDependents(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.DeleteBrand("Juniper"); err != nil {
		t.Errorf("expected success when deleting brand with no dependents, got: %v", err)
	}
}

func TestDeleteBrandCascade_SucceedsWithDependents(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	if err := repo.DeleteBrandCascade("Juniper"); err != nil {
		t.Errorf("expected cascade delete to succeed, got: %v", err)
	}

	brands, _ := repo.GetBrands()
	if brandSliceContains(brands, "Juniper") {
		t.Errorf("brand should have been deleted")
	}
}

// ---- ModelType ---- //

func TestDeleteModelType_BlocksWithDependentModel(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	if err := repo.DeleteModelType("Switch"); err == nil {
		t.Errorf("expected error when deleting device class referenced by a model, got nil")
	}
}

func TestDeleteModelType_SucceedsWithoutDependents(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.DeleteModelType("Switch"); err != nil {
		t.Errorf("expected success when deleting device class with no dependents, got: %v", err)
	}
}

// ---- ZoneType ---- //

func TestDeleteZoneType_BlocksWithDependentZone(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddZoneType("Building"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddZone("HQ", "", "", "", "Building"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}

	if err := repo.DeleteZoneType("Building"); err == nil {
		t.Errorf("expected error when deleting zone type referenced by a zone, got nil")
	}
}

func TestDeleteZoneType_SucceedsWithoutDependents(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddZoneType("Building"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.DeleteZoneType("Building"); err != nil {
		t.Errorf("expected success when deleting zone type with no dependents, got: %v", err)
	}
}

// ---- Zone ---- //

func TestDeleteZone_BlocksWithChildZone(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddZone("Parent", "", "", "", ""); err != nil {
		t.Fatalf("failed to add parent zone: %v", err)
	}

	zones, _ := repo.GetZones()
	var parentID string
	for _, z := range zones {
		if z.Name == "Parent" {
			parentID = z.ID
			break
		}
	}

	if err := repo.AddZone("Child", parentID, "", "", ""); err != nil {
		t.Fatalf("failed to add child zone: %v", err)
	}

	if err := repo.DeleteZone(parentID); err == nil {
		t.Errorf("expected error when deleting zone with child zones, got nil")
	}
}

func TestDeleteZone_BlocksWithDevice(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddZone("DC", "", "", "", ""); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}

	zones, _ := repo.GetZones()
	var zoneID string
	for _, z := range zones {
		if z.Name == "DC" {
			zoneID = z.ID
			break
		}
	}

	if err := repo.AddDevice("SW-01", "Catalyst 9300", zoneID, "", "", false, false); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	if err := repo.DeleteZone(zoneID); err == nil {
		t.Errorf("expected error when deleting zone with devices, got nil")
	}
}

func TestDeleteZone_SucceedsWithoutDependents(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddZone("EmptyZone", "", "", "", ""); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}

	zones, _ := repo.GetZones()
	var zoneID string
	for _, z := range zones {
		if z.Name == "EmptyZone" {
			zoneID = z.ID
			break
		}
	}

	if err := repo.DeleteZone(zoneID); err != nil {
		t.Errorf("expected success when deleting empty zone, got: %v", err)
	}
}

// ---- Model ---- //

func TestDeleteModel_BlocksWithDependentDevice(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	models, _ := repo.GetModels()
	var modelID string
	for _, m := range models {
		if m.Model == "Catalyst 9300" {
			modelID = m.ID
			break
		}
	}

	if err := repo.DeleteModel(modelID); err == nil {
		t.Errorf("expected error when deleting model referenced by a device, got nil")
	}
}

func TestDeleteModel_BlocksWithDependentModelPort(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/1", "0", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}

	models, _ := repo.GetModels()
	var modelID string
	for _, m := range models {
		if m.Model == "Catalyst 9300" {
			modelID = m.ID
			break
		}
	}

	if err := repo.DeleteModel(modelID); err == nil {
		t.Errorf("expected error when deleting model referenced by a model port, got nil")
	}
}

func TestDeleteModel_SucceedsWithoutDependents(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	models, _ := repo.GetModels()
	var modelID string
	for _, m := range models {
		if m.Model == "Catalyst 9300" {
			modelID = m.ID
			break
		}
	}

	if err := repo.DeleteModel(modelID); err != nil {
		t.Errorf("expected success when deleting model with no dependents, got: %v", err)
	}
}

// ---- Device ---- //

func TestDeleteDevice_BlocksWithDependentDevicePort(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/1", "0", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	devices, _ := repo.GetDevices()
	var deviceID string
	for _, d := range devices {
		if d.Label == "SW-01" {
			deviceID = d.ID
			break
		}
	}

	modelPorts, _ := repo.GetModelPorts()
	var modelPortID string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			modelPortID = mp.ID
			break
		}
	}

	if _, err := repo.AddDevicePort(deviceID, modelPortID, "", []e.PortVlanConfig{}); err != nil {
		t.Fatalf("failed to add device port: %v", err)
	}

	if err := repo.DeleteDevice(deviceID); err == nil {
		t.Errorf("expected error when deleting device with device ports, got nil")
	}
}

func TestDeleteDevice_BlocksWithDependentConnection(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/1", "0", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/2", "1", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device 1: %v", err)
	}
	if err := repo.AddDevice("SW-02", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device 2: %v", err)
	}

	devices, _ := repo.GetDevices()
	var device1ID, device2ID string
	for _, d := range devices {
		if d.Label == "SW-01" {
			device1ID = d.ID
		} else if d.Label == "SW-02" {
			device2ID = d.ID
		}
	}

	modelPorts, _ := repo.GetModelPorts()
	var port1ID, port2ID string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			port1ID = mp.ID
		} else if mp.Name == "Gi1/0/2" {
			port2ID = mp.ID
		}
	}

	// Connections require device ports to exist
	dp1ID, err := repo.AddDevicePort(device1ID, port1ID, "", []e.PortVlanConfig{})
	if err != nil {
		t.Fatalf("failed to add device port 1: %v", err)
	}
	dp2ID, err := repo.AddDevicePort(device2ID, port2ID, "", []e.PortVlanConfig{})
	if err != nil {
		t.Fatalf("failed to add device port 2: %v", err)
	}

	if err := repo.AddConnectionType("ethernet"); err != nil {
		t.Fatalf("failed to add connection type: %v", err)
	}
	if err := repo.AddConnection(dp1ID, dp2ID, "ethernet"); err != nil {
		t.Fatalf("failed to add connection: %v", err)
	}

	// Device has both device ports and a connection — deletion must be blocked
	if err := repo.DeleteDevice(device1ID); err == nil {
		t.Errorf("expected error when deleting device referenced by connections, got nil")
	}
}

func TestDeleteDevice_SucceedsWithoutDependents(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	devices, _ := repo.GetDevices()
	var deviceID string
	for _, d := range devices {
		if d.Label == "SW-01" {
			deviceID = d.ID
			break
		}
	}

	if err := repo.DeleteDevice(deviceID); err != nil {
		t.Errorf("expected success when deleting device with no dependents, got: %v", err)
	}
}

// ---- ModelPort ---- //

func TestDeleteModelPort_BlocksWithDependentDevicePort(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/1", "0", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	devices, _ := repo.GetDevices()
	var deviceID string
	for _, d := range devices {
		if d.Label == "SW-01" {
			deviceID = d.ID
			break
		}
	}

	modelPorts, _ := repo.GetModelPorts()
	var modelPortID string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			modelPortID = mp.ID
			break
		}
	}

	if _, err := repo.AddDevicePort(deviceID, modelPortID, "", []e.PortVlanConfig{}); err != nil {
		t.Fatalf("failed to add device port: %v", err)
	}

	if err := repo.DeleteModelPort(modelPortID); err == nil {
		t.Errorf("expected error when deleting model port referenced by a device port, got nil")
	}
}

func TestDeleteModelPort_SucceedsWithoutDependents(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/1", "0", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}

	modelPorts, _ := repo.GetModelPorts()
	var modelPortID string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			modelPortID = mp.ID
			break
		}
	}

	if err := repo.DeleteModelPort(modelPortID); err != nil {
		t.Errorf("expected success when deleting model port with no dependents, got: %v", err)
	}
}

// ---- DevicePort ---- //

func TestDeleteDevicePort_BlocksWithDependentConnection(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/1", "0", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port 1: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/2", "1", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port 2: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device 1: %v", err)
	}
	if err := repo.AddDevice("SW-02", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device 2: %v", err)
	}

	devices, _ := repo.GetDevices()
	var device1ID, device2ID string
	for _, d := range devices {
		if d.Label == "SW-01" {
			device1ID = d.ID
		} else if d.Label == "SW-02" {
			device2ID = d.ID
		}
	}

	modelPorts, _ := repo.GetModelPorts()
	var port1ID, port2ID string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			port1ID = mp.ID
		} else if mp.Name == "Gi1/0/2" {
			port2ID = mp.ID
		}
	}

	dp1ID, err := repo.AddDevicePort(device1ID, port1ID, "", []e.PortVlanConfig{})
	if err != nil {
		t.Fatalf("failed to add device port 1: %v", err)
	}
	dp2ID, err := repo.AddDevicePort(device2ID, port2ID, "", []e.PortVlanConfig{})
	if err != nil {
		t.Fatalf("failed to add device port 2: %v", err)
	}
	if err := repo.AddConnectionType("ethernet"); err != nil {
		t.Fatalf("failed to add connection type: %v", err)
	}
	if err := repo.AddConnection(dp1ID, dp2ID, "ethernet"); err != nil {
		t.Fatalf("failed to add connection: %v", err)
	}

	if err := repo.DeleteDevicePort(device1ID, port1ID); err == nil {
		t.Errorf("expected error when deleting device port referenced by a connection, got nil")
	}
}

func TestDeleteDevicePort_SucceedsWithoutDependents(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/1", "0", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	devices, _ := repo.GetDevices()
	var deviceID string
	for _, d := range devices {
		if d.Label == "SW-01" {
			deviceID = d.ID
			break
		}
	}

	modelPorts, _ := repo.GetModelPorts()
	var modelPortID string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			modelPortID = mp.ID
			break
		}
	}

	if _, err := repo.AddDevicePort(deviceID, modelPortID, "", []e.PortVlanConfig{}); err != nil {
		t.Fatalf("failed to add device port: %v", err)
	}

	if err := repo.DeleteDevicePort(deviceID, modelPortID); err != nil {
		t.Errorf("expected success when deleting device port with no connections, got: %v", err)
	}
}

// ---- Vlan ---- //

func TestDeleteVlan_CleansDeviceInterfaceVlanConfigs(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}
	if err := repo.AddVlan("20", "VLAN20", ""); err != nil {
		t.Fatalf("failed to add VLAN: %v", err)
	}

	devices, _ := repo.GetDevices()
	var deviceID string
	for _, d := range devices {
		if d.Label == "SW-01" {
			deviceID = d.ID
			break
		}
	}

	vlanConfigs := []e.PortVlanConfig{{VlanNumber: "20", Tagged: false}}
	if err := repo.AddDeviceInterface(deviceID, "Gi1/0/1.20", "", "", vlanConfigs, []string{}, "", ""); err != nil {
		t.Fatalf("failed to add device interface with VLAN config: %v", err)
	}

	vlans, _ := repo.GetVlans()
	var vlanDocID string
	for _, v := range vlans {
		if v.VlanID == "20" {
			vlanDocID = v.ID
			break
		}
	}

	if err := repo.DeleteVlan(vlanDocID); err != nil {
		t.Fatalf("expected VLAN delete to succeed, got: %v", err)
	}

	// Verify the device interface's vlan_configs no longer references VLAN 20
	ifaces, err := repo.GetDeviceInterfaces(deviceID)
	if err != nil {
		t.Fatalf("failed to get device interfaces: %v", err)
	}
	for _, iface := range ifaces {
		for _, vc := range iface.VlanConfigs {
			if vc.VlanNumber == "20" {
				t.Errorf("expected VLAN 20 to be removed from device interface vlan_configs, but it still exists")
			}
		}
	}
}

func TestDeleteVlan_SucceedsWithoutLocalVlans(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddVlan("30", "VLAN30", ""); err != nil {
		t.Fatalf("failed to add VLAN: %v", err)
	}

	vlans, _ := repo.GetVlans()
	var vlanDocID string
	for _, v := range vlans {
		if v.VlanID == "30" {
			vlanDocID = v.ID
			break
		}
	}

	if err := repo.DeleteVlan(vlanDocID); err != nil {
		t.Errorf("expected success when deleting VLAN, got: %v", err)
	}
}

// ---- Owner ---- //

func TestDeleteOwner_SetsNullOnZones(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddOwner("Acme Corp"); err != nil {
		t.Fatalf("failed to add owner: %v", err)
	}
	if err := repo.AddZone("HQ", "", "", "Acme Corp", ""); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}

	if err := repo.DeleteOwner("Acme Corp"); err != nil {
		t.Fatalf("expected success deleting owner (set-null), got: %v", err)
	}

	// Zone should still exist with empty owner
	zones, err := repo.GetZones()
	if err != nil {
		t.Fatalf("failed to get zones: %v", err)
	}
	found := false
	for _, z := range zones {
		if z.Name == "HQ" {
			found = true
			if z.Owner != "" {
				t.Errorf("expected zone owner to be cleared, got %q", z.Owner)
			}
			break
		}
	}
	if !found {
		t.Errorf("zone 'HQ' should still exist after owner deletion")
	}
}

func TestDeleteOwner_SetsNullOnDevices(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddOwner("Acme Corp"); err != nil {
		t.Fatalf("failed to add owner: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "Acme Corp", false, false); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	if err := repo.DeleteOwner("Acme Corp"); err != nil {
		t.Fatalf("expected success deleting owner (set-null), got: %v", err)
	}

	// Device should still exist with empty owner
	devices, err := repo.GetDevices()
	if err != nil {
		t.Fatalf("failed to get devices: %v", err)
	}
	found := false
	for _, d := range devices {
		if d.Label == "SW-01" {
			found = true
			if d.Owner != "" {
				t.Errorf("expected device owner to be cleared, got %q", d.Owner)
			}
			break
		}
	}
	if !found {
		t.Errorf("device 'SW-01' should still exist after owner deletion")
	}
}

func TestDeleteOwnerCascade_SetsNullNotCascadeDelete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddOwner("Acme Corp"); err != nil {
		t.Fatalf("failed to add owner: %v", err)
	}
	if err := repo.AddZone("HQ", "", "", "Acme Corp", ""); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "Acme Corp", false, false); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	if err := repo.DeleteOwnerCascade("Acme Corp"); err != nil {
		t.Fatalf("expected cascade delete to succeed, got: %v", err)
	}

	// Zone and device should still exist (set-null, not deleted)
	zones, _ := repo.GetZones()
	zoneFound := false
	for _, z := range zones {
		if z.Name == "HQ" {
			zoneFound = true
			if z.Owner != "" {
				t.Errorf("expected zone owner to be cleared, got %q", z.Owner)
			}
			break
		}
	}
	if !zoneFound {
		t.Errorf("zone 'HQ' should still exist after owner cascade deletion (set-null, not cascade-delete)")
	}

	devices, _ := repo.GetDevices()
	deviceFound := false
	for _, d := range devices {
		if d.Label == "SW-01" {
			deviceFound = true
			if d.Owner != "" {
				t.Errorf("expected device owner to be cleared, got %q", d.Owner)
			}
			break
		}
	}
	if !deviceFound {
		t.Errorf("device 'SW-01' should still exist after owner cascade deletion (set-null, not cascade-delete)")
	}
}
