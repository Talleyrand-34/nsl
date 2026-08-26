// SPDX-License-Identifier: MIT
// opnsense_renderer.go: OPNsense ConfigRenderer stub — registers itself in the
// global renderer registry on init so the push engine has a fallback entry.
package parsers

import "nsl-graph/internal/configparser"

func init() {
	configparser.DefaultRendererRegistry.RegisterRenderer(newOpnsenseRenderer())
}

func newOpnsenseRenderer() configparser.ConfigRenderer {
	return &opnsenseRenderer{}
}

func NewOpnsenseRendererForURL(baseURL, key, secret string) configparser.ConfigRenderer {
	return &opnsenseRenderer{}
}

type opnsenseRenderer struct{}

func (r *opnsenseRenderer) GetOsType() string                            { return "opnsense" }
func (r *opnsenseRenderer) SupportsDevice(d any) bool                  { return true }
func (r *opnsenseRenderer) Diff(intended, observed *configparser.ConfigData) []configparser.ConfigChange {
	return nil
}
func (r *opnsenseRenderer) Render(safety configparser.SafetyLevel, intended *configparser.ConfigData, sess configparser.Session, creds configparser.SSHCredentials) error {
	return configparser.ErrUnsupported{}
}
