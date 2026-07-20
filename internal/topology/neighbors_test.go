package topology

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading fixture %s: %v", name, err)
	}
	return string(b)
}

// The fixture is a real `/ip/neighbor/print terse` from the lab's MikroTik, and
// it happens to contain exactly the case that makes this table worth reading:
// one neighbour found by LLDP and one by MNDP. A collector that spoke only LLDP
// would see half the adjacencies this box knows about.
func TestParseRouterOSNeighbors(t *testing.T) {
	ev, err := parseRouterOSNeighbors(
		readFixture(t, "routeros-ip-neighbor.txt"), "10.0.50.12", "MikroTik-1")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(ev) != 2 {
		t.Fatalf("expected 2 neighbours, got %d", len(ev))
	}

	byName := map[string]NeighborEvidence{}
	for _, e := range ev {
		byName[e.RemoteSysName] = e
	}

	infix, ok := byName["infix-6d-ff-ff"]
	if !ok {
		t.Fatal("the LLDP-discovered Infix neighbour is missing")
	}
	if infix.Discovery != "lldp" {
		t.Errorf("Discovery = %q, want lldp", infix.Discovery)
	}
	if infix.LocalPort != "ether1" {
		t.Errorf("LocalPort = %q, want ether1", infix.LocalPort)
	}
	// `interface-name` is the REMOTE port and sits next to the local
	// `interface`; swapping them reverses every edge on the graph.
	if infix.RemotePort != "eth0" {
		t.Errorf("RemotePort = %q, want eth0 (the neighbour's port, not ours)", infix.RemotePort)
	}
	if infix.RemoteIP != "10.0.50.14" {
		t.Errorf("RemoteIP = %q", infix.RemoteIP)
	}
	if infix.RemoteChassisMAC != "0c:0d:fe:6e:00:00" {
		t.Errorf("RemoteChassisMAC = %q, want normalised lowercase", infix.RemoteChassisMAC)
	}

	mt, ok := byName["MikroTik"]
	if !ok {
		t.Fatal("the MNDP-discovered MikroTik neighbour is missing")
	}
	if mt.Discovery != "mndp" {
		t.Errorf("Discovery = %q, want mndp", mt.Discovery)
	}
	// This record's `version` value contains spaces and slashes
	// ("7.2 (stable) Mar/31/2022 09:11:50"). A tokenizer that split on
	// whitespace would swallow the keys that follow it.
	if mt.RemotePort != "ether1" {
		t.Errorf("RemotePort = %q — a space-bearing value ate the later keys", mt.RemotePort)
	}
}

// Infix runs stock lldpd, so the original json0 probe is the one that answers.
// Asserting it here keeps the probe list honest: if the first probe stopped
// working for Infix the failure should name Infix, not surface as a silent
// fallthrough to a RouterOS command.
func TestParseLLDPCLI_InfixFixture(t *testing.T) {
	ev, err := parseLLDPCLI(
		readFixture(t, "infix-lldpcli-json0.json"), SourceSSHLLDP, "10.0.50.14", "Infix-1")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(ev) != 1 {
		t.Fatalf("expected 1 neighbour, got %d", len(ev))
	}
	got := ev[0]
	if got.LocalPort != "eth0" {
		t.Errorf("LocalPort = %q", got.LocalPort)
	}
	if got.RemoteSysName != "MikroTik" {
		t.Errorf("RemoteSysName = %q", got.RemoteSysName)
	}
	if got.RemotePort != "ether1" {
		t.Errorf("RemotePort = %q", got.RemotePort)
	}
	if got.RemoteIP != "10.0.50.12" {
		t.Errorf("RemoteIP = %q", got.RemoteIP)
	}
}

