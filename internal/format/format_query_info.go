package format

import (
	"fmt"

	d "nslgraph/src/datastructs"
)

// PrintDeviceModelPortDetails prints the details of devices, models, and their ports.
func PrintDeviceModelPortDetails(data map[int][]d.DeviceModelPortDetails) {
	fmt.Println("DeviceID | DeviceName   | ModelID | ModelName    | ModelPortID   | ModelPortName")

	for deviceID, details := range data {
		for _, detail := range details {
			fmt.Printf(
				"%-9d | %-12s | %-7d | %-12s | %-13d | %s\n",
				deviceID,
				detail.DeviceLabel,
				detail.ModelID,
				detail.ModelName,
				detail.ModelPortID,
				detail.ModelPortName,
			)
		}
	}
}
