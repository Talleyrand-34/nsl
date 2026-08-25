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

	"github.com/spf13/cobra"
	"golang.org/x/term"

	cmd_pkg "nsl-graph/cmd"
	cmd_root "nsl-graph/cmd/root"
	util "nsl-graph/cmd/utils"
	e "nsl-graph/internal/repository/entities"
	"nsl-graph/internal/secret"
)

var (
	profHost          string
	profCommunity     string
	profSNMPVersion   string
	profSNMPPort      int
	profTimeout       int
	profScanSource    string
	profConfigSource  string
	profConfigFile    string
	profOsType        string
	profSSHUser       string
	profSSHPassword   string
	profSSHKeyFile    string
	profSSHPort       int
	profDiscrepancy   string
	profMergeConfigs  bool
	profConfigTimeout int
	profVLANAccuracy  int
	profGeneric       bool
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage saved scan profiles (per-host scan parameters)",
	Long: `Store reusable scanning parameters for a host so they are entered once and
never re-queried. A profile is auto-applied when you 'scan host <ip>' and a
profile whose host matches that IP exists; explicit flags always override.

The SSH password is encrypted at rest with a passphrase you supply; it is never
stored in clear and is decrypted only when an SSH scan actually needs it.`,
	RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() },
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved scan profiles",
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		profiles, err := service.GetScanProfiles()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if len(profiles) == 0 {
			fmt.Println("No scan profiles.")
			return
		}
		fmt.Printf("%-16s %-8s %-18s %-10s %-6s %-6s %-10s %s\n",
			"NAME", "KIND", "HOST", "COMMUNITY", "VER", "PORT", "SSH-USER", "SSH-PASS")
		for _, p := range profiles {
			pass := "no"
			if p.HasSSHPassword {
				pass = "yes (enc)"
			}
			kind := p.Kind
			if kind == "" {
				kind = "device"
			}
			fmt.Printf("%-16s %-8s %-18s %-10s %-6s %-6d %-10s %s\n",
				p.Name, kind, p.Host, p.SNMPCommunity, p.SNMPVersion, p.SNMPPort, p.SSHUser, pass)
		}
	},
}

var profileShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show a saved scan profile (SSH password is never printed)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		p, err := service.GetScanProfileByName(args[0])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if p == nil {
			fmt.Printf("No profile named %q\n", args[0])
			os.Exit(1)
		}
		pass := "not set"
		if p.SSHPassword != "" {
			pass = "set (encrypted)"
		}
		kind := p.Kind
		if kind == "" {
			kind = "device"
		}
		fmt.Printf("Name:               %s\n", p.Name)
		fmt.Printf("Kind:               %s\n", kind)
		fmt.Printf("Host:               %s\n", p.Host)
		fmt.Printf("SNMP community:     %s\n", p.SNMPCommunity)
		fmt.Printf("SNMP version:       %s\n", p.SNMPVersion)
		fmt.Printf("SNMP port:          %d\n", p.SNMPPort)
		fmt.Printf("Timeout (s):        %d\n", p.TimeoutSec)
		fmt.Printf("Scan source:        %s\n", p.ScanSource)
		fmt.Printf("Config source:      %s\n", p.ConfigSource)
		fmt.Printf("Config file:        %s\n", p.ConfigFile)
		fmt.Printf("OS type:        %s\n", p.OsType)
		fmt.Printf("SSH user:           %s\n", p.SSHUser)
		fmt.Printf("SSH password:       %s\n", pass)
		fmt.Printf("SSH key file:       %s\n", p.SSHKeyFile)
		fmt.Printf("SSH port:           %d\n", p.SSHPort)
		fmt.Printf("Discrepancy action: %s\n", p.DiscrepancyAction)
		fmt.Printf("Merge configs:      %t\n", p.MergeConfigs)
		fmt.Printf("Config timeout (s): %d\n", p.ConfigTimeout)
		fmt.Printf("VLAN accuracy:      %d\n", p.VLANAccuracy)
	},
}

var profileAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Create a scan profile",
	Long: `Create a scan profile storing the connection/scan parameters for a host.

A "device" profile (the default) is bound to a host (--host) and auto-applied
when that host is scanned. A "generic" profile (--generic) is not bound to a
host: it carries only reusable SSH credentials (--ssh-user plus a key/password)
and is used as an explicit SSH fallback for hosts without their own profile
(e.g. 'scan connections --subnet ... --generic-profile <name>').

If --ssh-password is given, it is encrypted by the credential vault (you are
asked to set or unlock its master passphrase). The vault is unlocked again
whenever an SSH scan needs the stored secret.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}

		kind := "device"
		if profGeneric {
			kind = "generic"
		}
		p := e.ScanProfile{
			Name:              args[0],
			Kind:              kind,
			Host:              profHost,
			SNMPCommunity:     profCommunity,
			SNMPVersion:       profSNMPVersion,
			SNMPPort:          profSNMPPort,
			TimeoutSec:        profTimeout,
			ScanSource:        profScanSource,
			ConfigSource:      profConfigSource,
			ConfigFile:        profConfigFile,
			OsType:            profOsType,
			SSHUser:           profSSHUser,
			SSHKeyFile:        profSSHKeyFile,
			SSHPort:           profSSHPort,
			DiscrepancyAction: profDiscrepancy,
			MergeConfigs:      profMergeConfigs,
			ConfigTimeout:     profConfigTimeout,
			VLANAccuracy:      profVLANAccuracy,
		}

		if profSSHPassword != "" {
			if err := ensureVaultUnlocked(service.Vault()); err != nil {
				fmt.Printf("Error unlocking credential vault: %v\n", err)
				os.Exit(1)
			}
			blob, err := service.Vault().Encrypt(profSSHPassword)
			if err != nil {
				fmt.Printf("Error encrypting SSH password: %v\n", err)
				os.Exit(1)
			}
			p.SSHPassword = blob
		}

		if err := service.AddScanProfile(p); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if kind == "generic" {
			fmt.Printf("Saved generic scan profile %q (SSH user %s).\n", p.Name, p.SSHUser)
		} else {
			fmt.Printf("Saved scan profile %q (host %s).\n", p.Name, p.Host)
		}
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a scan profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		service, err := util.GetServiceConnection(cmd_pkg.Srcdbpath)
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		if err := service.DeleteScanProfile(args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted scan profile %q.\n", args[0])
	},
}

// readSecret reads a line from the terminal without echo.
func readSecret(prompt string) (string, error) {
	fmt.Print(prompt)
	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(raw), err
}

// readPassphraseConfirmed prompts for a passphrase twice and checks they match.
func readPassphraseConfirmed() (string, error) {
	p1, err := readSecret("Master passphrase for the credential vault: ")
	if err != nil {
		return "", err
	}
	if p1 == "" {
		return "", fmt.Errorf("passphrase must not be empty")
	}
	p2, err := readSecret("Confirm passphrase: ")
	if err != nil {
		return "", err
	}
	if p1 != p2 {
		return "", fmt.Errorf("passphrases do not match")
	}
	return p1, nil
}

// ensureVaultUnlocked makes the credential vault usable for this CLI run. The
// vault's data key lives only in memory, so each invocation unlocks it from the
// persisted meta with the master passphrase; a fresh vault is set up on first
// use (passphrase entered twice).
func ensureVaultUnlocked(v *secret.Vault) error {
	if v.Unlocked() {
		return nil
	}
	if !v.Initialized() {
		fmt.Println("No credential vault yet — set a master passphrase to protect stored SSH secrets.")
		pass, err := readPassphraseConfirmed()
		if err != nil {
			return err
		}
		return v.Init(pass)
	}
	pass, err := readSecret("Master passphrase to unlock the credential vault: ")
	if err != nil {
		return err
	}
	return v.Unlock(pass)
}

func init() {
	cmd_root.ScanCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileShowCmd)
	profileCmd.AddCommand(profileAddCmd)
	profileCmd.AddCommand(profileDeleteCmd)

	f := profileAddCmd.Flags()
	f.BoolVar(&profGeneric, "generic", false, "Create a generic profile: reusable SSH credentials not bound to a host (requires --ssh-user; --host and SNMP fields are ignored)")
	f.StringVar(&profHost, "host", "", "Host IP / subnet / SSH alias the profile applies to (device profiles only)")
	f.StringVar(&profCommunity, "snmp-community", "public", "SNMP community string")
	f.StringVar(&profSNMPVersion, "snmp-version", "v2c", "SNMP version (v1, v2c)")
	f.IntVar(&profSNMPPort, "snmp-port", 161, "SNMP UDP port")
	f.IntVar(&profTimeout, "timeout", 0, "Scan timeout in seconds (0 = command default)")
	f.StringVar(&profScanSource, "scan-source", "snmp", "Primary scan source (snmp, ssh)")
	f.StringVar(&profConfigSource, "config-source", "none", "Config source (none, ssh, file, manual)")
	f.StringVar(&profConfigFile, "config-file", "", "Path to device configuration file")
	f.StringVar(&profOsType, "os-type", "", "OS type (opnsense, openwrt, fortinet)")
	f.StringVar(&profSSHUser, "ssh-user", "", "SSH username")
	f.StringVar(&profSSHPassword, "ssh-password", "", "SSH password (encrypted at rest with a passphrase)")
	f.StringVar(&profSSHKeyFile, "ssh-key", "", "Path to SSH private key file")
	f.IntVar(&profSSHPort, "ssh-port", 22, "SSH port")
	f.StringVar(&profDiscrepancy, "discrepancy-action", "prefer-snmp", "Conflict resolution (fail, prefer-snmp, prefer-config)")
	f.BoolVar(&profMergeConfigs, "merge-configs", false, "Merge SNMP data with configuration data")
	f.IntVar(&profConfigTimeout, "config-timeout", 60, "Configuration parsing timeout in seconds")
	f.IntVar(&profVLANAccuracy, "vlan-accuracy", 1, "VLAN detection accuracy level (1, 2)")
}
