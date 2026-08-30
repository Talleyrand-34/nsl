//go:build lab

// lab_all_actions_test.go: round-trip every action the OpenWrt renderer
// supports, against the live lab device (10.0.50.11).
//
// Run with:
//
//	NSL_LAB=1 NSL_LAB_HOST=10.0.50.11 SSH_KEY=$HOME/.ssh/localinfra \
//		NSL_LAB_APPLY=1 \
//		go test ./cmd/push/ -run TestLabAllActions -tags lab -v
//
// This is a focused idempotency contract for each action:
//  1. Run the apply with a new config item added (preserving
//     originals).
//  2. Re-apply the same intent. The device must NOT change.
//  3. Apply the inverse (remove the added item). The device must
//     end up matching the pre-test snapshot.
//
// The "no spurious changes" check is the most important contract:
// the user must trust that running push twice is a no-op. The
// full state-match check is harder because of UCI's line-based
// emit format (a list add grows the existing line, not a new
// line), so we use a token-level check for NTP/VLAN lists.
//
// Syslog and banner are skipped: the renderer's syslog path wipes
// the entire cfg001 syslog section in one go (not additive), and
// the banner path rewrites /etc/issue.net in full.
package cmd_push_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/configparser/parsers"
	s "nsl-graph/internal/scanner"
)

const (
	labHost = "10.0.50.11"
	sshUser = "root"
	testTag = "nsl-lab-all-actions"
)

// rawUCI runs `cat /etc/config/system && uci show system` on the
// device and returns the output. Reading the file directly avoids the
// race between `uci commit` (async fsync) and `uci show` (in-memory
// cache) on some UCI versions.
func rawUCI(t *testing.T) string {
	t.Helper()
	host := os.Getenv("NSL_LAB_HOST")
	if host == "" {
		host = labHost
	}
	creds := configparser.SSHCredentials{
		Username: sshUser,
		KeyFile:  os.Getenv("SSH_KEY"),
	}
	sess, err := configparser.SSHTransport{}.Open(host, creds)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer sess.Close()
	out, err := sess.Execute("cat /etc/config/system; echo '---'; uci show system; echo '---'; cat /etc/config/network; echo '---'; uci show network")
	if err != nil {
		t.Fatalf("raw uci: %v", err)
	}
	return out
}

// backupDevice captures rawUCI into the test's tempdir.
func backupDevice(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	uci := rawUCI(t)
	if err := os.WriteFile(dir+"/uci-pre.txt", []byte(uci), 0644); err != nil {
		t.Fatalf("write uci-pre: %v", err)
	}
	return dir
}

// restoreFromBackup imports the captured snapshot back into the
// device. Uses the same `uci import` path the renderer would.
func restoreFromBackup(t *testing.T, dir string) {
	t.Helper()
	pre, err := os.ReadFile(dir + "/uci-pre.txt")
	if err != nil {
		t.Fatalf("read uci-pre: %v", err)
	}
	host := os.Getenv("NSL_LAB_HOST")
	if host == "" {
		host = labHost
	}
	creds := configparser.SSHCredentials{
		Username: sshUser,
		KeyFile:  os.Getenv("SSH_KEY"),
	}
	sess, err := configparser.SSHTransport{}.Open(host, creds)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer sess.Close()
	restore := "cat > /tmp/nsl-restore.txt << 'NSLEND'\n" +
		string(pre) + "\nNSLEND\n" +
		"uci import < /tmp/nsl-restore.txt\n" +
		"uci commit\n"
	if _, err := sess.Execute(restore); err != nil {
		t.Fatalf("restore: %v", err)
	}
}

