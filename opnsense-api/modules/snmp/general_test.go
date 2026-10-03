// SPDX-License-Identifier: AGPL-3.0-or-later
package snmp

import (
	"context"
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
		assert.Equal(t, "/api/snmp/general/get", r.URL.Path)
		_, _ = w.Write([]byte(`{"general":{"enabled":"1","location":"Rack 4","contact":"noc@example.com","community":"public"}}`))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	resp, err := m.GeneralGet(context.Background())
	require.NoError(t, err)
	assert.Contains(t, string(resp.RawBody), `"location"`)
}

func TestGeneralSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/snmp/general/set", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"general"`)
		_, _ = w.Write([]byte(`{"result":"saved"}`))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	err := m.GeneralSet(context.Background(), GeneralSettings{
		Enabled:   true,
		Location:  "Rack 4",
		Community: "public",
	})
	require.NoError(t, err)
}
