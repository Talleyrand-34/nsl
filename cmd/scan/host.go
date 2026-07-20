/*
Copyright © 2025 Talleyrand-34 (t34@t34.dev)

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
package cmd_scan

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"golang.org/x/term"
	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"

	configparser "nsl-graph/internal/configparser"
	_ "nsl-graph/internal/configparser/parsers" // side-effect: registers all parsers
	q "nsl-graph/internal/repository/application"
	e "nsl-graph/internal/repository/entities"
	s "nsl-graph/internal/scanner"
)

var (
	hostTimeout      int
	hostCommunity    string
	hostSNMPVersion  string
	hostSNMPPort     uint16
	hostAutoImport   bool
	hostDefaultZone  string
	hostDefaultBrand string
	hostOutputFile   string
	hostApproveAll   bool
	hostVLANAccuracy int
	hostScanSource   string // "snmp" or "ssh"

	// Configuration parsing options
	hostConfigSource      string
	hostConfigFile        string
	hostOsType            string
	hostSSHUsername       string
	hostSSHPassword       string
	hostSSHKeyFile        string
	hostSSHPort           int
	hostDiscrepancyAction string
	hostMergeConfig       bool
	hostConfigTimeout     int

	hostProfile     string // use a named saved profile
	hostSaveProfile string // persist effective params as a profile

	hostHuman bool // -H: human-readable summary + interactive review
	hostRaw   bool // emit the raw ScanResult instead of the import plan
)

var hostScanCmd = &cobra.Command{
	Use:   "host [ip_or_ssh_alias]",
	Short: "Query a specific host via SNMP to collect inventory",
	Long: `Query a specific device to collect interfaces, MAC addresses, and VLAN assignments.

The argument can be an IP address or an SSH config alias from ~/.ssh/config.
When an alias is given, HostName/User/IdentityFile/Port are resolved automatically;
explicit flags (--ssh-user, --ssh-key, --ssh-port) always take precedence.

Use --scan-source to choose the primary protocol:
  snmp (default): query via SNMP; optionally augment with SSH/file config
  ssh:            skip SNMP entirely; use SSH configuration as the primary source

Examples:
  # SNMP scanning (default)
  nsl-graph scan host 192.168.1.1
  nsl-graph scan host 10.0.0.1 --snmp-community private
  nsl-graph scan host 192.168.1.1 --output device_scan.json
  nsl-graph scan host 172.16.1.10 --auto-import

  # SSH-only scanning via IP
  nsl-graph scan host 10.0.0.245 --scan-source ssh --ssh-user root --auto-import
  nsl-graph scan host 10.0.0.1   --scan-source ssh --ssh-user admin --os-type opnsense --auto-import

  # SSH-only scanning via ~/.ssh/config alias (resolves host/user/key automatically)
  nsl-graph scan host opnsense --scan-source ssh --auto-import
  nsl-graph scan host opnsense --auto-import          # scan-source defaults to ssh for aliases
  nsl-graph scan host opnsense --ssh-user admin       # override a single field

  # SNMP + SSH configuration augmentation
  nsl-graph scan host 10.0.0.1 --config-source ssh --ssh-user admin --merge-configs
  nsl-graph scan host 10.0.0.132 --config-source ssh --ssh-user admin --discrepancy-action prefer-config
  nsl-graph scan host 10.0.0.245 --config-source file --config-file openwrt.conf --auto-import`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ip := args[0]

		// If the argument is not a valid IP address, treat it as an ~/.ssh/config alias.
		if net.ParseIP(ip) == nil {
			entry, err := parseSSHConfigAlias(ip)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %q is not a valid IP address and could not be resolved from ~/.ssh/config: %v\n", ip, err)
				os.Exit(1)
			}
			if entry.HostName != "" {
				fmt.Fprintf(os.Stderr, "Resolved SSH alias %q → %s\n", ip, entry.HostName)
				ip = entry.HostName
			}
			// Apply resolved values only when the flag was not set explicitly
			if !cmd.Flags().Changed("ssh-user") && entry.User != "" {
				hostSSHUsername = entry.User
			}
			if !cmd.Flags().Changed("ssh-key") && entry.IdentityFile != "" {
				hostSSHKeyFile = entry.IdentityFile
			}
			if !cmd.Flags().Changed("ssh-port") && entry.Port != 0 {
				hostSSHPort = entry.Port
			}
			// Default scan-source to ssh when an alias is used
			if !cmd.Flags().Changed("scan-source") {
				hostScanSource = "ssh"
			}
		}

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}

		// Apply a saved scan profile (explicit --profile, else auto-matched by
		// host). A field is taken from the profile only when its flag was not set
		// explicitly; SSH string fields already resolved from ~/.ssh/config are
		// kept (alias > profile > default).
		var profile *e.ScanProfile
		if p, ok := service.ResolveScanProfile(ip, hostProfile); ok {
			profile = p
			fl := cmd.Flags()
			if !fl.Changed("snmp-community") && profile.SNMPCommunity != "" {
				hostCommunity = profile.SNMPCommunity
			}
			if !fl.Changed("snmp-version") && profile.SNMPVersion != "" {
				hostSNMPVersion = profile.SNMPVersion
			}
			if !fl.Changed("snmp-port") && profile.SNMPPort != 0 {
				hostSNMPPort = uint16(profile.SNMPPort)
			}
			if !fl.Changed("timeout") && profile.TimeoutSec != 0 {
				hostTimeout = profile.TimeoutSec
			}
			if !fl.Changed("scan-source") && profile.ScanSource != "" {
				hostScanSource = profile.ScanSource
			}
			if !fl.Changed("config-source") && profile.ConfigSource != "" {
				hostConfigSource = profile.ConfigSource
			}
			if !fl.Changed("config-file") && profile.ConfigFile != "" {
				hostConfigFile = profile.ConfigFile
			}
			if !fl.Changed("os-type") && profile.OsType != "" {
				hostOsType = profile.OsType
			}
			if !fl.Changed("ssh-user") && hostSSHUsername == "" {
				hostSSHUsername = profile.SSHUser
			}
			if !fl.Changed("ssh-key") && hostSSHKeyFile == "" {
				hostSSHKeyFile = profile.SSHKeyFile
			}
			if !fl.Changed("ssh-port") && profile.SSHPort != 0 {
				hostSSHPort = profile.SSHPort
			}
			if !fl.Changed("discrepancy-action") && profile.DiscrepancyAction != "" {
				hostDiscrepancyAction = profile.DiscrepancyAction
			}
			if !fl.Changed("merge-configs") && profile.MergeConfigs {
				hostMergeConfig = true
			}
			if !fl.Changed("config-timeout") && profile.ConfigTimeout != 0 {
				hostConfigTimeout = profile.ConfigTimeout
			}
			if !fl.Changed("vlan-accuracy") && profile.VLANAccuracy != 0 {
				hostVLANAccuracy = profile.VLANAccuracy
			}
			if hostProfile != "" {
				fmt.Fprintf(os.Stderr, "Using scan profile %q\n", profile.Name)
			} else {
				fmt.Fprintf(os.Stderr, "Auto-applied scan profile %q (host %s)\n", profile.Name, profile.Host)
			}
		} else if hostProfile != "" {
			fmt.Fprintf(os.Stderr, "Error: no scan profile named %q\n", hostProfile)
			os.Exit(1)
		}

		// Persist the effective parameters as a profile if requested.
		if hostSaveProfile != "" {
			if err := saveHostProfile(service, hostSaveProfile, ip); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving profile: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Saved scan profile %q (host %s).\n", hostSaveProfile, ip)
		}

		var device *s.SNMPDevice

		switch hostScanSource {
		case "ssh":
			if hostSSHUsername == "" {
				fmt.Fprintln(os.Stderr, "Error: --ssh-user is required when using --scan-source ssh")
				os.Exit(1)
			}
			if hostOsType == "" {
				fmt.Fprintf(os.Stderr, "Error: --os-type is required with --scan-source ssh\n")
				fmt.Fprintf(os.Stderr, "  Supported: %s\n", strings.Join(configparser.DefaultRegistry.ListParsers(), ", "))
				os.Exit(1)
			}
			parser, found := configparser.DefaultRegistry.GetParser(hostOsType)
			if !found {
				fmt.Fprintf(os.Stderr, "Error: unknown OS type %q\n  Supported: %s\n",
					hostOsType, strings.Join(configparser.DefaultRegistry.ListParsers(), ", "))
				os.Exit(1)
			}
			// A profile may carry an encrypted SSH password; unlock the credential
			// vault to decrypt it instead of prompting for the raw password.
			if hostSSHKeyFile == "" && hostSSHPassword == "" && profile != nil && profile.SSHPassword != "" {
				if err := ensureVaultUnlocked(service.Vault()); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to unlock credential vault: %v\n", err)
					os.Exit(1)
				}
				pw, err := service.Vault().Decrypt(profile.SSHPassword)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}
				hostSSHPassword = pw
			}
			if hostSSHKeyFile == "" && hostSSHPassword == "" {
				fmt.Fprintf(os.Stderr, "Password for %s@%s: ", hostSSHUsername, ip)
				raw, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Fprintln(os.Stderr)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to read password: %v\n", err)
					os.Exit(1)
				}
				hostSSHPassword = string(raw)
			}

			creds := configparser.SSHCredentials{
				Username: hostSSHUsername,
				Password: hostSSHPassword,
				KeyFile:  hostSSHKeyFile,
				Port:     hostSSHPort,
				Timeout:  time.Duration(hostConfigTimeout) * time.Second,
			}
			fmt.Fprintf(os.Stderr, "Scanning %s via SSH (user=%s, type=%s)...\n", ip, hostSSHUsername, hostOsType)
			rawConfig, err := configparser.FetchConfig(configparser.DefaultTransport, parser, ip, creds)
			if err != nil {
				fmt.Fprintf(os.Stderr, "SSH connection failed: %v\n", err)
				os.Exit(1)
			}
			stub := s.SNMPDevice{IP: ip, Reachable: true, SysName: ip}
			configData, err := parser.ParseConfig(rawConfig, stub)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Config parsing failed: %v\n", err)
				os.Exit(1)
			}
			device = configparser.ConfigDataToSNMPDevice(configData, ip)
			if hostHuman {
				fmt.Printf("\nSSH Results for %s:\n", ip)
				if device.SysName != ip {
					fmt.Printf("Name:     %s\n", device.SysName)
				}
				printDeviceInfo(device)
				printControlPlane(configData)
			}

		default: // "snmp"
			options := s.ScanOptions{
				Timeout: time.Duration(hostTimeout) * time.Second,
				SNMP: s.SNMPOptions{
					Community: hostCommunity,
					Version:   hostSNMPVersion,
					Port:      hostSNMPPort,
				},
			}

			fmt.Fprintf(os.Stderr, "Querying %s via SNMP (community=%s)...\n", ip, hostCommunity)

			var err error
			device, err = service.ScanDevice(ip, options)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Device scan failed: %v\n", err)
				os.Exit(1)
			}

			if !device.Reachable {
				fmt.Fprintf(os.Stderr, "Device %s did not respond to SNMP\n", ip)
				return
			}

			if hostHuman {
				fmt.Printf("\nSNMP Results for %s:\n", ip)
				if device.SysName != "" {
					fmt.Printf("Name:     %s\n", device.SysName)
				}
				if device.SysDescr != "" {
					descr := device.SysDescr
					if len(descr) > 80 {
						descr = descr[:80] + "..."
					}
					fmt.Printf("Descr:    %s\n", descr)
				}
				if device.SysLocation != "" {
					fmt.Printf("Location: %s\n", device.SysLocation)
				}
				if device.SysContact != "" {
					fmt.Printf("Contact:  %s\n", device.SysContact)
				}
				printDeviceInfo(device)

				if len(device.Neighbors) > 0 {
					fmt.Printf("\nNeighbors (%d):\n", len(device.Neighbors))
					for _, n := range device.Neighbors {
						fmt.Printf(
							"  [%s] local:%s → remote:%s (%s)",
							n.Protocol, n.LocalPort, n.RemoteName, n.RemotePort,
						)
						if n.RemoteIP != "" {
							fmt.Printf(" IP:%s", n.RemoteIP)
						}
						fmt.Println()
					}
				}
			}
		}

		scanResult := &s.ScanResult{
			ID:        fmt.Sprintf("host_%s_%d", ip, time.Now().Unix()),
			Subnet:    ip + "/32", // Single host as /32
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Devices:   []s.SNMPDevice{*device},
		}

		// --raw: emit the raw ScanResult (back-compat with the old --output).
		if hostRaw {
			if err := emitScanJSON(scanResult, hostOutputFile); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
			return
		}

		devices, err := service.DiscoverDevices(scanResult)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Device discovery failed: %v\n", err)
			os.Exit(1)
		}
		if len(devices) == 0 {
			fmt.Fprintln(os.Stderr, "No devices discovered from scan result.")
			return
		}
		if hostDefaultZone != "" {
			devices[0].SuggestedZone = hostDefaultZone
		}

		importOptions := s.ImportOptions{
			AutoImport:        true,
			CreateZones:       true,
			DefaultZone:       hostDefaultZone,
			DefaultBrand:      hostDefaultBrand,
			SkipExisting:      true,
			InteractiveVLANs:  true,
			VLANAccuracyLevel: hostVLANAccuracy,

			// Configuration parsing options. When --scan-source ssh is used, SSH is
			// already the primary source.
			ConfigSource:       hostConfigSource,
			ConfigFile:         hostConfigFile,
			OsType:             hostOsType,
			SSHUsername:        hostSSHUsername,
			SSHPassword:        hostSSHPassword,
			SSHKeyFile:         hostSSHKeyFile,
			SSHPort:            hostSSHPort,
			DiscrepancyAction:  hostDiscrepancyAction,
			MergeWithConfig:    hostMergeConfig,
			ParseConfigTimeout: hostConfigTimeout,
		}

		// When using --scan-source ssh without --merge-configs, disable config merging.
		if hostScanSource == "ssh" && !hostMergeConfig {
			importOptions.ConfigSource = "none"
			importOptions.MergeWithConfig = false
		}
		// When --merge-configs is used, default to prefer-config unless explicitly set.
		if hostMergeConfig && hostDiscrepancyAction == "prefer-snmp" {
			importOptions.DiscrepancyAction = "prefer-config"
		}

		// --auto-import: import directly (non-interactive); status to stderr.
		if hostAutoImport {
			if err := importSingleDeviceWithVLANMapping(service, devices[0], importOptions); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to import device: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Device '%s' imported successfully!\n", devices[0].SuggestedName)
			return
		}

		// Default: emit the editable import plan; -H: interactive review.
		if err := emitOrReviewDevices(service, devices, hostHuman, hostOutputFile, importOptions); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	},
}

// ifTypeLabel returns a short human-readable label for an SNMP ifType value.
func ifTypeLabel(ifType int) string {
	switch ifType {
	case s.IfTypeEthernetCsmacd:
		return "ethernet"
	case s.IfTypeLag:
		return "lag"
	case s.IfTypeIEEE80211:
		return "wifi-radio"
	case s.IfTypeLoopback:
		return "loopback"
	case s.IfTypeTunnel:
		return "tunnel"
	case s.IfTypePropVirtual:
		return "virtual"
	default:
		return fmt.Sprintf("type%d", ifType)
	}
}

func printDeviceInfo(device *s.SNMPDevice) {
	// Print physical ports from SwitchPorts (e.g., OpenWrt board.json data)
	if len(device.SwitchPorts) > 0 {
		fmt.Printf("\nPhysical Ports (%d):\n", len(device.SwitchPorts))
		printSwitchPortsTable(device.SwitchPorts)
	}

	// Print WiFi radios from interfaces
	var radios []s.DeviceInterface
	for _, iface := range device.Interfaces {
		if iface.IfType == s.IfTypeIEEE80211 {
			radios = append(radios, iface)
		}
	}
	if len(radios) > 0 {
		fmt.Printf("\nWiFi Radios (%d):\n", len(radios))
		printIfaceTable(radios, false)
	}

	// Print logical interfaces
	var logical []s.DeviceInterface
	for _, iface := range device.Interfaces {
		if iface.IfType != s.IfTypeIEEE80211 {
			logical = append(logical, iface)
		}
	}

	// Build MAC → physical port name map for parent resolution
	macToPort := make(map[string]string)
	for _, iface := range device.Interfaces {
		if iface.MAC == "" {
			continue
		}
		if _, seen := macToPort[iface.MAC]; !seen {
			macToPort[iface.MAC] = iface.Name
		} else if len(iface.IPAddresses) > 0 {
			macToPort[iface.MAC] = iface.Name
		}
	}

	if len(logical) > 0 {
		fmt.Printf("\nLogical Interfaces (%d):\n", len(logical))
		printIfaceTable(logical, true)
	}
}

// printControlPlane reports how the device learns routes.
//
// On a topology where every variant shares one physical shape and one
// addressing plan — the lab's static/OSPF/iBGP rings are exactly that — the
// interface table is identical across all three and says nothing about which
// design is deployed. This is the part of a scan that does.
func printControlPlane(cd *configparser.ConfigData) {
	plane := cd.ControlPlane()
	if len(plane) == 0 {
		return
	}

	fmt.Printf("\nControl Plane: %s\n", strings.Join(plane, " + "))
	for _, p := range cd.RoutingProtocols {
		if !p.Enabled {
			continue
		}
		line := "  " + p.Type
		if p.Instance != "" {
			line += fmt.Sprintf(" [%s]", p.Instance)
		}
		if p.RouterID != "" {
			line += fmt.Sprintf("  router-id %s", p.RouterID)
		}
		if p.LocalAS != "" {
			line += fmt.Sprintf("  AS %s", p.LocalAS)
			if p.HasIBGP() {
				line += " (iBGP)"
			}
		}
		if p.VRF != "" {
			line += fmt.Sprintf("  vrf %s", p.VRF)
		}
		fmt.Println(line)

		for _, a := range p.Areas {
			detail := ""
			switch {
			case len(a.Networks) > 0:
				detail = strings.Join(a.Networks, ", ")
			case len(a.Interfaces) > 0:
				detail = strings.Join(a.Interfaces, ", ")
			}
			areaLine := fmt.Sprintf("    area %s", a.ID)
			if a.Type != "" && a.Type != "default" {
				areaLine += fmt.Sprintf(" (%s)", a.Type)
			}
			if detail != "" {
				areaLine += ": " + detail
			}
			fmt.Println(areaLine)
		}
		for _, n := range p.Neighbors {
			nb := fmt.Sprintf("    neighbor %s  remote-as %s", n.Address, n.RemoteAS)
			if n.IsIBGP(p.LocalAS) {
				nb += " (internal)"
			}
			if n.UpdateSource != "" {
				nb += fmt.Sprintf("  via %s", n.UpdateSource)
			}
			if n.Description != "" {
				nb += fmt.Sprintf("  %q", n.Description)
			}
			fmt.Println(nb)
		}
		if len(p.Networks) > 0 {
			fmt.Printf("    networks: %s\n", strings.Join(p.Networks, ", "))
		}
		if len(p.Redistribute) > 0 {
			fmt.Printf("    redistribute: %s\n", strings.Join(p.Redistribute, ", "))
		}
	}

	if n := len(cd.Routes); n > 0 {
		fmt.Printf("  static routes: %d\n", n)
	}
}

func printSwitchPortsTable(ports []s.PhysicalPortInfo) {
	const (
		wName   = 22
		wType   = 10
		wMAC    = 17
		wStatus = 6
		wInfo   = 20
		wIPs    = 32
	)
	hdr := fmt.Sprintf("  %-*s %-*s %-*s %-*s %-*s %-*s",
		wName, "Name", wType, "Type", wMAC, "MAC", wStatus, "Status", wInfo, "Info", wIPs, "IPs")
	sep := fmt.Sprintf("  %-*s %-*s %-*s %-*s %-*s %-*s",
		wName, strings.Repeat("-", wName),
		wType, strings.Repeat("-", wType),
		wMAC, strings.Repeat("-", wMAC),
		wStatus, strings.Repeat("-", wStatus),
		wInfo, strings.Repeat("-", wInfo),
		wIPs, strings.Repeat("-", wIPs))
	fmt.Println(hdr)
	fmt.Println(sep)

	for _, port := range ports {
		status := "down"
		if port.LinkStatus == "up" {
			status = "up"
		}

		info := ""
		if port.Role != "" {
			info = port.Role
		}

		if len(info) > wInfo {
			info = info[:wInfo-1] + "…"
		}

		mac := "-"
		if port.MAC != "" {
			mac = port.MAC
		}

		var vlanInfo string
		for _, vlan := range port.VLANs {
			if vlanInfo != "" {
				vlanInfo += " "
			}
			tag := "U"
			if vlan.Tagged {
				tag = "T"
			}
			vlanInfo += fmt.Sprintf("%s:%s", vlan.VID, tag)
		}

		ips := "-"
		if vlanInfo != "" {
			ips = vlanInfo
		}

		fmt.Printf("  %-*s %-*s %-*s %-*s %-*s %-*s\n",
			wName, port.Name,
			wType, "ethernet",
			wMAC, mac,
			wStatus, status,
			wInfo, info,
			wIPs, ips)
	}
}

func printIfaceTable(ifaces []s.DeviceInterface, showParent bool) {
	const (
		wName   = 22
		wType   = 10
		wMAC    = 17
		wStatus = 6
		wInfo   = 20
		wIPs    = 32
		wParent = 14
	)
	hdr := fmt.Sprintf("  %-*s %-*s %-*s %-*s %-*s %-*s",
		wName, "Name", wType, "Type", wMAC, "MAC", wStatus, "Status", wInfo, "Info", wIPs, "IPs")
	sep := fmt.Sprintf("  %-*s %-*s %-*s %-*s %-*s %-*s",
		wName, strings.Repeat("-", wName),
		wType, strings.Repeat("-", wType),
		wMAC, strings.Repeat("-", wMAC),
		wStatus, strings.Repeat("-", wStatus),
		wInfo, strings.Repeat("-", wInfo),
		wIPs, strings.Repeat("-", wIPs))
	if showParent {
		hdr += fmt.Sprintf(" %-*s", wParent, "Parent")
		sep += fmt.Sprintf(" %-*s", wParent, strings.Repeat("-", wParent))
	}
	fmt.Println(hdr)
	fmt.Println(sep)

	// Build MAC → physical port name map for parent resolution in SNMP mode
	macToPort := make(map[string]string)
	for _, iface := range ifaces {
		if iface.MAC == "" {
			continue
		}
		if _, seen := macToPort[iface.MAC]; !seen {
			macToPort[iface.MAC] = iface.Name
		} else if len(iface.IPAddresses) > 0 {
			macToPort[iface.MAC] = iface.Name
		}
	}

	for _, iface := range ifaces {
		status := "down"
		if iface.OperStatus == 1 {
			status = "up"
		}

		// Info column: wifi band for radios, SSID+security for wifi-iface, and
		// otherwise the operator's own description — on a firewall whose ports
		// are all named vtnetN, "RINGOWRTO (opt1)" is the only thing in the row
		// that says what the link actually is.
		info := ""
		if iface.WifiBand != "" {
			info = iface.WifiBand
		} else if iface.WifiSSID != "" {
			info = iface.WifiSSID
			if iface.WifiSecurity != "" && iface.WifiSecurity != "open" {
				info += " (" + iface.WifiSecurity + ")"
			}
		} else if iface.Description != "" {
			info = iface.Description
		}
		if len(info) > wInfo {
			info = info[:wInfo-1] + "…"
		}

		ips := ""
		if len(iface.IPAddresses) > 0 {
			ips = strings.Join(iface.IPAddresses, ", ")
			if len(ips) > wIPs {
				ips = ips[:wIPs-1] + "…"
			}
		}

		mac := iface.MAC
		if mac == "" {
			mac = "-"
		}

		line := fmt.Sprintf("  %-*s %-*s %-*s %-*s %-*s %-*s",
			wName, iface.Name,
			wType, ifTypeLabel(iface.IfType),
			wMAC, mac,
			wStatus, status,
			wInfo, info,
			wIPs, ips)

		if showParent {
			parent := iface.Parent
			if parent == "" && iface.MAC != "" {
				// Inferring the parent from a shared MAC finds the physical port
				// a subinterface sits on — but a physical port shares its MAC
				// with itself, so the lookup returns the row's own name. That is
				// not a parent relationship, it is the absence of one.
				if p := macToPort[iface.MAC]; p != iface.Name {
					parent = p
				}
			}
			line += fmt.Sprintf(" %-*s", wParent, parent)
		}
		fmt.Println(line)

		// Sub-line: VLANs
		if len(iface.VLANs) > 0 {
			var vlanParts []string
			for _, v := range iface.VLANs {
				tag := "U"
				if v.Tagged {
					tag = "T"
				}
				vlanParts = append(vlanParts, v.VLANNumber+":"+tag)
			}
			fmt.Printf("  %*s VLANs: %s\n", wName, "", strings.Join(vlanParts, "  "))
		}
	}
}

// importSingleDeviceWithVLANMapping imports a device with interactive VLAN mapping
func importSingleDeviceWithVLANMapping(
	service q.NetServiceInt,
	device s.DiscoveredDevice,
	options s.ImportOptions,
) error {
	// Analyze device for import plan
	plan, err := service.AnalyzeDeviceForImport(device)
	if err != nil {
		return fmt.Errorf("failed to analyze device: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Analysis: %s\n", plan.Summary)

	// Show interface details with IP-VLAN mappings if any
	for i, interfacePlan := range plan.InterfacePlans {
		iface := interfacePlan.Interface
		fmt.Fprintf(os.Stderr, "  Interface %d: %s", i+1, iface.Name)

		if len(iface.IPAddresses) > 0 {
			fmt.Fprintf(os.Stderr, " (IPs: %v)", iface.IPAddresses)
		}

		if len(interfacePlan.IPMappings) > 0 {
			fmt.Fprintf(os.Stderr, "\n    Proposed VLAN mappings:")
			for _, mapping := range interfacePlan.IPMappings {
				fmt.Fprintf(os.Stderr,
					"\n      %s → VLAN %s (%s)",
					mapping.IP,
					mapping.VLANNumber,
					mapping.Confidence,
				)
			}
		}
		fmt.Fprintln(os.Stderr)
	}

	if plan.RequiresInput {
		if hostApproveAll {
			fmt.Fprintf(os.Stderr, "Auto-approving VLAN mapping (--approve-all flag)\n")
		} else {
			fmt.Fprintf(os.Stderr, "\nApprove this VLAN mapping? (y/n): ")
			var response string
			fmt.Scanln(&response)

			if strings.ToLower(strings.TrimSpace(response)) != "y" && strings.ToLower(strings.TrimSpace(response)) != "yes" {
				fmt.Fprintln(os.Stderr, "Import cancelled by user.")
				return nil
			}
		}
	}

	// Execute the approved plan
	if err := service.ExecuteApprovedImportPlan(plan, options); err != nil {
		return fmt.Errorf("failed to execute import plan: %w", err)
	}

	return nil
}

// saveHostProfile persists the effective scan parameters as a named profile.
// If an SSH password was provided explicitly, it is encrypted by the credential
// vault before storage.
func saveHostProfile(service q.NetServiceInt, name, host string) error {
	p := e.ScanProfile{
		Name:              name,
		Host:              host,
		SNMPCommunity:     hostCommunity,
		SNMPVersion:       hostSNMPVersion,
		SNMPPort:          int(hostSNMPPort),
		TimeoutSec:        hostTimeout,
		ScanSource:        hostScanSource,
		ConfigSource:      hostConfigSource,
		ConfigFile:        hostConfigFile,
		OsType:            hostOsType,
		SSHUser:           hostSSHUsername,
		SSHKeyFile:        hostSSHKeyFile,
		SSHPort:           hostSSHPort,
		DiscrepancyAction: hostDiscrepancyAction,
		MergeConfigs:      hostMergeConfig,
		ConfigTimeout:     hostConfigTimeout,
		VLANAccuracy:      hostVLANAccuracy,
	}
	if hostSSHPassword != "" {
		if err := ensureVaultUnlocked(service.Vault()); err != nil {
			return err
		}
		blob, err := service.Vault().Encrypt(hostSSHPassword)
		if err != nil {
			return err
		}
		p.SSHPassword = blob
	}
	return service.AddScanProfile(p)
}

func init() {
	cmd_root.ScanCmd.AddCommand(hostScanCmd)

	hostScanCmd.Flags().
		StringVar(&hostScanSource, "scan-source", "snmp", "Primary scan protocol: snmp (default) or ssh (skip SNMP, use SSH config as primary source)")
	hostScanCmd.Flags().
		IntVarP(&hostTimeout, "timeout", "t", 10, "Timeout in seconds for the SNMP query")
	hostScanCmd.Flags().
		StringVar(&hostCommunity, "snmp-community", "public", "SNMP community string")
	hostScanCmd.Flags().StringVar(&hostSNMPVersion, "snmp-version", "v2c", "SNMP version (v1, v2c)")
	hostScanCmd.Flags().Uint16Var(&hostSNMPPort, "snmp-port", 161, "SNMP UDP port")

	hostScanCmd.Flags().
		StringVarP(&hostOutputFile, "output", "o", "", "Write the JSON output to this file instead of stdout")
	hostScanCmd.Flags().
		BoolVarP(&hostHuman, "human", "H", false, "Human-readable summary + interactive review (default output is the plan JSON)")
	hostScanCmd.Flags().
		BoolVar(&hostRaw, "raw", false, "Emit the raw ScanResult JSON instead of the editable import plan")

	hostScanCmd.Flags().
		BoolVar(&hostAutoImport, "auto-import", false, "Automatically import discovered device")
	hostScanCmd.Flags().
		StringVar(&hostDefaultZone, "default-zone", "", "Default zone for the device (defaults to Generic)")
	hostScanCmd.Flags().
		StringVar(&hostDefaultBrand, "default-brand", "", "Default brand for unidentified device")
	hostScanCmd.Flags().
		BoolVar(&hostApproveAll, "approve-all", false, "Automatically approve all VLAN mappings without prompting (only with --auto-import)")
	hostScanCmd.Flags().
		IntVar(&hostVLANAccuracy, "vlan-accuracy", 1, "VLAN detection accuracy level (1=interface names only, 2=include IP heuristics)")

	// Configuration parsing flags
	hostScanCmd.Flags().
		StringVar(&hostConfigSource, "config-source", "none", "Configuration source (none, ssh, file, manual)")
	hostScanCmd.Flags().
		StringVar(&hostConfigFile, "config-file", "", "Path to device configuration file (when using file source)")
	hostScanCmd.Flags().
		StringVar(&hostOsType, "os-type", "", "Device OS type (opnsense, openwrt, fortinet, cisco) - auto-detected if not specified")
	hostScanCmd.Flags().
		StringVar(&hostSSHUsername, "ssh-user", "", "SSH username for configuration retrieval")
	hostScanCmd.Flags().
		StringVar(&hostSSHPassword, "ssh-password", "", "SSH password for configuration retrieval")
	hostScanCmd.Flags().StringVar(&hostSSHKeyFile, "ssh-key", "", "Path to SSH private key file")
	hostScanCmd.Flags().IntVar(&hostSSHPort, "ssh-port", 22, "SSH port for configuration retrieval")
	hostScanCmd.Flags().
		StringVar(&hostDiscrepancyAction, "discrepancy-action", "prefer-snmp", "Action when SNMP and config data conflict (fail, prefer-snmp, prefer-config)")
	hostScanCmd.Flags().
		BoolVar(&hostMergeConfig, "merge-configs", false, "Merge SNMP data with configuration data")
	hostScanCmd.Flags().
		IntVar(&hostConfigTimeout, "config-timeout", 60, "Configuration parsing timeout in seconds")

	hostScanCmd.Flags().
		StringVar(&hostProfile, "profile", "", "Use a saved scan profile by name (else auto-matched by host)")
	hostScanCmd.Flags().
		StringVar(&hostSaveProfile, "save-profile", "", "Save the effective scan parameters as a profile with this name")
}
