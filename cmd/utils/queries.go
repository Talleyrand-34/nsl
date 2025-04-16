package cmd_utils

import (
	libsql "database/sql"

	"github.com/sirupsen/logrus"
)

func getAllConnectionsInfo(logger *logrus.Logger, db *libsql.DB) {
	// connections, err := customsql.FetchConnections(db)
	// if err != nil {
	// 	logger.Error(fmt.Sprintf("Error fetching connections: %v", err))
	// 	return
	// }
	//
	// formattedConnections := d2.FormatConnections(connections)
	// if logger.IsLevelEnabled(logrus.DebugLevel) { // Print only if logging level is DEBUG
	// 	logger.Debug(formattedConnections)
	// }
	// // Fetch devices and format them
	// devices, err := customsql.FetchDevices(db)
	// if err != nil {
	// 	logger.Error(fmt.Sprintf("Error fetching devices: %v", err))
	// 	return
	// }
	//
	// formattedDevices := d2.FormatDevices(devices)
	// if logger.IsLevelEnabled(logrus.DebugLevel) { // Print only if logging level is DEBUG
	// 	logger.Debug(formattedDevices)
	// }
	//
	// // Combine formatted connections and devices into a single D2 script
	// d2Script := formattedConnections + "\n" + formattedDevices
	//
	// d2.WriteDiagram(d2Script, outpath, outFile, outImage)
}

func getAllValidConnectionsInfo(logger *logrus.Logger, db *libsql.DB) {
	// connections, err := customsql.FetchValidConnections(db)
	// if err != nil {
	// 	logger.Error(fmt.Sprintf("Error fetching connections: %v", err))
	// 	return
	// }
	//
	// formattedConnections := d2.FormatConnections(connections)
	// if logger.IsLevelEnabled(logrus.DebugLevel) { // Print only if logging level is DEBUG
	// 	logger.Debug(formattedConnections)
	// }
	// // Fetch devices and format them
	// devices, err := customsql.FetchDevices(db)
	// if err != nil {
	// 	logger.Error(fmt.Sprintf("Error fetching devices: %v", err))
	// 	return
	// }
	//
	// formattedDevices := d2.FormatDevices(devices)
	// if logger.IsLevelEnabled(logrus.DebugLevel) { // Print only if logging level is DEBUG
	// 	logger.Debug(formattedDevices)
	// }
	//
	// // Combine formatted connections and devices into a single D2 script
	// d2Script := formattedConnections + "\n" + formattedDevices
	//
	// d2.WriteDiagram(d2Script, outpath, outFile, outImage)
}
