// SPDX-License-Identifier: AGPL-3.0-or-later
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
// package application set ups an interface for access to database and transforms outputs to json
package application

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"

	configparser "nsl-graph/internal/configparser"
	"nsl-graph/internal/datastore"
	fmtd2 "nsl-graph/internal/format"
	"nsl-graph/internal/observ"
	d "nsl-graph/internal/repository/domain"
	e "nsl-graph/internal/repository/entities"
	p "nsl-graph/internal/push"
	s "nsl-graph/internal/scanner"
	"nsl-graph/internal/secret"
	"nsl-graph/internal/topology"
	"nsl-graph/internal/yang/canon"
)

// vaultIdleTimeout auto-locks the credential vault after this much inactivity.
// vaultIdleTimeout auto-locks the credential vault after this much inactivity.
const vaultIdleTimeout = 15 * time.Minute

type NetService struct {
	netRepo d.NetRepository
	vault   *secret.Vault
}

type NetServiceInt interface {
	// Vault returns the server-side credential vault.
	Vault() *secret.Vault
	// Brand operations
	AddBrand(brandName string) error
	GetBrands() ([]e.Brand, error)
	UpdateBrand(brandId string, newBrandName string) error
	DeleteBrand(brandName string) error

	// ModelType operations
	AddModelType(modelTypeName string) error
	GetModelTypes() ([]e.ModelType, error)
	UpdateModelType(modelTypeId string, newModelTypeName string) error
	DeleteModelType(modelTypeName string) error

	// OsType operations
	AddOsType(osTypeName string) error
	GetOsTypes() ([]e.OsType, error)
	UpdateOsType(osTypeId string, newOsTypeName string) error
	DeleteOsType(osTypeName string) error
	EnsureOsType(osTypeName string) error

	// ZoneType operations
	AddZoneType(zoneTypeName string) error
	GetZonetypes() ([]e.ZoneType, error)
	UpdateZoneType(zoneTypeId string, newZoneTypeName string) error
	DeleteZoneType(zoneTypeName string) error

	// Owner operations
	AddOwner(ownerName string) error
	GetOwners() ([]e.Owner, error)
	UpdateOwner(ownerId string, newOwnerName string) error
	DeleteOwner(ownerName string) error

	// Zone operations
	AddZone(
		zoneName string,
		fatherZoneName string,
		fatherZoneId string,
		ownerName string,
		zoneTypeName string,
	) error
	GetZones() ([]e.Zone, error)
	UpdateZone(
		zoneId string,
		newZoneName string,
		newFatherZoneId string,
		newZoneTypeId string,
		newOwnerId string,
	) error
	DeleteZone(zoneId string) error

	// Model operations
	AddModel(
		modelName string,
		brandName string,
		modelTypeName string,
		osTypeName string,
	) error
	GetModels() ([]e.ModelDevice, error)
	UpdateModel(
		modelId string,
		newModelName string,
		newBrandId string,
		newModelTypeId string,
		newOsTypeId string,
	) error
	DeleteModel(modelId string) error

	// Device operations
	AddDevice(
		deviceLabel string,
		modelName string,
		zoneId string,
		zoneName string,
		ownerName string,
		isUnmanaged bool,
		isInvisible bool,
	) error
	GetDevices() ([]e.Device, error)
	UpdateDevice(
		deviceId string,
		newDeviceLabel string,
		newModelId string,
		newZoneId string,
		newOwnerId string,
		isUnmanaged *bool,
	) error
	UpdateDeviceIPs(deviceId string, ips []string) error
	UpdateDeviceProfile(deviceId string, profile string) error
	MigrateDeviceModel(deviceId string, newModelId string, portMap map[string]string) error
	DeleteDevice(deviceId string) error

	// ModelPort operations
	AddModelPort(
		portName string,
		positionX string,
		positionY string,
		modelName string,
		allowMultipleConnections bool,
		portType string,
		band string,
	) error
	GetModelPorts() ([]e.ModelPort, error)
	UpdateModelPort(
		modelPortId string,
		newPortName string,
		newPositionX string,
		newPositionY string,
		newModelId string,
		newAllowMultipleConnections bool,
	) error
	DeleteModelPort(modelPortId string) error

	// DevicePort operations
	AddDevicePort(deviceId string, modelPortId string, macAddress string, vlanConfigs []e.PortVlanConfig) (string, error)
	GetDevicePorts() ([]e.DevicePort, error)
	GetDevicePortByIDs(deviceId string, modelPortId string) (*e.DevicePort, error)
	UpdateDevicePort(deviceId string, modelPortId string, macAddress string, vlanConfigs []e.PortVlanConfig) error
	DeleteDevicePort(deviceId string, modelPortId string) error

	// ConnectionType operations
	AddConnectionType(connectionTypeName string) error
	GetConnectionTypes() ([]e.ConnectionType, error)
	UpdateConnectionType(connectionTypeId string, newConnectionTypeName string) error
	DeleteConnectionType(connectionTypeName string) error

	// Connection operations
	AddConnection(
		fromDeviceportID string,
		toDeviceportID string,
		connectionType string,
		discoveredVia ...string,
	) error
	GetConnections() ([]e.Connection, error)

	// DiscoverConnections collects multi-source L2/L1 evidence from the given
	// targets and correlates it into connection edges; ImportConnectionEdgesChecked
	// persists the ones the commit rules allow, with their evidence.
	DiscoverConnections(targets []topology.Target, only string) (*topology.ConnectionScanResult, error)
	// ImportConnectionEdgesChecked stages edges, validates them against the commit
	// rules, and commits only those that pass — returning every violation that stopped
	// the rest. This is where "a weak link may not be committed unreviewed" is
	// enforced, for every caller rather than only the CLI.
	ImportConnectionEdgesChecked(c datastore.Candidate) (int, []datastore.Violation, error)
	// CreatePlaceholderForIntermediaries materializes one shared placeholder
	// unmanaged device (in a placeholder zone) for unknown device(s) detected
	// between known hosts, wiring each observing endpoint to it.
	CreatePlaceholderForIntermediaries(intermediaries []topology.Intermediary) (PlaceholderResult, error)
	// DiscoverConnectionsByMode builds targets (from-db/profiles/subnet) and runs
	// discovery without interactive prompts — used by the HTTP API.
	DiscoverConnectionsByMode(opts ConnectionScanOptions, em observ.Emitter) (*topology.ConnectionScanResult, error)
	UpdateConnection(
		connectionId string,
		newFromDeviceportID string,
		newToDeviceportID string,
		connectionType string,
	) error
	DeleteConnection(connectionId string) error

	// VLAN operations
	AddVlan(vlanID string, vlanName string, ipSegment string) error
	GetVlans() ([]e.Vlan, error)
	UpdateVlan(vlanId string, newVlanID string, newVlanName string) error
	UpdateVlanIPSegment(vlanId string, ipSegment string) error
	DeleteVlan(vlanId string) error

	// Cascade deletion operations
	DeleteBrandCascade(brandName string) error
	DeleteModelTypeCascade(modelTypeName string) error
	DeleteZoneTypeCascade(zoneTypeName string) error
	DeleteOwnerCascade(ownerName string) error
	DeleteZoneCascade(zoneId string) error
	DeleteModelCascade(modelId string) error
	DeleteDeviceCascade(deviceId string) error
	DeleteModelPortCascade(modelPortId string) error
	DeleteDevicePortCascade(deviceId string, modelPortId string) error
	DeleteConnectionCascade(connectionId string) error
	DeleteVlanCascade(vlanId string) error

	// DeviceInterface operations
	AddDeviceInterface(deviceID, name, description, parent string, vlanConfigs []e.PortVlanConfig, ips []string, wifiSSID, wifiSecurity string) error
	GetDeviceInterfaces(deviceID string) ([]e.DeviceInterface, error)
	GetAllDeviceInterfaces() ([]e.DeviceInterface, error)
	UpdateDeviceInterface(id string, vlanConfigs []e.PortVlanConfig) error
	DeleteDeviceInterface(id string) error

	// InterfacePort operations
	AddInterfacePort(interfaceID, deviceID, modelPortID string) error
	GetInterfacesForPort(deviceID, modelPortID string) ([]e.DeviceInterface, error)
	GetPortsForInterface(interfaceID string) ([]e.DevicePort, error)
	GetAllInterfacePorts() ([]e.InterfacePort, error)
	DeleteInterfacePort(interfaceID, deviceID, modelPortID string) error

	// Special operations
	GetAllPortsDevice(deviceId string) ([]e.DevicePort, error)
	GetAllPortsAll() ([]e.DevicePort, error)
	ExportAllStructs() []byte

	// ExportYANG renders the specification as RFC 7951 JSON, valid against the
	// standard YANG models (RFC 8345/8944, IEEE 802.1Q) plus the nsl-* augments.
	// The warnings list values the domain model holds but the schema rejects; they
	// are dropped from the output rather than failing the export.
	ExportYANG() ([]byte, []string, error)

	// IntendedTree is the specification — what the network SHOULD be.
	IntendedTree() (*canon.Root, []string, error)
	// ObservedTree is a scan — what the network IS — resolved against the
	// specification so the same device carries the same identity in both.
	ObservedTree(devices []s.SNMPDevice) (*canon.Root, []string, error)
	// DiffAgainstScan reports how the network differs from its specification.
	DiffAgainstScan(devices []s.SNMPDevice) ([]datastore.Change, []string, error)

	// Network scanning operations
	RunScan(opts RunScanOptions, em observ.Emitter) ([]s.DiscoveredDevice, error)
	ScanNetwork(subnet string, options s.ScanOptions) (*s.ScanResult, error)
	ScanDevice(ip string, options s.ScanOptions) (*s.SNMPDevice, error)
	ScanDeviceViaSSH(ip, osType string, creds configparser.SSHCredentials) (*s.SNMPDevice, error)
	DiscoverDevices(scanResult *s.ScanResult) ([]s.DiscoveredDevice, error)
	ImportScanResults(devices []s.DiscoveredDevice, options s.ImportOptions) error
	ImportDiscoveredDevices(devices []s.DiscoveredDevice, options s.ImportOptions) error

	// New interactive VLAN mapping methods
	AnalyzeDeviceForImport(discovered s.DiscoveredDevice) (s.DeviceImportPlan, error)
	RegenerateVLANPlans(plan s.DeviceImportPlan) s.DeviceImportPlan
	ExecuteApprovedImportPlan(plan s.DeviceImportPlan, options s.ImportOptions) error

	// Model management
	EnsureModelExists(modelName, brandName, defaultBrand, osTypeName string) error

	// Scan profiles. Encryption/decryption of the SSH password goes through the
	// vault (unlocked once); the service stores the blob as-is and sanitizes it
	// out of listings.
	AddScanProfile(p e.ScanProfile) error
	UpdateScanProfile(p e.ScanProfile) error
	GetScanProfiles() ([]e.ScanProfile, error)                     // SSHPassword blanked, HasSSHPassword set
	GetScanProfileByName(name string) (*e.ScanProfile, error)      // raw (with blob) — in-process use
	ResolveScanProfile(target, name string) (*e.ScanProfile, bool) // by name, else auto-match by host
	DeleteScanProfile(name string) error

	// ProfileDevice rows bind a profile to N hosts.
	AddProfileDevice(d e.ProfileDevice) error
	GetProfileDevices(profileName string) ([]e.ProfileDevice, error)
	DeleteProfileDevice(profileName, host string) error
	DeleteAllProfileDevices(profileName string) error
	AppendPushRun(run p.PushRun) (string, error)
	AllPushRuns() ([]p.PushRun, error)
	// PushRun audit trail (push-config-tab plan phase 3).
}

