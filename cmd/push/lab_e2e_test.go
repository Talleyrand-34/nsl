//go:build lab

// lab_e2e_test.go: end-to-end live test against OpenWrt-1.
//
// Run with:
//
//	NSL_LAB=1 NSL_LAB_HOST=10.0.50.11 SSH_KEY=$HOME/.ssh/localinfra \
//	  go test ./cmd/push/ -run TestLabE2E -tags lab -v
//
// What it does:
//   1. Connects to the device over SSH.
//   2. Fetches the current UCI configuration.
//   3. Parses it into ConfigData.
//   4. Runs Render(SafetyDryRun) — must succeed with zero SSH commands.
//   5. (Staged only, off by default) Render(SafetyStaged) writes a patch file.
//   6. (Apply only, off by default) Render(SafetyApply) would mutate the device.

package cmd_push_test

import (
	"os"
	"testing"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/configparser/parsers"
	s "nsl-graph/internal/scanner"
)

func TestLabE2E(t *testing.T) {
	if os.Getenv("NSL_LAB") == "" {
		t.Skip("set NSL_LAB=1 to run against real GNS3 lab")
	}

	host := os.Getenv("NSL_LAB_HOST")
	if host == "" {
		host = "10.0.50.11"
	}

	creds := configparser.SSHCredentials{
		Username: "root",
		KeyFile:  os.Getenv("SSH_KEY"),
	}
	transport := configparser.SSHTransport{}
	sess, err := transport.Open(host, creds)
	if err != nil {
		t.Fatalf("connect to %s: %v", host, err)
	}
	defer sess.Close()

	r := parsers.NewOpenWrtRenderer()
	raw, err := r.Fetch(sess)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	t.Logf("fetched %d bytes from %s", len(raw), host)

	parser := parsers.NewOpenWrtParser()
	cfg, err := parser.ParseConfig(raw, s.SNMPDevice{SysDescr: "Linux OpenWrt", SysName: "OpenWrt"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	t.Logf("parsed %d interfaces", len(cfg.Interfaces))
	for _, i := range cfg.Interfaces {
		t.Logf("  %s (%s): %d VLANs", i.Name, i.Type, len(i.VLANs))
	}

	// DryRun must succeed and NOT touch the device.
	if err := r.Render(configparser.SafetyDryRun, cfg, sess, creds); err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	t.Log("OK: DryRun completed (no SSH mutation)")

	// Optional Staged/Apply — opt in with NSL_LAB_APPLY=1 (off by default).
	if os.Getenv("NSL_LAB_APPLY") != "" {
		// Build an intent with one new VLAN tagged on br-lan (or first non-lo interface).
		intent := *cfg
		intent.Interfaces = append(intent.Interfaces, configparser.ConfigInterface{
			Name: "vlan-lab-e2e", Type: "bridge", Enabled: true,
			Description: "lab e2e test interface",
			VLANs: []configparser.ConfigVLAN{
				{ID: "999", Tagged: true},
			},
		})
		if err := r.Render(configparser.SafetyApply, &intent, sess, creds); err != nil {
			t.Fatalf("apply: %v", err)
		}
		t.Log("OK: Apply completed (device mutated)")
	} else {
		t.Log("skipping Apply (set NSL_LAB_APPLY=1 to actually mutate)")
	}
}

var _ = configparser.SSHCredentials{}
