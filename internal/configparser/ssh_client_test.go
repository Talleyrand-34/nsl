package configparser

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

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSSHAddress covers the IPv6 defect that `go vet` reported as
//
//	address format "%s:%d" does not work with IPv6
//
// fmt.Sprintf("%s:%d", "::1", 22) yields "::1:22", which is not an address any
// dialler can parse: the colons of the IPv6 literal are indistinguishable from the
// port separator. The host must be bracketed -- "[::1]:22" -- which is exactly what
// net.JoinHostPort exists to do.
func TestSSHAddress(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{"IPv4 with explicit port", "10.0.2.20", 2222, "10.0.2.20:2222"},
		{"IPv4 defaults to 22", "10.0.2.20", 0, "10.0.2.20:22"},
		{"hostname defaults to 22", "opnsense", 0, "opnsense:22"},
		{"hostname with explicit port", "opnsense", 2222, "opnsense:2222"},

		// The regression. Without bracketing these are unparseable.
		{"IPv6 loopback defaults to 22", "::1", 0, "[::1]:22"},
		{"IPv6 with explicit port", "::1", 2222, "[::1]:2222"},
		{"IPv6 full address", "fe80::1ff:fe23:4567:890a", 22, "[fe80::1ff:fe23:4567:890a]:22"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sshAddress(tt.host, tt.port)
			assert.Equal(t, tt.want, got)

			// Whatever we build must round-trip through the standard library, or a
			// dial with it cannot succeed. This is the property that actually
			// matters; the string comparison above just pins the spelling.
			gotHost, gotPort, err := net.SplitHostPort(got)
			require.NoError(t, err, "address %q is not parseable by net.SplitHostPort", got)
			assert.Equal(t, tt.host, gotHost)
			assert.NotEmpty(t, gotPort)
		})
	}
}
