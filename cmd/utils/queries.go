
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
