package scanning

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	configparser "nsl-graph/internal/configparser"
	q "nsl-graph/internal/repository/application"
	e "nsl-graph/internal/repository/entities"
	s "nsl-graph/internal/scanner"
	"nsl-graph/internal/secret"
)

type ScanNetworkRequest struct {
	Subnet      string `json:"subnet"`
	Timeout     int    `json:"timeout,omitempty"`
	Community   string `json:"community,omitempty"`
	SNMPVersion string `json:"snmp_version,omitempty"`
	SNMPPort    uint16 `json:"snmp_port,omitempty"`
	Profile     string `json:"profile,omitempty"` // saved profile to apply (else auto-match by subnet)
}

type ScanNetworkResponse struct {
	ScanID    string               `json:"scan_id"`
	Subnet    string               `json:"subnet"`
	StartTime time.Time            `json:"start_time"`
	EndTime   time.Time            `json:"end_time"`
	Duration  string               `json:"duration"`
	Total     int                  `json:"total"`
	Devices   []s.DiscoveredDevice `json:"devices"`
}

type ScanHostRequest struct {
	IP          string `json:"ip"`
	Timeout     int    `json:"timeout,omitempty"`
	Community   string `json:"community,omitempty"`
	SNMPVersion string `json:"snmp_version,omitempty"`
	SNMPPort    uint16 `json:"snmp_port,omitempty"`
	Profile     string `json:"profile,omitempty"` // saved profile to apply (else auto-match by ip)
}

