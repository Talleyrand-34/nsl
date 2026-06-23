/*
Copyright © 2026 Talleyrand-34 (t34@t34.dev)

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

package configparser

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SSHConfigHost is one resolved Host block from an OpenSSH config file. It maps a
// host alias to the connection details a scan needs: the real address
// (HostName), the login User, the IdentityFile path, and an optional Port.
type SSHConfigHost struct {
	Alias        string
	HostName     string
	User         string
	IdentityFile string
	Port         int // 0 means unspecified
}

// ParseSSHConfig parses an OpenSSH config from r into one entry per Host block.
// Only exact host aliases are kept (glob patterns are recorded verbatim as the
// alias and matched literally). Directives are case-insensitive; both
// "key value" and "key = value" separators are accepted. Parsing never fails on
// unknown directives — they are ignored.
func ParseSSHConfig(r io.Reader) ([]SSHConfigHost, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var hosts []SSHConfigHost
	var cur *SSHConfigHost
	flush := func() {
		if cur != nil {
			hosts = append(hosts, *cur)
			cur = nil
		}
	}
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.ReplaceAll(line, "=", " ")
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		if key == "host" {
			flush()
			cur = &SSHConfigHost{Alias: value}
			continue
		}
		if cur == nil {
			continue
		}
		switch key {
		case "hostname":
			cur.HostName = value
		case "user":
			cur.User = value
		case "identityfile":
			cur.IdentityFile = value
		case "port":
			if p, err := strconv.Atoi(value); err == nil {
				cur.Port = p
			}
		}
	}
	flush()
	return hosts, nil
}

// ParseSSHConfigFile reads and parses an OpenSSH config file, expanding a leading
// "~" in IdentityFile paths against the current user's home directory.
func ParseSSHConfigFile(path string) ([]SSHConfigHost, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	hosts, err := ParseSSHConfig(f)
	if err != nil {
		return nil, err
	}
	if home, err := os.UserHomeDir(); err == nil {
		for i := range hosts {
			hosts[i].IdentityFile = ExpandHome(hosts[i].IdentityFile, home)
		}
	}
	return hosts, nil
}

// MatchSSHConfig returns the entry that applies to target, preferring an exact
// HostName match (target is usually a discovered IP) and falling back to an exact
// Alias match. Returns nil when nothing matches.
func MatchSSHConfig(hosts []SSHConfigHost, target string) *SSHConfigHost {
	for i := range hosts {
		if strings.EqualFold(hosts[i].HostName, target) {
			return &hosts[i]
		}
	}
	for i := range hosts {
		if strings.EqualFold(hosts[i].Alias, target) {
			return &hosts[i]
		}
	}
	return nil
}

// ExpandHome replaces a leading "~" in path with home.
func ExpandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	if path == "~" {
		return home
	}
	return path
}
