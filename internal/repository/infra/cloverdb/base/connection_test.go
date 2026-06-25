package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- Connection Tests --- //

func TestConnection_AddAndGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddBrand("Cisco"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("Catalyst 9300", "Cisco", "Switch"); err != nil {
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

	// Get IDs
	devices, _ := repo.GetDevices()
	var device1Id, device2Id string
	for _, d := range devices {
		if d.Label == "SW-01" {
			device1Id = d.ID
		} else if d.Label == "SW-02" {
			device2Id = d.ID
		}
	}

	modelPorts, _ := repo.GetModelPorts()
	var port1Id, port2Id string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			port1Id = mp.ID
		} else if mp.Name == "Gi1/0/2" {
			port2Id = mp.ID
		}
	}

	// Add device ports (required for connections)
	dp1Id, err := repo.AddDevicePort(device1Id, port1Id, "", []e.PortVlanConfig{})
	if err != nil {
		t.Fatalf("failed to add device port 1: %v", err)
	}
	dp2Id, err := repo.AddDevicePort(device2Id, port2Id, "", []e.PortVlanConfig{})
	if err != nil {
		t.Fatalf("failed to add device port 2: %v", err)
	}

	// Add connection
	if err := repo.AddConnectionType("ethernet"); err != nil {
		t.Fatalf("failed to add connection type: %v", err)
	}
	if err := repo.AddConnection(dp1Id, dp2Id, "ethernet"); err != nil {
		t.Errorf("failed to add connection: %v", err)
	}

	// Get and verify
	connections, err := repo.GetConnections()
	if err != nil {
		t.Errorf("failed to get connections: %v", err)
	}

	if len(connections) == 0 {
		t.Errorf("expected at least one connection, got none")
	}
}

func TestConnection_AddDuplicatePort(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddBrand("Cisco"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("Catalyst 9300", "Cisco", "Switch"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/1", "0", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port 1: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/2", "1", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port 2: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/3", "2", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port 3: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device 1: %v", err)
	}
	if err := repo.AddDevice("SW-02", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device 2: %v", err)
	}
	if err := repo.AddDevice("SW-03", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device 3: %v", err)
	}

	// Get IDs
	devices, _ := repo.GetDevices()
	var device1Id, device2Id, device3Id string
	for _, d := range devices {
		if d.Label == "SW-01" {
			device1Id = d.ID
		} else if d.Label == "SW-02" {
			device2Id = d.ID
		} else if d.Label == "SW-03" {
			device3Id = d.ID
		}
	}

	modelPorts, _ := repo.GetModelPorts()
	var port1Id, port2Id, port3Id string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			port1Id = mp.ID
		} else if mp.Name == "Gi1/0/2" {
			port2Id = mp.ID
		} else if mp.Name == "Gi1/0/3" {
			port3Id = mp.ID
		}
	}

	// Add device ports (required for connections)
	dp1Id, err := repo.AddDevicePort(device1Id, port1Id, "", []e.PortVlanConfig{})
	if err != nil {
		t.Fatalf("failed to add device port 1: %v", err)
	}
	dp2Id, err := repo.AddDevicePort(device2Id, port2Id, "", []e.PortVlanConfig{})
	if err != nil {
		t.Fatalf("failed to add device port 2: %v", err)
	}
	dp3Id, err := repo.AddDevicePort(device3Id, port3Id, "", []e.PortVlanConfig{})
	if err != nil {
		t.Fatalf("failed to add device port 3: %v", err)
	}

	// Add first connection
	if err := repo.AddConnectionType("ethernet"); err != nil {
		t.Fatalf("failed to add connection type: %v", err)
	}
	if err := repo.AddConnection(dp1Id, dp2Id, "ethernet"); err != nil {
		t.Fatalf("failed to add first connection: %v", err)
	}

	// Try to add second connection using same port - should fail
	if err := repo.AddConnection(dp1Id, dp3Id, "ethernet"); err == nil {
		t.Errorf("expected error when reusing port, got nil")
	}
}

func TestConnection_Delete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddBrand("Cisco"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("Catalyst 9300", "Cisco", "Switch"); err != nil {
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

	// Get IDs
	devices, _ := repo.GetDevices()
	var device1Id, device2Id string
	for _, d := range devices {
		if d.Label == "SW-01" {
			device1Id = d.ID
		} else if d.Label == "SW-02" {
			device2Id = d.ID
		}
	}

	modelPorts, _ := repo.GetModelPorts()
	var port1Id, port2Id string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			port1Id = mp.ID
		} else if mp.Name == "Gi1/0/2" {
			port2Id = mp.ID
		}
	}

	// Add device ports (required for connections)
	dp1Id, err := repo.AddDevicePort(device1Id, port1Id, "", []e.PortVlanConfig{})
	if err != nil {
		t.Fatalf("failed to add device port 1: %v", err)
	}
	dp2Id, err := repo.AddDevicePort(device2Id, port2Id, "", []e.PortVlanConfig{})
	if err != nil {
		t.Fatalf("failed to add device port 2: %v", err)
	}

	// Add connection
	if err := repo.AddConnectionType("ethernet"); err != nil {
		t.Fatalf("failed to add connection type: %v", err)
	}
	if err := repo.AddConnection(dp1Id, dp2Id, "ethernet"); err != nil {
		t.Fatalf("failed to add connection: %v", err)
	}

	// Get connection ID
	connections, _ := repo.GetConnections()
	if len(connections) == 0 {
		t.Fatalf("expected at least one connection")
	}
	connectionId := connections[0].ID

	// Delete connection
	if err := repo.DeleteConnection(connectionId); err != nil {
		t.Errorf("failed to delete connection: %v", err)
	}

	// Verify deletion
	connectionsAfterDelete, err := repo.GetConnections()
	if err != nil {
		t.Errorf("failed to get connections after deletion: %v", err)
	}

	if len(connectionsAfterDelete) > 0 {
		t.Errorf("connection should have been deleted, but got %d connections", len(connectionsAfterDelete))
	}
}
