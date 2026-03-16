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
package cmd_scan

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// sshConfigEntry holds the resolved fields for a single SSH config host alias.
type sshConfigEntry struct {
	HostName     string // actual IP or hostname to connect to
	User         string
	IdentityFile string
	Port         int // 0 means not specified in config
}

// parseSSHConfigAlias looks up alias in ~/.ssh/config and returns the resolved
// connection details. Explicit SSH flags passed by the caller take precedence
// over the values returned here.
//
// Only exact Host matches are supported (no glob patterns). The match is
// case-insensitive.
func parseSSHConfigAlias(alias string) (*sshConfigEntry, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	configPath := filepath.Join(home, ".ssh", "config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", configPath, err)
	}

	entry := &sshConfigEntry{}
	inBlock := false
	found := false

	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)

		// Skip blank lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split key and value (allow '=' or whitespace as separator)
		line = strings.ReplaceAll(line, "=", " ")
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		if key == "host" {
			if inBlock {
				// We've left the matched block — stop parsing
				break
			}
			if strings.EqualFold(value, alias) {
				inBlock = true
				found = true
			}
			continue
		}

		if !inBlock {
			continue
		}

		switch key {
		case "hostname":
			entry.HostName = value
		case "user":
			entry.User = value
		case "identityfile":
			entry.IdentityFile = expandHome(value, home)
		case "port":
			if p, err := strconv.Atoi(value); err == nil {
				entry.Port = p
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("alias %q not found in %s", alias, configPath)
	}

	return entry, nil
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	if path == "~" {
		return home
	}
	return path
}
