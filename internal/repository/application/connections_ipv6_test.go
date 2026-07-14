package application

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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// listenLoopback starts a listener on addr and returns its host and port.
// It SKIPS the test when the address family is unavailable, so a host without IPv6
// does not turn this into a spurious failure.
func listenLoopback(t *testing.T, addr string) (host string, port int) {
	t.Helper()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Skipf("cannot listen on %s (address family unavailable here): %v", addr, err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	h, p, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)
	n, err := strconv.Atoi(p)
	require.NoError(t, err)

	return h, n
}

// TestTcpOpen_IPv6 is the end-to-end half of the IPv6 address fix; TestSSHAddress in
// internal/configparser is the unit half.
//
// tcpOpen built its address with fmt.Sprintf("%s:%d", ...), which renders the IPv6
// loopback as "::1:22" -- an address net.Dial rejects outright ("too many colons in
// address"). tcpOpen therefore reported every IPv6 host as unreachable, and
// SweepSubnet silently skipped them. net.JoinHostPort brackets the literal:
// "[::1]:22".
func TestTcpOpen_IPv6(t *testing.T) {
	host, port := listenLoopback(t, "[::1]:0")

	assert.True(t, tcpOpen(host, port, 2*time.Second),
		"tcpOpen must reach a listener on the IPv6 loopback")
}

func TestTcpOpen_IPv4(t *testing.T) {
	host, port := listenLoopback(t, "127.0.0.1:0")

	assert.True(t, tcpOpen(host, port, 2*time.Second),
		"tcpOpen must reach a listener on the IPv4 loopback")
}

// TestTcpOpen_Closed guards against the fix being "achieved" by making tcpOpen return
// true unconditionally.
func TestTcpOpen_Closed(t *testing.T) {
	// Bind a port, learn it, then release it -- so it is almost certainly closed.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	_, p, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)
	port, err := strconv.Atoi(p)
	require.NoError(t, err)
	require.NoError(t, ln.Close())

	assert.False(t, tcpOpen("127.0.0.1", port, 500*time.Millisecond),
		"tcpOpen must report a closed port as unreachable")
}