type ImportDevicesRequest struct {
	Devices []s.DiscoveredDevice `json:"devices"`
	Options s.ImportOptions      `json:"options"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func ScanNetworkHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method_not_allowed", Message: "Only POST method is allowed"})
			return
		}

		var req ScanNetworkRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_json", Message: err.Error()})
			return
		}

		if req.Subnet == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "missing_subnet", Message: "subnet is required"})
			return
		}

		// Apply a saved profile (explicit or auto-matched by subnet); only empty
		// request fields are filled (explicit values win). SNMP-only.
		if p, ok := service.ResolveScanProfile(req.Subnet, req.Profile); ok {
			if req.Community == "" {
				req.Community = p.SNMPCommunity
			}
			if req.SNMPVersion == "" {
				req.SNMPVersion = p.SNMPVersion
			}
			if req.SNMPPort == 0 {
				req.SNMPPort = uint16(p.SNMPPort)
			}
			if req.Timeout == 0 {
				req.Timeout = p.TimeoutSec
			}
		} else if req.Profile != "" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "profile_not_found", Message: fmt.Sprintf("no scan profile named %q", req.Profile)})
			return
		}

		timeout := time.Duration(req.Timeout) * time.Second
		if req.Timeout == 0 {
			timeout = 30 * time.Second
		}

		options := s.ScanOptions{
			Subnet:  req.Subnet,
			Timeout: timeout,
			SNMP: s.SNMPOptions{
				Community: req.Community,
				Version:   req.SNMPVersion,
				Port:      req.SNMPPort,
			},
		}

		log.Printf("Starting SNMP network scan of %s", req.Subnet)

		scanResult, err := service.ScanNetwork(req.Subnet, options)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "scan_failed", Message: err.Error()})
			return
		}

		devices, err := service.DiscoverDevices(scanResult)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "discovery_failed", Message: err.Error()})
			return
		}

		response := ScanNetworkResponse{
			ScanID:    scanResult.ID,
			Subnet:    scanResult.Subnet,
			StartTime: scanResult.StartTime,
			EndTime:   scanResult.EndTime,
			Duration:  scanResult.EndTime.Sub(scanResult.StartTime).String(),
			Total:     len(devices),
			Devices:   devices,
		}

		log.Printf("Scan completed: %d devices discovered", len(devices))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func ScanHostHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method_not_allowed", Message: "Only POST method is allowed"})
			return
		}

		var req ScanHostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_json", Message: err.Error()})
			return
		}

		if req.IP == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "missing_ip", Message: "IP address is required"})
			return
		}

		// Apply a saved profile (explicit or auto-matched by ip); only empty
		// request fields are filled (explicit values win). The API scan is
		// SNMP-only, so the SSH password is never touched here.
		if p, ok := service.ResolveScanProfile(req.IP, req.Profile); ok {
			if req.Community == "" {
				req.Community = p.SNMPCommunity
			}
			if req.SNMPVersion == "" {
				req.SNMPVersion = p.SNMPVersion
			}
			if req.SNMPPort == 0 {
				req.SNMPPort = uint16(p.SNMPPort)
			}
			if req.Timeout == 0 {
				req.Timeout = p.TimeoutSec
			}
		} else if req.Profile != "" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "profile_not_found", Message: fmt.Sprintf("no scan profile named %q", req.Profile)})
			return
		}

		timeout := time.Duration(req.Timeout) * time.Second
		if req.Timeout == 0 {
			timeout = 10 * time.Second
		}

		options := s.ScanOptions{
			Timeout: timeout,
			SNMP: s.SNMPOptions{
				Community: req.Community,
				Version:   req.SNMPVersion,
				Port:      req.SNMPPort,
			},
		}

		log.Printf("Starting SNMP device scan of %s", req.IP)

		device, err := service.ScanDevice(req.IP, options)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "scan_failed", Message: err.Error()})
			return
		}

		discoverer := s.NewDeviceDiscoverer()
		brand, model, class := discoverer.ClassifyDevice(*device)

		discovered := s.DiscoveredDevice{
			Device:        *device,
			Brand:         brand,
			Model:         model,
			DeviceClass:   class,
			SuggestedName: discoverer.GenerateDeviceName(*device, class),
			SuggestedZone: discoverer.SuggestZone(*device),
		}

		log.Printf("Device scan completed: %s (%s) - %s %s", device.IP, device.SysName, brand, model)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(discovered)
	}
}

func ImportDevicesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method_not_allowed", Message: "Only POST method is allowed"})
			return
		}

		var req ImportDevicesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_json", Message: err.Error()})
			return
		}

		if len(req.Devices) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "no_devices", Message: "No devices provided for import"})
			return
		}

		log.Printf("Importing %d devices", len(req.Devices))

		if err := service.ImportScanResults(req.Devices, req.Options); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "import_failed", Message: err.Error()})
			return
		}

		response := map[string]interface{}{
			"success":          true,
			"imported_devices": len(req.Devices),
			"message":          fmt.Sprintf("Successfully imported %d devices", len(req.Devices)),
		}

		log.Printf("Import completed successfully")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// ImportScanFileHandler imports devices from a raw SNMP scan-result file (the
// same JSON format consumed by `nsl-graph scan import <file>`). The body is a
// scanner.ScanResult; the handler classifies the devices server-side via
// DiscoverDevices and then imports them, mirroring the CLI's --auto-import path.
func ImportScanFileHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method_not_allowed", Message: "Only POST method is allowed"})
			return
		}

		var scanResult s.ScanResult
		if err := json.NewDecoder(r.Body).Decode(&scanResult); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_json", Message: err.Error()})
			return
		}

		if len(scanResult.Devices) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "no_devices", Message: "Scan result contains no devices"})
			return
		}

		devices, err := service.DiscoverDevices(&scanResult)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "discovery_failed", Message: err.Error()})
			return
		}

		if len(devices) == 0 {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":          true,
				"imported_devices": 0,
				"message":          "No reachable devices to import",
			})
			return
		}

		// Defaults mirror the CLI `scan import --auto-import` behaviour.
		options := s.ImportOptions{
			AutoImport:        true,
			CreateZones:       true,
			DefaultZone:       "Discovered",
			SkipExisting:      true,
			InteractiveVLANs:  false,
			VLANAccuracyLevel: 2,
		}

		log.Printf("Importing %d devices from uploaded scan file", len(devices))

		if err := service.ImportScanResults(devices, options); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "import_failed", Message: err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":          true,
			"imported_devices": len(devices),
			"message":          fmt.Sprintf("Successfully imported %d devices", len(devices)),
		})
	}
}

// scanProfileRequest is the POST/PUT body for /scan/profiles. ssh_password is
// accepted in clear on input only and is immediately encrypted with passphrase;
// neither is ever stored or returned.
type scanProfileRequest struct {
	Name              string `json:"name"`
	Host              string `json:"host"`
	SNMPCommunity     string `json:"snmp_community"`
	SNMPVersion       string `json:"snmp_version"`
	SNMPPort          int    `json:"snmp_port"`
	TimeoutSec        int    `json:"timeout_sec"`
	ScanSource        string `json:"scan_source"`
	ConfigSource      string `json:"config_source"`
	ConfigFile        string `json:"config_file"`
	DeviceType        string `json:"device_type"`
	SSHUser           string `json:"ssh_user"`
	SSHPassword       string `json:"ssh_password"`
	SSHKeyFile        string `json:"ssh_key_file"`
	SSHKey            string `json:"ssh_key"` // PEM private-key content (uploaded)
	SSHPort           int    `json:"ssh_port"`
	DiscrepancyAction string `json:"discrepancy_action"`
	MergeConfigs      bool   `json:"merge_configs"`
	ConfigTimeout     int    `json:"config_timeout"`
	VLANAccuracy      int    `json:"vlan_accuracy"`
	Passphrase        string `json:"passphrase"`
}

func (req scanProfileRequest) toEntity() (e.ScanProfile, error) {
	p := e.ScanProfile{
		Name:              req.Name,
		Host:              req.Host,
		SNMPCommunity:     req.SNMPCommunity,
		SNMPVersion:       req.SNMPVersion,
		SNMPPort:          req.SNMPPort,
		TimeoutSec:        req.TimeoutSec,
		ScanSource:        req.ScanSource,
		ConfigSource:      req.ConfigSource,
		ConfigFile:        req.ConfigFile,
		DeviceType:        req.DeviceType,
		SSHUser:           req.SSHUser,
		SSHKeyFile:        req.SSHKeyFile,
		SSHPort:           req.SSHPort,
		DiscrepancyAction: req.DiscrepancyAction,
		MergeConfigs:      req.MergeConfigs,
		ConfigTimeout:     req.ConfigTimeout,
		VLANAccuracy:      req.VLANAccuracy,
	}
	if req.SSHPassword != "" {
		if req.Passphrase == "" {
			return p, fmt.Errorf("a passphrase is required to store an SSH password")
		}
		blob, err := secret.Encrypt(req.SSHPassword, req.Passphrase)
		if err != nil {
			return p, err
		}
		p.SSHPassword = blob
	}
	if req.SSHKey != "" {
		if req.Passphrase == "" {
			return p, fmt.Errorf("a passphrase is required to store an SSH private key")
		}
		blob, err := secret.Encrypt(req.SSHKey, req.Passphrase)
		if err != nil {
			return p, err
		}
		p.SSHKey = blob
	}
	return p, nil
}

// ScanProfilesHandler exposes CRUD over saved scan profiles. GET never returns
// the SSH password (only `has_ssh_password`).
func ScanProfilesHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusOK)

		case http.MethodGet:
			profiles, err := service.GetScanProfiles()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "list_failed", Message: err.Error()})
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(profiles)

		case http.MethodPost, http.MethodPut:
			var req scanProfileRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_json", Message: err.Error()})
				return
			}
			if req.Name == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "missing_name", Message: "name is required"})
				return
			}
			p, err := req.toEntity()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_profile", Message: err.Error()})
				return
			}
			if r.Method == http.MethodPut {
				// Preserve existing encrypted secrets on metadata-only updates.
				if p.SSHPassword == "" || p.SSHKey == "" {
					if existing, _ := service.GetScanProfileByName(p.Name); existing != nil {
						if p.SSHPassword == "" {
							p.SSHPassword = existing.SSHPassword
						}
						if p.SSHKey == "" {
							p.SSHKey = existing.SSHKey
						}
					}
				}
				if err := service.UpdateScanProfile(p); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(ErrorResponse{Error: "save_failed", Message: err.Error()})
					return
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"message": "updated", "name": p.Name})
				return
			}
			if err := service.AddScanProfile(p); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "save_failed", Message: err.Error()})
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"message": "created", "name": p.Name})

		case http.MethodDelete:
			name := r.URL.Query().Get("name")
			if name == "" {
				var body struct {
					Name string `json:"name"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
					name = body.Name
				}
			}
			if name == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "missing_name", Message: "name is required"})
				return
			}
			if err := service.DeleteScanProfile(name); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "delete_failed", Message: err.Error()})
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "deleted", "name": name})

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method_not_allowed"})
		}
	}
}

