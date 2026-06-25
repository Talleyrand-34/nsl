/*
Copyright © 2026 Talleyrand-34 (t34@t34.dev)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package application

import (
	"fmt"
	"strings"

	"nsl-graph/internal/topology"
)

const (
	// placeholderZoneName is the well-known zone that holds placeholder devices
	// standing in for unknown infrastructure detected between known hosts.
	placeholderZoneName = "Unknown infrastructure"
	// placeholderModelName is the model used for placeholder unmanaged devices.
	placeholderModelName = "Unmanaged Switch"
)

// PlaceholderResult summarizes what CreatePlaceholderForIntermediaries created.
type PlaceholderResult struct {
	Zone        string `json:"zone"`
	Device      string `json:"device"`
	Connections int    `json:"connections"`
}

// CreatePlaceholderForIntermediaries materializes a single shared placeholder
// unmanaged device representing the unknown device(s) detected "in the middle"
// of the given intermediaries, and wires each observing "device:port" endpoint to
// it. Because the placeholder is unmanaged, VLANs replicate across its ports,
// which is what makes it useful for VLAN attestation/documentation of the gap.
func (ns *NetService) CreatePlaceholderForIntermediaries(intermediaries []topology.Intermediary) (PlaceholderResult, error) {
	var res PlaceholderResult
	if len(intermediaries) == 0 {
		return res, fmt.Errorf("no intermediaries provided")
	}

	// Union of every observing endpoint across the selected intermediaries, plus
	// a human-readable provenance note describing the gap.
	var seenBy []string
	var notes []string
	for _, im := range intermediaries {
		seenBy = append(seenBy, im.SeenBy...)
		desc := im.MAC
		if im.Vendor != "" {
			desc = im.Vendor + " " + im.MAC
		}
		notes = append(notes, desc)
	}
	seenBy = dedup(seenBy)
	if len(seenBy) == 0 {
		return res, fmt.Errorf("intermediaries have no observed endpoints to attach")
	}
	provenance := "placeholder: intermediary " + strings.Join(notes, ", ")

	// 1. Ensure the placeholder zone exists (idempotent by name).
	zoneID, err := ns.ensurePlaceholderZone()
	if err != nil {
		return res, err
	}
	res.Zone = placeholderZoneName

	// 2. Ensure the placeholder model (and its brand/class) and the "ethernet"
	// connection type exist — model and connection creation require their
	// referenced rows to be present.
	if err := ns.ensureBrand("Unknown"); err != nil {
		return res, fmt.Errorf("ensure brand: %w", err)
	}
	if err := ns.ensureDeviceClass("Switch"); err != nil {
		return res, fmt.Errorf("ensure device class: %w", err)
	}
	if err := ns.AddModel(placeholderModelName, "Unknown", "Switch"); err != nil &&
		!strings.Contains(err.Error(), "already exists") {
		return res, fmt.Errorf("create placeholder model: %w", err)
	}
	if err := ns.ensureConnectionType("ethernet"); err != nil {
		return res, fmt.Errorf("ensure connection type: %w", err)
	}

	// 3. Create the shared placeholder device (unmanaged). Generate the label up
	// front so we can find the device id afterwards.
	label, err := ns.generateUnmanagedName()
	if err != nil {
		return res, err
	}
	if err := ns.AddDevice(label, placeholderModelName, zoneID, placeholderZoneName, "Discovered", true, false); err != nil {
		return res, fmt.Errorf("create placeholder device: %w", err)
	}
	res.Device = label

	deviceID := ""
	if devs, err := ns.GetDevices(); err == nil {
		for _, d := range devs {
			if d.Name == label {
				deviceID = d.ID
				break
			}
		}
	}
	if deviceID == "" {
		return res, fmt.Errorf("placeholder device %q not found after creation", label)
	}

	// 4. For each observing endpoint, create a dedicated placeholder port and
	// connect the real endpoint to it.
	var errs []string
	for i, ep := range seenBy {
		realPortID, err := ns.resolveOrCreatePort("", ep)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", ep, err))
			continue
		}

		portName := fmt.Sprintf("p%d", i+1)
		if err := ns.AddModelPort(portName, "0", "0", placeholderModelName, true, "ethernet", ""); err != nil &&
			!strings.Contains(err.Error(), "already exists") {
			errs = append(errs, fmt.Sprintf("%s: model port: %v", ep, err))
			continue
		}
		modelPortID := ""
		if mports, err := ns.GetModelPorts(); err == nil {
			for _, mp := range mports {
				if mp.Model == placeholderModelName && mp.Name == portName {
					modelPortID = mp.ID
					break
				}
			}
		}
		if modelPortID == "" {
			errs = append(errs, fmt.Sprintf("%s: placeholder model port %q not found", ep, portName))
			continue
		}

		phPortID, err := ns.AddDevicePort(deviceID, modelPortID, "", nil)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: placeholder port: %v", ep, err))
			continue
		}

		if err := ns.AddConnection(realPortID, phPortID, "ethernet", provenance); err != nil {
			errs = append(errs, fmt.Sprintf("%s: connect: %v", ep, err))
			continue
		}
		res.Connections++
	}

	if len(errs) > 0 {
		return res, fmt.Errorf("%d endpoint(s) skipped/failed:\n  %s", len(errs), strings.Join(errs, "\n  "))
	}
	return res, nil
}

// ensurePlaceholderZone returns the id of the placeholder zone, creating it if
// it doesn't already exist.
func (ns *NetService) ensurePlaceholderZone() (string, error) {
	zones, err := ns.GetZones()
	if err != nil {
		return "", err
	}
	for _, z := range zones {
		if z.Name == placeholderZoneName {
			return z.ID, nil
		}
	}
	// Strict AddZone requires the referenced owner and zone type to exist.
	if err := ns.ensureOwner("Discovered"); err != nil {
		return "", err
	}
	if err := ns.ensureZoneType("Unknown"); err != nil {
		return "", err
	}
	if err := ns.AddZone(placeholderZoneName, "", "", "Discovered", "Unknown"); err != nil {
		return "", fmt.Errorf("create placeholder zone: %w", err)
	}
	zones, err = ns.GetZones()
	if err != nil {
		return "", err
	}
	for _, z := range zones {
		if z.Name == placeholderZoneName {
			return z.ID, nil
		}
	}
	return "", fmt.Errorf("placeholder zone not found after creation")
}