func NewNetService(netRepository d.NetRepository) NetServiceInt {
	v := secret.NewVault(netRepository.GetVaultMeta, netRepository.SetVaultMeta, vaultIdleTimeout)
	ns := &NetService{netRepo: netRepository, vault: v}
	// Seed the OS-type catalogue from the available config parsers so the
	// dropdowns are populated out of the box (idempotent).
	for _, name := range configparser.DefaultRegistry.ListParsers() {
		_ = ns.EnsureOsType(name)
	}
	return ns
}

// Vault returns the server-side credential vault (unlock/lock + encrypt/decrypt).
func (ns *NetService) Vault() *secret.Vault {
	return ns.vault
}

// --- Scan profiles ----------------------------------------------------------

// normalizeProfileKind defaults Kind to "device" and enforces the invariants of
// each kind: a generic profile is not bound to a host (it only carries SSH
// credentials, used as a fallback) and must name an SSH user.
func normalizeProfileKind(p *e.ScanProfile) error {
	if p.Kind == "" {
		p.Kind = "device"
	}
	switch p.Kind {
	case "device":
		return nil
	case "generic":
		p.Host = "" // generic profiles are not IP-bound
		if p.SSHUser == "" {
			return fmt.Errorf("a generic profile requires an SSH user")
		}
		return nil
	default:
		return fmt.Errorf("invalid profile kind %q (want \"device\" or \"generic\")", p.Kind)
	}
}

func (ns *NetService) AddScanProfile(p e.ScanProfile) error {
	if err := normalizeProfileKind(&p); err != nil {
		return err
	}
	return ns.netRepo.AddScanProfile(p)
}

func (ns *NetService) UpdateScanProfile(p e.ScanProfile) error {
	if err := normalizeProfileKind(&p); err != nil {
		return err
	}
	return ns.netRepo.UpdateScanProfile(p)
}

// GetScanProfiles returns all profiles with the encrypted SSH password removed
// and HasSSHPassword set, suitable for API/UI listing.
func (ns *NetService) GetScanProfiles() ([]e.ScanProfile, error) {
	profiles, err := ns.netRepo.GetScanProfiles()
	if err != nil {
		return nil, err
	}
	for i := range profiles {
		profiles[i].HasSSHPassword = profiles[i].SSHPassword != ""
		profiles[i].HasSSHKey = profiles[i].SSHKey != ""
		profiles[i].SSHPassword = ""
		profiles[i].SSHKey = ""
	}
	return profiles, nil
}

// GetScanProfileByName returns the raw stored profile (including the encrypted
// SSH password blob) for in-process use by the CLI.
func (ns *NetService) GetScanProfileByName(name string) (*e.ScanProfile, error) {
	return ns.netRepo.GetScanProfileByName(name)
}

func (ns *NetService) DeleteScanProfile(name string) error {
	return ns.netRepo.DeleteScanProfile(name)
}

// --- ProfileDevice rows -----------------------------------------------------

// AddProfileDevice attaches a host to a profile. Validates that the profile
// exists (returning an error if not) and that the host string is non-empty;
// uniqueness on (profile_name, host) is enforced by the storage layer.
func (ns *NetService) AddProfileDevice(d e.ProfileDevice) error {
	if d.ProfileName == "" {
		return fmt.Errorf("profile_name is required")
	}
	if d.Host == "" {
		return fmt.Errorf("host is required")
	}
	if _, err := ns.GetScanProfileByName(d.ProfileName); err != nil {
		return err
	}
	if d.SSHProfileName != "" {
		if _, err := ns.GetScanProfileByName(d.SSHProfileName); err != nil {
			return fmt.Errorf("ssh_profile_name: %w", err)
		}
	}
	// Inline custom: when SSHUser is set without an override profile, the row
	// must carry either a password or a private-key blob — never just the user.
	if d.SSHProfileName == "" && d.SSHUser != "" && d.SSHPassword == "" && d.SSHKey == "" {
		return fmt.Errorf("inline SSH credentials on row %q require a password or private key", d.Host)
	}
	return ns.netRepo.AddProfileDevice(d)
}

func (ns *NetService) GetProfileDevices(profileName string) ([]e.ProfileDevice, error) {
	return ns.netRepo.GetProfileDevices(profileName)
}

func (ns *NetService) DeleteProfileDevice(profileName, host string) error {
	return ns.netRepo.DeleteProfileDevice(profileName, host)
}

func (ns *NetService) DeleteAllProfileDevices(profileName string) error {
	return ns.netRepo.DeleteAllProfileDevices(profileName)
}

// PushRun audit-trail methods. Phase 3 of push-config-tab plan: every push
// invocation goes through the engine, which records one row per run via
// AppendPushRun. The HTTP handler at GET /api/v1/push/history reads back
// via AllPushRuns to render the audit log on the Push config tab.
func (ns *NetService) AppendPushRun(run p.PushRun) (string, error) {
	if run.DeviceID == "" {
		return "", fmt.Errorf("device_id is required")
	}
	if run.StartedAt.IsZero() {
		return "", fmt.Errorf("started_at is required")
	}
	if run.ID == "" {
		run.ID = uuid.NewString()
	}
	return run.ID, ns.netRepo.AppendPushRun(run)
}

func (ns *NetService) AllPushRuns() ([]p.PushRun, error) {
	return ns.netRepo.AllPushRuns()
}

// ResolveScanProfile loads a profile by explicit name, or — when name is empty —
// auto-matches one whose Host equals target. Returns the raw profile (the caller
// decrypts the SSH password only if/when an SSH scan needs it).
func (ns *NetService) ResolveScanProfile(target, name string) (*e.ScanProfile, bool) {
	if name != "" {
		p, err := ns.netRepo.GetScanProfileByName(name)
		if err != nil || p == nil {
			return nil, false
		}
		return p, true
	}
	if target == "" {
		return nil, false
	}
	p, err := ns.netRepo.GetScanProfileByHost(target)
	if err != nil || p == nil {
		return nil, false
	}
	if p.Kind == "generic" {
		return nil, false // generic profiles are never auto-matched by host
	}
	return p, true
}

func (ns *NetService) AddBrand(brandName string) error {
	return ns.netRepo.AddBrand(brandName)
}

func (ns *NetService) GetBrands() ([]e.Brand, error) {
	return ns.netRepo.GetBrands()
}

func (ns *NetService) UpdateBrand(brandId string, newBrandName string) error {
	return ns.netRepo.UpdateBrand(brandId, newBrandName)
}

func (ns *NetService) AddModelType(modelTypeName string) error {
	return ns.netRepo.AddModelType(modelTypeName)
}

func (ns *NetService) GetModelTypes() ([]e.ModelType, error) {
	return ns.netRepo.GetModelTypes()
}

func (ns *NetService) AddOsType(osTypeName string) error {
	return ns.netRepo.AddOsType(osTypeName)
}

func (ns *NetService) GetOsTypes() ([]e.OsType, error) {
	return ns.netRepo.GetOsTypes()
}

func (ns *NetService) UpdateOsType(osTypeId string, newOsTypeName string) error {
	return ns.netRepo.UpdateOsType(osTypeId, newOsTypeName)
}

func (ns *NetService) DeleteOsType(osTypeName string) error {
	return ns.netRepo.DeleteOsType(osTypeName)
}

// EnsureOsType creates the OS type if it doesn't exist yet (idempotent).
func (ns *NetService) EnsureOsType(osTypeName string) error {
	if osTypeName == "" {
		return nil
	}
	for _, o := range func() []e.OsType { os, _ := ns.netRepo.GetOsTypes(); return os }() {
		if o.Name == osTypeName {
			return nil
		}
	}
	return ns.netRepo.AddOsType(osTypeName)
}

func (ns *NetService) AddZoneType(zoneTypeName string) error {
	return ns.netRepo.AddZoneType(zoneTypeName)
}

func (ns *NetService) GetZonetypes() ([]e.ZoneType, error) {
	return ns.netRepo.GetZonetypes()
}

func (ns *NetService) AddOwner(owner string) error {
	return ns.netRepo.AddOwner(owner)
}

func (ns *NetService) GetOwners() ([]e.Owner, error) {
	return ns.netRepo.GetOwners()
}

func (ns *NetService) AddZone(
	name string,
	fatherid string,
	father string,
	owner string,
	zonename string,
) error {
	return ns.netRepo.AddZone(name, fatherid, father, owner, zonename)
}

func (ns *NetService) GetZones() ([]e.Zone, error) {
	return ns.netRepo.GetZones()
}

func (ns *NetService) GetModels() ([]e.ModelDevice, error) {
	return ns.netRepo.GetModels()
}

func (ns *NetService) AddModelPort(
	name string,
	posx string,
	posy string,
	modelName string,
	allowMultipleConnections bool,
	portType string,
	band string,
) error {
	return ns.netRepo.AddModelPort(name, posx, posy, modelName, allowMultipleConnections, portType, band)
}

func (ns *NetService) GetModelPorts() ([]e.ModelPort, error) {
	return ns.netRepo.GetModelPorts()
}

func (ns *NetService) AddModel(
	modelName string,
	brandName string,
	modelTypeName string,
	osTypeName string,
) error {
	return ns.netRepo.AddModel(modelName, brandName, modelTypeName, osTypeName)
}

func (ns *NetService) GetDevices() ([]e.Device, error) {
	return ns.netRepo.GetDevices()
}

