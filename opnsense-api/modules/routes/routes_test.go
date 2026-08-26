// SPDX-License-Identifier: MIT
package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t34/opnsense-api/opnsense"
)

func TestRouteAdd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/routes/routes/add_item", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))
		resp := `{"result":"saved","uuid":"route-abc123"}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	uuid, err := m.RouteAdd(context.Background(), RouteAdd{
		Network:   "192.168.5.0/24",
		Gateway:   "10.0.0.1",
		Interface: "wan",
		Descr:     "nsl-graph push",
	})
	require.NoError(t, err)
	assert.Equal(t, "route-abc123", uuid)
}

func TestRouteDel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/routes/routes/del_item/route-abc123", r.URL.Path)
		resp := `{"result":"deleted"}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	err := m.RouteDel(context.Background(), "route-abc123")
	require.NoError(t, err)
}

func TestRouteSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/routes/routes/search_item", r.URL.Path)
		// OPNsense wraps rows in a response object: {"rows": [...]}
		resp := `{"rows":[{"uuid":"r1","network":"10.0.0.0/8","gateway":"10.0.0.1"}]}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	result, err := m.RouteSearch(context.Background())
	require.NoError(t, err)
	// Decode the wrapper to extract rows.
	var wrapper struct {
		Rows []map[string]interface{} `json:"rows"`
	}
	require.NoError(t, json.Unmarshal(result.RawBody, &wrapper))
	assert.Equal(t, "r1", wrapper.Rows[0]["uuid"])
}

func TestRouteReconfigure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/routes/routes/reconfigure", r.URL.Path)
		resp := `{"result":"ok"}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := opnsense.NewClient(srv.URL, "key", "secret")
	m := New(c)
	err := m.RouteReconfigure(context.Background())
	require.NoError(t, err)
}
