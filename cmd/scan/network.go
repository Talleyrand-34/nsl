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
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
	q "nsl-graph/internal/repository/application"
	e "nsl-graph/internal/repository/entities"
	s "nsl-graph/internal/scanner"
)

var (
	scanSubnet       string
	scanTimeout      int
	scanCommunity    string
	scanSNMPVersion  string
	scanSNMPPort     uint16
	scanOutputFile   string
	scanAutoImport   bool
	scanMergeIPs     bool
	scanReviewMode   bool
	scanCreateZones  bool
	scanDefaultZone  string
	scanDefaultBrand string
	scanSkipExisting bool

	scanProfile     string
	scanSaveProfile string
	scanHuman       bool
	scanRaw         bool
)

var networkScanCmd = &cobra.Command{
	Use:   "network [subnet...]",
	Short: "Scan one or more network subnets to discover devices via SNMP",
	Long: `Scan one or more network subnets using SNMP to collect device inventory
(interfaces, MACs, VLANs). Pass several CIDRs as separate arguments or as a
single comma-separated value.

Output is JSON on stdout by default — an editable import plan, ready for
"scan import" (status goes to stderr, so it pipes cleanly). Pass -H/--human for
a readable summary plus interactive review, or --raw to emit the raw ScanResult.

Examples:
  nsl-graph scan network 192.168.1.0/24 > plan.json
  nsl-graph scan network 192.168.1.0/24 10.0.0.0/24 > plan.json
  nsl-graph scan network 192.168.1.0/24,10.0.0.0/24 --community private -H
  nsl-graph scan network 10.0.0.0/24 --community public --auto-import
  nsl-graph scan network 172.16.1.0/24 --raw -o scan_results.json`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		subnet := strings.Join(args, ",")
		if scanSubnet != "" {
			subnet = scanSubnet
		}

		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}

		// Apply a saved scan profile (explicit --profile, else auto-matched by
		// subnet). SNMP-only fields; explicit flags override.
		if p, ok := service.ResolveScanProfile(subnet, scanProfile); ok {
			fl := cmd.Flags()
			if !fl.Changed("community") && p.SNMPCommunity != "" {
				scanCommunity = p.SNMPCommunity
			}
			if !fl.Changed("snmp-version") && p.SNMPVersion != "" {
				scanSNMPVersion = p.SNMPVersion
			}
			if !fl.Changed("snmp-port") && p.SNMPPort != 0 {
				scanSNMPPort = uint16(p.SNMPPort)
			}
			if !fl.Changed("timeout") && p.TimeoutSec != 0 {
				scanTimeout = p.TimeoutSec
			}
			fmt.Fprintf(os.Stderr, "Applied scan profile %q\n", p.Name)
		} else if scanProfile != "" {
			fmt.Fprintf(os.Stderr, "Error: no scan profile named %q\n", scanProfile)
			os.Exit(1)
		}

		if scanSaveProfile != "" {
			p := e.ScanProfile{
				Name:          scanSaveProfile,
				Host:          subnet,
				SNMPCommunity: scanCommunity,
				SNMPVersion:   scanSNMPVersion,
				SNMPPort:      int(scanSNMPPort),
				TimeoutSec:    scanTimeout,
				ScanSource:    "snmp",
			}
			if err := service.AddScanProfile(p); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving profile: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Saved scan profile %q (host %s).\n", scanSaveProfile, subnet)
		}

		options := s.ScanOptions{
			Subnet:  subnet,
			Timeout: time.Duration(scanTimeout) * time.Second,
			SNMP: s.SNMPOptions{
				Community: scanCommunity,
				Version:   scanSNMPVersion,
				Port:      scanSNMPPort,
			},
		}

		fmt.Fprintf(os.Stderr, "Starting SNMP scan of %s (community=%s, version=%s)...\n",
			subnet, scanCommunity, scanSNMPVersion)

		scanResult, err := service.ScanNetwork(subnet, options)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Network scan failed: %v\n", err)
			os.Exit(1)
		}

		duration := scanResult.EndTime.Sub(scanResult.StartTime)
		fmt.Fprintf(os.Stderr, "Scan completed in %v\n", duration)
		fmt.Fprintf(os.Stderr, "Found %d reachable devices\n", len(scanResult.Devices))

		// --raw: emit the raw ScanResult (back-compat with the old output).
		if scanRaw {
			if err := emitScanJSON(scanResult, scanOutputFile); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
			return
		}

		if len(scanResult.Devices) == 0 {
			fmt.Fprintln(os.Stderr, "No SNMP-reachable devices found.")
			return
		}

		devices, err := service.DiscoverDevices(scanResult)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Device discovery failed: %v\n", err)
			os.Exit(1)
		}

		importOptions := s.ImportOptions{
			AutoImport:        true,
			MergeIPs:          scanMergeIPs,
			CreateZones:       scanCreateZones,
			DefaultZone:       scanDefaultZone,
			DefaultBrand:      scanDefaultBrand,
			SkipExisting:      scanSkipExisting,
			InteractiveVLANs:  true,
			VLANAccuracyLevel: 2,
		}

		// --auto-import: bulk import non-interactively (status to stderr).
		if scanAutoImport {
			if err := importDevicesWithVLANMapping(service, devices, importOptions); err != nil {
				fmt.Fprintf(os.Stderr, "Import failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintln(os.Stderr, "Devices imported successfully!")
			return
		}

		// Default: emit the editable import plan; -H/--review: interactive review.
		if err := emitOrReviewDevices(service, devices, scanHuman || scanReviewMode, scanOutputFile, importOptions); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	},
}