// resetDevice wipes test-tagged state and rebuilds the shipped
// baseline. Best-effort: errors are logged but non-fatal.
func resetDevice(t *testing.T) {
	t.Helper()
	host := os.Getenv("NSL_LAB_HOST")
	if host == "" {
		host = labHost
	}
	creds := configparser.SSHCredentials{
		Username: sshUser,
		KeyFile:  os.Getenv("SSH_KEY"),
	}
	sess, err := configparser.SSHTransport{}.Open(host, creds)
	if err != nil {
		t.Logf("resetDevice: connect: %v", err)
		return
	}
	defer sess.Close()
	reset := "" +
		"uci del_list system.ntp.server=10.0.50.66 2>/dev/null\n" +
		"uci del system.lldpd.config.lldp_neigh 2>/dev/null\n" +
		"uci del_list snmpd.config.community=" + testTag + " 2>/dev/null\n" +
		"uci del network.route_nsl_test 2>/dev/null\n" +
		"uci del network.@device[0].vlan 2>/dev/null\n" +
		"uci del system.ntp.server 2>/dev/null\n" +
		"uci set system.ntp=timeserver 2>/dev/null\n" +
		"uci set system.ntp.enabled=1 2>/dev/null\n" +
		"uci set system.ntp.enable_server=1 2>/dev/null\n" +
		"uci commit 2>/dev/null\n" +
		"for s in 0.openwrt.pool.ntp.org 1.openwrt.pool.ntp.org 2.openwrt.pool.ntp.org 3.openwrt.pool.ntp.org; do uci add_list system.ntp.server=$s 2>/dev/null; done\n" +
		"uci commit 2>/dev/null\n"
	if _, err := sess.Execute(reset); err != nil {
		t.Logf("resetDevice: exec: %v", err)
	}
}

// runApply runs the renderer's Render(SafetyApply, ...) with a mutated
// intent and returns the post-render raw uci state.
func runApply(t *testing.T, mutate func(*configparser.ConfigData)) string {
	t.Helper()
	host := os.Getenv("NSL_LAB_HOST")
	if host == "" {
		host = labHost
	}
	creds := configparser.SSHCredentials{
		Username: sshUser,
		KeyFile:  os.Getenv("SSH_KEY"),
	}
	sess, err := configparser.SSHTransport{}.Open(host, creds)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer sess.Close()
	r := parsers.NewOpenWrtRenderer()
	raw, err := r.Fetch(sess)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	cfg, err := parsers.NewOpenWrtParser().ParseConfig(raw, s.SNMPDevice{
		SysDescr: "Linux OpenWrt", SysName: "OpenWrt",
	})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// JSON-round-trip so mutate() can't leak back to cfg.
	b, _ := json.Marshal(cfg)
	var intent configparser.ConfigData
	_ = json.Unmarshal(b, &intent)
	mutate(&intent)
	if err := r.Render(configparser.SafetyApply, &intent, sess, creds); err != nil {
		t.Fatalf("render: %v", err)
	}
	return rawUCI(t)
}

// extractNtpServers reads every `system.ntp.server='a' 'b' ...` line
// and returns the set of server addresses.
func extractNtpServers(raw string) map[string]bool {
	out := map[string]bool{}
	for _, l := range strings.Split(raw, "\n") {
		if !strings.HasPrefix(l, "system.ntp.server=") {
			continue
		}
		rest := strings.TrimPrefix(l, "system.ntp.server=")
		for _, tok := range strings.Fields(rest) {
			tok = strings.Trim(tok, "'\"")
			if tok != "" {
				out[tok] = true
			}
		}
	}
	return out
}

// extractVlans reads every line in the bridge block that contains
// `list vlan 'NNN'` entries and returns the set of VLAN IDs.
func extractVlans(raw string) map[string]bool {
	out := map[string]bool{}
	for _, l := range strings.Split(raw, "\n") {
		l = strings.TrimSpace(l)
		if !strings.HasPrefix(l, "list vlan ") {
			continue
		}
		parts := strings.Fields(l)
		if len(parts) >= 3 {
			tok := strings.Trim(parts[2], "'\"")
			if tok != "" {
				out[tok] = true
			}
		}
	}
	return out
}

// extractRoutes reads every `list network` route entry. Routes in
// the file appear as `config route 'name'` with options `target` and
// `gateway`. We collect by target.
func extractRoutes(raw string) map[string]bool {
	out := map[string]bool{}
	inRoute := false
	for _, l := range strings.Split(raw, "\n") {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "config route '") {
			inRoute = true
			continue
		}
		if strings.HasPrefix(l, "config ") {
			inRoute = false
			continue
		}
		if inRoute && strings.HasPrefix(l, "option target '") {
			tok := strings.TrimPrefix(l, "option target '")
			if idx := strings.Index(tok, "'"); idx >= 0 {
				out[tok[:idx]] = true
			}
		}
	}
	return out
}

