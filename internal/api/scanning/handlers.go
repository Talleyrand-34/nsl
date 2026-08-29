package scanning

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	configparser "nsl-graph/internal/configparser"
	"nsl-graph/internal/observ"
	q "nsl-graph/internal/repository/application"
	e "nsl-graph/internal/repository/entities"
	s "nsl-graph/internal/scanner"
	"nsl-graph/internal/secret"
)

// startAsyncScan creates a tracked run, executes fn in a goroutine, and replies
// immediately with the scan_id. The client polls GET /scan/status?scan_id= for
// progress, granular events, and (on completion) the result. fn returns the same
// payload the endpoint used to return synchronously, so result rendering is
// unchanged.
func startAsyncScan(w http.ResponseWriter, kind, title string, fn func(run *observ.Run) (any, error)) {
	run := observ.Runs.NewRun(kind, title)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				run.Fail(fmt.Errorf("panic: %v", p))
			}
		}()
		res, err := fn(run)
		if err != nil {
			run.Fail(err)
		} else {
			run.Finish(res)
		}
	}()
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"scan_id": run.ID})
}

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

		startAsyncScan(w, "network", req.Subnet, func(run *observ.Run) (any, error) {
			options.OnProgress = func(done, total int, ip string, reachable bool) {
				run.Progress(done, total)
				if reachable {
					run.Emit("info", "host responded to SNMP", "ip", ip, "done", done, "total", total)
				}
			}
			scanResult, err := service.ScanNetwork(req.Subnet, options)
			if err != nil {
				return nil, err
			}
			devices, err := service.DiscoverDevices(scanResult)
			if err != nil {
				return nil, err
			}
			return ScanNetworkResponse{
				ScanID:    scanResult.ID,
				Subnet:    scanResult.Subnet,
				StartTime: scanResult.StartTime,
				EndTime:   scanResult.EndTime,
				Duration:  scanResult.EndTime.Sub(scanResult.StartTime).String(),
				Total:     len(devices),
				Devices:   devices,
			}, nil
		})
	}
}

// ScanRunRequest is the body for the unified /scan/run endpoint. method is
// "snmp" or "ssh"; target is a single IP (single scan) or a CIDR / comma-list
// (batch). For SSH, the profile supplies the credentials and os_type; the
// stored secrets are decrypted by the (already unlocked) server vault.
type ScanRunRequest struct {
	Method      string `json:"method"`
	Target      string `json:"target"`
	Community   string `json:"community,omitempty"`
	SNMPVersion string `json:"snmp_version,omitempty"`
	SNMPPort    uint16 `json:"snmp_port,omitempty"`
	Timeout     int    `json:"timeout,omitempty"`
	Profile     string `json:"profile,omitempty"`
	OsType      string `json:"os_type,omitempty"`
}

