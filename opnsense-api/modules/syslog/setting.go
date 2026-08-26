// SPDX-License-Identifier: MIT
package syslog

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/t34/opnsense-api/opnsense"
)

// GeneralSettings is the payload for GET /api/syslog/settings/get and
// POST /api/syslog/settings/set (general section).
type GeneralSettings struct {
	Enabled      string `json:"enabled"`
	PreserveFQDN string `json:"preservefqdn"`
}

// GeneralGet calls GET /api/syslog/settings/get and returns the general settings.
func (m *Module) GeneralGet(ctx context.Context) (*GeneralSettings, error) {
	resp, err := m.c.Do(ctx, opnsense.Request{
		Method: "GET",
		Path:   "/api/syslog/settings/get",
	})
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		General GeneralSettings `json:"general"`
	}
	if err := json.Unmarshal(resp.RawBody, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.General, nil
}

// GeneralSet calls POST /api/syslog/settings/set with the general settings.
// OPNsense nests these under a "general" key.
func (m *Module) GeneralSet(ctx context.Context, settings GeneralSettings) error {
	body, err := json.Marshal(map[string]any{"general": settings})
	if err != nil {
		return err
	}
	_, err = m.c.Do(ctx, opnsense.Request{
		Method: "POST",
		Path:   "/api/syslog/settings/set",
		Body:   strings.NewReader(string(body)),
	})
	return err
}
