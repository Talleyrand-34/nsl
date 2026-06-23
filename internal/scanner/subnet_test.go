package scanner

import (
	"reflect"
	"testing"
)

func TestSplitSubnets(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"10.0.0.0/24", []string{"10.0.0.0/24"}},
		{"10.0.0.0/24,10.0.1.0/24", []string{"10.0.0.0/24", "10.0.1.0/24"}},
		{" 10.0.0.0/24 ,  10.0.1.0/24 ", []string{"10.0.0.0/24", "10.0.1.0/24"}},
		{"10.0.0.0/24 10.0.1.0/24", []string{"10.0.0.0/24", "10.0.1.0/24"}},
		{"10.0.0.0/24,10.0.0.0/24", []string{"10.0.0.0/24"}}, // dedup
		{"   ", nil},
		{"", nil},
	}
	for _, c := range cases {
		got := SplitSubnets(c.in)
		if len(got) == 0 && len(c.want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("SplitSubnets(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestEnumerateSubnet_MultiCIDR(t *testing.T) {
	// Two adjacent /30s plus a bare IP; /30 drops network+broadcast (2 hosts each).
	ips, err := enumerateSubnet("10.0.0.0/30, 10.0.1.0/30, 10.0.2.5")
	if err != nil {
		t.Fatalf("enumerateSubnet: %v", err)
	}
	want := []string{"10.0.0.1", "10.0.0.2", "10.0.1.1", "10.0.1.2", "10.0.2.5"}
	if !reflect.DeepEqual(ips, want) {
		t.Errorf("enumerateSubnet multi = %v, want %v", ips, want)
	}
}

func TestEnumerateSubnet_DedupAcrossCIDRs(t *testing.T) {
	// Overlapping ranges must not yield duplicate addresses.
	ips, err := enumerateSubnet("10.0.0.0/30,10.0.0.0/29")
	if err != nil {
		t.Fatalf("enumerateSubnet: %v", err)
	}
	seen := map[string]bool{}
	for _, ip := range ips {
		if seen[ip] {
			t.Fatalf("duplicate address %s in %v", ip, ips)
		}
		seen[ip] = true
	}
}

func TestEnumerateSubnet_Empty(t *testing.T) {
	if _, err := enumerateSubnet("  "); err == nil {
		t.Error("expected an error for an empty spec")
	}
}
