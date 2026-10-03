// SPDX-License-Identifier: AGPL-3.0-or-later
// Package ntp wraps the OPNsense /api/ntp/settings endpoints.
package ntp

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

// NTPSettings is the payload for GET /api/ntp/settings/get and
// POST /api/ntp/settings/set. OPNsense nests everything under "general".
type NTPSettings struct {
	Enable      string   `json:"enable"`   // "1" or "0"
	Timeservers []string `json:"timeservers"`
	Timezone    string   `json:"timezone"`
}

// NTPGet calls GET /api/ntp/settings/get and returns the raw JSON for caller-side decoding.
func (m *Module) NTPGet(ctx context.Context) (*opnsense.GenericResponse, error) {
	return m.c.Do(ctx, opnsense.Request{
		Method: "GET",
		Path:   "/api/ntp/settings/get",
	})
}

// NTPSet calls POST /api/ntp/settings/set with the given payload.
// OPNsense requires the fields nested under a "general" key, so we marshal manually.
func (m *Module) NTPSet(ctx context.Context, settings NTPSettings) error {
	body, err := json.Marshal(map[string]any{"general": settings})
	if err != nil {
		return err
	}
	_, err = m.c.Do(ctx, opnsense.Request{
		Method: "POST",
		Path:   "/api/ntp/settings/set",
		Body:   strings.NewReader(string(body)),
	})
	return err
}
