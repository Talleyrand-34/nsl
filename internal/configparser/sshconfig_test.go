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
	"strings"
	"testing"
)

const sampleSSHConfig = `
# a comment
Host gw
    HostName 10.0.2.241
    User root
    IdentityFile ~/.ssh/lab_ed25519
    Port 2222

Host openwrt-1
    Hostname = 10.0.2.246
    User = admin
    IdentityFile ~/.ssh/openwrt.key
`

func TestParseSSHConfig_MultiHost(t *testing.T) {
	hosts, err := ParseSSHConfig(strings.NewReader(sampleSSHConfig))
	if err != nil {
		t.Fatalf("ParseSSHConfig: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d: %+v", len(hosts), hosts)
	}
	gw := hosts[0]
	if gw.Alias != "gw" || gw.HostName != "10.0.2.241" || gw.User != "root" {
		t.Errorf("gw block parsed wrong: %+v", gw)
	}
	if gw.IdentityFile != "~/.ssh/lab_ed25519" || gw.Port != 2222 {
		t.Errorf("gw identity/port parsed wrong: %+v", gw)
	}
	// "key = value" separator must be accepted too.
	ow := hosts[1]
	if ow.HostName != "10.0.2.246" || ow.User != "admin" || ow.IdentityFile != "~/.ssh/openwrt.key" {
		t.Errorf("openwrt-1 block parsed wrong: %+v", ow)
	}
}

func TestMatchSSHConfig(t *testing.T) {
	hosts, _ := ParseSSHConfig(strings.NewReader(sampleSSHConfig))

	// Match by HostName (an IP, as discovered in a sweep).
	if m := MatchSSHConfig(hosts, "10.0.2.246"); m == nil || m.Alias != "openwrt-1" {
		t.Errorf("expected HostName match for openwrt-1, got %+v", m)
	}
	// Fall back to alias match.
	if m := MatchSSHConfig(hosts, "gw"); m == nil || m.HostName != "10.0.2.241" {
		t.Errorf("expected alias match for gw, got %+v", m)
	}
	// No match.
	if m := MatchSSHConfig(hosts, "192.0.2.1"); m != nil {
		t.Errorf("expected no match, got %+v", m)
	}
}

func TestExpandHome(t *testing.T) {
	if got := ExpandHome("~/.ssh/id", "/home/u"); got != "/home/u/.ssh/id" {
		t.Errorf("ExpandHome(~/.ssh/id) = %q", got)
	}
	if got := ExpandHome("/abs/path", "/home/u"); got != "/abs/path" {
		t.Errorf("ExpandHome(/abs/path) should be unchanged, got %q", got)
	}
}