// The prober must tell "this vendor does not understand the question" apart from
// "this device has no neighbours". Both come back on stdout, and only the second
// should end the probing — treating a rejection as an answer would stop the
// search before the probe that actually works is ever tried.
func TestProbeOutputIsUsable(t *testing.T) {
	unusable := map[string]string{
		"empty":               "",
		"whitespace":          "   \n  ",
		"posix shell":         "-ash: lldpcli: not found",
		"bash":                "bash: lldpcli: command not found",
		"missing binary":      "ls: /usr/sbin/lldpcli: No such file or directory",
		"vyos restricted cli": "\n  Invalid command: [lldpcli]\n",
		"routeros rejection":  "syntax error (line 1 column 16)",
		"fortios rejection":   "command parse error before 'lldp'\nCommand fail. Return code -61",
		"lldpd not running":   "2026-07-20T00:46:51 [WARN/control] unable to connect to socket /run/lldpd.socket: No such file or directory",
	}
	for name, out := range unusable {
		if probeOutputIsUsable(out) {
			t.Errorf("%s: should be rejected as unusable, got usable", name)
		}
	}

	usable := map[string]string{
		"empty lldp table": `{"lldp":[{}]}`,
		"routeros record":  "0 interface=ether1 identity=x discovered-by=mndp",
	}
	for name, out := range usable {
		if !probeOutputIsUsable(out) {
			t.Errorf("%s: should be usable, got rejected", name)
		}
	}
}

func TestParseRouterOSIdentityLine(t *testing.T) {
	if got := parseRouterOSIdentityLine("  name: MikroTik\n"); got != "MikroTik" {
		t.Errorf("got %q, want MikroTik", got)
	}
	if got := parseRouterOSIdentityLine("nothing here"); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

// When nothing answers, the message has to name a cause the operator can act
// on. The probes are ordered, so the last one to fail is RouterOS's — and
// telling someone with a VyOS box that /ip/neighbor/print does not exist is
// true, useless, and points at the wrong device entirely.
func TestNoNeighborSourceError_DiagnosesStoppedLLDPD(t *testing.T) {
	err := noNeighborSourceError([]string{
		"lldpcli -f json0 show neighbors: -ash: lldpcli: not found",
		"/usr/sbin/lldpcli -f json0 show neighbors: [WARN/control] unable to connect to socket /run/lldpd.socket: No such file or directory",
		"/ip/neighbor/print terse without-paging: vbash: /ip/neighbor/print: No such file or directory",
	})
	msg := err.Error()
	if !strings.Contains(msg, "lldpd is installed but not running") {
		t.Errorf("should diagnose the stopped daemon, got: %s", msg)
	}
	if strings.Contains(msg, "/ip/neighbor") {
		t.Errorf("should not blame the RouterOS probe on a non-RouterOS device: %s", msg)
	}
}

// With no such clue, listing what was tried is the best available answer.
func TestNoNeighborSourceError_FallsBackToTheAttemptList(t *testing.T) {
	err := noNeighborSourceError([]string{"lldpcli ...: not found"})
	if !strings.Contains(err.Error(), "tried") {
		t.Errorf("expected the attempt list, got: %s", err)
	}
	if noNeighborSourceError(nil) == nil {
		t.Error("an empty attempt list must still produce an error")
	}
}

// A vendor that rejects every probe must be reported as such, not as a device
// with no links.
//
// This is the failure the grammar check exists to prevent. FortiOS answers an
// unknown command with prose — and its wording varies by command, "command parse
// error" for one and "Unknown action 0" for another — which the RouterOS parser
// happily reduced to zero records. The last probe then "succeeded" with an empty
// result, and a box that had refused every question was reported as simply
// having no neighbours.
func TestParseRouterOSNeighbors_RejectsForeignOutput(t *testing.T) {
	foreign := []string{
		"FGT30D3X15012871 # Unknown action 0",
		"FGT30D3X15012871 # \ncommand parse error before 'lldp'\nCommand fail. Return code -61",
		`{"lldp":[{"interface":[]}]}`,
		"-ash: /ip/neighbor/print: not found",
	}
	for _, out := range foreign {
		if _, err := parseRouterOSNeighbors(out, "h", "d"); err == nil {
			t.Errorf("should refuse non-RouterOS output, accepted: %.60q", out)
		}
	}
}

func TestLooksLikeTerseOutput(t *testing.T) {
	if !looksLikeTerseOutput(readFixture(t, "routeros-ip-neighbor.txt")) {
		t.Error("the real RouterOS fixture must be recognised")
	}
	// A legend with no records is RouterOS talking, but it carries no records —
	// so it is not accepted as an answer either.
	if looksLikeTerseOutput("Flags: X - disabled\n") {
		t.Error("a bare legend has no records and must not count as an answer")
	}
	if looksLikeTerseOutput("Unknown action 0") {
		t.Error("prose is not terse output")
	}
}
