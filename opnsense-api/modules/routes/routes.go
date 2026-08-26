// SPDX-License-Identifier: MIT
// Package routes wraps the OPNsense /api/routes/routes/* endpoints.
//
// The real OPNsense endpoint is /api/routes/routes/addroute (singular verb),
// not /api/routes/routes/add_item — see
// https://docs.opnsense.org/development/api/core/routes.html and the forum
// thread that nailed down the JSON shape: gateway is a *name string*
// referencing /api/routing/settings/searchGateway, not a nested object.
package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/t34/opnsense-api/opnsense"
)

// Module is the typed entry point for the routes module.
type Module struct {
	c *opnsense.Client
}

func New(c *opnsense.Client) *Module { return &Module{c: c} }

// RouteAdd is the typed payload for POST /api/routes/routes/addroute.
//
// Network is the destination CIDR ("10.0.50.0/24"). Gateway is the gateway
// *name* (configured under System → Gateways → Single), not an IP address.
// The wrapper resolves gateway-name → IP at apply-time via the routing module.
type RouteAdd struct {
	Network  string `json:"network"`
	Gateway  string `json:"gateway"`
	Descr    string `json:"descr,omitempty"`
	Disabled string `json:"disabled,omitempty"` // "0" | "1"
}

// RouteAdd calls POST /api/routes/routes/addroute. The OPNsense API returns
// 200 OK with {"result":"saved"} on success but does not include the new
// UUID; the caller re-fetches via SearchRoute to obtain one when needed
// (e.g. for subsequent set/del).
func (m *Module) RouteAdd(ctx context.Context, r RouteAdd) error {
	body, err := json.Marshal(map[string]any{"route": r})
	if err != nil {
		return err
	}
	resp, err := m.c.Do(ctx, opnsense.Request{
		Method: "POST", Path: "/api/routes/routes/addroute",
		Body:   strings.NewReader(string(body)),
	})
	if err != nil {
		return err
	}
	if !routeSuccess(resp) {
		return fmt.Errorf("opnsense: %s: status=%q msg=%q",
			"/api/routes/routes/addroute", resp.Status, resp.Msg)
	}
	return nil
}

// RouteDel calls POST /api/routes/routes/delroute/{uuid}.
func (m *Module) RouteDel(ctx context.Context, uuid string) error {
	resp, err := m.c.Do(ctx, opnsense.Request{
		Method: "POST", Path: "/api/routes/routes/delroute/" + uuid,
	})
	if err != nil {
		return err
	}
	if !routeSuccess(resp) {
		return fmt.Errorf("opnsense: %s: status=%q msg=%q",
			"/api/routes/routes/delroute", resp.Status, resp.Msg)
	}
	return nil
}

// RouteApply calls POST /api/routes/routes/reconfigure, activating pending
// route changes. The equivalent of OpenWrt's `uci commit network`.
func (m *Module) RouteApply(ctx context.Context) error {
	resp, err := m.c.Do(ctx, opnsense.Request{
		Method: "POST", Path: "/api/routes/routes/reconfigure",
	})
	if err != nil {
		return err
	}
	if !routeSuccess(resp) {
		return fmt.Errorf("opnsense: %s: status=%q msg=%q",
			"/api/routes/routes/reconfigure", resp.Status, resp.Msg)
	}
	return nil
}

// SearchRoute calls GET /api/routes/routes/searchroute and returns the raw
// rows for caller-side decoding. Each row carries the route UUID under "uuid"
// plus the same fields as RouteAdd (network, gateway, descr, disabled).
func (m *Module) SearchRoute(ctx context.Context) (*opnsense.GenericResponse, error) {
	return m.c.Do(ctx, opnsense.Request{
		Method: "GET", Path: "/api/routes/routes/searchroute",
	})
}

// routeSuccess reports whether a parsed response carries a canonical OPNsense
// success status. The API is permissive — 200 OK with status="failed" is a
// real failure that the caller must see. 4xx/5xx errors are surfaced by
// Client.Do itself via the retry layer.
func routeSuccess(r *opnsense.GenericResponse) bool {
	switch r.Status {
	case "ok", "saved", "deleted":
		return true
	}
	return false
}