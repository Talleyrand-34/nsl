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
	"sort"
	"strings"
	"sync"
	"time"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/observ"
	e "nsl-graph/internal/repository/entities"
	s "nsl-graph/internal/scanner"
)

// RunScanOptions controls a unified device scan: the method (snmp|ssh) is
// explicit; single vs batch is inferred from Target — a bare IP is a single
// host, while a CIDR or a comma/space-separated list is a batch.
type RunScanOptions struct {
	Method      string `json:"method"`
	Target      string `json:"target"`
	Community   string `json:"community"`
	SNMPVersion string `json:"snmp_version"`
	SNMPPort    uint16 `json:"snmp_port"`
	TimeoutSec  int    `json:"timeout_sec"`
	Profile     string `json:"profile"` // SSH: a profile (device or generic) supplying the credentials
	OsType      string `json:"os_type"` // SSH: overrides the profile's os_type (required for generic profiles)
}

// RunScan discovers devices for importation. It only orchestrates existing scan
// primitives (no new scanner logic): SNMP via ScanDevice/ScanNetwork, SSH via a
// device profile's credentials and ScanDeviceViaSSH (a CIDR is swept for
// SSH-open hosts first). The result is always a classified, uniform list ready
// for the analyze/import flow.
func (ns *NetService) RunScan(opts RunScanOptions, em observ.Emitter) ([]s.DiscoveredDevice, error) {
	if em == nil {
		em = observ.Discard
	}
	target := strings.TrimSpace(opts.Target)
	tokens := s.SplitSubnets(target)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("a target IP or CIDR is required")
	}
	batch := len(tokens) > 1 || strings.Contains(target, "/")

	method := strings.ToLower(strings.TrimSpace(opts.Method))
	if method == "" {
		method = "snmp"
	}
	em.Emit("info", "scan starting", "method", method, "target", target, "batch", batch)

	switch method {
	case "snmp":
		return ns.runSNMPScan(opts, target, batch, em)
	case "ssh":
		return ns.runSSHScan(opts, tokens, batch, em)
	default:
		return nil, fmt.Errorf("unknown scan method %q (want \"snmp\" or \"ssh\")", opts.Method)
	}
}

// runSNMPScan handles the SNMP method: a single host (ScanDevice) or a swept
// CIDR (ScanNetwork). A saved profile (explicit or auto-matched by target) fills
// any unset SNMP parameters, mirroring the per-endpoint handlers.
func (ns *NetService) runSNMPScan(opts RunScanOptions, target string, batch bool, em observ.Emitter) ([]s.DiscoveredDevice, error) {
	community, version, port, timeoutSec := opts.Community, opts.SNMPVersion, opts.SNMPPort, opts.TimeoutSec
	profileName := "" // tie discovered devices to whatever profile supplied the scan
	if p, ok := ns.ResolveScanProfile(target, opts.Profile); ok {
		profileName = p.Name
		em.Emit("info", "applied scan profile", "profile", p.Name)
		if community == "" {
			community = p.SNMPCommunity
		}
		if version == "" {
			version = p.SNMPVersion
		}
		if port == 0 {
			port = uint16(p.SNMPPort)
		}
		if timeoutSec == 0 {
			timeoutSec = p.TimeoutSec
		}
	}
	timeout := time.Duration(timeoutSec) * time.Second
	if timeoutSec == 0 {
		timeout = 10 * time.Second
		if batch {
			timeout = 30 * time.Second
		}
	}
	options := s.ScanOptions{
		Timeout: timeout,
		SNMP:    s.SNMPOptions{Community: community, Version: version, Port: port},
	}

	if batch {
		em.Emit("info", "sweeping subnet over SNMP", "community", community, "version", version)
		options.OnProgress = func(done, total int, ip string, reachable bool) {
			em.Progress(done, total)
			if reachable {
				em.Emit("info", "host responded to SNMP", "ip", ip, "done", done, "total", total)
			} else {
				em.Emit("debug", "no SNMP response", "ip", ip, "done", done, "total", total)
			}
		}
		res, err := ns.ScanNetwork(target, options)
		if err != nil {
			return nil, err
		}
		devs, err := ns.DiscoverDevices(res)
		if err == nil {
			em.Emit("info", "snmp sweep complete", "devices", len(devs))
		}
		return tagProfile(devs, profileName, ""), err
	}
	em.Emit("info", "scanning host over SNMP", "ip", target)
	em.Progress(0, 1)
	dev, err := ns.ScanDevice(target, options)
	if err != nil {
		return nil, err
	}
	em.Progress(1, 1)
	if !dev.Reachable {
		em.Emit("warn", "host did not respond to SNMP", "ip", target)
	} else {
		em.Emit("info", "host responded", "ip", target, "sysname", dev.SysName)
	}
	devs, err := ns.DiscoverDevices(&s.ScanResult{Devices: []s.SNMPDevice{*dev}})
	return tagProfile(devs, profileName, ""), err
}

// tagProfile records the scan-profile name (and, for SSH scans, the OS type used)
// on each discovered device so both are persisted when the device is imported.
func tagProfile(devs []s.DiscoveredDevice, profile, osType string) []s.DiscoveredDevice {
	for i := range devs {
		if profile != "" {
			devs[i].Profile = profile
		}
		if osType != "" {
			devs[i].OsType = osType
		}
	}
	return devs
}