func (ns *NetService) AddDevice(
	label string,
	model string,
	zoneId string,
	zoneName string,
	owner string,
	isUnmanaged bool,
	isInvisible bool,
) error {
	// Generate name for unmanaged devices if not provided
	if isUnmanaged && label == "" {
		generatedLabel, err := ns.generateUnmanagedName()
		if err != nil {
			return err
		}
		label = generatedLabel
	}
	return ns.netRepo.AddDevice(label, model, zoneId, zoneName, owner, isUnmanaged, isInvisible)
}

func (ns *NetService) generateUnmanagedName() (string, error) {
	devices, err := ns.netRepo.GetDevices()
	if err != nil {
		return "", fmt.Errorf("failed to get devices for name generation: %w", err)
	}

	// Collect existing "unmanagedXX" names
	existingNumbers := make(map[int]bool)
	for _, d := range devices {
		if len(d.Label) > 9 && d.Label[:9] == "unmanaged" {
			var num int
			_, err := fmt.Sscanf(d.Label[9:], "%d", &num)
			if err == nil {
				existingNumbers[num] = true
			}
		}
	}

	// Find the first available number starting from 1
	for i := 1; ; i++ {
		if !existingNumbers[i] {
			return fmt.Sprintf("unmanaged%02d", i), nil
		}
	}
}

func (ns *NetService) UpdateDeviceProfile(deviceId string, profile string) error {
	return ns.netRepo.UpdateDeviceProfile(deviceId, profile)
}

func (ns *NetService) UpdateDeviceIPs(deviceId string, ips []string) error {
	return ns.netRepo.UpdateDeviceIPs(deviceId, ips)
}

func (ns *NetService) MigrateDeviceModel(deviceId string, newModelId string, portMap map[string]string) error {
	return ns.netRepo.MigrateDeviceModel(deviceId, newModelId, portMap)
}

func (ns *NetService) GetDevicePorts() ([]e.DevicePort, error) {
	zones, err := ns.netRepo.GetDevicePorts()
	return zones, err
}

func (ns *NetService) GetDevicePortByIDs(deviceid string, modelportid string) (*e.DevicePort, error) {
	return ns.netRepo.GetDevicePortByIDs(deviceid, modelportid)
}

func (ns *NetService) AddDevicePort(
	deviceid string,
	modelportid string,
	macAddress string,
	vlanConfigs []e.PortVlanConfig,
) (string, error) {
	if err := validatePortVlans(deviceid, modelportid, vlanConfigs); err != nil {
		return "", err
	}
	return ns.netRepo.AddDevicePort(deviceid, modelportid, macAddress, vlanConfigs)
}

func (ns *NetService) UpdateDevicePort(deviceid string, modelportid string, macAddress string, vlanConfigs []e.PortVlanConfig) error {
	if err := validatePortVlans(deviceid, modelportid, vlanConfigs); err != nil {
		return err
	}
	return ns.netRepo.UpdateDevicePort(deviceid, modelportid, macAddress, vlanConfigs)
}

// validatePortVlans refuses a VLAN configuration the commit rules reject: an identifier
// outside the 802.1Q range, or the same VLAN configured both tagged and untagged.
//
// The second is why this exists. A port either tags a VLAN's frames on egress or it does
// not; recording both is a contradiction, not a duplicate, and it means one of the two is
// wrong. test-dbs/real.db contains nine of them -- almost certainly from merging the
// 802.1Q egress-port and untagged-port sets on import without reconciling them -- and
// nothing in the model could refuse them, so they were simply stored.
//
// The YANG schema cannot catch this: vlan-id is the KEY of vlan-membership, so the
// contradiction is unrepresentable there and would be silently collapsed on export. It
// can only be caught on the way in.
func validatePortVlans(deviceid, portid string, vlanConfigs []e.PortVlanConfig) error {
	if len(vlanConfigs) == 0 {
		return nil
	}

	var c datastore.Candidate
	vlans := make([]datastore.VLANMembership, 0, len(vlanConfigs))
	for _, vc := range vlanConfigs {
		vlans = append(vlans, datastore.VLANMembership{VLAN: vc.VlanNumber, Tagged: vc.Tagged})
	}
	c.StagePort(deviceid, portid, vlans)

	return datastore.ViolationsError(datastore.Validate(c))
}

func (ns *NetService) AddConnectionType(connectionTypeName string) error {
	return ns.netRepo.AddConnectionType(connectionTypeName)
}

func (ns *NetService) GetConnectionTypes() ([]e.ConnectionType, error) {
	return ns.netRepo.GetConnectionTypes()
}

func (ns *NetService) GetConnections() ([]e.Connection, error) {
	connections, err := ns.netRepo.GetConnections()
	if err != nil {
		return nil, err
	}

	devicePorts, _ := ns.GetDevicePorts()
	allInterfaces, _ := ns.GetAllDeviceInterfaces()
	ifacePorts, _ := ns.GetAllInterfacePorts()

	for i := range connections {
		vlans, missing := fmtd2.GetConnectionVlanInfo(connections[i], devicePorts, allInterfaces, ifacePorts)
		connections[i].Vlans = vlans
		connections[i].MissingVlans = missing
	}

	return connections, nil
}

func (ns *NetService) AddConnection(
	fromDeviceportID string,
	toDeviceportID string,
	connectionType string,
	discoveredVia ...string,
) error {
	return ns.netRepo.AddConnection(fromDeviceportID, toDeviceportID, connectionType, discoveredVia...)
}

// ensureConnectionType creates the named connection type if it doesn't already
// exist, so callers (e.g. scans) can guarantee the strict AddConnection
// dependency is satisfied. A no-op (nil) if it already exists.
func (ns *NetService) ensureConnectionType(name string) error {
	if name == "" {
		return fmt.Errorf("connection type name is required")
	}
	types, err := ns.GetConnectionTypes()
	if err != nil {
		return err
	}
	for _, t := range types {
		if t.Name == name {
			return nil
		}
	}
	return ns.AddConnectionType(name)
}

// ensureBrand creates the named brand if it doesn't already exist.
func (ns *NetService) ensureBrand(name string) error {
	if name == "" {
		return fmt.Errorf("brand name is required")
	}
	brands, err := ns.GetBrands()
	if err != nil {
		return err
	}
	for _, b := range brands {
		if b.Name == name {
			return nil
		}
	}
	return ns.AddBrand(name)
}

// ensureModelType creates the named device class if it doesn't already exist.
func (ns *NetService) ensureModelType(name string) error {
	if name == "" {
		return fmt.Errorf("device class name is required")
	}
	classes, err := ns.GetModelTypes()
	if err != nil {
		return err
	}
	for _, c := range classes {
		if c.Name == name {
			return nil
		}
	}
	return ns.AddModelType(name)
}

// ensureZoneType creates the named zone type if it doesn't already exist.
func (ns *NetService) ensureZoneType(name string) error {
	if name == "" {
		return fmt.Errorf("zone type name is required")
	}
	types, err := ns.GetZonetypes()
	if err != nil {
		return err
	}
	for _, t := range types {
		if t.Name == name {
			return nil
		}
	}
	return ns.AddZoneType(name)
}

// ensureOwner creates the named owner if it doesn't already exist.
func (ns *NetService) ensureOwner(name string) error {
	if name == "" {
		return fmt.Errorf("owner name is required")
	}
	props, err := ns.GetOwners()
	if err != nil {
		return err
	}
	for _, p := range props {
		if p.Name == name {
			return nil
		}
	}
	return ns.AddOwner(name)
}

func (ns *NetService) GetAllPortsAll() ([]e.DevicePort, error) {
	return ns.netRepo.GetAllPortsAll()
}

// DeviceInterface method implementations

func (ns *NetService) AddDeviceInterface(deviceID, name, description, parent string, vlanConfigs []e.PortVlanConfig, ips []string, wifiSSID, wifiSecurity string) error {
	return ns.netRepo.AddDeviceInterface(deviceID, name, description, parent, vlanConfigs, ips, wifiSSID, wifiSecurity)
}

func (ns *NetService) GetDeviceInterfaces(deviceID string) ([]e.DeviceInterface, error) {
	return ns.netRepo.GetDeviceInterfaces(deviceID)
}

func (ns *NetService) GetAllDeviceInterfaces() ([]e.DeviceInterface, error) {
	return ns.netRepo.GetAllDeviceInterfaces()
}

func (ns *NetService) UpdateDeviceInterface(id string, vlanConfigs []e.PortVlanConfig) error {
	return ns.netRepo.UpdateDeviceInterface(id, vlanConfigs)
}

func (ns *NetService) DeleteDeviceInterface(id string) error {
	return ns.netRepo.DeleteDeviceInterface(id)
}

func (ns *NetService) AddInterfacePort(interfaceID, deviceID, modelPortID string) error {
	return ns.netRepo.AddInterfacePort(interfaceID, deviceID, modelPortID)
}

func (ns *NetService) GetAllInterfacePorts() ([]e.InterfacePort, error) {
	return ns.netRepo.GetAllInterfacePorts()
}

func (ns *NetService) DeleteInterfacePort(interfaceID, deviceID, modelPortID string) error {
	return ns.netRepo.DeleteInterfacePort(interfaceID, deviceID, modelPortID)
}

// GetInterfacesForPort returns all DeviceInterfaces that are linked to the given physical port
func (ns *NetService) GetInterfacesForPort(deviceID, modelPortID string) ([]e.DeviceInterface, error) {
	ifacePorts, err := ns.netRepo.GetInterfacePortsByPort(deviceID, modelPortID)
	if err != nil {
		return nil, fmt.Errorf("GetInterfacesForPort: %w", err)
	}

	result := make([]e.DeviceInterface, 0, len(ifacePorts))
	for _, ip := range ifacePorts {
		ifaceList, err := ns.netRepo.GetDeviceInterfaces(ip.DeviceID)
		if err != nil {
			continue
		}
		for _, iface := range ifaceList {
			if iface.ID == ip.InterfaceID {
				result = append(result, iface)
				break
			}
		}
	}
	return result, nil
}

// GetPortsForInterface returns all DevicePorts that are linked to the given logical interface
func (ns *NetService) GetPortsForInterface(interfaceID string) ([]e.DevicePort, error) {
	ifacePorts, err := ns.netRepo.GetInterfacePortsByInterface(interfaceID)
	if err != nil {
		return nil, fmt.Errorf("GetPortsForInterface: %w", err)
	}

	result := make([]e.DevicePort, 0, len(ifacePorts))
	for _, ip := range ifacePorts {
		dp, err := ns.netRepo.GetDevicePortByIDs(ip.DeviceID, ip.ModelPortID)
		if err != nil {
			continue
		}
		result = append(result, *dp)
	}
	return result, nil
}

func (ns *NetService) GetAllPortsDevice(deviceid string) ([]e.DevicePort, error) {
	return ns.netRepo.GetAllPortsDevice(deviceid)
}

