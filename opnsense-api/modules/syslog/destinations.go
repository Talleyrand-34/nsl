// SPDX-License-Identifier: AGPL-3.0-or-later
// Package syslog wraps the OPNsense /api/syslog/settings endpoints.
package syslog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/t34/opnsense-api/opnsense"
)

// Module wraps the OPNsense syslog destinations API.
type Module struct {
	c *opnsense.Client
}

// New returns a new syslog module.
func New(c *opnsense.Client) *Module { return &Module{c: c} }

// Destination represents one remote syslog destination in OPNsense.
type Destination struct {
	UUID        string `json:"uuid"`
	Address     string `json:"address"`
	Port        string `json:"port"`
	Transport   string `json:"transport"` // udp or tcp
	Facility    string `json:"facility"`
	Program     string `json:"program"`
	Level       string `json:"level"`
	Certificate string `json:"certificate"`
	Enabled     string `json:"enabled"`
}

// SearchDestinations calls GET /api/syslog/settings/search_destinations
// and returns the list of configured destinations.
func (m *Module) SearchDestinations(ctx context.Context) ([]Destination, error) {
	resp, err := m.c.Do(ctx, opnsense.Request{
		Method: "GET",
		Path:   "/api/syslog/settings/search_destinations",
	})
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Rows []Destination `json:"rows"`
	}
	if err := json.Unmarshal(resp.RawBody, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Rows, nil
}

// AddDestination calls POST /api/syslog/settings/add_destination with the
// destination fields as form values.
func (m *Module) AddDestination(ctx context.Context, dest Destination) (string, error) {
	form := url.Values{}
	form.Set("address", dest.Address)
	form.Set("port", dest.Port)
	form.Set("transport", dest.Transport)
	form.Set("facility", dest.Facility)
	form.Set("program", dest.Program)
	form.Set("level", dest.Level)
	form.Set("certificate", dest.Certificate)
	form.Set("enabled", dest.Enabled)

	resp, err := m.c.Do(ctx, opnsense.Request{
		Method:      http.MethodPost,
		Path:        "/api/syslog/settings/add_destination",
		ContentType: "application/x-www-form-urlencoded",
		Body:        strings.NewReader(form.Encode()),
	})
	if err != nil {
		return "", err
	}
	var result struct {
		Result string `json:"result"`
		UUID   string `json:"uuid"`
	}
	if err := json.Unmarshal(resp.RawBody, &result); err != nil {
		return "", err
	}
	return result.UUID, nil
}

// DelDestination calls POST /api/syslog/settings/del_destination/$uuid.
func (m *Module) DelDestination(ctx context.Context, uuid string) error {
	_, err := m.c.Do(ctx, opnsense.Request{
		Method: http.MethodPost,
		Path:   "/api/syslog/settings/del_destination/" + uuid,
	})
	return err
}

// SetDestination calls POST /api/syslog/settings/set with the syslog
// configuration as JSON. This performs a full replace of the destinations list.
func (m *Module) SetDestination(ctx context.Context, enabled bool, preserveFQDN bool, destinations []Destination) error {
	body, err := json.Marshal(map[string]any{
		"general": map[string]any{
			"enabled":      boolToString(enabled),
			"preservefqdn": boolToString(preserveFQDN),
		},
		"destinations": destinations,
	})
	if err != nil {
		return err
	}
	_, err = m.c.Do(ctx, opnsense.Request{
		Method: http.MethodPost,
		Path:   "/api/syslog/settings/set",
		Body:   strings.NewReader(string(body)),
	})
	return err
}

func boolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
