// SPDX-License-Identifier: AGPL-3.0-or-later
// Package system wraps OPNsense /api/system/general endpoints.
package system

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

// BannerResponse is the GET /api/system/general/get response.
type BannerResponse struct {
	Banner string `json:"banner"`
}

// GeneralGet calls GET /api/system/general/get and returns the raw JSON.
func (m *Module) GeneralGet(ctx context.Context) (*opnsense.GenericResponse, error) {
	return m.c.Do(ctx, opnsense.Request{
		Method: "GET",
		Path:   "/api/system/general/get",
	})
}

// GeneralSet calls POST /api/system/general/set with the given login banner.
// OPNsense stores the banner under a "general" wrapper key.
func (m *Module) GeneralSet(ctx context.Context, banner string) error {
	payload := map[string]any{"general": map[string]any{"banner": banner}}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = m.c.Do(ctx, opnsense.Request{
		Method: "POST",
		Path:   "/api/system/general/set",
		Body:   strings.NewReader(string(body)),
	})
	return err
}
