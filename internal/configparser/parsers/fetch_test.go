package parsers

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
)

// fakeSession answers commands from a table instead of a device.
//
// Every fetch strategy in this package is a chain of fallbacks -- the interesting,
// failure-handling half of talking to a switch -- and until the transport seam existed
// not one of them could be tested, because each parser opened its own SSH connection to
// a real host. This is what the split bought.
type fakeSession struct {
	out  map[string]string // command -> output
	fail map[string]error  // command -> error (output may still be returned alongside)
	ran  []string
}

func newFakeSession() *fakeSession {
	return &fakeSession{out: map[string]string{}, fail: map[string]error{}}
}

func (f *fakeSession) Execute(command string) (string, error) {
	f.ran = append(f.ran, command)
	if err, ok := f.fail[command]; ok {
		return f.out[command], err
	}
	if out, ok := f.out[command]; ok {
		return out, nil
	}
	// An unknown command is a device that does not have it -- which is exactly what the
	// fallbacks exist for.
	return "", fmt.Errorf("sh: %s: not found", command)
}

func (f *fakeSession) Close() error { return nil }

// -----------------------------------------------------------------------------
// FreeBSD
// -----------------------------------------------------------------------------

func TestFreeBSDFetch_PrefersTheCIDRFormat(t *testing.T) {
	sess := newFakeSession()
	sess.out["hostname"] = "fw1\n"
	sess.out["ifconfig -a -f inet:cidr,inet6:cidr"] = "igc0: flags=8843\n\tinet 10.0.2.1/24\n"

	got, err := NewFreeBSDParser().Fetch(sess)

	require.NoError(t, err)
	assert.Contains(t, got, "HOSTNAME:fw1")
	assert.Contains(t, got, "10.0.2.1/24", "the CIDR form carries the prefix length")
	assert.NotContains(t, sess.ran, "ifconfig -a", "no need for the fallback")
}

// Older hosts have no -f flag. The fallback loses the netmask, which is a real cost --
// VLAN/subnet inference is degraded for them -- but it is better than no interfaces.
func TestFreeBSDFetch_FallsBackToPlainIfconfig(t *testing.T) {
	sess := newFakeSession()
	sess.out["hostname"] = "old-box"
	sess.fail["ifconfig -a -f inet:cidr,inet6:cidr"] = fmt.Errorf("ifconfig: unknown option -- f")
	sess.out["ifconfig -a"] = "em0: flags=8843\n\tinet 10.0.2.1 netmask 0xffffff00\n"

	got, err := NewFreeBSDParser().Fetch(sess)

	require.NoError(t, err)
	assert.Contains(t, got, "netmask 0xffffff00")
}

// Some hosts accept the flag and return nothing. An empty answer is a failure, not a
// configuration with no interfaces.
func TestFreeBSDFetch_TreatsEmptyCIDROutputAsAFailure(t *testing.T) {
	sess := newFakeSession()
	sess.out["hostname"] = "fw1"
	sess.out["ifconfig -a -f inet:cidr,inet6:cidr"] = "   \n"
	sess.out["ifconfig -a"] = "em0: flags=8843\n"

	got, err := NewFreeBSDParser().Fetch(sess)

	require.NoError(t, err)
	assert.Contains(t, got, "em0")
	assert.Contains(t, sess.ran, "ifconfig -a")
}

// The interfaces are what we came for; a missing hostname is not worth failing over.
func TestFreeBSDFetch_SurvivesAMissingHostname(t *testing.T) {
	sess := newFakeSession()
	sess.fail["hostname"] = fmt.Errorf("command not permitted")
	sess.out["ifconfig -a -f inet:cidr,inet6:cidr"] = "em0: flags=8843\n"

	got, err := NewFreeBSDParser().Fetch(sess)

	require.NoError(t, err)
	assert.Contains(t, got, "HOSTNAME:\n")
	assert.Contains(t, got, "em0")
}

func TestFreeBSDFetch_FailsWhenIfconfigIsUnavailable(t *testing.T) {
	sess := newFakeSession()
	sess.out["hostname"] = "fw1"

	_, err := NewFreeBSDParser().Fetch(sess)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ifconfig")
}

// -----------------------------------------------------------------------------
// Fortinet
// -----------------------------------------------------------------------------

