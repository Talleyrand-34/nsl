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
