package specops

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpecOps_ExportAllStructs(t *testing.T) {
	repo, repo2, err := setupTestRepository(t)
	require.NoError(t, err)
	defer repo.Close()

	// Prepare referenced data: Proprietary, Brand, DeviceClass, Model, ZoneType, Zone
	require.NoError(t, repo2.AddProprietary("IT Department"))
	require.NoError(t, repo2.AddBrand("Fortinet"))
	require.NoError(t, repo2.AddDeviceClass("Router"))
	require.NoError(t, repo2.AddZoneType("Physical"))
	require.NoError(t, repo2.AddZone("HQ", "", "", "IT Department", "Physical"))
	require.NoError(t, repo2.AddModel("F100", "Fortinet", "Router"))
	require.NoError(t, repo2.AddModel("F200", "Fortinet", "Router"))

	// Add model ports for both models
	require.NoError(t, repo2.AddModelPort("eth0", "10", "20", "F100"))
	require.NoError(t, repo2.AddModelPort("eth1", "15", "25", "F200"))

	// Add devices
	require.NoError(t, repo2.AddDevice("MainRouter", "F100", "", "HQ", "IT Department"))
	require.NoError(t, repo2.AddDevice("BackupRouter", "F200", "", "HQ", "IT Department"))

	// Export all structs
	all, err := repo.ExportAllStructs()
	require.NoError(t, err)

	// --- Assertions ---

	// Brands
	require.Len(t, all.Brands, 1)
	require.Equal(t, "Fortinet", all.Brands[0].Name)

	// Proprietaries
	require.Len(t, all.Proprietaries, 1)
	require.Equal(t, "IT Department", all.Proprietaries[0].Proprietary)

	// DeviceClasses
	require.Len(t, all.DeviceClasses, 1)
	require.Equal(t, "Router", all.DeviceClasses[0].Name)

	// ZoneTypes
	require.Len(t, all.ZoneTypes, 1)
	require.Equal(t, "Physical", all.ZoneTypes[0].LocationType)

	// Zones
	require.Len(t, all.Zones, 1)
	require.Equal(t, "HQ", all.Zones[0].Name)

	// Models
	require.Len(t, all.ModelDevices, 2)
	models := map[string]bool{}
	for _, m := range all.ModelDevices {
		models[m.Model] = true
	}
	require.True(t, models["F100"])
	require.True(t, models["F200"])

	// ModelPorts
	require.Len(t, all.ModelPorts, 2)
	portNames := map[string]bool{}
	for _, mp := range all.ModelPorts {
		portNames[mp.Name] = true
	}
	require.True(t, portNames["eth0"])
	require.True(t, portNames["eth1"])

	// Devices
	require.Len(t, all.Devices, 2)
	deviceLabels := map[string]bool{}
	for _, d := range all.Devices {
		deviceLabels[d.Label] = true
	}
	require.True(t, deviceLabels["MainRouter"])
	require.True(t, deviceLabels["BackupRouter"])

	// DevicePorts, Connections, ConnectionTypes should be empty (since not added)
	require.Len(t, all.DevicePorts, 0)
	require.Len(t, all.Connections, 0)
	require.Len(t, all.ConnectionTypes, 0)
}
