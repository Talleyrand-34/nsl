package format

// // Test FormatDevices function.
// func TestFormatDevices(t *testing.T) {
// 	devices := []d.Device{
// 		{
// 			Label:         "DeviceA",
// 			ZoneHierarchy: []string{"Zone1", "Zone2"},
// 			Shape:         "rectangle",
// 			Ports: map[string]d.Port{
// 				"1": {ID: "1", Name: "Port1", PositionX: 10, PositionY: 20},
// 				"2": {ID: "2", Name: "Port2", PositionX: 30, PositionY: 40},
// 			},
// 		},
// 	}
//
// 	expected := `Zone1.Zone2.DeviceA: {
//   shape: rectangle
//   label: "DeviceA"
//   1: "Port1"
//   2: "Port2"
// }
// `
//
// 	result := FormatDevices(devices)
// 	if result != expected {
// 		t.Errorf("FormatDevices() = \n%s\nExpected:\n%s", result, expected)
// 	}
// }
//
// // Test FormatDevicesToJSON function.
// func TestFormatDevicesToJSON(t *testing.T) {
// 	devices := []d.Device{
// 		{
// 			Label:         "DeviceA",
// 			ZoneHierarchy: []string{"Zone1", "Zone2"},
// 			Shape:         "rectangle",
// 			Ports: map[string]d.Port{
// 				"1": {ID: "1", Name: "Port1", PositionX: 10, PositionY: 20},
// 				"2": {ID: "2", Name: "Port2", PositionX: 30, PositionY: 40},
// 			},
// 		},
// 	}
//
// 	expected := `{
//   "Zone1.Zone2.DeviceA": {
//     "1": {
//       "ID": "1",
//       "Name": "Port1",
//       "PositionX": 10,
//       "PositionY": 20
//     },
//     "2": {
//       "ID": "2",
//       "Name": "Port2",
//       "PositionX": 30,
//       "PositionY": 40
//     }
//   }
// }`
//
// 	result, err := FormatDevicesToJSON(devices)
// 	if err != nil {
// 		t.Fatalf("FormatDevicesToJSON() returned an error: %v", err)
// 	}
//
// 	if result != expected {
// 		t.Errorf("FormatDevicesToJSON() = \n%s\nExpected:\n%s", result, expected)
// 	}
// }
//
// // Test FormatConnections function.
// func TestFormatConnections(t *testing.T) {
// 	connections := []d.Connection{
// 		{
// 			ConnectionID: 1,
// 			FromZone:     []string{"Zone1", "Zone2"},
// 			FromDevice:   "DeviceA",
// 			FromPort:     d.Port{ID: "1", Name: "Port1"},
// 			ToZone:       []string{"Zone3"},
// 			ToDevice:     "DeviceB",
// 			ToPort:       d.Port{ID: "2", Name: "Port2"},
// 		},
// 	}
//
// 	expected := `Zone1.Zone2.DeviceA.Port1 -- Zone3.DeviceB.Port2
// `
//
// 	result := FormatConnections(connections)
// 	if result != expected {
// 		t.Errorf("FormatConnections() = \n%s\nExpected:\n%s", result, expected)
// 	}
// }
