// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package scanner

import (
	"regexp"
	"strconv"
	"strings"
)

// VLANInference represents the result of VLAN inference for an interface
type VLANInference struct {
	VLANNumbers []string // Detected VLAN number(s)
	Accuracy    int      // Confidence level (1 = highest)
	Method      string   // Description of inference method used
	Notes       string   // Additional notes about conflicts or multiple detections
}

// InferVLANFromInterface analyzes interface name and IPs to infer VLAN information
func InferVLANFromInterface(name string, ips []string) VLANInference {
	nameVLAN := extractVLANFromName(name)
	ipVLAN := extractVLANFromIP(ips)

	return combineInferences(nameVLAN, ipVLAN)
}

// extractVLANFromName extracts VLAN number from interface name
func extractVLANFromName(name string) VLANInference {
	if name == "" {
		return VLANInference{VLANNumbers: []string{}, Accuracy: 0}
	}

	// Pattern for vlan interfaces: vlan0.X, vlanX, vlan.X, etc.
	vlanPatterns := []string{
		`vlan0?\.(\d+)`,  // vlan0.10, vlan.10
		`vlan(\d+)`,      // vlan10
		`.*vlan.*?(\d+)`, // any variation with vlan and numbers
	}

	for _, pattern := range vlanPatterns {
		re := regexp.MustCompile(`(?i)` + pattern) // case insensitive
		matches := re.FindStringSubmatch(name)
		if len(matches) > 1 {
			vlanNum := matches[1]

			// Fix VLAN 0 to VLAN 1 (VLAN 0 is reserved and typically means native VLAN 1)
			if vlanNum == "0" {
				vlanNum = "1"
			}

			return VLANInference{
				VLANNumbers: []string{vlanNum},
				Accuracy:    1,
				Method:      "name",
				Notes:       "",
			}
		}
	}

	return VLANInference{VLANNumbers: []string{}, Accuracy: 0}
}

// extractVLANFromIP extracts VLAN number from IP address patterns
func extractVLANFromIP(ips []string) VLANInference {
	if len(ips) == 0 {
		return VLANInference{VLANNumbers: []string{}, Accuracy: 0}
	}

	var detectedVLANs []string

	// Pattern: 10.0.X.1 suggests VLAN X
	ipPattern := regexp.MustCompile(`^10\.0\.(\d+)\.1$`)

	for _, ip := range ips {
		matches := ipPattern.FindStringSubmatch(ip)
		if len(matches) > 1 {
			vlanNum := matches[1]

			// Fix VLAN 0 to VLAN 1 (VLAN 0 is reserved and typically means native VLAN 1)
			if vlanNum == "0" {
				vlanNum = "1"
			}

			// Avoid duplicates
			if !contains(detectedVLANs, vlanNum) {
				detectedVLANs = append(detectedVLANs, vlanNum)
			}
		}
	}

	if len(detectedVLANs) > 0 {
		return VLANInference{
			VLANNumbers: detectedVLANs,
			Accuracy:    2,
			Method:      "ip",
			Notes:       "",
		}
	}

	return VLANInference{VLANNumbers: []string{}, Accuracy: 0}
}

// combineInferences combines name-based and IP-based inferences
func combineInferences(nameInf, ipInf VLANInference) VLANInference {
	// No inference from either method
	if len(nameInf.VLANNumbers) == 0 && len(ipInf.VLANNumbers) == 0 {
		return VLANInference{
			VLANNumbers: []string{},
			Accuracy:    0,
			Method:      "none",
			Notes:       "",
		}
	}

	// Only name-based inference
	if len(nameInf.VLANNumbers) > 0 && len(ipInf.VLANNumbers) == 0 {
		return nameInf
	}

	// Only IP-based inference
	if len(nameInf.VLANNumbers) == 0 && len(ipInf.VLANNumbers) > 0 {
		return ipInf
	}

	// Both methods have results - check for agreement
	nameVLAN := nameInf.VLANNumbers[0]
	ipVLAN := ipInf.VLANNumbers[0]

	if nameVLAN == ipVLAN {
		// Perfect match - highest confidence
		return VLANInference{
			VLANNumbers: []string{nameVLAN},
			Accuracy:    1,
			Method:      "name+ip",
			Notes:       "confirmed by both methods",
		}
	}

	// Conflict between methods
	combinedVLANs := append(nameInf.VLANNumbers, ipInf.VLANNumbers...)
	notes := "conflict: name=" + nameVLAN + ", ip=" + ipVLAN

	return VLANInference{
		VLANNumbers: combinedVLANs,
		Accuracy:    2,
		Method:      "name+ip",
		Notes:       notes,
	}
}

// FormatVLANDisplay formats VLAN inference results for display
func (vi VLANInference) FormatVLANDisplay() string {
	if len(vi.VLANNumbers) == 0 || vi.Accuracy == 0 {
		return "N/A"
	}

	if len(vi.VLANNumbers) == 1 {
		return vi.VLANNumbers[0]
	}

	// Multiple VLANs or conflict
	if strings.Contains(vi.Notes, "conflict") {
		return strings.Join(vi.VLANNumbers, "/")
	}

	return strings.Join(vi.VLANNumbers, ",")
}

// FormatAccuracyDisplay formats accuracy level for display
func (vi VLANInference) FormatAccuracyDisplay() string {
	if vi.Accuracy == 0 {
		return "-"
	}
	return strconv.Itoa(vi.Accuracy)
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
