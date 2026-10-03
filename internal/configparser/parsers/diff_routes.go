// SPDX-License-Identifier: AGPL-3.0-or-later
// diff_routes.go: shared diff logic for static routes.
package parsers

import (
	"fmt"
	"nsl-graph/internal/configparser"
)

// diffRoutes compares intended vs observed routes and returns add/delete changes.
func diffRoutes(intended, observed *configparser.ConfigData) []configparser.ConfigChange {
	if intended == nil {
		return nil
	}
	// A nil Routes means the caller did not speak about routes at all, and that
	// must not be read as "delete every route the device has". Only a non-nil
	// list — the empty one included — states the full intended route set, and
	// only then are deletions computed. Without this a preview that changes
	// nothing but NTP also proposes wiping the routing table, because an
	// unmentioned section and an intentionally empty one looked identical.
	if intended.Routes == nil {
		return nil
	}
	if observed == nil {
		var out []configparser.ConfigChange
		for _, r := range intended.Routes {
			out = append(out, routeAddChange(r))
		}
		return out
	}

	obsMap := routeMap(observed.Routes)
	var out []configparser.ConfigChange

	for _, r := range intended.Routes {
		key := routeKey(r)
		if _, ok := obsMap[key]; !ok {
			out = append(out, routeAddChange(r))
		}
	}

	intMap := routeMap(intended.Routes)
	for _, r := range observed.Routes {
		key := routeKey(r)
		if _, ok := intMap[key]; !ok {
			out = append(out, routeDelChange(r))
		}
	}
	return out
}

func routeKey(r configparser.ConfigRoute) string {
	// Network + gateway uniquely identifies a route. Interface is metadata
	// for the apply step (OpenWrt UCI needs it; OPNsense ignores it) and
	// varies across vendor models, so we don't include it in the key.
	return r.Network + "|" + r.Gateway
}

func routeMap(routes []configparser.ConfigRoute) map[string]configparser.ConfigRoute {
	m := make(map[string]configparser.ConfigRoute, len(routes))
	for _, r := range routes {
		m[routeKey(r)] = r
	}
	return m
}

func routeAddChange(r configparser.ConfigRoute) configparser.ConfigChange {
	return configparser.ConfigChange{
		Kind:   "route-add",
		Path:   r.Network,
		New:    r.Network,
		Old:    r.Gateway,
		Patch: []string{
			"network=" + r.Network,
			"gateway=" + r.Gateway,
			"interface=" + r.Interface,
		},
	}
}

func routeDelChange(r configparser.ConfigRoute) configparser.ConfigChange {
	return configparser.ConfigChange{
		Kind:   "route-del",
		Path:   r.Network,
		Old:    r.Gateway,
		Patch:  []string{fmt.Sprintf("network=%s gateway=%s", r.Network, r.Gateway)},
	}
}