// importDevicesWithVLANMapping imports devices using interactive VLAN mapping
func importDevicesWithVLANMapping(service q.NetServiceInt, devices []s.DiscoveredDevice, options s.ImportOptions) error {
	for i, device := range devices {
		fmt.Fprintf(os.Stderr, "\n=== Device %d/%d: %s (%s) ===\n", i+1, len(devices), device.SuggestedName, device.Device.IP)

		// Use the same logic as scan host for consistency
		if err := importSingleDeviceWithVLANMapping(service, device, options); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to import device %s: %v\n", device.SuggestedName, err)
			if !options.ReviewMode {
				return err
			}
		} else {
			fmt.Fprintf(os.Stderr, "✓ Device %s imported successfully\n", device.SuggestedName)
		}
	}
	return nil
}

func init() {
	cmd_root.ScanCmd.AddCommand(networkScanCmd)

	networkScanCmd.Flags().StringVar(&scanSubnet, "subnet", "", "Network subnet(s) to scan, comma-separated (alternative to positional args)")
	networkScanCmd.Flags().IntVarP(&scanTimeout, "timeout", "t", 30, "Timeout in seconds for the scan")
	networkScanCmd.Flags().StringVar(&scanCommunity, "community", "public", "SNMP community string")
	networkScanCmd.Flags().StringVar(&scanSNMPVersion, "snmp-version", "v2c", "SNMP version (v1, v2c)")
	networkScanCmd.Flags().Uint16Var(&scanSNMPPort, "snmp-port", 161, "SNMP UDP port")

	networkScanCmd.Flags().StringVarP(&scanOutputFile, "output", "o", "", "Write the JSON output to this file instead of stdout")
	networkScanCmd.Flags().BoolVarP(&scanHuman, "human", "H", false, "Human-readable summary + interactive review (default output is the plan JSON)")
	networkScanCmd.Flags().BoolVar(&scanRaw, "raw", false, "Emit the raw ScanResult JSON instead of the editable import plan")

	networkScanCmd.Flags().BoolVar(&scanAutoImport, "auto-import", false, "Automatically import discovered devices")
	networkScanCmd.Flags().BoolVar(&scanMergeIPs, "merge-ips", false, "Merge IP addresses with existing devices")
	networkScanCmd.Flags().BoolVar(&scanReviewMode, "review", false, "Review devices before importing")
	networkScanCmd.Flags().BoolVar(&scanCreateZones, "create-zones", true, "Create zones for discovered devices")
	networkScanCmd.Flags().StringVar(&scanDefaultZone, "default-zone", "Discovered", "Default zone for discovered devices")
	networkScanCmd.Flags().StringVar(&scanDefaultBrand, "default-brand", "", "Default brand for unidentified devices")
	networkScanCmd.Flags().BoolVar(&scanSkipExisting, "skip-existing", true, "Skip devices with existing IP addresses")

	networkScanCmd.Flags().StringVar(&scanProfile, "profile", "", "Use a saved scan profile by name (else auto-matched by subnet)")
	networkScanCmd.Flags().StringVar(&scanSaveProfile, "save-profile", "", "Save the effective SNMP parameters as a profile with this name")
}
