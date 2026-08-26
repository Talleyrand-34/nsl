// SPDX-License-Identifier: MIT
// Package interfaces wraps the OPNsense /api/interfaces/* endpoints.
package interfaces

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/t34/opnsense-api/opnsense"
)

// Module is the typed entry point for the interfaces module.
type Module struct {
	c *opnsense.Client
}

func New(c *opnsense.Client) *Module { return &Module{c: c} }

// VLANAdd is the typed payload for POST /api/interfaces/vlan/add.
type VLANAdd struct {
	If    string `json:"if"`
	Tag   int    `json:"tag"`
	PCP   int    `json:"pcp,omitempty"`
	Descr string `json:"descr,omitempty"`
}

// VLANAdd calls POST /api/interfaces/vlan/add and returns the new VLAN's UUID.
func (m *Module) VLANAdd(ctx context.Context, v VLANAdd) (string, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	resp, err := m.c.Do(ctx, opnsense.Request{
		Method: "POST", Path: "/api/interfaces/vlan/add",
		Body: strings.NewReader(string(body)),
	})
	if err != nil {
		return "", err
	}
	return resp.UUID, nil
}

// VLANDel calls POST /api/interfaces/vlan/del/{uuid}.
func (m *Module) VLANDel(ctx context.Context, uuid string) error {
	_, err := m.c.Do(ctx, opnsense.Request{
		Method: "POST", Path: "/api/interfaces/vlan/del/" + uuid,
	})
	return err
}

// OverviewCommit calls POST /api/interfaces/overview/commit, applying pending
// interface changes. Equivalent to OpenWrt `uci commit`.
func (m *Module) OverviewCommit(ctx context.Context) error {
	_, err := m.c.Do(ctx, opnsense.Request{
		Method: "POST", Path: "/api/interfaces/overview/commit",
	})
	return err
}

// OverviewList calls GET /api/interfaces/overview/list and returns the raw
// JSON for caller-side decoding.
func (m *Module) OverviewList(ctx context.Context) (*opnsense.GenericResponse, error) {
	return m.c.Do(ctx, opnsense.Request{
		Method: "GET", Path: "/api/interfaces/overview/list",
	})
}
