package parsers

import (
	"net"
	"strings"
)

// netmaskToPrefix converts a dotted-decimal subnet mask to a CIDR prefix length.
// Returns -1 if the mask is not a valid IPv4 mask.
func netmaskToPrefix(mask string) int {
	ip := net.ParseIP(mask)
	if ip == nil {
		return -1
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return -1
	}
	ones, bits := net.IPMask(ip4).Size()
	if bits == 0 {
		return -1 // non-contiguous mask
	}
	return ones
}

// stripUCIQuotes removes surrounding single quotes from a UCI value.
func stripUCIQuotes(s string) string {
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}
	return s
}

// unquotedFields splits a UCI value into whitespace-separated tokens, stripping
// surrounding single/double quotes from each token. This is needed when a list
// is rendered on a single line (e.g. ports='eth0' 'eth1' 'eth2'), where stripping
// only the outer pair leaves inner tokens like `'eth1'` or `eth0'`.
func unquotedFields(s string) []string {
	fields := strings.Fields(s)
	out := fields[:0]
	for _, f := range fields {
		if f = strings.Trim(f, "'\""); f != "" {
			out = append(out, f)
		}
	}
	return out
}

// splitUCIListValue extracts the list items from a UCI option value.
// UCI emits list entries either as separate lines (`option[]=value`) — which
// the parser handles in the list-regex branch — or, for DSA bridges,
// collapsed onto a single line with space-separated quoted values
// (`option='a' 'b' 'c'`). This helper handles the single-line form: it
// splits on whitespace, strips surrounding quotes, and returns the items.
// A single item is returned as a one-element slice (still triggers the
// list branch, which is correct for any multi-value key the parser
// wants as a list). An empty string returns nil.
func splitUCIListValue(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	// Split on whitespace, strip quotes from each token. unquotedFields
	// already does this.
	items := unquotedFields(s)
	if len(items) == 0 {
		return nil
	}
	return items
}