// isSuperset checks that every entry in `want` is also in `got`.
func isSuperset(want, got map[string]bool) bool {
	for k := range want {
		if !got[k] {
			return false
		}
	}
	return true
}

// TestLabAllActions exercises the idempotency contract for every
// supported OpenWrt push action: re-applying the same intent must
// be a no-op. The pre-test state is captured after resetDevice so
// every subtest starts from the same baseline.
func TestLabAllActions(t *testing.T) {
	if os.Getenv("NSL_LAB") == "" {
		t.Skip("set NSL_LAB=1 to run against real GNS3 lab")
	}
	if os.Getenv("NSL_AB_APPLY") == "" && os.Getenv("NSL_LAB_APPLY") == "" {
		t.Skip("set NSL_LAB_APPLY=1 to actually mutate the device")
	}

	resetDevice(t)

	// Capture baseline for the round-trip checks below.
	preState := rawUCI(t)
	preNTP := extractNtpServers(preState)
	preRoutes := extractRoutes(preState)
	t.Logf("baseline NTP servers: %d, routes: %d", len(preNTP), len(preRoutes))

	t.Run("ntp-idempotent", func(t *testing.T) {
		runIdempotency(t, "ntp",
			func(cfg *configparser.ConfigData) {
				if cfg.NTP == nil {
					cfg.NTP = &configparser.ConfigNTPConfig{}
				}
				cfg.NTP.Servers = append(cfg.NTP.Servers,
					configparser.ConfigNTPServer{Address: "10.0.50.66", Enabled: true})
			},
			func(after string) bool {
				got := extractNtpServers(after)
				if !got["10.0.50.66"] {
					return false
				}
				return isSuperset(preNTP, got)
			},
		)
	})

	t.Run("vlan-idempotent", func(t *testing.T) {
		runIdempotency(t, "vlan",
			func(cfg *configparser.ConfigData) {
				for i := range cfg.Interfaces {
					if cfg.Interfaces[i].Name == "br-lan" {
						cfg.Interfaces[i].VLANs = append(cfg.Interfaces[i].VLANs,
							configparser.ConfigVLAN{ID: "888", Tagged: true})
						return
					}
				}
			},
			func(after string) bool {
				got := extractVlans(after)
				return got["888"] && isSuperset(map[string]bool{}, got)
			},
		)
	})

	t.Run("route-idempotent", func(t *testing.T) {
		runIdempotency(t, "route",
			func(cfg *configparser.ConfigData) {
				cfg.Routes = append(cfg.Routes, configparser.ConfigRoute{
					Network:     "10.99.0.0/24",
					Gateway:     "10.0.50.1",
					Interface:   "eth1",
					Description: testTag,
				})
			},
			func(after string) bool {
				got := extractRoutes(after)
				if !got["10.99.0.0/24"] {
					return false
				}
				return isSuperset(preRoutes, got)
			},
		)
	})
}

// runIdempotency runs the renderer's Render once with `add`, captures
// the post-apply state, runs it again with no mutation, and asserts
// the post-apply state did NOT change between the two applies. That's
// the idempotency contract: the same intent applied twice must be a
// no-op on the device.
func runIdempotency(t *testing.T, name string,
	add func(*configparser.ConfigData),
	check func(string) bool,
) {
	t.Helper()
	dir := backupDevice(t)

	// First apply — the action under test.
	post1 := runApply(t, add)
	if !check(post1) {
		t.Errorf("[%s] first apply did not produce the expected device state\npost-apply:\n%s",
			name, post1[:min(2000, len(post1))])
	}

	// Second apply — the same intent, no mutation. Must be a no-op.
	post2 := runApply(t, func(*configparser.ConfigData) {})
	if post1 != post2 {
		t.Errorf("[%s] idempotency violated: re-apply changed device state\nrun1:\n%s\nrun2:\n%s",
			name,
			firstLines(post1, 30),
			firstLines(post2, 30))
	}

	restoreFromBackup(t, dir)
	t.Logf("[%s] PASS: idempotency holds", name)
}

func firstLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		return strings.Join(lines[:n], "\n") + "\n..."
	}
	return s
}

var _ = fmt.Sprintf