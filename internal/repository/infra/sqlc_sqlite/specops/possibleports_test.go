package specops

import (
	"database/sql"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"nsl-graph/internal/repository/infra/sqlc_sqlite/basicops"
)

func setupTestRepository(
	t *testing.T,
) (SpecOpsSQLiteRepository, basicops.BasicOpsSQLiteRepository, error) {
	// Use in-memory DB for tests, or a temp file
	db, err := sql.Open("sqlite3", ":memory:")
	// db, err := sql.Open("sqlite3", "/tmp/test.db")
	// db, err := sql.Open("sqlite3", "file:memdb1?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	repo, err := NewSQLiteRepositoryFromDB(db)
	// repo, err := NewSQLiteRepository("/tmp/test.db")
	if err != nil {
		return SpecOpsSQLiteRepository{}, basicops.BasicOpsSQLiteRepository{}, err
	}
	repo2, err := basicops.NewSQLiteRepositoryFromDB(db)
	// repo, err := NewSQLiteRepository("/tmp/test.db")
	if err != nil {
		return SpecOpsSQLiteRepository{}, basicops.BasicOpsSQLiteRepository{}, err
	}

	return repo, repo2, nil
}

func TestSpecOps_GetAllPortsAllAndDevice(t *testing.T) {
	repo, repo2, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare referenced data: Proprietary, Brand, DeviceClass, Model, ZoneType, Zone
	if err := repo2.AddProprietary("IT Department"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo2.AddBrand("Fortinet"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo2.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo2.AddZoneType("Physical"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo2.AddZone("HQ", "", "", "IT Department", "Physical"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}
	if err := repo2.AddModel("F100", "Fortinet", "Router"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo2.AddModel("F200", "Fortinet", "Router"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Add model ports for both models
	if err := repo2.AddModelPort("eth0", "10", "20", "F100"); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo2.AddModelPort("eth1", "15", "25", "F200"); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}

	// Add devices
	if err := repo2.AddDevice("MainRouter", "F100", "", "HQ", "IT Department"); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}
	if err := repo2.AddDevice("BackupRouter", "F200", "", "HQ", "IT Department"); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	// Retrieve devices and model ports for ID reference
	devices, _ := repo2.GetDevices()
	modelPorts, _ := repo2.GetModelPorts()
	if len(devices) < 2 || len(modelPorts) < 2 {
		t.Fatalf("expected at least two devices and two model ports")
	}
	deviceID1 := devices[0].ID
	deviceID2 := devices[1].ID
	modelPortID1 := modelPorts[0].ID
	modelPortID2 := modelPorts[1].ID

	// --- Test GetAllPortsAll ---
	allPorts, _ := repo.GetAllPortsAll()
	expectedAll := map[string]bool{
		// All combinations: 2 devices x 2 ports
		key(deviceID1, modelPortID1): true,
		key(deviceID1, modelPortID2): true,
		key(deviceID2, modelPortID1): true,
		key(deviceID2, modelPortID2): true,
	}
	if len(allPorts) != 4 {
		t.Errorf("expected 4 device-port combinations, got %d: %+v", len(allPorts), allPorts)
	}
	for _, dp := range allPorts {
		k := key(dp.DeviceID, dp.ModelID)
		if !expectedAll[k] {
			t.Errorf(
				"unexpected device-port combination: device_id=%d, model_id=%d",
				dp.DeviceID,
				dp.ModelID,
			)
		}
		delete(expectedAll, k)
	}
	if len(expectedAll) != 0 {
		t.Errorf("missing device-port combinations: %+v", expectedAll)
	}

	// --- Test GetAllPortsDevice for deviceID1 ---
	allPortsDev1, _ := repo.GetAllPortsDevice(strconv.Itoa(deviceID1))
	expectedDev1 := map[string]bool{
		key(deviceID1, modelPortID1): true,
		key(deviceID1, modelPortID2): true,
	}
	if len(allPortsDev1) != 2 {
		t.Errorf(
			"expected 2 device-port combinations for device 1, got %d: %+v",
			len(allPortsDev1),
			allPortsDev1,
		)
	}
	for _, dp := range allPortsDev1 {
		k := key(dp.DeviceID, dp.ModelID)
		if !expectedDev1[k] {
			t.Errorf(
				"unexpected device-port combination for device 1: device_id=%d, model_id=%d",
				dp.DeviceID,
				dp.ModelID,
			)
		}
		delete(expectedDev1, k)
	}
	if len(expectedDev1) != 0 {
		t.Errorf("missing device-port combinations for device 1: %+v", expectedDev1)
	}

	// --- Test GetAllPortsDevice with invalid ID ---
	allPortsInvalid, _ := repo.GetAllPortsDevice("notanumber")
	if len(allPortsInvalid) != 0 {
		t.Errorf("expected 0 results for invalid device id, got %+v", allPortsInvalid)
	}
}

// helper to make map keys
func key(deviceID, modelID int) string {
	return strconv.Itoa(deviceID) + ":" + strconv.Itoa(modelID)
}
