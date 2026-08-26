// SPDX-License-Identifier: MIT
// Package routes wraps the OPNsense /api/routes/routes/* endpoints.
package routes

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/t34/opnsense-api/opnsense"
)

// Module is the typed entry point for the routes module.
type Module struct {
	c *opnsense.Client
}

// New returns a new routes module.
func New(c *opnsense.Client) *Module { return &Module{c: c} }

// RouteAdd is the typed payload for POST /api/routes/routes/add_item.
type RouteAdd struct {
	Network   string `json:"network"`
	Gateway   string `json:"gateway"`
	Interface string `json:"interface"`
	Descr     string `json:"descr,omitempty"`
	Disabled  string `json:"disabled,omitempty"`
}

// RouteAdd calls POST /api/routes/routes/add_item and returns the new route's UUID.
func (m *Module) RouteAdd(ctx context.Context, r RouteAdd) (string, error) {
	body, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	resp, err := m.c.Do(ctx, opnsense.Request{
		Method: "POST", Path: "/api/routes/routes/add_item",
		Body: strings.NewReader(string(body)),
	})
	if err != nil {
		return "", err
	}
	return resp.UUID, nil
}

// RouteDel calls POST /api/routes/routes/del_item/{uuid}.
func (m *Module) RouteDel(ctx context.Context, uuid string) error {
	_, err := m.c.Do(ctx, opnsense.Request{
		Method: "POST", Path: "/api/routes/routes/del_item/" + uuid,
	})
	return err
}

// RouteSearch calls GET /api/routes/routes/search_item.
func (m *Module) RouteSearch(ctx context.Context) (*opnsense.GenericResponse, error) {
	return m.c.Do(ctx, opnsense.Request{
		Method: "GET", Path: "/api/routes/routes/search_item",
	})
}

// RouteReconfigure calls POST /api/routes/routes/reconfigure.
func (m *Module) RouteReconfigure(ctx context.Context) error {
	_, err := m.c.Do(ctx, opnsense.Request{
		Method: "POST", Path: "/api/routes/routes/reconfigure",
	})
	return err
}