func (ns *NetService) ExportAllStructs() []byte {
	export, err := ns.netRepo.ExportAllStructs()
	if err != nil {
		log.Printf("Error exporting: %v", err)
		return []byte("{\"error\": \"Failed to export\"}")
	}
	jsonData, err := json.Marshal(export)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return []byte("[]") // Return empty JSON array on error
	}
	return jsonData
}

func (ns *NetService) DeleteBrand(brand string) error {
	return ns.netRepo.DeleteBrand(brand)
}

func (ns *NetService) DeleteModelType(name string) error {
	return ns.netRepo.DeleteModelType(name)
}

func (ns *NetService) DeleteZoneType(locationType string) error {
	return ns.netRepo.DeleteZoneType(locationType)
}

func (ns *NetService) DeleteOwner(owner string) error {
	return ns.netRepo.DeleteOwner(owner)
}

func (ns *NetService) DeleteZone(id string) error {
	return ns.netRepo.DeleteZone(id)
}

func (ns *NetService) DeleteModel(id string) error {
	return ns.netRepo.DeleteModel(id)
}

func (ns *NetService) DeleteDevice(id string) error {
	return ns.netRepo.DeleteDevice(id)
}

func (ns *NetService) DeleteModelPort(id string) error {
	return ns.netRepo.DeleteModelPort(id)
}

func (ns *NetService) DeleteDevicePort(deviceID, modelPortID string) error {
	return ns.netRepo.DeleteDevicePort(deviceID, modelPortID)
}

func (ns *NetService) DeleteConnection(id string) error {
	return ns.netRepo.DeleteConnection(id)
}

func (ns *NetService) UpdateModelType(modelTypeId string, newModelTypeName string) error {
	return ns.netRepo.UpdateModelType(modelTypeId, newModelTypeName)
}

func (ns *NetService) UpdateZoneType(zoneTypeId string, newZoneTypeName string) error {
	return ns.netRepo.UpdateZoneType(zoneTypeId, newZoneTypeName)
}

func (ns *NetService) UpdateOwner(ownerId string, newOwnerName string) error {
	return ns.netRepo.UpdateOwner(ownerId, newOwnerName)
}

func (ns *NetService) UpdateZone(
	zoneId string,
	newZoneName string,
	newFatherZoneId string,
	newZoneTypeId string,
	newOwnerId string,
) error {
	return ns.netRepo.UpdateZone(zoneId, newZoneName, newFatherZoneId, newZoneTypeId, newOwnerId)
}

func (ns *NetService) UpdateModel(
	modelId string,
	newModelName string,
	newBrandId string,
	newModelTypeId string,
	newOsTypeId string,
) error {
	return ns.netRepo.UpdateModel(modelId, newModelName, newBrandId, newModelTypeId, newOsTypeId)
}

func (ns *NetService) UpdateDevice(
	deviceId string,
	newDeviceLabel string,
	newModelId string,
	newZoneId string,
	newOwnerId string,
	isUnmanaged *bool,
) error {
	return ns.netRepo.UpdateDevice(deviceId, newDeviceLabel, newModelId, newZoneId, newOwnerId, isUnmanaged)
}

func (ns *NetService) UpdateModelPort(
	modelPortId string,
	newPortName string,
	newPositionX string,
	newPositionY string,
	newModelId string,
	newAllowMultipleConnections bool,
) error {
	return ns.netRepo.UpdateModelPort(modelPortId, newPortName, newPositionX, newPositionY, newModelId, newAllowMultipleConnections)
}

func (ns *NetService) UpdateConnectionType(connectionTypeId string, newConnectionTypeName string) error {
	return ns.netRepo.UpdateConnectionType(connectionTypeId, newConnectionTypeName)
}

func (ns *NetService) DeleteConnectionType(connectionTypeName string) error {
	return ns.netRepo.DeleteConnectionType(connectionTypeName)
}

func (ns *NetService) UpdateConnection(
	connectionId string,
	newFromDeviceportID string,
	newToDeviceportID string,
	connectionType string,
) error {
	return ns.netRepo.UpdateConnection(connectionId, newFromDeviceportID, newToDeviceportID, connectionType)
}

func (ns *NetService) AddVlan(vlanID string, vlanName string, ipSegment string) error {
	return ns.netRepo.AddVlan(vlanID, vlanName, ipSegment)
}

func (ns *NetService) UpdateVlanIPSegment(vlanId string, ipSegment string) error {
	return ns.netRepo.UpdateVlanIPSegment(vlanId, ipSegment)
}

func (ns *NetService) GetVlans() ([]e.Vlan, error) {
	return ns.netRepo.GetVlans()
}

func (ns *NetService) UpdateVlan(vlanId string, newVlanID string, newVlanName string) error {
	return ns.netRepo.UpdateVlan(vlanId, newVlanID, newVlanName)
}

func (ns *NetService) DeleteVlan(vlanId string) error {
	return ns.netRepo.DeleteVlan(vlanId)
}

// Cascade deletion method implementations
func (ns *NetService) DeleteBrandCascade(brandName string) error {
	return ns.netRepo.DeleteBrandCascade(brandName)
}

func (ns *NetService) DeleteModelTypeCascade(modelTypeName string) error {
	return ns.netRepo.DeleteModelTypeCascade(modelTypeName)
}

func (ns *NetService) DeleteZoneTypeCascade(zoneTypeName string) error {
	return ns.netRepo.DeleteZoneTypeCascade(zoneTypeName)
}

func (ns *NetService) DeleteOwnerCascade(ownerName string) error {
	return ns.netRepo.DeleteOwnerCascade(ownerName)
}

func (ns *NetService) DeleteZoneCascade(zoneId string) error {
	return ns.netRepo.DeleteZoneCascade(zoneId)
}

func (ns *NetService) DeleteModelCascade(modelId string) error {
	return ns.netRepo.DeleteModelCascade(modelId)
}

func (ns *NetService) DeleteDeviceCascade(deviceId string) error {
	return ns.netRepo.DeleteDeviceCascade(deviceId)
}

func (ns *NetService) DeleteModelPortCascade(modelPortId string) error {
	return ns.netRepo.DeleteModelPortCascade(modelPortId)
}

func (ns *NetService) DeleteDevicePortCascade(deviceId string, modelPortId string) error {
	return ns.netRepo.DeleteDevicePortCascade(deviceId, modelPortId)
}

func (ns *NetService) DeleteConnectionCascade(connectionId string) error {
	return ns.netRepo.DeleteConnectionCascade(connectionId)
}

func (ns *NetService) DeleteVlanCascade(vlanId string) error {
	return ns.netRepo.DeleteVlanCascade(vlanId)
}

// Network scanning method implementations

func (ns *NetService) ScanNetwork(subnet string, options s.ScanOptions) (*s.ScanResult, error) {
	scanner := s.NewSNMPScanner()

	options.Subnet = subnet
	if options.SNMP.Community == "" {
		options.SNMP.Community = "public"
	}
	if options.SNMP.Version == "" {
		options.SNMP.Version = "v2c"
	}

	result, err := scanner.Scan(options)
	if err != nil {
		return nil, fmt.Errorf("network scan failed: %w", err)
	}

	return result, nil
}

func (ns *NetService) ScanDevice(ip string, options s.ScanOptions) (*s.SNMPDevice, error) {
	scanner := s.NewSNMPScanner()

	if options.SNMP.Community == "" {
		options.SNMP.Community = "public"
	}
	if options.SNMP.Version == "" {
		options.SNMP.Version = "v2c"
	}

	device, err := scanner.ScanDevice(ip, options)
	if err != nil {
		return nil, fmt.Errorf("device scan failed: %w", err)
	}

	return device, nil
}

func (ns *NetService) DiscoverDevices(scanResult *s.ScanResult) ([]s.DiscoveredDevice, error) {
	discoverer := s.NewDeviceDiscoverer()

	var devices []s.DiscoveredDevice
	for _, dev := range scanResult.Devices {
		if !dev.Reachable {
			continue
		}
		brand, model, class := discoverer.ClassifyDevice(dev)
		devices = append(devices, s.DiscoveredDevice{
			Device:        dev,
			Brand:         brand,
			Model:         model,
			ModelType:     class,
			SuggestedName: discoverer.GenerateDeviceName(dev, class),
			SuggestedZone: discoverer.SuggestZone(dev),
		})
	}

	log.Printf("Discovered %d devices", len(devices))
	return devices, nil
}

func (ns *NetService) ImportScanResults(devices []s.DiscoveredDevice, options s.ImportOptions) error {
	return ns.ImportDiscoveredDevices(devices, options)
}

func (ns *NetService) ImportDiscoveredDevices(devices []s.DiscoveredDevice, options s.ImportOptions) error {
	if err := ns.ensureRequiredEntities(options); err != nil {
		return fmt.Errorf("failed to setup required entities: %w", err)
	}

	for _, device := range devices {
		if options.SkipExisting && ns.deviceExists(device.Device.IP) {
			continue
		}

		if err := ns.importSingleDevice(device, options); err != nil {
			log.Printf("Failed to import device %s: %v", device.SuggestedName, err)
			if !options.ReviewMode {
				return err
			}
		}
	}

	return nil
}

