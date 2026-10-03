// SPDX-License-Identifier: AGPL-3.0-or-later
// Package lldp wraps OPNsense /api/lldp/service endpoints.
package lldp

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/t34/opnsense-api/opnsense"
)

type Module struct {
	c *opnsense.Client
}

func New(c *opnsense.Client) *Module { return &Module{c: c} }

// LLDPServiceSettings is the payload for GET/POST /api/lldp/service.
type LLDPServiceSettings struct {
	Enabled bool `json:"enabled"` // "1"/"0" marshalled by Client
}

// ServiceGet calls GET /api/lldp/service/get.
func (m *Module) ServiceGet(ctx context.Context) (*opnsense.GenericResponse, error) {
	return m.c.Do(ctx, opnsense.Request{
		Method: "GET",
		Path:   "/api/lldp/service/get",
	})
}

// ServiceSet calls POST /api/lldp/service/set.
func (m *Module) ServiceSet(ctx context.Context, settings LLDPServiceSettings) error {
	body, err := json.Marshal(map[string]any{"general": settings})
	if err != nil {
		return err
	}
	_, err = m.c.Do(ctx, opnsense.Request{
		Method: "POST",
		Path:   "/api/lldp/service/set",
		Body:   strings.NewReader(string(body)),
	})
	return err
}

// ServiceReconfigure calls POST /api/lldp/service/reconfigure.
func (m *Module) ServiceReconfigure(ctx context.Context) error {
	_, err := m.c.Do(ctx, opnsense.Request{
		Method: "POST",
		Path:   "/api/lldp/service/reconfigure",
	})
	return err
}
