// SPDX-License-Identifier: AGPL-3.0-or-later
package scanning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	q "nsl-graph/internal/repository/application"
	"nsl-graph/internal/repository/entities"
)

// profileDeviceServiceStub is a hand-rolled stub satisfying q.NetServiceInt
// for the ProfileDevicesHandler. We embed the interface so we get every
// method, then override only the ones the handler exercises.
type profileDeviceServiceStub struct {
	q.NetServiceInt
	profiles map[string]*entities.ScanProfile // by name
	devices  map[string][]entities.ProfileDevice
}

func newProfileDeviceServiceStub() *profileDeviceServiceStub {
	return &profileDeviceServiceStub{
		profiles: map[string]*entities.ScanProfile{},
		devices:  map[string][]entities.ProfileDevice{},
	}
}

func (s *profileDeviceServiceStub) GetScanProfileByName(name string) (*entities.ScanProfile, error) {
	p, ok := s.profiles[name]
	if !ok {
		return nil, nil
	}
	clone := *p
	return &clone, nil
}

func (s *profileDeviceServiceStub) AddProfileDevice(d entities.ProfileDevice) error {
	if _, ok := s.profiles[d.ProfileName]; !ok {
		return fmt.Errorf("no scan profile named %q", d.ProfileName)
	}
	for _, existing := range s.devices[d.ProfileName] {
		if existing.Host == d.Host {
			return fmt.Errorf("device %q is already in profile %q", d.Host, d.ProfileName)
		}
	}
	s.devices[d.ProfileName] = append(s.devices[d.ProfileName], d)
	return nil
}

func (s *profileDeviceServiceStub) GetProfileDevices(profileName string) ([]entities.ProfileDevice, error) {
	out := make([]entities.ProfileDevice, len(s.devices[profileName]))
	copy(out, s.devices[profileName])
	return out, nil
}

func (s *profileDeviceServiceStub) DeleteProfileDevice(profileName, host string) error {
	devs := s.devices[profileName]
	for i, d := range devs {
		if d.Host == host {
			s.devices[profileName] = append(devs[:i], devs[i+1:]...)
			return nil
		}
	}
	return nil // DELETE is idempotent in the handler
}

func setupDeviceRouter(s *profileDeviceServiceStub) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/scan/profiles/{name}/devices", ProfileDevicesHandler(s)).Methods("GET", "POST", "DELETE", "OPTIONS")
	return r
}

func TestProfileDevices_GET_Empty(t *testing.T) {
	svc := newProfileDeviceServiceStub()
	svc.profiles["lab"] = &entities.ScanProfile{Name: "lab", Kind: "device", Host: "10.0.0.1"}
	r := setupDeviceRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/scan/profiles/lab/devices", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty array, got %v", got)
	}
}

func TestProfileDevices_POST_ThenList(t *testing.T) {
	svc := newProfileDeviceServiceStub()
	svc.profiles["lab"] = &entities.ScanProfile{Name: "lab", Kind: "device", Host: "10.0.0.1"}
	r := setupDeviceRouter(svc)
	body, _ := json.Marshal(map[string]string{"host": "10.0.0.10", "ssh_profile_name": "creds-pool"})
	req := httptest.NewRequest(http.MethodPost, "/scan/profiles/lab/devices", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/scan/profiles/lab/devices", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET expected 200, got %d", w.Code)
	}
	var got []map[string]any
	_ = json.NewDecoder(w.Body).Decode(&got)
	if len(got) != 1 {
		t.Fatalf("expected 1 device, got %d: %#v", len(got), got)
	}
	if got[0]["host"] != "10.0.0.10" {
		t.Errorf("wrong host: %#v", got[0])
	}
	if got[0]["ssh_profile_name"] != "creds-pool" {
		t.Errorf("wrong ssh_profile_name: %#v", got[0])
	}
}

func TestProfileDevices_POST_DuplicateHost_Returns409(t *testing.T) {
	svc := newProfileDeviceServiceStub()
	svc.profiles["lab"] = &entities.ScanProfile{Name: "lab", Kind: "device", Host: "10.0.0.1"}
	r := setupDeviceRouter(svc)
	body, _ := json.Marshal(map[string]string{"host": "10.0.0.10"})
	for i := range 2 {
		req := httptest.NewRequest(http.MethodPost, "/scan/profiles/lab/devices", bytes.NewReader(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		switch i {
		case 0:
			if w.Code != http.StatusCreated {
				t.Fatalf("first POST expected 201, got %d: %s", w.Code, w.Body.String())
			}
		case 1:
			if w.Code != http.StatusConflict {
				t.Fatalf("second POST expected 409, got %d: %s", w.Code, w.Body.String())
			}
		}
	}
}

func TestProfileDevices_DELETE(t *testing.T) {
	svc := newProfileDeviceServiceStub()
	svc.profiles["lab"] = &entities.ScanProfile{Name: "lab", Kind: "device", Host: "10.0.0.1"}
	svc.devices["lab"] = []entities.ProfileDevice{
		{ProfileName: "lab", Host: "10.0.0.10"},
		{ProfileName: "lab", Host: "10.0.0.11"},
	}
	r := setupDeviceRouter(svc)
	req := httptest.NewRequest(http.MethodDelete, "/scan/profiles/lab/devices?host=10.0.0.10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if len(svc.devices["lab"]) != 1 || svc.devices["lab"][0].Host != "10.0.0.11" {
		t.Fatalf("wrong devices left: %#v", svc.devices["lab"])
	}
}

func TestProfileDevices_DELETE_MissingHost(t *testing.T) {
	svc := newProfileDeviceServiceStub()
	svc.profiles["lab"] = &entities.ScanProfile{Name: "lab", Kind: "device"}
	r := setupDeviceRouter(svc)
	req := httptest.NewRequest(http.MethodDelete, "/scan/profiles/lab/devices", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}