func (ns *NetService) ensureRequiredEntities(options s.ImportOptions) error {
	brands, _ := ns.GetBrands()
	brandExists := func(name string) bool {
		for _, b := range brands {
			if b.Name == name {
				return true
			}
		}
		return false
	}

	requiredBrands := []string{"Generic", "Discovered", "Linux", "Juniper", "Aruba", "Ubiquiti", "Fortinet", "Palo Alto", "MikroTik", "HP", "Microsoft", "BSD"}
	for _, brandName := range requiredBrands {
		if !brandExists(brandName) {
			ns.AddBrand(brandName)
		}
	}

	classes, _ := ns.GetModelTypes()
	classExists := func(name string) bool {
		for _, c := range classes {
			if c.Name == name {
				return true
			}
		}
		return false
	}

	requiredClasses := []string{"Switch", "Router", "Server", "Workstation", "Printer", "Generic", "Access Point", "Firewall"}
	for _, modelTypeName := range requiredClasses {
		if !classExists(modelTypeName) {
			ns.AddModelType(modelTypeName)
		}
	}

	// Create all possible models from device discovery
	models, _ := ns.GetModels()
	modelExists := func(name string) bool {
		for _, m := range models {
			if m.Model == name {
				return true
			}
		}
		return false
	}

	requiredModels := []string{
		"Juniper Device", "Aruba Device", "UniFi Device", "FortiGate", "PAN Device",
		"RouterOS Device", "ProCurve Switch", "Linux Server", "Windows Server",
		"BSD Server", "Network Printer", "Network Device",
	}
	for _, modelName := range requiredModels {
		if !modelExists(modelName) {
			// Determine brand and device class for each model
			var brandName, modelTypeName string
			switch modelName {
			case "Juniper Device":
				brandName, modelTypeName = "Juniper", "Router"
			case "Aruba Device":
				brandName, modelTypeName = "Aruba", "Access Point"
			case "UniFi Device":
				brandName, modelTypeName = "Ubiquiti", "Access Point"
			case "FortiGate":
				brandName, modelTypeName = "Fortinet", "Firewall"
			case "PAN Device":
				brandName, modelTypeName = "Palo Alto", "Firewall"
			case "RouterOS Device":
				brandName, modelTypeName = "MikroTik", "Router"
			case "ProCurve Switch":
				brandName, modelTypeName = "HP", "Switch"
			case "Linux Server":
				brandName, modelTypeName = "Linux", "Server"
			case "Windows Server":
				brandName, modelTypeName = "Microsoft", "Server"
			case "BSD Server":
				brandName, modelTypeName = "BSD", "Server"
			case "Network Printer":
				brandName, modelTypeName = "Generic", "Printer"
			case "Network Device":
				brandName, modelTypeName = "Generic", "Generic"
			}

			if err := ns.AddModel(modelName, brandName, modelTypeName, ""); err != nil {
				log.Printf("Warning: failed to create model %s: %v", modelName, err)
			}
		}
	}

	owners, _ := ns.GetOwners()
	ownerExists := func(name string) bool {
		for _, p := range owners {
			if p.Name == name {
				return true
			}
		}
		return false
	}

	if !ownerExists("Discovered") {
		ns.AddOwner("Discovered")
	}

	// Ensure required zone types exist
	zoneTypes, _ := ns.GetZonetypes()
	zoneTypeExists := func(name string) bool {
		for _, zt := range zoneTypes {
			if zt.Name == name {
				return true
			}
		}
		return false
	}

	requiredZoneTypes := []string{"Unknown", "Office", "Data Center", "Remote", "Cloud"}
	for _, zoneTypeName := range requiredZoneTypes {
		if !zoneTypeExists(zoneTypeName) {
			ns.AddZoneType(zoneTypeName)
		}
	}

	if options.CreateZones {
		zones, _ := ns.GetZones()
		zoneExists := func(name string) bool {
			for _, z := range zones {
				if z.Name == name {
					return true
				}
			}
			return false
		}

		// Ensure the owner and zone types referenced below exist first, so the
		// strict AddZone dependency checks pass (it never silently drops a ref).
		_ = ns.ensureOwner("Discovered")
		_ = ns.ensureZoneType("Office")
		_ = ns.ensureZoneType("Unknown")

		requiredZones := []string{"Generic", "Discovered", "LAN", "Internal", "External", "Private"}
		for _, zoneName := range requiredZones {
			if !zoneExists(zoneName) {
				// Use "Unknown" zone type for Generic zone, "Office" for others
				zoneType := "Office"
				if zoneName == "Generic" {
					zoneType = "Unknown"
				}
				if err := ns.AddZone(zoneName, "", "", "Discovered", zoneType); err != nil {
					log.Printf("Warning: failed to create zone %q: %v", zoneName, err)
				}
			}
		}
	}

	return nil
}

func (ns *NetService) deviceExists(ip string) bool {
	devices, err := ns.GetDevices()
	if err != nil {
		return false
	}

	for _, device := range devices {
		// Check if IP matches any device interface
		ifaces, err := ns.GetDeviceInterfaces(device.ID)
		if err != nil {
			continue
		}
		for _, iface := range ifaces {
			for _, ifaceIP := range iface.IPAddresses {
				if ifaceIP == ip {
					return true
				}
			}
		}
	}

	return false
}

func (ns *NetService) importSingleDevice(discovered s.DiscoveredDevice, options s.ImportOptions) error {
	zoneName := discovered.SuggestedZone
	if options.DefaultZone != "" {
		zoneName = options.DefaultZone
	}
	// Default to "Generic" zone if no zone is specified
	if zoneName == "" {
		zoneName = "Generic"
	}

	// Ensure model exists before adding device
	err := ns.EnsureModelExists(discovered.Model, discovered.Brand, options.DefaultBrand, discovered.OsType)
	if err != nil {
		return fmt.Errorf("failed to ensure model exists: %w", err)
	}

	err = ns.AddDevice(
		discovered.SuggestedName,
		discovered.Model,
		"",
		zoneName,
		"Discovered",
		false,
		false,
	)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("a device named %q is already in the database — it was probably imported before; "+
				"rename it in the scan results, or delete the existing device first (overwriting is not supported)", discovered.SuggestedName)
		}
		return fmt.Errorf("failed to add device: %w", err)
	}

	// Locate the freshly-created device (labels are unique) to attach the scan
	// profile it was discovered with, and its ports.
	devices, _ := ns.GetDevices()
	var deviceID string
	for _, device := range devices {
		if device.Label == discovered.SuggestedName {
			deviceID = device.ID
			break
		}
	}

	if deviceID != "" && discovered.Profile != "" {
		if err := ns.UpdateDeviceProfile(deviceID, discovered.Profile); err != nil {
			log.Printf("Failed to set scan profile for device %s: %v", discovered.SuggestedName, err)
		}
	}

	if deviceID != "" && (len(discovered.Device.Interfaces) > 0 || len(discovered.Device.SwitchPorts) > 0) {
		if err := ns.createDevicePortsForDevice(deviceID, discovered); err != nil {
			log.Printf("Failed to create ports for device %s: %v", discovered.SuggestedName, err)
		}
	}

	return nil
}

func (ns *NetService) createDevicePortsForDevice(deviceID string, discovered s.DiscoveredDevice) error {

	models, err := ns.GetModels()
	if err != nil {
		return err
	}

	var modelName string
	for _, model := range models {
		if model.Model == discovered.Model {
			modelName = model.Model
			break
		}
	}

	if modelName == "" {
		modelName = discovered.Model
		if modelName == "" {
			modelName = "Generic Model"
		}
		if err := ns.AddModel(modelName, discovered.Brand, discovered.ModelType, discovered.OsType); err != nil {
			return err
		}
	}

	portIdx := 0
	for _, iface := range discovered.Device.Interfaces {
		if !iface.IsPhysicalPort() {
			continue
		}

		portName := iface.Name
		if portName == "" {
			if iface.IsWifiRadio() {
				portName = fmt.Sprintf("radio%d", portIdx)
			} else {
				portName = fmt.Sprintf("eth%d", portIdx)
			}
		}

		portType := ""
		band := ""
		if iface.IsWifiRadio() {
			portType = "wifi"
			band = iface.WifiBand
		}

		if err := ns.AddModelPort(portName, fmt.Sprintf("%d", portIdx), "0", modelName, false, portType, band); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
		portIdx++

		modelPorts, _ := ns.GetModelPorts()
		var modelPortID string
		for _, mp := range modelPorts {
			if mp.Name == portName && mp.Model == modelName {
				modelPortID = mp.ID
				break
			}
		}

		if modelPortID == "" {
			continue
		}

		// Create the physical port (no VLAN configs on the port itself)
		if _, err := ns.AddDevicePort(deviceID, modelPortID, iface.MAC, nil); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}

		// Build VLAN configs from SNMP data
		vlanConfigs := buildVlanConfigs(iface.VLANs)

		// Create a DeviceInterface for this logical interface and link it to the physical port
		if len(vlanConfigs) > 0 {
			if err := ns.AddDeviceInterface(deviceID, iface.Name, "", iface.Parent, vlanConfigs, iface.IPAddresses, "", ""); err != nil {
				log.Printf("Warning: failed to create device interface %s for device %s: %v", iface.Name, deviceID, err)
			} else {
				// Look up the newly created interface by name so we can link it
				ifaceList, err := ns.GetDeviceInterfaces(deviceID)
				if err == nil {
					for _, di := range ifaceList {
						if di.Name == iface.Name && di.DeviceID == deviceID {
							if err := ns.AddInterfacePort(di.ID, deviceID, modelPortID); err != nil {
								log.Printf("Warning: failed to link interface %s to port %s: %v", di.ID, modelPortID, err)
							}
							break
						}
					}
				}
			}
		}
	}

	// Create DeviceInterface records for virtual interfaces (wifi-iface SSIDs and logical/VLAN UCI interfaces)
	for _, iface := range discovered.Device.Interfaces {
		if iface.IsPhysicalPort() {
			continue
		}
		hasData := iface.WifiSSID != "" || len(iface.IPAddresses) > 0 || len(iface.VLANs) > 0
		if !hasData {
			continue
		}
		vlanConfigs := buildVlanConfigs(iface.VLANs)
		if err := ns.AddDeviceInterface(deviceID, iface.Name, "", iface.Parent, vlanConfigs, iface.IPAddresses, iface.WifiSSID, iface.WifiSecurity); err != nil {
			log.Printf("Warning: failed to create interface %s for device %s: %v", iface.Name, deviceID, err)
		}
	}

	return nil
}

// buildVlanConfigs converts scanner VLAN memberships to entity VLAN configs.
func buildVlanConfigs(vlans []s.VLANMembership) []e.PortVlanConfig {
	if len(vlans) == 0 {
		return nil
	}
	out := make([]e.PortVlanConfig, 0, len(vlans))
	for _, v := range vlans {
		out = append(out, e.PortVlanConfig{VlanNumber: v.VLANNumber, Tagged: v.Tagged})
	}
	return out
}

// AnalyzeDeviceForImport creates a comprehensive import plan for a discovered device
func (ns *NetService) AnalyzeDeviceForImport(discovered s.DiscoveredDevice) (s.DeviceImportPlan, error) {
	plan := s.DeviceImportPlan{
		Device:         discovered,
		InterfacePlans: make([]s.InterfaceImportPlan, 0, len(discovered.Device.Interfaces)),
		HasConflicts:   false,
		RequiresInput:  false,
	}

	// Get existing VLANs for reference
	existingVLANs, err := ns.GetVlans()
	if err != nil {
		return plan, fmt.Errorf("failed to get existing VLANs: %w", err)
	}

	for _, iface := range discovered.Device.Interfaces {
		// Skip physical ports - they're handled separately in createDevicePortsWithPlan
		// Physical ports should not have InterfacePlans created for them
		if iface.IsPhysicalPort() {
			continue
		}

		interfacePlan, err := ns.analyzeInterfaceForImport(iface, existingVLANs)
		if err != nil {
			return plan, fmt.Errorf("failed to analyze interface %s: %w", iface.Name, err)
		}

		plan.InterfacePlans = append(plan.InterfacePlans, interfacePlan)
		if interfacePlan.RequiresUserInput() {
			plan.RequiresInput = true
		}
	}

	plan.Summary = ns.generateImportSummary(plan)
	return plan, nil
}

