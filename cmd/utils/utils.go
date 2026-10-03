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
package cmd_utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	c "nsl-graph/cmd"
	q "nsl-graph/internal/repository/application"
	e "nsl-graph/internal/repository/entities"

	infra "nsl-graph/internal/repository/infra/cloverdb/base"
)

// ParseVlanConfigs parses --vlan-configs entries of the form "100:tagged" /
// "100:untagged" (or just "100", which defaults to tagged) into PortVlanConfigs.
// Empty entries are skipped.
func ParseVlanConfigs(specs []string) []e.PortVlanConfig {
	var out []e.PortVlanConfig
	for _, s := range specs {
		if s == "" {
			continue
		}
		parts := strings.Split(s, ":")
		tagged := true
		if len(parts) == 2 && strings.ToLower(parts[1]) == "untagged" {
			tagged = false
		}
		out = append(out, e.PortVlanConfig{VlanNumber: parts[0], Tagged: tagged})
	}
	return out
}

func OpenDatabaseConnectionClover() {
	os.Mkdir(c.Srcdbpath, 0755)
}

// Return repositoryDB connection using global config
func ServiceConnection() (q.NetServiceInt, error) {
	return GetServiceConnection(c.Srcdbpath)
}

// Return repositoryDB connection with custom database path
func GetServiceConnection(dbPath string) (q.NetServiceInt, error) {
	backend := strings.ToLower(c.Backend)

	switch backend {
	case "cloverdb", "clover", "":
		// Use CloverDB backend (the only supported backend).
		// NewCloverRepository creates the directory if needed and logs the outcome.
		baseRepo, err := infra.NewCloverRepository(dbPath)
		if err != nil {
			log.Fatalf("Error connecting to CloverDB: %v", err)
		}

		// Create service with base repository
		service := q.NewNetService(baseRepo)
		return service, nil

	default:
		return nil, fmt.Errorf("unsupported backend type: %s (supported: cloverdb)", backend)
	}
}

// PrintPrettyJSON prints the JSON data in a human-readable format.
func PrintPrettyJSON(jsonData string) error {
	var prettyData map[string]interface{}

	// Unmarshal the JSON string into a map for formatting
	err := json.Unmarshal([]byte(jsonData), &prettyData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Marshal the map back into a pretty-printed JSON string
	prettyJSON, err := json.MarshalIndent(prettyData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(prettyJSON))
	return nil
}

func PrintStringArrayPrettyJson(arr []string) {
	b, _ := json.MarshalIndent(arr, "", "  ")
	fmt.Println(string(b))
}

// flagproc gets the values of the given flag names from the command and returns them as a slice of strings.
func Flagproc(cmd *cobra.Command, flagNames []string) []string {
	values := make([]string, len(flagNames))
	for i, flag := range flagNames {
		val, err := cmd.Flags().GetString(flag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading flag '%v': %v\n", flag, err)
			os.Exit(1)
		}
		values[i] = val
	}
	return values
}