// ScanHostSSHRequest scans a host over SSH using a saved profile's stored
// (encrypted) SSH credentials, unlocked with a passphrase.
type ScanHostSSHRequest struct {
	Profile    string `json:"profile"`
	Passphrase string `json:"passphrase"`
	IP         string `json:"ip,omitempty"` // optional override of the profile's host
}

// ScanHostSSHHandler runs an SSH/config scan of a host. SSH credentials come
// from a saved scan profile; the profile's encrypted SSH password is decrypted
// with the supplied passphrase (server-side only).
func ScanHostSSHHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method_not_allowed", Message: "Only POST method is allowed"})
			return
		}

		var req ScanHostSSHRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_json", Message: err.Error()})
			return
		}
		if req.Profile == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "missing_profile", Message: "a profile name is required for SSH scans"})
			return
		}

		profile, err := service.GetScanProfileByName(req.Profile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "lookup_failed", Message: err.Error()})
			return
		}
		if profile == nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "profile_not_found", Message: fmt.Sprintf("no scan profile named %q", req.Profile)})
			return
		}

		ip := profile.Host
		if req.IP != "" {
			ip = req.IP
		}
		if ip == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "missing_host", Message: "profile has no host and no ip override was given"})
			return
		}
		if profile.DeviceType == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "missing_device_type", Message: "the profile must set a device_type for SSH scans"})
			return
		}

		password := ""
		if profile.SSHPassword != "" {
			pw, err := secret.Decrypt(profile.SSHPassword, req.Passphrase)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "incorrect_passphrase", Message: err.Error()})
				return
			}
			password = pw
		}
		privateKey := ""
		if profile.SSHKey != "" {
			pk, err := secret.Decrypt(profile.SSHKey, req.Passphrase)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "incorrect_passphrase", Message: err.Error()})
				return
			}
			privateKey = pk
		}

		port := profile.SSHPort
		if port == 0 {
			port = 22
		}
		timeout := time.Duration(profile.ConfigTimeout) * time.Second
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		creds := configparser.SSHCredentials{
			Username:   profile.SSHUser,
			Password:   password,
			KeyFile:    profile.SSHKeyFile,
			PrivateKey: privateKey,
			Port:       port,
			Timeout:    timeout,
		}

		log.Printf("Starting SSH scan of %s (profile=%s, type=%s)", ip, profile.Name, profile.DeviceType)

		device, err := service.ScanDeviceViaSSH(ip, profile.DeviceType, creds)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "scan_failed", Message: err.Error()})
			return
		}

		discoverer := s.NewDeviceDiscoverer()
		brand, model, class := discoverer.ClassifyDevice(*device)
		discovered := s.DiscoveredDevice{
			Device:        *device,
			Brand:         brand,
			Model:         model,
			DeviceClass:   class,
			SuggestedName: discoverer.GenerateDeviceName(*device, class),
			SuggestedZone: discoverer.SuggestZone(*device),
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(discovered)
	}
}

