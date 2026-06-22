package scanning

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	q "nsl-graph/internal/repository/application"
	s "nsl-graph/internal/scanner"
)

type ScanNetworkRequest struct {
	Subnet      string `json:"subnet"`
	Timeout     int    `json:"timeout,omitempty"`
	Community   string `json:"community,omitempty"`
	SNMPVersion string `json:"snmp_version,omitempty"`
	SNMPPort    uint16 `json:"snmp_port,omitempty"`
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