// ScanDeviceViaSSH retrieves and parses a device's configuration over SSH and
// returns it as an SNMPDevice (interfaces, VLANs, IPs). This mirrors the CLI
// `--scan-source ssh` path but is free of stdin/stdout, so it is usable from an
// API handler.
func (ns *NetService) ScanDeviceViaSSH(ip, osType string, creds configparser.SSHCredentials) (*s.SNMPDevice, error) {
	parser, found := configparser.DefaultRegistry.GetParser(osType)
	if !found {
		return nil, fmt.Errorf("unsupported OS type %q (supported: %s)",
			osType, strings.Join(configparser.DefaultRegistry.ListParsers(), ", "))
	}
	rawConfig, err := configparser.FetchConfig(configparser.DefaultTransport, parser, ip, creds)
	if err != nil {
		return nil, fmt.Errorf("SSH connection failed: %w", err)
	}
	stub := s.SNMPDevice{IP: ip, Reachable: true, SysName: ip}
	configData, err := parser.ParseConfig(rawConfig, stub)
	if err != nil {
		return nil, fmt.Errorf("config parsing failed: %w", err)
	}
	return configparser.ConfigDataToSNMPDevice(configData, ip), nil
}

// RegenerateVLANPlans recomputes each interface plan's VLANsToCreate/Update from
// its (possibly user-edited) IPMappings, so changes to vlan-id / ip-segment made
// by a client stay consistent with the VLANs actually created on execute.
func (ns *NetService) RegenerateVLANPlans(plan s.DeviceImportPlan) s.DeviceImportPlan {
	existingVLANs, err := ns.GetVlans()
	if err != nil {
		return plan
	}
	existingVLANMap := make(map[string]e.Vlan)
	for _, vlan := range existingVLANs {
		existingVLANMap[vlan.VlanID] = vlan
	}
	for i := range plan.InterfacePlans {
		plan.InterfacePlans[i].VLANsToCreate = make([]s.VLANPlan, 0)
		plan.InterfacePlans[i].VLANsToUpdate = make([]s.VLANPlan, 0)
		ns.generateVLANPlans(&plan.InterfacePlans[i], existingVLANMap)
	}
	plan.Summary = ns.generateImportSummary(plan)
	return plan
}

// analyzeInterfaceForImport analyzes a single interface and suggests IP-VLAN mappings
func (ns *NetService) analyzeInterfaceForImport(iface s.DeviceInterface, existingVLANs []e.Vlan) (s.InterfaceImportPlan, error) {
	plan := s.InterfaceImportPlan{
		Interface:     iface,
		IPMappings:    make([]s.IPVLANMapping, 0, len(iface.IPAddresses)),
		VLANsToCreate: make([]s.VLANPlan, 0),
		VLANsToUpdate: make([]s.VLANPlan, 0),
	}

	// Create a map of existing VLANs for quick lookup
	existingVLANMap := make(map[string]e.Vlan)
	for _, vlan := range existingVLANs {
		existingVLANMap[vlan.VlanID] = vlan
	}

	// PRIORITY 1: Use VLAN inference from interface names (accuracy 1)
	vlanInference := s.InferVLANFromInterface(iface.Name, iface.IPAddresses)

	// Analyze each IP address using VLAN inference first
	for _, ip := range iface.IPAddresses {
		mapping := ns.suggestIPVLANMappingWithInference(ip, iface.VLANs, existingVLANs, vlanInference)
		// Populate subnet CIDR from SNMP netmask if available
		if iface.IPNetmasks != nil {
			if mask, ok := iface.IPNetmasks[ip]; ok {
				mapping.Subnet = s.NetmaskToCIDR(ip, mask)
			}
		}
		plan.IPMappings = append(plan.IPMappings, mapping)
	}

	// Generate VLAN plans based on mappings
	ns.generateVLANPlans(&plan, existingVLANMap)

	return plan, nil
}

// suggestIPVLANMapping suggests a VLAN mapping for a given IP address
func (ns *NetService) suggestIPVLANMapping(ip string, vlanMemberships []s.VLANMembership, existingVLANs []e.Vlan) s.IPVLANMapping {
	// Method 1: Check if IP belongs to existing VLAN IP segments
	if exactMapping := ns.findExactVLANMatch(ip, existingVLANs); exactMapping.VLANNumber != "" {
		return exactMapping
	}

	// Method 2: Use interface VLAN memberships (prefer untagged)
	if interfaceMapping := ns.mapToInterfaceVLAN(ip, vlanMemberships); interfaceMapping.VLANNumber != "" {
		return interfaceMapping
	}

	// Method 3: Apply RFC1918 heuristics
	if heuristicMapping := ns.applyRFC1918Heuristic(ip, existingVLANs); heuristicMapping.VLANNumber != "" {
		return heuristicMapping
	}

	// Method 4: Default suggestion
	return ns.getDefaultMapping(ip, vlanMemberships)
}

// findExactVLANMatch checks if IP belongs to existing VLAN IP segments
func (ns *NetService) findExactVLANMatch(ip string, existingVLANs []e.Vlan) s.IPVLANMapping {
	for _, vlan := range existingVLANs {
		if vlan.IPSegment != "" && ns.ipBelongsToSegment(ip, vlan.IPSegment) {
			return s.IPVLANMapping{
				IP:         ip,
				VLANNumber: vlan.VlanID,
				Confidence: "exact",
				Reason:     fmt.Sprintf("IP belongs to existing VLAN %s segment %s", vlan.VlanID, vlan.IPSegment),
				IsNewVLAN:  false,
			}
		}
	}
	return s.IPVLANMapping{}
}

// mapToInterfaceVLAN maps IP to interface VLAN memberships
func (ns *NetService) mapToInterfaceVLAN(ip string, vlanMemberships []s.VLANMembership) s.IPVLANMapping {
	// Prefer untagged VLAN (native VLAN)
	for _, vlan := range vlanMemberships {
		if !vlan.Tagged {
			return s.IPVLANMapping{
				IP:           ip,
				VLANNumber:   vlan.VLANNumber,
				Confidence:   "suggested",
				Reason:       fmt.Sprintf("Interface has untagged VLAN %s", vlan.VLANNumber),
				IsNewVLAN:    false, // Will be checked later
				OriginalVLAN: vlan.VLANNumber,
			}
		}
	}

	// If no untagged, suggest first tagged VLAN
	if len(vlanMemberships) > 0 {
		vlan := vlanMemberships[0]
		return s.IPVLANMapping{
			IP:           ip,
			VLANNumber:   vlan.VLANNumber,
			Confidence:   "suggested",
			Reason:       fmt.Sprintf("Interface has tagged VLAN %s (no untagged found)", vlan.VLANNumber),
			IsNewVLAN:    false,
			OriginalVLAN: vlan.VLANNumber,
		}
	}

	return s.IPVLANMapping{}
}

// applyRFC1918Heuristic applies RFC1918 private network heuristics
func (ns *NetService) applyRFC1918Heuristic(ip string, existingVLANs []e.Vlan) s.IPVLANMapping {
	// Parse IP to determine private network class
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return s.IPVLANMapping{}
	}

	var suggestedVLAN string
	var reason string

	// Check RFC1918 ranges and suggest VLAN based on network
	if parsedIP.IsPrivate() {
		octets := strings.Split(ip, ".")
		if len(octets) >= 2 {
			switch octets[0] {
			case "10":
				suggestedVLAN = "10"
				reason = "RFC1918 10.0.0.0/8 network suggests VLAN 10"
			case "172":
				if secondOctet := octets[1]; secondOctet >= "16" && secondOctet <= "31" {
					suggestedVLAN = "172"
					reason = "RFC1918 172.16.0.0/12 network suggests VLAN 172"
				}
			case "192":
				if octets[1] == "168" {
					suggestedVLAN = "192"
					reason = "RFC1918 192.168.0.0/16 network suggests VLAN 192"
				}
			}
		}
	}

	if suggestedVLAN != "" {
		// Check if suggested VLAN exists
		vlanExists := false
		for _, vlan := range existingVLANs {
			if vlan.VlanID == suggestedVLAN {
				vlanExists = true
				break
			}
		}

		return s.IPVLANMapping{
			IP:         ip,
			VLANNumber: suggestedVLAN,
			Confidence: "heuristic",
			Reason:     reason,
			IsNewVLAN:  !vlanExists,
		}
	}

	return s.IPVLANMapping{}
}

// getDefaultMapping provides a default mapping when no other methods work
func (ns *NetService) getDefaultMapping(ip string, vlanMemberships []s.VLANMembership) s.IPVLANMapping {
	// Suggest generic VLAN (-1) for preservation
	return s.IPVLANMapping{
		IP:         ip,
		VLANNumber: "-1",
		Confidence: "suggested",
		Reason:     "No specific VLAN detected, suggest generic preservation",
		IsNewVLAN:  true,
	}
}

// suggestIPVLANMappingWithInference suggests a VLAN mapping using VLAN inference as the primary method
func (ns *NetService) suggestIPVLANMappingWithInference(ip string, vlanMemberships []s.VLANMembership, existingVLANs []e.Vlan, vlanInference s.VLANInference) s.IPVLANMapping {
	// METHOD 1: Use VLAN inference from interface name (highest accuracy)
	if len(vlanInference.VLANNumbers) > 0 && vlanInference.Accuracy == 1 {
		// Use the first VLAN number from inference (highest confidence)
		vlanNum := vlanInference.VLANNumbers[0]
		reason := fmt.Sprintf("Interface name %s suggests VLAN %s", vlanInference.Method, vlanNum)
		if vlanInference.Notes != "" {
			reason += " (" + vlanInference.Notes + ")"
		}

		return s.IPVLANMapping{
			IP:           ip,
			VLANNumber:   vlanNum,
			Confidence:   "exact",
			Reason:       reason,
			IsNewVLAN:    !ns.vlanExists(vlanNum, existingVLANs),
			OriginalVLAN: vlanNum,
		}
	}

	// METHOD 2: Check if IP belongs to existing VLAN IP segments
	if exactMapping := ns.findExactVLANMatch(ip, existingVLANs); exactMapping.VLANNumber != "" {
		return exactMapping
	}

	// METHOD 3: Use interface VLAN memberships (prefer untagged)
	if interfaceMapping := ns.mapToInterfaceVLAN(ip, vlanMemberships); interfaceMapping.VLANNumber != "" {
		return interfaceMapping
	}

	// METHOD 4: Use VLAN inference with lower accuracy (if accuracy level permits)
	if len(vlanInference.VLANNumbers) > 0 && vlanInference.Accuracy == 2 {
		vlanNum := vlanInference.VLANNumbers[0]
		reason := fmt.Sprintf("IP pattern suggests VLAN %s", vlanNum)
		if vlanInference.Notes != "" {
			reason += " (" + vlanInference.Notes + ")"
		}

		return s.IPVLANMapping{
			IP:           ip,
			VLANNumber:   vlanNum,
			Confidence:   "heuristic",
			Reason:       reason,
			IsNewVLAN:    !ns.vlanExists(vlanNum, existingVLANs),
			OriginalVLAN: vlanNum,
		}
	}

	// METHOD 5: Apply RFC1918 heuristics (if no interface inference available)
	if heuristicMapping := ns.applyRFC1918Heuristic(ip, existingVLANs); heuristicMapping.VLANNumber != "" {
		return heuristicMapping
	}

	// METHOD 6: Default suggestion
	return ns.getDefaultMapping(ip, vlanMemberships)
}

