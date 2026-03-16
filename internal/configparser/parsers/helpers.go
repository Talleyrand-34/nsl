package parsers

import "net"

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