// AnalyzeDeviceHandler returns the import plan (editable IP/VLAN/ip-segment
// tuples) for a discovered device.
func AnalyzeDeviceHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method_not_allowed", Message: "Only POST method is allowed"})
			return
		}

		var req struct {
			Device s.DiscoveredDevice `json:"device"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_json", Message: err.Error()})
			return
		}

		plan, err := service.AnalyzeDeviceForImport(req.Device)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "analyze_failed", Message: err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(plan)
	}
}

// ExecuteImportPlanHandler imports a (possibly user-edited) device import plan.
// VLAN create/update plans are regenerated from the edited IP mappings first so
// edits to vlan-id / ip-segment are applied consistently.
func ExecuteImportPlanHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method_not_allowed", Message: "Only POST method is allowed"})
			return
		}

		var req struct {
			Plan    s.DeviceImportPlan `json:"plan"`
			Options s.ImportOptions    `json:"options"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_json", Message: err.Error()})
			return
		}

		// Defaults mirror the CLI `scan import --auto-import` path.
		opts := req.Options
		if opts.DefaultZone == "" {
			opts.DefaultZone = "Discovered"
		}
		opts.AutoImport = true
		opts.CreateZones = true
		if opts.VLANAccuracyLevel == 0 {
			opts.VLANAccuracyLevel = 2
		}

		plan := service.RegenerateVLANPlans(req.Plan)
		if err := service.ExecuteApprovedImportPlan(plan, opts); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "import_failed", Message: err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Imported %s", plan.Device.SuggestedName),
		})
	}
}

func GetScanStatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		scanID := r.URL.Query().Get("scan_id")
		if scanID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "missing_scan_id", Message: "scan_id parameter is required"})
			return
		}

		status := map[string]interface{}{
			"scan_id": scanID,
			"status":  "completed",
			"message": "Scan status tracking not implemented",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(status)
	}
}

func ValidateSubnetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		subnet := r.URL.Query().Get("subnet")
		if subnet == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "missing_subnet", Message: "subnet parameter is required"})
			return
		}

		valid := validateSubnetFormat(subnet)

		response := map[string]interface{}{
			"subnet": subnet,
			"valid":  valid,
		}

		if !valid {
			response["message"] = "Invalid subnet format. Expected formats: 192.168.1.0/24, 10.0.0.1"
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func validateSubnetFormat(subnet string) bool {
	subnet = strings.TrimSpace(subnet)

	if strings.Contains(subnet, "/") {
		return true
	}

	parts := strings.Split(subnet, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if num, err := strconv.Atoi(part); err != nil || num < 0 || num > 255 {
			return false
		}
	}

	return true
}