// ipBelongsToSegment checks if an IP belongs to a given segment (IP or CIDR)
func (ns *NetService) ipBelongsToSegment(ip, segment string) bool {
	// Handle exact IP match
	if ip == segment {
		return true
	}

	// Handle CIDR notation
	if strings.Contains(segment, "/") {
		_, network, err := net.ParseCIDR(segment)
		if err != nil {
			return false
		}
		parsedIP := net.ParseIP(ip)
		return parsedIP != nil && network.Contains(parsedIP)
	}

	return false
}

// EnsureModelExists checks if a model exists and creates it with fallback defaults if not
func (ns *NetService) EnsureModelExists(modelName, brandName, defaultBrand, osTypeName string) error {
	// Check if model already exists
	models, err := ns.GetModels()
	if err != nil {
		return fmt.Errorf("failed to get models: %w", err)
	}

	// Look for existing model
	for _, model := range models {
		if model.Model == modelName {
			return nil // Model already exists
		}
	}

	// Model doesn't exist, create it with fallback defaults

	// Determine brand to use - prioritize detected brand, then default, then "Generic"
	finalBrand := brandName
	if finalBrand == "" || finalBrand == "Unknown" {
		if defaultBrand != "" {
			finalBrand = defaultBrand
		} else {
			finalBrand = "Generic"
		}
	}

	// Ensure brand exists
	err = ns.AddBrand(finalBrand)
	if err != nil {
		// Brand might already exist, that's ok
		// Continue with device class creation
	}

	// Ensure "Router" device class exists (fallback device class)
	err = ns.AddModelType("Router")
	if err != nil {
		// Device class might already exist, that's ok
		// Continue with model creation
	}

	// Create the model with fallback defaults (os_type is auto-registered if new)
	err = ns.AddModel(modelName, finalBrand, "Router", osTypeName)
	if err != nil {
		return fmt.Errorf("failed to create model %s: %w", modelName, err)
	}

	return nil
}

// generateVLANPlans creates VLAN creation/update plans based on IP mappings
func (ns *NetService) generateVLANPlans(plan *s.InterfaceImportPlan, existingVLANMap map[string]e.Vlan) {
	vlanUpdates := make(map[string][]string)   // VLAN ID -> new IP segments
	vlanCreations := make(map[string][]string) // VLAN ID -> IP segments

	for _, mapping := range plan.IPMappings {
		// Use the subnet CIDR as the segment when available, otherwise fall back to bare IP
		segment := mapping.IP
		if mapping.Subnet != "" {
			segment = mapping.Subnet
		}
		if mapping.IsNewVLAN {
			vlanCreations[mapping.VLANNumber] = append(vlanCreations[mapping.VLANNumber], segment)
		} else {
			if _, exists := existingVLANMap[mapping.VLANNumber]; exists {
				vlanUpdates[mapping.VLANNumber] = append(vlanUpdates[mapping.VLANNumber], segment)
			} else {
				// VLAN doesn't exist, need to create it
				vlanCreations[mapping.VLANNumber] = append(vlanCreations[mapping.VLANNumber], segment)
			}
		}
	}

	// Create VLAN creation plans
	for vlanID, ipSegments := range vlanCreations {
		vlanName := ns.generateVLANName(vlanID)
		plan.VLANsToCreate = append(plan.VLANsToCreate, s.VLANPlan{
			VLANNumber:   vlanID,
			VLANName:     vlanName,
			IPSegmentIDs: ipSegments,
			Action:       "create",
		})
	}

	// Create VLAN update plans
	for vlanID, newSegments := range vlanUpdates {
		if existingVLAN, exists := existingVLANMap[vlanID]; exists {
			plan.VLANsToUpdate = append(plan.VLANsToUpdate, s.VLANPlan{
				VLANNumber:       vlanID,
				VLANName:         existingVLAN.VlanName,
				IPSegmentIDs:     newSegments,
				Action:           "update",
				ExistingSegments: []string{existingVLAN.IPSegment},
			})
		}
	}
}

// generateVLANName creates a descriptive name for a VLAN
func (ns *NetService) generateVLANName(vlanID string) string {
	switch vlanID {
	case "-1":
		return "Generic Network Segment"
	case "10":
		return "VLAN 10 (Private Network)"
	case "172":
		return "VLAN 172 (Private Network)"
	case "192":
		return "VLAN 192 (Private Network)"
	default:
		if vlanID[0] == '-' {
			return fmt.Sprintf("Custom Segment %s", vlanID)
		}
		return fmt.Sprintf("VLAN %s", vlanID)
	}
}

// generateImportSummary creates a human-readable summary of the import plan
func (ns *NetService) generateImportSummary(plan s.DeviceImportPlan) string {
	totalIPs := 0
	totalVLANsToCreate := 0
	totalVLANsToUpdate := 0

	for _, interfacePlan := range plan.InterfacePlans {
		totalIPs += len(interfacePlan.IPMappings)
		totalVLANsToCreate += len(interfacePlan.VLANsToCreate)
		totalVLANsToUpdate += len(interfacePlan.VLANsToUpdate)
	}

	return fmt.Sprintf("Device: %s, Interfaces: %d, IPs: %d, VLANs to create: %d, VLANs to update: %d",
		plan.Device.SuggestedName, len(plan.InterfacePlans), totalIPs, totalVLANsToCreate, totalVLANsToUpdate)
}

// ExecuteApprovedImportPlan executes an approved import plan
func (ns *NetService) ExecuteApprovedImportPlan(plan s.DeviceImportPlan, options s.ImportOptions) error {
	// First ensure required entities exist (zone types, zones, brands, etc.)
	if err := ns.ensureRequiredEntities(options); err != nil {
		return fmt.Errorf("failed to ensure required entities: %w", err)
	}

	// Create/update VLANs
	if err := ns.executeVLANPlans(plan); err != nil {
		return fmt.Errorf("failed to execute VLAN plans: %w", err)
	}

	// Then create the device and its ports with proper VLAN configurations
	// (importSingleDeviceWithPlan already returns descriptive, user-facing errors).
	if err := ns.importSingleDeviceWithPlan(plan.Device, plan, options); err != nil {
		return err
	}

	return nil
}

// executeVLANPlans executes VLAN creation and update plans
func (ns *NetService) executeVLANPlans(plan s.DeviceImportPlan) error {
	// Create new VLANs
	for _, interfacePlan := range plan.InterfacePlans {
		for _, vlanPlan := range interfacePlan.VLANsToCreate {
			// Use first IP segment if available, otherwise empty string
			ipSegment := ""
			if len(vlanPlan.IPSegmentIDs) > 0 {
				ipSegment = vlanPlan.IPSegmentIDs[0]
			}
			if err := ns.AddVlan(vlanPlan.VLANNumber, vlanPlan.VLANName, ipSegment); err != nil {
				if !strings.Contains(err.Error(), "already exists") {
					return fmt.Errorf("failed to create VLAN %s: %w", vlanPlan.VLANNumber, err)
				}
			}
		}

		// Update existing VLANs
		for _, vlanPlan := range interfacePlan.VLANsToUpdate {
			// Use first IP segment if available, otherwise empty string
			ipSegment := ""
			if len(vlanPlan.IPSegmentIDs) > 0 {
				ipSegment = vlanPlan.IPSegmentIDs[0]
			}
			if err := ns.UpdateVlanIPSegment(vlanPlan.VLANNumber, ipSegment); err != nil {
				return fmt.Errorf("failed to update VLAN %s: %w", vlanPlan.VLANNumber, err)
			}
		}
	}

	return nil
}

// importSingleDeviceWithPlan imports a device using the approved import plan
func (ns *NetService) importSingleDeviceWithPlan(discovered s.DiscoveredDevice, plan s.DeviceImportPlan, options s.ImportOptions) error {
	zoneName := discovered.SuggestedZone
	if options.DefaultZone != "" {
		zoneName = options.DefaultZone
	}
	// Default to "Generic" zone if no zone is specified
	if zoneName == "" {
		zoneName = "Generic"
	}

	// Ensure model exists before adding device
	err := ns.EnsureModelExists(discovered.Model, discovered.Brand, options.DefaultBrand, discovered.OsType)
	if err != nil {
		return fmt.Errorf("failed to ensure model exists: %w", err)
	}

	err = ns.AddDevice(
		discovered.SuggestedName,
		discovered.Model,
		"",
		zoneName,
		"Discovered",
		false,
		false,
	)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("a device named %q is already in the database — it was probably imported before; "+
				"rename it in the scan results, or delete the existing device first (overwriting is not supported)", discovered.SuggestedName)
		}
		return fmt.Errorf("failed to add device: %w", err)
	}

	// Locate the freshly-created device (labels are unique) to attach the scan
	// profile it was discovered with, and its ports.
	devices, _ := ns.GetDevices()
	var deviceID string
	for _, device := range devices {
		if device.Label == discovered.SuggestedName {
			deviceID = device.ID
			break
		}
	}

	if deviceID != "" && discovered.Profile != "" {
		if err := ns.UpdateDeviceProfile(deviceID, discovered.Profile); err != nil {
			log.Printf("Failed to set scan profile for device %s: %v", discovered.SuggestedName, err)
		}
	}

	if deviceID != "" && (len(discovered.Device.Interfaces) > 0 || len(discovered.Device.SwitchPorts) > 0) {
		if err := ns.createDevicePortsWithPlan(deviceID, discovered, plan); err != nil {
			log.Printf("Failed to create ports for device %s: %v", discovered.SuggestedName, err)
		}
	}

	return nil
}

