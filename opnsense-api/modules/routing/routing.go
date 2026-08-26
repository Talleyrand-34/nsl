// SPDX-License-Identifier: MIT
// Package routing wraps the OPNsense /api/routing/settings/* endpoints.
//
// The push pipeline needs gateway-name → IP resolution to build the route
// payload. The routes module accepts gateway *names*, not IPs, so callers
// fetch the gateway list and look the name up before calling RouteAdd.
package routing

import (
	"context"
	"encoding/json"

	"github.com/t34/opnsense-api/opnsense"
)

// Module is the typed entry point for the routing settings module.
type Module struct {
	c *opnsense.Client
}

func New(c *opnsense.Client) *Module { return &Module{c: c} }

// Gateway describes one gateway as returned by /api/routing/settings/searchGateway.
// Only the fields the push pipeline needs are decoded; the API carries more
// (monitor, default route, etc.) that are ignored here.
type Gateway struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Address     string `json:"address"`     // "10.0.0.1"
	GatewayType string `json:"gateway_type"` // "ipv4" | "ipv6"
	Description string `json:"descr"`
}

// SearchGateway calls GET /api/routing/settings/searchGateway and returns
// the parsed rows. The API returns {"rows":[...]} with one entry per
// configured gateway.
func (m *Module) SearchGateway(ctx context.Context) ([]Gateway, error) {
	resp, err := m.c.Do(ctx, opnsense.Request{
		Method: "GET", Path: "/api/routing/settings/searchGateway",
	})
	if err != nil {
		return nil, err
	}
	var env struct {
		Rows []Gateway `json:"rows"`
	}
	if err := json.Unmarshal(resp.RawBody, &env); err != nil {
		return nil, err
	}
	return env.Rows, nil
}