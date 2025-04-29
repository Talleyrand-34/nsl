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
package specops

//
// import (
// 	"context"
// 	"testing"
//
// 	"github.com/stretchr/testify/require"
// )
//
// func TestSpecOps_ImportAllStructs(t *testing.T) {
// 	// Setup FIRST repo and fill with data
// 	repo1, repo2, err := setupTestRepository(t)
// 	require.NoError(t, err)
// 	defer repo1.Close()
//
// 	// Fill repo1 with some data
// 	require.NoError(t, repo2.AddProprietary("IT Department"))
// 	require.NoError(t, repo2.AddBrand("Fortinet"))
// 	require.NoError(t, repo2.AddDeviceClass("Router"))
// 	require.NoError(t, repo2.AddZoneType("Physical"))
// 	require.NoError(t, repo2.AddZone("HQ", "", "", "IT Department", "Physical"))
// 	require.NoError(t, repo2.AddModel("F100", "Fortinet", "Router"))
// 	require.NoError(t, repo2.AddModel("F200", "Fortinet", "Router"))
// 	require.NoError(t, repo2.AddModelPort("eth0", "10", "20", "F100"))
// 	require.NoError(t, repo2.AddModelPort("eth1", "15", "25", "F200"))
// 	require.NoError(t, repo2.AddDevice("MainRouter", "F100", "", "HQ", "IT Department"))
// 	require.NoError(t, repo2.AddDevice("BackupRouter", "F200", "", "HQ", "IT Department"))
// 	// require.NoError(t, repo2.AddPolicy("AllowAll", "Allow all traffic", ""))
//
// 	// Export from repo1
// 	ctx := context.Background()
// 	all, err := repo1.ExportAllStructs()
// 	require.NoError(t, err)
//
// 	// Setup SECOND repo (fresh, empty)
// 	repo3, _, err := setupTestRepository(t)
// 	require.NoError(t, err)
// 	defer repo3.Close()
//
// 	// Import into repo3
// 	err = repo3.ImportAllStructs(ctx, all)
// 	require.NoError(t, err)
//
// 	// Export again from repo3 for comparison
// 	all2, err := repo3.ExportAllStructs()
// 	require.NoError(t, err)
//
// 	// Check that key data is present after import
// 	require.Equal(t, len(all.Brands), len(all2.Brands))
// 	require.Equal(t, all.Brands[0].Name, all2.Brands[0].Name)
//
// 	require.Equal(t, len(all.Proprietaries), len(all2.Proprietaries))
// 	require.Equal(t, all.Proprietaries[0].Proprietary, all2.Proprietaries[0].Proprietary)
//
// 	require.Equal(t, len(all.DeviceClasses), len(all2.DeviceClasses))
// 	require.Equal(t, all.DeviceClasses[0].Name, all2.DeviceClasses[0].Name)
//
// 	require.Equal(t, len(all.ZoneTypes), len(all2.ZoneTypes))
// 	require.Equal(t, all.ZoneTypes[0].LocationType, all2.ZoneTypes[0].LocationType)
//
// 	require.Equal(t, len(all.Zones), len(all2.Zones))
// 	require.Equal(t, all.Zones[0].Name, all2.Zones[0].Name)
//
// 	require.Equal(t, len(all.ModelDevices), len(all2.ModelDevices))
// 	require.Equal(t, all.ModelDevices[0].Model, all2.ModelDevices[0].Model)
// 	require.Equal(t, all.ModelDevices[1].Model, all2.ModelDevices[1].Model)
//
// 	require.Equal(t, len(all.ModelPorts), len(all2.ModelPorts))
// 	require.Equal(t, all.ModelPorts[0].Name, all2.ModelPorts[0].Name)
// 	require.Equal(t, all.ModelPorts[1].Name, all2.ModelPorts[1].Name)
//
// 	require.Equal(t, len(all.Devices), len(all2.Devices))
// 	require.Equal(t, all.Devices[0].Label, all2.Devices[0].Label)
// 	require.Equal(t, all.Devices[1].Label, all2.Devices[1].Label)
//
// 	require.Equal(t, len(all.Policies), len(all2.Policies))
// 	require.Equal(t, all.Policies[0].Name, all2.Policies[0].Name)
//
// 	// You can add more detailed checks as needed
// }
