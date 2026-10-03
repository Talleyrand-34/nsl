// transport.go: the seam between HOW we reach a device and WHAT we say to it.
// SPDX-License-Identifier: AGPL-3.0-or-later
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

import "fmt"

// Session is an open connection to a device, on which commands can be run.
//
// This is the seam. Before it existed, every OS parser was also an SSH client: five
// copies of NewSSHClient/Connect/Execute/Close, one per vendor, differing only in the
// commands they ran. The connection logic is not OS knowledge and does not belong in a
// parser -- but the COMMANDS are, and they stay there.
//
// It matters for two reasons beyond tidiness:
//
//   - the write path can reuse it. A renderer that pushes UCI over SSH needs exactly
//     this, and without the seam it would have grown its own SSH client, the way every
//     parser already had.
//
//   - the fetch strategies become testable. Every parser's fallback chain -- OPNsense's
//     `cat /conf/config.xml` falling back to `configctl`, FreeBSD's CIDR ifconfig falling
//     back to plain, Fortinet's targeted commands falling back to the full config -- used
//     to require a real device on the far end of a real TCP connection, so none of them
//     was ever tested. A fake Session tests them all.
//
// *SSHClient satisfies this as it stands; no adapter is needed.
type Session interface {
	// Execute runs a command and returns its combined output. A non-nil error may still
	// come with output -- callers rely on that to report what the device said.
	Execute(command string) (string, error)
	Close() error
}

// Transport opens a Session to a device.
type Transport interface {
	// Name identifies the transport in logs and errors ("ssh").
	Name() string
	Open(host string, creds SSHCredentials) (Session, error)
}

// SSHTransport reaches devices over SSH. It is the only transport there is, and by
// design: SNMP is read-only forever (writable MIBs were never deployed -- see RFC 3535),
// so a device is configured over SSH or not at all.
type SSHTransport struct{}

func (SSHTransport) Name() string { return "ssh" }

func (SSHTransport) Open(host string, creds SSHCredentials) (Session, error) {
	client := NewSSHClient(creds)
	if err := client.Connect(host); err != nil {
		return nil, fmt.Errorf("ssh: connecting to %s: %w", host, err)
	}
	return client, nil
}

// DefaultTransport is what the read path uses, and what the write path will.
var DefaultTransport Transport = SSHTransport{}

// FetchConfig opens a session to host and lets the parser assemble the device's raw
// configuration from it.
//
// The division of labour: the transport knows how to reach the device and nothing about
// it; the parser knows what to ask and nothing about how the connection is made.
func FetchConfig(t Transport, parser ConfigParser, host string, creds SSHCredentials) (string, error) {
	sess, err := t.Open(host, creds)
	if err != nil {
		return "", err
	}
	defer sess.Close()

	raw, err := parser.Fetch(sess)
	if err != nil {
		return "", fmt.Errorf("%s: fetching %s configuration from %s: %w",
			t.Name(), parser.GetOsType(), host, err)
	}
	return raw, nil
}
