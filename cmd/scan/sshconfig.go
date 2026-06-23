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
package cmd_scan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nsl-graph/internal/configparser"
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
	hosts, err := configparser.ParseSSHConfigFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", configPath, err)
	}
	for _, h := range hosts {
		if strings.EqualFold(h.Alias, alias) {
			return &sshConfigEntry{
				HostName:     h.HostName,
				User:         h.User,
				IdentityFile: h.IdentityFile,
				Port:         h.Port,
			}, nil
		}
	}
	return nil, fmt.Errorf("alias %q not found in %s", alias, configPath)
}
