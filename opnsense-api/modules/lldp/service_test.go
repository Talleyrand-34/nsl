// SPDX-License-Identifier: AGPL-3.0-or-later
package lldp

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

func TestServiceGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/lldp/service/get", r.URL.Path)
		_, _ = w.Write([]byte(`{"general":{"enabled":"1"}}`))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	resp, err := m.ServiceGet(context.Background())
	require.NoError(t, err)
	assert.Contains(t, string(resp.RawBody), `"enabled"`)
}

func TestServiceSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/lldp/service/set", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"enabled"`)
		_, _ = w.Write([]byte(`{"result":"saved"}`))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	err := m.ServiceSet(context.Background(), LLDPServiceSettings{Enabled: true})
	require.NoError(t, err)
}

func TestServiceReconfigure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/lldp/service/reconfigure", r.URL.Path)
		_, _ = w.Write([]byte(`{"result":"ok"}`))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	err := m.ServiceReconfigure(context.Background())
	require.NoError(t, err)
}
