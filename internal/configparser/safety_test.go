// safety_test.go: TDD for the SafetyLevel enum.
//
// Every write API (renderer.Render, cmd/push/*) takes a SafetyLevel. The rule is:
//
//	zero value == SafetyDryRun. A miscompiled binary that drops the flag must NOT apply.
package configparser

import "testing"

func TestSafetyLevel_ZeroIsDryRun(t *testing.T) {
	var s SafetyLevel
	if s != SafetyDryRun {
		t.Fatalf("zero SafetyLevel must equal SafetyDryRun to keep the default-no-apply contract; got %v", s)
	}
}

func TestSafetyLevel_String(t *testing.T) {
	cases := []struct {
		in   SafetyLevel
		want string
	}{
		{SafetyDryRun, "dry-run"},
		{SafetyStaged, "staged"},
		{SafetyApply, "apply"},
		{SafetyLevel(99), "unknown"},
	}
	for _, c := range cases {
		if got := c.in.String(); got != c.want {
			t.Errorf("SafetyLevel(%d).String() = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseSafetyLevel(t *testing.T) {
	cases := []struct {
		in   string
		want SafetyLevel
		err  bool
	}{
		// The three flag spellings every CLI/HTTP caller will reach for.
		{"dry-run", SafetyDryRun, false},
		{"staged", SafetyStaged, false},
		{"apply", SafetyApply, false},
		// Aliases.
		{"dryrun", SafetyDryRun, false},
		{"DRY-RUN", SafetyDryRun, false}, // case-insensitive on the way in
		{"", SafetyDryRun, false},       // empty defaults to dry-run, never to apply
		// Rejections.
		{"yes", 0, true},
		{"commit", 0, true},
	}
	for _, c := range cases {
		got, err := ParseSafetyLevel(c.in)
		if (err != nil) != c.err {
			t.Errorf("ParseSafetyLevel(%q) err=%v, want err=%v", c.in, err, c.err)
			continue
		}
		if !c.err && got != c.want {
			t.Errorf("ParseSafetyLevel(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