// createDevicePortsWithPlan creates device ports using the approved import plan
func (ns *NetService) createDevicePortsWithPlan(deviceID string, discovered s.DiscoveredDevice, plan s.DeviceImportPlan) error {
	models, err := ns.GetModels()
	if err != nil {
		return err
	}

	var modelName string
	for _, model := range models {
		if model.Model == discovered.Model {
			modelName = model.Model
			break
		}
	}

	if modelName == "" {
		modelName = discovered.Model
		if modelName == "" {
			modelName = "Generic Model"
		}
		if err := ns.AddModel(modelName, discovered.Brand, discovered.ModelType, discovered.OsType); err != nil {
			return err
		}
	}

	// First pass: collect WiFi radio names from phy*-ap* interfaces
	// OpenWrt doesn't expose physical radios in SNMP, so we infer them
	wifiRadios := ns.collectWifiRadios(discovered.Device.Interfaces)

	// Create physical WiFi radio ports and DevicePort records
	for radioIdx, radioName := range wifiRadios {
		// Determine band from AP interfaces (heuristic: check if any AP on this radio has 5GHz or 6GHz)
		band := "2.4GHz" // default
		// Find the MAC address from the first AP interface on this radio
		var radioMAC string
		for _, iface := range discovered.Device.Interfaces {
			if strings.HasPrefix(iface.Name, radioName+"-ap") {
				if iface.WifiBand != "" {
					band = iface.WifiBand
				}
				if radioMAC == "" && iface.MAC != "" {
					radioMAC = iface.MAC
				}
				break
			}
		}

		if err := ns.AddModelPort(radioName, fmt.Sprintf("%d", radioIdx), "0", modelName, false, "wifi", band); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}

		// Create DevicePort for the WiFi radio
		modelPorts, _ := ns.GetModelPorts()
		var modelPortID string
		for _, mp := range modelPorts {
			if mp.Name == radioName && mp.Model == modelName {
				modelPortID = mp.ID
				break
			}
		}

		if modelPortID != "" {
			if _, err := ns.AddDevicePort(deviceID, modelPortID, radioMAC, nil); err != nil {
				if !strings.Contains(err.Error(), "already exists") {
					return err
				}
			}
		}
	}

	// createdPorts tracks physical port names already created for this device so the
	// switch-port (board.json) and interface (UCI/SNMP) passes can't create the same
	// kernel port twice (e.g. "eth1" appearing in both sources).
	createdPorts := make(map[string]bool)

	// Process physical switch ports from configuration (e.g., OpenWrt board.json)
	for _, port := range discovered.Device.SwitchPorts {
		portName := port.Name
		if portName == "" {
			portName = fmt.Sprintf("%s%d", port.Role, port.PortNumber)
		}
		if createdPorts[portName] {
			continue
		}

		portType := "ethernet"
		band := ""
		if port.Role == "wan" {
			portType = "wan"
		} else if port.Role == "lan" {
			portType = "lan"
		}

		if err := ns.AddModelPort(portName, fmt.Sprintf("%d", port.PortNumber), "0", modelName, false, portType, band); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}

		// Create DevicePort for the switch port
		modelPorts, _ := ns.GetModelPorts()
		var modelPortID string
		for _, mp := range modelPorts {
			if mp.Name == portName && mp.Model == modelName {
				modelPortID = mp.ID
				break
			}
		}

		if modelPortID != "" {
			var vlanConfigs []e.PortVlanConfig
			for _, vlan := range port.VLANs {
				vlanConfigs = append(vlanConfigs, e.PortVlanConfig{
					VlanNumber: vlan.VID,
					Tagged:     vlan.Tagged,
				})
			}
			if _, err := ns.AddDevicePort(deviceID, modelPortID, port.MAC, vlanConfigs); err != nil {
				if !strings.Contains(err.Error(), "already exists") {
					return err
				}
			}
			createdPorts[portName] = true
		}
	}

	// Process physical ports from interfaces
	portIdx := 0
	for _, iface := range discovered.Device.Interfaces {
		if !iface.IsPhysicalPort() {
			continue
		}

		portName := iface.Name
		if portName == "" {
			if iface.IsWifiRadio() {
				portName = fmt.Sprintf("radio%d", portIdx)
			} else {
				portName = fmt.Sprintf("eth%d", portIdx)
			}
		}
		// Already created as a switch port (kernel name) above — don't duplicate.
		if createdPorts[portName] {
			continue
		}
		createdPorts[portName] = true

		portType := ""
		band := ""
		if iface.IsWifiRadio() {
			portType = "wifi"
			band = iface.WifiBand
		}

		if err := ns.AddModelPort(portName, fmt.Sprintf("%d", portIdx), "0", modelName, false, portType, band); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
		portIdx++

		modelPorts, _ := ns.GetModelPorts()
		var modelPortID string
		for _, mp := range modelPorts {
			if mp.Name == portName && mp.Model == modelName {
				modelPortID = mp.ID
				break
			}
		}

		if modelPortID == "" {
			continue
		}

		// Find matching interface plan to convert IP mappings to VLAN data
		var interfacePlan *s.InterfaceImportPlan
		for _, iPlan := range plan.InterfacePlans {
			if iPlan.Interface.Name == iface.Name {
				interfacePlan = &iPlan
				break
			}
		}

		// Convert IP mappings to VLAN memberships if plan exists
		if interfacePlan != nil && len(iface.VLANs) == 0 {
			vlanMap := make(map[string]bool)
			for _, mapping := range interfacePlan.IPMappings {
				// Only process valid VLAN numbers (skip -1 and empty)
				if mapping.VLANNumber != "" && mapping.VLANNumber != "-1" {
					vlanMap[mapping.VLANNumber] = true
				}
			}

			// Populate interface VLANs array from plan mappings
			for vlanNum := range vlanMap {
				iface.VLANs = append(iface.VLANs, s.VLANMembership{
					VLANNumber: vlanNum,
					Tagged:     true, // Default to tagged - can be enhanced with better logic
				})
			}
		}

		// Use VLAN configurations (now populated from plan if needed)
		vlanConfigs := make([]e.PortVlanConfig, 0, len(iface.VLANs))
		for _, v := range iface.VLANs {
			vlanConfigs = append(vlanConfigs, e.PortVlanConfig{
				VlanNumber: v.VLANNumber,
				Tagged:     v.Tagged,
			})
		}

		if _, err := ns.AddDevicePort(deviceID, modelPortID, iface.MAC, vlanConfigs); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}

	// Create DeviceInterface records for virtual interfaces (wifi-iface SSIDs and logical/VLAN UCI interfaces)
	// Build a map of physical port names to modelPortIDs for linking
	physicalPortMap := make(map[string]string) // port name -> modelPortID
	modelPorts, _ := ns.GetModelPorts()
	for _, iface := range discovered.Device.Interfaces {
		if !iface.IsPhysicalPort() {
			continue
		}
		portName := iface.Name
		if portName == "" {
			if iface.IsWifiRadio() {
				portName = fmt.Sprintf("radio%d", portIdx)
			} else {
				portName = fmt.Sprintf("eth%d", portIdx)
			}
		}
		// Find the modelPortID for this physical port
		for _, mp := range modelPorts {
			if mp.Name == portName && mp.Model == modelName {
				physicalPortMap[portName] = mp.ID
				break
			}
		}
	}

	// Add SwitchPorts (from board.json) to physicalPortMap so bridges can link via parent chain
	for _, port := range discovered.Device.SwitchPorts {
		portName := port.Name
		if portName == "" {
			portName = fmt.Sprintf("%s%d", port.Role, port.PortNumber)
		}
		// Find the modelPortID for this switch port
		for _, mp := range modelPorts {
			if mp.Name == portName && mp.Model == modelName {
				physicalPortMap[portName] = mp.ID
				break
			}
		}
	}

	for _, iface := range discovered.Device.Interfaces {
		if iface.IsPhysicalPort() {
			continue
		}
		vlanConfigs := buildVlanConfigs(iface.VLANs)
		// Pass the Parent field from scanner interface
		if err := ns.AddDeviceInterface(deviceID, iface.Name, "", iface.Parent, vlanConfigs, iface.IPAddresses, iface.WifiSSID, iface.WifiSecurity); err != nil {
			log.Printf("Warning: failed to create interface %s for device %s: %v", iface.Name, deviceID, err)
			continue
		}

		// Find the newly created interface and link it to the appropriate physical port
		// Traverse parent chain to find the root physical port
		rootPortName := iface.Parent
		for rootPortName != "" {
			if modelPortID, exists := physicalPortMap[rootPortName]; exists {
				// Found the root physical port - link this interface to it
				ifaceList, err := ns.GetDeviceInterfaces(deviceID)
				if err == nil {
					for _, di := range ifaceList {
						if di.Name == iface.Name && di.DeviceID == deviceID {
							if err := ns.AddInterfacePort(di.ID, deviceID, modelPortID); err != nil {
								log.Printf("Warning: failed to link interface %s to port %s: %v", di.ID, modelPortID, err)
							}
							break
						}
					}
				}
				break
			}
			// Need to traverse up the parent chain to find the physical port
			parentIface := findInterfaceByName(discovered.Device.Interfaces, rootPortName)
			if parentIface == nil {
				break
			}
			if parentIface.Parent == "" || parentIface.IsPhysicalPort() {
				break
			}
			rootPortName = parentIface.Parent
		}
	}

	return nil
}

// collectWifiRadios extracts WiFi radio names from phy*-ap* interface patterns
// e.g., from [phy0-ap0, phy0-ap1, phy1-ap0] returns ["phy0", "phy1"]
func (ns *NetService) collectWifiRadios(interfaces []s.DeviceInterface) []string {
	radioSet := make(map[string]bool)
	for _, iface := range interfaces {
		if strings.HasPrefix(iface.Name, "phy") {
			parts := strings.Split(iface.Name, "-")
			if len(parts) >= 1 {
				radioSet[parts[0]] = true
			}
		}
	}

	radios := make([]string, 0, len(radioSet))
	for radio := range radioSet {
		radios = append(radios, radio)
	}
	return radios
}

// vlanExists checks if a VLAN number exists in the provided slice of existing VLANs
func (ns *NetService) vlanExists(vlanNumber string, existingVLANs []e.Vlan) bool {
	for _, vlan := range existingVLANs {
		if vlan.VlanID == vlanNumber {
			return true
		}
	}
	return false
}

// findInterfaceByName finds a scanner interface by name from a list of interfaces
func findInterfaceByName(interfaces []s.DeviceInterface, name string) *s.DeviceInterface {
	for i := range interfaces {
		if interfaces[i].Name == name {
			return &interfaces[i]
		}
	}
	return nil
}