// ScanRunHandler runs a unified device scan (SNMP or SSH; single or batch,
// inferred from the target) and returns the classified devices to import.
func ScanRunHandler(service q.NetServiceInt) http.HandlerFunc {
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

		var req ScanRunRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_json", Message: err.Error()})
			return
		}
		if strings.TrimSpace(req.Target) == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "missing_target", Message: "a target IP or CIDR is required"})
			return
		}

		method := req.Method
		if method == "" {
			method = "snmp"
		}
		startAsyncScan(w, "run", fmt.Sprintf("%s %s", method, req.Target), func(run *observ.Run) (any, error) {
			devices, err := service.RunScan(q.RunScanOptions{
				Method:      req.Method,
				Target:      req.Target,
				Community:   req.Community,
				SNMPVersion: req.SNMPVersion,
				SNMPPort:    req.SNMPPort,
				TimeoutSec:  req.Timeout,
				Profile:     req.Profile,
				OsType:      req.OsType,
			}, run)
			if err != nil {
				return nil, err
			}
			return map[string]any{"devices": devices, "total": len(devices)}, nil
		})
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

		startAsyncScan(w, "host", req.IP, func(run *observ.Run) (any, error) {
			run.Emit("info", "scanning host over SNMP", "ip", req.IP)
			run.Progress(0, 1)
			device, err := service.ScanDevice(req.IP, options)
			if err != nil {
				return nil, err
			}
			run.Progress(1, 1)
			discoverer := s.NewDeviceDiscoverer()
			brand, model, class := discoverer.ClassifyDevice(*device)
			run.Emit("info", "device scan completed", "ip", device.IP, "sysname", device.SysName, "brand", brand, "model", model)
			return s.DiscoveredDevice{
				Device:        *device,
				Brand:         brand,
				Model:         model,
				ModelType:     class,
				SuggestedName: discoverer.GenerateDeviceName(*device, class),
				SuggestedZone: discoverer.SuggestZone(*device),
			}, nil
		})
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

		slog.Info("importing devices", "count", len(req.Devices))

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

		slog.Info("import completed", "count", len(req.Devices))

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

		slog.Info("importing devices from uploaded scan file", "count", len(devices))

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
// accepted in clear on input only and is immediately encrypted by the server
// vault (which must be unlocked); the clear value is never stored or returned.
type scanProfileRequest struct {
	Name              string `json:"name"`
	Kind              string `json:"kind"` // "device" (default) or "generic"
	Host              string `json:"host"`
	SNMPCommunity     string `json:"snmp_community"`
	SNMPVersion       string `json:"snmp_version"`
	SNMPPort          int    `json:"snmp_port"`
	TimeoutSec        int    `json:"timeout_sec"`
	ScanSource        string `json:"scan_source"`
	ConfigSource      string `json:"config_source"`
	ConfigFile        string `json:"config_file"`
	OsType            string `json:"os_type"`
	SSHUser           string `json:"ssh_user"`
	SSHPassword       string `json:"ssh_password"`
	SSHKeyFile        string `json:"ssh_key_file"`
	SSHKey            string `json:"ssh_key"` // PEM private-key content (uploaded)
	SSHPort           int    `json:"ssh_port"`
	DiscrepancyAction string `json:"discrepancy_action"`
	MergeConfigs      bool   `json:"merge_configs"`
	ConfigTimeout     int    `json:"config_timeout"`
	VLANAccuracy      int    `json:"vlan_accuracy"`

	// Hosts is the list of devices attached to this profile (multi-device
	// form). Empty for a generic profile. The overlap-detection pass scans
	// this list against existing profiles so the operator sees "10.0.0.245 is
	// already in profile 'lab-prod'" before persisting.
	Hosts []string `json:"hosts,omitempty"`
}

func (req scanProfileRequest) toEntity(v *secret.Vault) (e.ScanProfile, error) {
	p := e.ScanProfile{
		Name:              req.Name,
		Kind:              req.Kind,
		Host:              req.Host,
		SNMPCommunity:     req.SNMPCommunity,
		SNMPVersion:       req.SNMPVersion,
		SNMPPort:          req.SNMPPort,
		TimeoutSec:        req.TimeoutSec,
		ScanSource:        req.ScanSource,
		ConfigSource:      req.ConfigSource,
		ConfigFile:        req.ConfigFile,
		OsType:            req.OsType,
		SSHUser:           req.SSHUser,
		SSHKeyFile:        req.SSHKeyFile,
		SSHPort:           req.SSHPort,
		DiscrepancyAction: req.DiscrepancyAction,
		MergeConfigs:      req.MergeConfigs,
		ConfigTimeout:     req.ConfigTimeout,
		VLANAccuracy:      req.VLANAccuracy,
	}
	if req.SSHPassword != "" {
		blob, err := v.Encrypt(req.SSHPassword)
		if err != nil {
			return p, fmt.Errorf("store SSH password: %w (unlock the vault first)", err)
		}
		p.SSHPassword = blob
	}
	if req.SSHKey != "" {
		blob, err := v.Encrypt(req.SSHKey)
		if err != nil {
			return p, fmt.Errorf("store SSH private key: %w (unlock the vault first)", err)
		}
		p.SSHKey = blob
	}
	return p, nil
}