// runSSHScan handles the SSH method using a saved profile's credentials. The
// profile may be a device profile or a generic (credential-only) one reused
// across hosts; in the latter case the caller supplies the OS type. The
// target overrides the profile's host; a CIDR is swept for SSH-open hosts and
// each is read with the same (homogeneous) os_type.
func (ns *NetService) runSSHScan(opts RunScanOptions, tokens []string, batch bool, em observ.Emitter) ([]s.DiscoveredDevice, error) {
	if opts.Profile == "" {
		return nil, fmt.Errorf("an SSH scan requires a profile supplying SSH credentials")
	}
	profile, err := ns.GetScanProfileByName(opts.Profile)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, fmt.Errorf("no scan profile named %q", opts.Profile)
	}
	// OS type: an explicit value wins (required for generic profiles, which
	// carry only credentials); otherwise fall back to a device profile's own.
	osType := strings.TrimSpace(opts.OsType)
	if osType == "" {
		osType = profile.OsType
	}
	if osType == "" {
		return nil, fmt.Errorf("an SSH scan needs a OS type — set one (the %q profile does not carry a os_type)", opts.Profile)
	}
	creds := profileSSHCreds(profile, ns.vault)
	if creds == nil {
		return nil, fmt.Errorf("could not unlock SSH credentials for profile %q (unlock the vault, and check it has an SSH user plus a key or password)", opts.Profile)
	}
	timeout := time.Duration(profile.ConfigTimeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	creds.Timeout = timeout
	sshPort := profile.SSHPort
	if sshPort == 0 {
		sshPort = 22
	}
	em.Emit("info", "ssh scan", "profile", profile.Name, "os_type", osType)

	var ips []string
	var perHostCreds []configparser.SSHCredentials // parallel to ips; empty entry = use the profile default
	devRows, rowsErr := ns.GetProfileDevices(profile.Name)
	if rowsErr != nil {
		return nil, rowsErr
	}
	if len(devRows) > 0 {
		em.Emit("info", "multi-device profile: iterating attached devices", "rows", len(devRows))
		for _, r := range devRows {
			ips = append(ips, r.Host)
			perHostCreds = append(perHostCreds, perRowOverrideCreds(r, profile, ns, em))
		}
	} else if batch {
		em.Emit("info", "sweeping subnet for open SSH ports")
		for _, h := range SweepSubnet(opts.Target, opts.Community, opts.SNMPVersion, timeout, sshPort, em) {
			if h.SSH {
				ips = append(ips, h.IP)
			}
		}
		em.Emit("info", "ssh-reachable hosts found", "count", len(ips))
	} else {
		ips = []string{tokens[0]}
	}
	// Read each host's config over SSH concurrently (bounded) — the hosts are
	// independent, so one slow box no longer blocks the rest.
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		devices  []s.SNMPDevice
		firstErr error
		done     int
	)
	total := len(ips)
	sem := make(chan struct{}, 16)
	em.Progress(0, total)
	for i, ip := range ips {
		wg.Add(1)
		go func(i int, ip string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Per-host SSH override wins over the profile-wide default.
			effective := *creds
			if i < len(perHostCreds) && perHostCreds[i].Username != "" {
				effective = perHostCreds[i]
			}

			em.Emit("info", "reading config over SSH", "ip", ip)
			dev, err := ns.ScanDeviceViaSSH(ip, osType, effective)

			mu.Lock()
			done++
			d := done
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
			} else {
				dev.IP = ip
				dev.Reachable = true
				devices = append(devices, *dev)
			}
			mu.Unlock()

			em.Progress(d, total)
			if err != nil {
				em.Emit("warn", "SSH scan failed for host", "ip", ip, "error", err.Error())
			} else {
				em.Emit("info", "device read over SSH", "ip", ip, "sysname", dev.SysName)
			}
		}(i, ip)
	}
	wg.Wait()

	// Stable, IP-sorted output (completion order is nondeterministic).
	sort.Slice(devices, func(i, j int) bool { return devices[i].IP < devices[j].IP })
	if len(devices) == 0 {
		if firstErr != nil {
			return nil, fmt.Errorf("SSH scan failed: %w", firstErr)
		}
		return nil, nil // no SSH-reachable hosts in the target
	}
	devs, err := ns.DiscoverDevices(&s.ScanResult{Devices: devices})
	return tagProfile(devs, profile.Name, osType), err
}

// perRowOverrideCreds resolves the effective SSH credentials for a single
// ProfileDevice row. When the row names another profile (via SSHProfileName),
// that profile's credentials win; otherwise the zero value is returned and
// the caller falls back to the row's parent profile's credentials.
func perRowOverrideCreds(row e.ProfileDevice, parent *e.ScanProfile, ns *NetService, em observ.Emitter) configparser.SSHCredentials {
	if row.SSHProfileName == "" {
		return configparser.SSHCredentials{}
	}
	override, err := ns.GetScanProfileByName(row.SSHProfileName)
	if err != nil || override == nil {
		em.Emit("warn", "row SSH override profile not found; falling back to parent", "row_host", row.Host, "missing_profile", row.SSHProfileName)
		return configparser.SSHCredentials{}
	}
	if creds := profileSSHCreds(override, ns.vault); creds != nil {
		em.Emit("info", "row SSH override applied", "row_host", row.Host, "override_profile", override.Name)
		return *creds
	}
	return configparser.SSHCredentials{}
}
