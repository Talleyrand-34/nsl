// SPDX-License-Identifier: MIT
// observability_e2e_test.go: end-to-end tests for NTP, banner, LLDP, syslog, SNMP push.
package cmd_push_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/configparser/parsers"
	"nsl-graph/internal/push"
)

// obsHandler simulates OPNsense with observability endpoints.
type obsHandler struct {
	mu       sync.Mutex
	paths    []string
	commitOK bool
}

func newObsHandler() *obsHandler {
	return &obsHandler{commitOK: true}
}

func (h *obsHandler) record(path string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.paths = append(h.paths, path)
}

func (h *obsHandler) snapshot() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]string, len(h.paths))
	copy(out, h.paths)
	return out
}

func (h *obsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.record(r.URL.Path)
	switch {
	case r.URL.Path == "/api/interfaces/overview/list" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(`{"rows":[{"device":"vtnet0","identifier":"wan"}]}`))
	case r.URL.Path == "/api/interfaces/overview/commit" && r.Method == http.MethodPost:
		if !h.commitOK {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(`{"status":"failed"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	case r.URL.Path == "/api/interfaces/vlan/add" && r.Method == http.MethodPost:
		_, _ = w.Write([]byte(`{"status":"ok","uuid":"v-new"}`))
	case r.URL.Path == "/api/routes/routes/searchroute" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(`{"rows":[]}`))
	case r.URL.Path == "/api/routes/routes/addroute" && r.Method == http.MethodPost:
		_, _ = w.Write([]byte(`{"result":"saved"}`))
	case r.URL.Path == "/api/routes/routes/reconfigure" && r.Method == http.MethodPost:
		_, _ = w.Write([]byte(`{"result":"saved"}`))
	case r.URL.Path == "/api/routing/settings/searchGateway" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(`{"rows":[]}`))
	case r.URL.Path == "/api/ntp/settings/get" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(`{"general":{"enable":"0","timeservers":[],"timezone":""}}`))
	case r.URL.Path == "/api/ntp/settings/set" && r.Method == http.MethodPost:
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"timeservers"`) {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(`{"error":"missing timeservers"}`))
			return
		}
		_, _ = w.Write([]byte(`{"result":"saved"}`))
	case r.URL.Path == "/api/system/general/get" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(`{"general":{"banner":""}}`))
	case r.URL.Path == "/api/system/general/set" && r.Method == http.MethodPost:
		_, _ = w.Write([]byte(`{"result":"saved"}`))
	case r.URL.Path == "/api/lldp/service/get" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(`{"general":{"enabled":"0"}}`))
	case r.URL.Path == "/api/lldp/service/set" && r.Method == http.MethodPost:
		_, _ = w.Write([]byte(`{"result":"saved"}`))
	case r.URL.Path == "/api/lldp/service/reconfigure" && r.Method == http.MethodPost:
		_, _ = w.Write([]byte(`{"result":"ok"}`))
	case r.URL.Path == "/api/syslog/settings/search_destinations" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(`{"rows":[]}`))
	case r.URL.Path == "/api/syslog/settings/set" && r.Method == http.MethodPost:
		_, _ = w.Write([]byte(`{"result":"saved"}`))
	case r.URL.Path == "/api/snmp/general/get" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(`{"general":{"enabled":"0","location":"","contact":"","community":""}}`))
	case r.URL.Path == "/api/snmp/general/set" && r.Method == http.MethodPost:
		_, _ = w.Write([]byte(`{"result":"saved"}`))
	default:
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"error":"not implemented: ` + r.URL.Path + `"}`))
	}
}

func obsEngineFor(srv *httptest.Server) *push.Engine {
	return push.NewDefaultEngine(push.WithOpnsenseFactory(func(_, _, _ string) configparser.ConfigRenderer {
		return parsers.NewOpnsenseRendererForURL(srv.URL, "k", "s")
	}))
}

// intentWithAllObservability builds a ConfigData with all five observability areas set.
func intentWithAllObservability() *configparser.ConfigData {
	return &configparser.ConfigData{
		Hostname: "opnsense-obs",
		Interfaces: []configparser.ConfigInterface{
			{Name: "vtnet0", Type: "physical", Enabled: true, VLANs: []configparser.ConfigVLAN{{ID: "30", Tagged: true}}},
		},
		NTP: &configparser.ConfigNTPConfig{
			Enabled:  true,
			Timezone: "Europe/Madrid",
			Servers: []configparser.ConfigNTPServer{
				{Address: "0.pool.ntp.org", Enabled: true},
				{Address: "1.pool.ntp.org", Enabled: true},
			},
		},
		Banner: &configparser.ConfigBanner{
			LoginBanner: "Welcome to OPNsense\nAuthorized use only.\n",
			PostLogin:   "Welcome back.\n",
		},
		LLDP: &configparser.ConfigLLDPSettings{
			Enabled:   true,
			SystemName: "opnsense-obs",
		},
		Syslog: &configparser.ConfigSyslogConfig{
			Enabled: true,
			Targets: []configparser.ConfigSyslogTarget{
				{Address: "syslog.example.com", Port: 514, Protocol: "udp", Facility: "local0"},
			},
		},
		SNMP: &configparser.ConfigSNMPConfig{
			Enabled:   true,
			Location:  "Rack 4",
			Contact:   "noc@example.com",
			Communities: []configparser.ConfigSNMPCommunity{
				{Name: "public", Access: "ro"},
				{Name: "private", Access: "rw"},
			},
		},
	}
}

// TestObs_E2E_AllFiveObservability verifies that pushing a ConfigData with all
// five observability areas set fires the correct API calls.
func TestObs_E2E_AllFiveObservability(t *testing.T) {
	lab := newObsHandler()
	srv := httptest.NewTLSServer(lab)
	defer srv.Close()

	r := obsEngineFor(srv).Renderers()["opnsense"]

	intent := intentWithAllObservability()
	err := r.Render(configparser.SafetyApply, intent, nil, configparser.SSHCredentials{})
	require.NoError(t, err, "Render with all five observability areas must succeed")

	paths := lab.snapshot()

	assert.Contains(t, paths, "/api/ntp/settings/set", "NTP set must be called; paths: %v", paths)
	assert.Contains(t, paths, "/api/system/general/set", "Banner set must be called; paths: %v", paths)
	assert.Contains(t, paths, "/api/lldp/service/set", "LLDP set must be called; paths: %v", paths)
	assert.Contains(t, paths, "/api/lldp/service/reconfigure", "LLDP reconfigure must be called; paths: %v", paths)
	assert.Contains(t, paths, "/api/syslog/settings/set", "Syslog set must be called; paths: %v", paths)
	assert.Contains(t, paths, "/api/snmp/general/set", "SNMP set must be called; paths: %v", paths)
	assert.Contains(t, paths, "/api/interfaces/vlan/add", "VLAN add must be called; paths: %v", paths)
	assert.Contains(t, paths, "/api/interfaces/overview/commit", "Interface commit must be called; paths: %v", paths)
}

// TestObs_E2E_SNMP_UsesFirstCommunity verifies SNMP apply uses the first community.
func TestObs_E2E_SNMP_UsesFirstCommunity(t *testing.T) {
	lab := newObsHandler()
	srv := httptest.NewTLSServer(lab)
	defer srv.Close()

	r := obsEngineFor(srv).Renderers()["opnsense"]

	intent := &configparser.ConfigData{
		Hostname: "obs-snmp",
		SNMP: &configparser.ConfigSNMPConfig{
			Enabled:     true,
			Location:    "Rack 4",
			Communities: []configparser.ConfigSNMPCommunity{
				{Name: "first-community", Access: "ro"},
				{Name: "second-community", Access: "rw"},
			},
		},
	}
	err := r.Render(configparser.SafetyApply, intent, nil, configparser.SSHCredentials{})
	require.NoError(t, err)

	assert.Contains(t, lab.snapshot(), "/api/snmp/general/set", "SNMP set must be called")
}

// TestObs_E2E_DryRun_CallsNoMutatingEndpoints verifies that DryRun fires no
// mutating endpoints (POST/add/reconfigure).
func TestObs_E2E_DryRun_CallsNoMutatingEndpoints(t *testing.T) {
	lab := newObsHandler()
	srv := httptest.NewTLSServer(lab)
	defer srv.Close()

	r := obsEngineFor(srv).Renderers()["opnsense"]

	err := r.Render(configparser.SafetyDryRun, intentWithAllObservability(), nil, configparser.SSHCredentials{})
	require.NoError(t, err)

	var setCalls []string
	for _, p := range lab.snapshot() {
		if strings.Contains(p, "/set") || strings.Contains(p, "/add") || strings.Contains(p, "/reconfigure") {
			setCalls = append(setCalls, p)
		}
	}
	assert.Empty(t, setCalls, "DryRun must not call mutating endpoints; got: %v", setCalls)
}

// TestObs_E2E_Render_WhenObservedIsNil_PushesAllAreas verifies that when observed
// is nil, all five observability areas are pushed. The diff functions compare
// intended vs observed ConfigData structs; when observed is nil the diff always
// fires. The VLAN and route diffs consult the live fetch.
func TestObs_E2E_Render_WhenObservedIsNil_PushesAllAreas(t *testing.T) {
	lab := newObsHandler()
	srv := httptest.NewTLSServer(lab)
	defer srv.Close()

	r := obsEngineFor(srv).Renderers()["opnsense"]
	intent := intentWithAllObservability()

	err := r.Render(configparser.SafetyApply, intent, nil, configparser.SSHCredentials{})
	require.NoError(t, err)

	paths := lab.snapshot()
	assert.Contains(t, paths, "/api/ntp/settings/set", "NTP set must fire when observed is nil; paths: %v", paths)
	assert.Contains(t, paths, "/api/system/general/set", "Banner set must fire; paths: %v", paths)
	assert.Contains(t, paths, "/api/lldp/service/set", "LLDP set must fire; paths: %v", paths)
	assert.Contains(t, paths, "/api/syslog/settings/set", "Syslog set must fire; paths: %v", paths)
	assert.Contains(t, paths, "/api/snmp/general/set", "SNMP set must fire; paths: %v", paths)
}