func TestFortinetFetch_PrefersTargetedCommands(t *testing.T) {
	sess := newFakeSession()
	sess.out["show system interface"] = "config system interface\n    edit \"port1\"\nend"
	sess.out["show system global"] = "config system global\n    set hostname \"fg1\"\nend"

	got, err := NewFortinetParser().Fetch(sess)

	require.NoError(t, err)
	assert.Contains(t, got, "port1")
	assert.Contains(t, got, "hostname")
	assert.NotContains(t, sess.ran, "show full-configuration",
		"the flat output is preferred; the nested full dump is a last resort")
}

func TestFortinetFetch_FallsBackToFullConfiguration(t *testing.T) {
	sess := newFakeSession()
	sess.fail["show system interface"] = fmt.Errorf("command parse error")
	sess.out["show full-configuration"] = "config system interface\n    edit \"port1\"\nend"

	got, err := NewFortinetParser().Fetch(sess)

	require.NoError(t, err)
	assert.Contains(t, got, "port1")
}

// A device that gives us interfaces but refuses `show system global` is still usable:
// we lose only the hostname.
func TestFortinetFetch_ToleratesAMissingSystemGlobal(t *testing.T) {
	sess := newFakeSession()
	sess.out["show system interface"] = "config system interface\n    edit \"port1\"\nend"
	sess.fail["show system global"] = fmt.Errorf("permission denied")

	got, err := NewFortinetParser().Fetch(sess)

	require.NoError(t, err)
	assert.Contains(t, got, "port1")
}

func TestFortinetFetch_FailsWhenBothRoutesFail(t *testing.T) {
	sess := newFakeSession()
	sess.fail["show system interface"] = fmt.Errorf("parse error")

	_, err := NewFortinetParser().Fetch(sess)

	require.Error(t, err)
}

// -----------------------------------------------------------------------------
// Cisco
// -----------------------------------------------------------------------------

func TestCiscoFetch_FallsBackToShowConfig(t *testing.T) {
	sess := newFakeSession()
	// `show running-config` needs enable mode on most devices.
	sess.fail["show running-config"] = fmt.Errorf("%% Invalid input detected")
	sess.out["show config"] = "interface GigabitEthernet0/1\n"

	got, err := NewCiscoParser().Fetch(sess)

	require.NoError(t, err)
	assert.Contains(t, got, "GigabitEthernet0/1")
}

func TestCiscoFetch_FailsWhenNeitherCommandWorks(t *testing.T) {
	_, err := NewCiscoParser().Fetch(newFakeSession())
	require.Error(t, err)
}

// -----------------------------------------------------------------------------
// OpenWrt
// -----------------------------------------------------------------------------

// OpenWrt's config is spread across eight commands, and the output is re-split by
// ParseConfig on the "# <command>" headers this writes. The eighth is frr.conf:
// UCI cannot express a dynamic routing protocol, so on a routed OpenWrt the whole
// control plane lives in a file UCI knows nothing about.
func TestOpenWrtFetch_RunsEveryCommandAndTagsEachBlock(t *testing.T) {
	sess := newFakeSession()
	sess.out["uci show network"] = "network.lan=interface\n"
	sess.out["uci show wireless"] = "wireless.radio0=wifi-device\n"
	sess.out["uci show firewall"] = ""
	sess.out["uci show system"] = ""
	sess.out["uci show dhcp"] = ""
	sess.out["cat /etc/board.json"] = `{"model":{"id":"x"}}`
	sess.out["swconfig dev switch0 show"] = "VLAN 1:\n"
	sess.out["cat /etc/frr/frr.conf"] = "router ospf\n"

	got, err := NewOpenWrtParser().Fetch(sess)

	require.NoError(t, err)
	assert.Len(t, sess.ran, 8)
	assert.Contains(t, got, "# uci show network")
	assert.Contains(t, got, "network.lan=interface")
	assert.Contains(t, got, "# swconfig dev switch0 show")
	assert.Contains(t, got, "# cat /etc/frr/frr.conf")
}

// swconfig is absent on DSA-based devices and /etc/board.json on older ones. A failing
// command is recorded in-band and the rest still parses -- refusing the whole device over
// a missing swconfig would lose every modern OpenWrt box.
func TestOpenWrtFetch_SurvivesAMissingCommand(t *testing.T) {
	sess := newFakeSession()
	sess.out["uci show network"] = "network.lan=interface\n"
	sess.fail["swconfig dev switch0 show"] = fmt.Errorf("swconfig: not found")

	got, err := NewOpenWrtParser().Fetch(sess)

	require.NoError(t, err, "a missing swconfig must not fail the whole fetch")
	assert.Contains(t, got, "network.lan=interface", "the rest of the config survives")
	assert.Contains(t, got, "# Error executing swconfig dev switch0 show")
}
