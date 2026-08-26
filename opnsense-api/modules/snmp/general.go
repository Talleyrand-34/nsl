// SPDX-License-Identifier: MIT
// Package snmp wraps OPNsense /api/snmp/general endpoints.
package snmp

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

// GeneralSettings is the payload for GET/POST /api/snmp/general.
type GeneralSettings struct {
	Enabled     bool   `json:"enabled"`
	Location    string `json:"location"`
	Contact     string `json:"contact"`
	Community   string `json:"community"` // primary community string
	BindTo      string `json:"bind_to_interface,omitempty"`
}

// GeneralGet calls GET /api/snmp/general/get.
func (m *Module) GeneralGet(ctx context.Context) (*opnsense.GenericResponse, error) {
	return m.c.Do(ctx, opnsense.Request{
		Method: "GET",
		Path:   "/api/snmp/general/get",
	})
}

// GeneralSet calls POST /api/snmp/general/set with the given settings.
// OPNsense nests settings under a "general" key.
func (m *Module) GeneralSet(ctx context.Context, settings GeneralSettings) error {
	payload := map[string]any{"general": settings}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = m.c.Do(ctx, opnsense.Request{
		Method: "POST",
		Path:   "/api/snmp/general/set",
		Body:   strings.NewReader(string(body)),
	})
	return err
}
