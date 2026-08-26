// SPDX-License-Identifier: MIT
// Package lldp wraps OPNsense /api/lldp/interface endpoints.
package lldp

import (
	"context"

	"github.com/t34/opnsense-api/opnsense"
)

// InterfaceModule provides per-interface LLDP configuration.
type InterfaceModule struct {
	c *opnsense.Client
}

func NewInterface(c *opnsense.Client) *InterfaceModule { return &InterfaceModule{c: c} }

// InterfaceSearch calls GET /api/lldp/interface/search_item.
func (m *InterfaceModule) InterfaceSearch(ctx context.Context) (*opnsense.GenericResponse, error) {
	return m.c.Do(ctx, opnsense.Request{
		Method: "GET",
		Path:   "/api/lldp/interface/search_item",
	})
}
