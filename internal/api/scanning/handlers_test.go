// SPDX-License-Identifier: MIT
package scanning

import (
	"reflect"
	"testing"

	q "nsl-graph/internal/repository/application"
	"nsl-graph/internal/repository/entities"
)

// stubService implements just the methods detectProfileOverlaps touches
// via the q.NetServiceInt interface. The handler signature accepts
// q.NetServiceInt, so this stub works as-is.
type stubService struct {
	q.NetServiceInt
	profiles []*entities.ScanProfile
}

func (s *stubService) GetScanProfiles() ([]entities.ScanProfile, error) {
	out := make([]entities.ScanProfile, len(s.profiles))
	for i, p := range s.profiles {
		out[i] = *p
	}
	return out, nil
}

func TestDetectProfileOverlaps_EmptyHosts(t *testing.T) {
	s := &stubService{profiles: nil}
	if got := detectProfileOverlaps(s, "lab-prod", nil); got != nil {
		t.Fatalf("nil hosts -> nil overlaps, got %#v", got)
	}
	if got := detectProfileOverlaps(s, "lab-prod", []string{}); got != nil {
		t.Fatalf("empty hosts -> nil overlaps, got %#v", got)
	}
}

func TestDetectProfileOverlaps_NoMatches(t *testing.T) {
	s := &stubService{profiles: []*entities.ScanProfile{
		{Name: "other", Host: "10.0.0.99"},
	}}
	got := detectProfileOverlaps(s, "lab-new", []string{"10.0.0.1"})
	if len(got) != 0 {
		t.Fatalf("host not in any profile -> no overlaps, got %#v", got)
	}
}

func TestDetectProfileOverlaps_FindsCrossProfileMatch(t *testing.T) {
	// 10.0.0.245 is in two profiles (lab-prod + lab-staging);
	// 10.0.0.246 is in lab-prod; 10.0.0.247 is new.
	// Expected: 3 overlap entries, one per (host, other_profile) pair.
	s := &stubService{profiles: []*entities.ScanProfile{
		{Name: "lab-prod", Host: "10.0.0.245"},
		{Name: "lab-prod", Host: "10.0.0.246"},
		{Name: "lab-staging", Host: "10.0.0.245"},
	}}
	got := detectProfileOverlaps(s, "lab-temp", []string{"10.0.0.245", "10.0.0.247", "10.0.0.246"})
	if len(got) != 3 {
		t.Fatalf("expected 3 overlaps, got %d: %#v", len(got), got)
	}
	byHost := map[string][]string{}
	for _, o := range got {
		byHost[o.Host] = append(byHost[o.Host], o.OtherProfile)
	}
	if !contains(byHost["10.0.0.245"], "lab-prod") || !contains(byHost["10.0.0.245"], "lab-staging") {
		t.Errorf("10.0.0.245 should overlap with lab-prod AND lab-staging, got %#v", byHost)
	}
	if !contains(byHost["10.0.0.246"], "lab-prod") {
		t.Errorf("10.0.0.246 should overlap with lab-prod, got %#v", byHost)
	}
	if _, ok := byHost["10.0.0.247"]; ok {
		t.Errorf("10.0.0.247 is a new host, should not appear, got %#v", byHost)
	}
}

func TestDetectProfileOverlaps_ExcludesSameProfile(t *testing.T) {
	// If the operator edits an existing profile and adds hosts that are
	// already in it, the overlap list must NOT report self-overlaps.
	s := &stubService{profiles: []*entities.ScanProfile{
		{Name: "lab-prod", Host: "10.0.0.245"},
		{Name: "lab-prod", Host: "10.0.0.246"},
	}}
	got := detectProfileOverlaps(s, "lab-prod", []string{"10.0.0.245", "10.0.0.246"})
	if len(got) != 0 {
		t.Fatalf("same-profile hosts must be excluded, got %#v", got)
	}
}

func TestDetectProfileOverlaps_Deduplicates(t *testing.T) {
	s := &stubService{profiles: []*entities.ScanProfile{
		{Name: "p1", Host: "10.0.0.1"},
		{Name: "p2", Host: "10.0.0.1"},
	}}
	got := detectProfileOverlaps(s, "new", []string{"10.0.0.1"})
	if len(got) != 2 {
		t.Fatalf("expected 2 unique (p1 + p2) overlap entries, got %d: %#v", len(got), got)
	}
	seen := map[string]bool{}
	for _, o := range got {
		key := o.Host + "\x00" + o.OtherProfile
		if seen[key] {
			t.Errorf("duplicate entry: %#v", o)
		}
		seen[key] = true
	}
	if !reflect.DeepEqual(got[0].Host, "10.0.0.1") {
		t.Errorf("first overlap host wrong: %#v", got[0])
	}
	if got[0].OtherProfile != "p1" && got[0].OtherProfile != "p2" {
		t.Errorf("first overlap other_profile wrong: %#v", got[0])
	}
}

func contains(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}