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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	s "nsl-graph/internal/scanner"
)

// --- doubles ------------------------------------------------------------------

type recordingSession struct {
	closed bool
	ran    []string
}

func (r *recordingSession) Execute(command string) (string, error) {
	r.ran = append(r.ran, command)
	return "output of " + command, nil
}
func (r *recordingSession) Close() error { r.closed = true; return nil }

type fakeTransport struct {
	sess     *recordingSession
	openErr  error
	openedAs string
	creds    SSHCredentials
}

func (f *fakeTransport) Name() string { return "fake" }
func (f *fakeTransport) Open(host string, creds SSHCredentials) (Session, error) {
	if f.openErr != nil {
		return nil, f.openErr
	}
	f.openedAs = host
	f.creds = creds
	f.sess = &recordingSession{}
	return f.sess, nil
}

// stubParser is the minimum a ConfigParser must be. Only Fetch matters here.
type stubParser struct {
	fetch func(Session) (string, error)
}

func (stubParser) GetOsType() string                { return "stub" }
func (stubParser) SupportsDevice(s.SNMPDevice) bool { return true }
func (p stubParser) Fetch(sess Session) (string, error) {
	return p.fetch(sess)
}
func (stubParser) ParseConfig(string, s.SNMPDevice) (*ConfigData, error) { return &ConfigData{}, nil }
func (stubParser) ValidateConfig(*ConfigData) []error                    { return nil }

// --- tests --------------------------------------------------------------------

// The division of labour: the transport reaches the device and knows nothing about it;
// the parser knows what to ask and nothing about how the connection was made.
func TestFetchConfig_OpensASessionAndLetsTheParserDriveIt(t *testing.T) {
	tr := &fakeTransport{}
	parser := stubParser{fetch: func(sess Session) (string, error) {
		return sess.Execute("show me the config")
	}}

	got, err := FetchConfig(tr, parser, "10.0.2.20", SSHCredentials{Username: "root", Port: 2222})

	require.NoError(t, err)
	assert.Equal(t, "output of show me the config", got)
	assert.Equal(t, "10.0.2.20", tr.openedAs)
	assert.Equal(t, "root", tr.creds.Username)
	assert.Equal(t, []string{"show me the config"}, tr.sess.ran)
}

// The session must be closed even when the parser gives up half-way, or a failing scan
// across a subnet leaks a connection per host.
func TestFetchConfig_ClosesTheSessionWhenTheParserFails(t *testing.T) {
	tr := &fakeTransport{}
	parser := stubParser{fetch: func(Session) (string, error) {
		return "", fmt.Errorf("device said no")
	}}

	_, err := FetchConfig(tr, parser, "10.0.2.20", SSHCredentials{})

	require.Error(t, err)
	assert.True(t, tr.sess.closed, "the session must be closed on the failure path too")
}

func TestFetchConfig_ClosesTheSessionOnSuccess(t *testing.T) {
	tr := &fakeTransport{}
	parser := stubParser{fetch: func(Session) (string, error) { return "config", nil }}

	_, err := FetchConfig(tr, parser, "10.0.2.20", SSHCredentials{})

	require.NoError(t, err)
	assert.True(t, tr.sess.closed)
}

// The error must say which transport, which OS and which host — a scan sweeps a subnet,
// so "connection refused" on its own names nothing.
func TestFetchConfig_ErrorNamesTheTransportOsAndHost(t *testing.T) {
	tr := &fakeTransport{}
	parser := stubParser{fetch: func(Session) (string, error) {
		return "", fmt.Errorf("device said no")
	}}

	_, err := FetchConfig(tr, parser, "10.0.2.20", SSHCredentials{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fake")
	assert.Contains(t, err.Error(), "stub")
	assert.Contains(t, err.Error(), "10.0.2.20")
	assert.Contains(t, err.Error(), "device said no")
}

func TestFetchConfig_PropagatesAConnectionFailure(t *testing.T) {
	tr := &fakeTransport{openErr: fmt.Errorf("connection refused")}
	parser := stubParser{fetch: func(Session) (string, error) {
		t.Fatal("the parser must not run when the connection failed")
		return "", nil
	}}

	_, err := FetchConfig(tr, parser, "10.0.2.20", SSHCredentials{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

// *SSHClient satisfies Session as it stands — that is why the split needed no adapter,
// and why it is cheap. If this stops compiling, the seam has drifted.
func TestSSHClientIsASession(t *testing.T) {
	var _ Session = (*SSHClient)(nil)
	var _ Transport = SSHTransport{}
	assert.Equal(t, "ssh", SSHTransport{}.Name())
}
