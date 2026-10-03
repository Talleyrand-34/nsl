// SPDX-License-Identifier: AGPL-3.0-or-later
package system

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
		assert.Equal(t, "/api/system/general/get", r.URL.Path)
		resp := `{"general":{"banner":"Welcome to OPNsense\n---\n"}}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	result, err := m.GeneralGet(context.Background())
	require.NoError(t, err)
	var wrapper struct {
		General struct {
			Banner string `json:"banner"`
		} `json:"general"`
	}
	require.NoError(t, json.Unmarshal(result.RawBody, &wrapper))
	assert.Equal(t, "Welcome to OPNsense\n---\n", wrapper.General.Banner)
}

func TestGeneralSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/system/general/set", r.URL.Path)
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
	err := m.GeneralSet(context.Background(), "Welcome to OPNsense\n")
	require.NoError(t, err)
}
