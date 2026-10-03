// SPDX-License-Identifier: AGPL-3.0-or-later
package ntp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t34/opnsense-api/opnsense"
)

func TestNTPGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/ntp/settings/get", r.URL.Path)
		resp := `{"general":{"enable":"1","timeservers":["0.pool.ntp.org","1.pool.ntp.org"],"timezone":"Europe/Madrid"}}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	result, err := m.NTPGet(context.Background())
	require.NoError(t, err)
	var wrapper struct {
		General struct {
			Enable      string   `json:"enable"`
			Timeservers []string `json:"timeservers"`
			Timezone    string   `json:"timezone"`
		} `json:"general"`
	}
	require.NoError(t, json.Unmarshal(result.RawBody, &wrapper))
	assert.Equal(t, "1", wrapper.General.Enable)
	assert.Equal(t, "Europe/Madrid", wrapper.General.Timezone)
}

func TestNTPSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/ntp/settings/set", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		// Verify the JSON contains the expected fields under "general"
		var payload map[string]map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(body, &payload))
		_, hasGeneral := payload["general"]
		assert.True(t, hasGeneral, "payload should have 'general' key: %s", string(body))
		resp := `{"result":"saved"}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	err := m.NTPSet(context.Background(), NTPSettings{
		Enable:      "1",
		Timeservers: []string{"0.pool.ntp.org"},
		Timezone:    "UTC",
	})
	require.NoError(t, err)
}