// profileHostOverlap flags a host that's already attached to a different
// profile. Surfaced in the POST response so the multi-device form can warn
// the operator before persisting.
type profileHostOverlap struct {
	Host        string `json:"host"`
	OtherProfile string `json:"other_profile"`
}

// detectProfileOverlaps walks every saved profile and reports any host
// in hosts that's already attached to a profile other than thisProfile.
// Same profile is excluded (the operator is editing their own profile).
func detectProfileOverlaps(service q.NetServiceInt, thisProfile string, hosts []string) []profileHostOverlap {
	if len(hosts) == 0 {
		return nil
	}
	want := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		want[h] = true
	}
	all, err := service.GetScanProfiles()
	if err != nil {
		return nil
	}
	seen := make(map[string]bool)
	var out []profileHostOverlap
	for _, p := range all {
		if p.Name == thisProfile {
			continue
		}
		if want[p.Host] {
			key := p.Host + "\x00" + p.Name
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, profileHostOverlap{Host: p.Host, OtherProfile: p.Name})
		}
	}
	return out
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
			p, err := req.toEntity(service.Vault())
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
			overlaps := detectProfileOverlaps(service, p.Name, req.Hosts)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message":  "created",
				"name":     p.Name,
				"overlaps": overlaps,
			})

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
// (encrypted) SSH credentials, decrypted by the unlocked server vault.
type ScanHostSSHRequest struct {
	Profile string `json:"profile"`
	IP      string `json:"ip,omitempty"` // optional override of the profile's host
}

// ScanHostSSHHandler runs an SSH/config scan of a host. SSH credentials come
// from a saved scan profile; the profile's encrypted SSH secrets are decrypted
// by the server vault (which must be unlocked).
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
		if profile.OsType == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "missing_os_type", Message: "the profile must set a os_type for SSH scans"})
			return
		}

		password := ""
		if profile.SSHPassword != "" {
			pw, err := service.Vault().Decrypt(profile.SSHPassword)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "vault_locked", Message: err.Error()})
				return
			}
			password = pw
		}
		privateKey := ""
		if profile.SSHKey != "" {
			pk, err := service.Vault().Decrypt(profile.SSHKey)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "vault_locked", Message: err.Error()})
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

		startAsyncScan(w, "host-ssh", ip, func(run *observ.Run) (any, error) {
			run.Emit("info", "reading config over SSH", "ip", ip, "profile", profile.Name, "os_type", profile.OsType)
			run.Progress(0, 1)
			device, err := service.ScanDeviceViaSSH(ip, profile.OsType, creds)
			if err != nil {
				return nil, err
			}
			run.Progress(1, 1)
			discoverer := s.NewDeviceDiscoverer()
			brand, model, class := discoverer.ClassifyDevice(*device)
			run.Emit("info", "device read over SSH", "ip", ip, "sysname", device.SysName)
			return s.DiscoveredDevice{
				Device:        *device,
				Brand:         brand,
				Model:         model,
				ModelType:     class,
				SuggestedName: discoverer.GenerateDeviceName(*device, class),
				SuggestedZone: discoverer.SuggestZone(*device),
			}, nil
		})
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
			// A name clash with an existing device is a conflict, not a server error.
			if strings.Contains(err.Error(), "already in the database") || strings.Contains(err.Error(), "already exists") {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "device_exists", Message: err.Error()})
				return
			}
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

// GetScanStatusHandler reports the live status of an async scan: its state,
// progress counters, the events since the client's cursor (?since=<seq>), and —
// once completed — the result payload (or the error if it failed).
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
		since := 0
		if v := r.URL.Query().Get("since"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				since = n
			}
		}

		run, ok := observ.Runs.Get(scanID)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "scan_not_found", Message: fmt.Sprintf("no scan with id %q (it may have expired)", scanID)})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(run.Snapshot(since))
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
	// Accept one or more comma/space-separated CIDRs or IPs; all must be valid.
	tokens := s.SplitSubnets(subnet)
	if len(tokens) == 0 {
		return false
	}
	for _, tok := range tokens {
		if !validateSingleSubnet(tok) {
			return false
		}
	}
	return true
}

func validateSingleSubnet(subnet string) bool {
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
