// SPDX-License-Identifier: MIT
// Package firewall wraps the OPNsense /api/firewall/* endpoints.
// Phase 1: stub only — interfaces and routes are wired.
package firewall

import "github.com/t34/opnsense-api/opnsense"

type Module struct {
	c *opnsense.Client
}

func New(c *opnsense.Client) *Module { return &Module{c: c} }
