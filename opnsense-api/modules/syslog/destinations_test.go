// SPDX-License-Identifier: AGPL-3.0-or-later
package syslog

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t34/opnsense-api/opnsense"
)

func TestSearchDestinations(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/syslog/settings/search_destinations", r.URL.Path)
		resp := `{"rows":[{"uuid":"abc","address":"10.0.0.1","port":"514","transport":"udp","facility":"local0","program":"","level":"","certificate":"","enabled":"1"}]}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	dests, err := m.SearchDestinations(context.Background())
	require.NoError(t, err)
	require.Len(t, dests, 1)
	assert.Equal(t, "abc", dests[0].UUID)
	assert.Equal(t, "10.0.0.1", dests[0].Address)
	assert.Equal(t, "514", dests[0].Port)
	assert.Equal(t, "udp", dests[0].Transport)
}

func TestAddDestination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/syslog/settings/add_destination", r.URL.Path)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		params, err := url.ParseQuery(string(body))
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.1", params.Get("address"))
		assert.Equal(t, "514", params.Get("port"))
		assert.Equal(t, "udp", params.Get("transport"))
		resp := `{"result":"saved","uuid":"new-uuid"}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	uuid, err := m.AddDestination(context.Background(), Destination{
		Address:   "10.0.0.1",
		Port:      "514",
		Transport: "udp",
		Facility:  "local0",
		Enabled:   "1",
	})
	require.NoError(t, err)
	assert.Equal(t, "new-uuid", uuid)
}

func TestDelDestination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/syslog/settings/del_destination/uuid-to-delete", r.URL.Path)
		resp := `{"result":"deleted"}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	err := m.DelDestination(context.Background(), "uuid-to-delete")
	require.NoError(t, err)
}

func TestSetDestination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/syslog/settings/set", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		general, ok := payload["general"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "1", general["enabled"])
		assert.Equal(t, "0", general["preservefqdn"])
		resp := `{"result":"saved"}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	err := m.SetDestination(context.Background(), true, false, nil)
	require.NoError(t, err)
}
