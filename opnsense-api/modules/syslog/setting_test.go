// SPDX-License-Identifier: MIT
package syslog

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

func TestGeneralGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/syslog/settings/get", r.URL.Path)
		resp := `{"general":{"enabled":"1","preservefqdn":"1"}}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	result, err := m.GeneralGet(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "1", result.Enabled)
	assert.Equal(t, "1", result.PreserveFQDN)
}

func TestGeneralSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/syslog/settings/set", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
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
	err := m.GeneralSet(context.Background(), GeneralSettings{
		Enabled:      "1",
		PreserveFQDN: "1",
	})
	require.NoError(t, err)
}
