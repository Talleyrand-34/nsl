// SPDX-License-Identifier: MIT
// openwrt_renderer.go: OpenWrt UCI renderer for nsl-graph push.
package parsers

import (
	"fmt"
	"strings"

	"nsl-graph/internal/configparser"
	s "nsl-graph/internal/scanner"
)

type OpenWrtRenderer struct{}

func NewOpenWrtRenderer() *OpenWrtRenderer { return &OpenWrtRenderer{} }

func init() { configparser.DefaultRendererRegistry.RegisterRenderer(NewOpenWrtRenderer()) }

func (r *OpenWrtRenderer) GetOsType() string { return "openwrt" }

func (r *OpenWrtRenderer) SupportsDevice(d any) bool {
	if dev, ok := d.(s.SNMPDevice); ok {
		return strings.Contains(strings.ToLower(dev.SysDescr), "openwrt")
	}
	return false
}

func (r *OpenWrtRenderer) Fetch(sess configparser.Session) (string, error) {
	return "", nil
}

func (r *OpenWrtRenderer) Diff(intended, observed *configparser.ConfigData) []configparser.ConfigChange {
	var out []configparser.ConfigChange
	out = append(out, diffVLANs(intended, observed)...)
	out = append(out, diffRoutes(intended, observed)...)
	return out
}

func (r *OpenWrtRenderer) Render(safety configparser.SafetyLevel, intended *configparser.ConfigData, sess configparser.Session, creds configparser.SSHCredentials) error {
	if safety == configparser.SafetyDryRun {
		return nil
	}
	diffs := r.Diff(intended, nil)
	for _, d := range diffs {
		cmds, err := r.renderChange(d)
		if err != nil {
			return err
		}
		for _, cmd := range cmds {
			if _, err := sess.Execute(cmd); err != nil {
				return fmt.Errorf("openwrt-renderer: %s: %w", cmd, err)
			}
		}
	}
	return nil
}

func (r *OpenWrtRenderer) renderChange(d configparser.ConfigChange) ([]string, error) {
	switch d.Kind {
	case "route-add":
		return r.routeAddCommands(d)
	case "route-del":
		return r.routeDelCommands(d)
	default:
		return nil, nil
	}
}

func (r *OpenWrtRenderer) routeAddCommands(d configparser.ConfigChange) ([]string, error) {
	name := "route_" + sanitizeUCIName(d.New)
	gw := patchField(d.Patch, "gateway=")
	iface := patchField(d.Patch, "interface=")
	return []string{
		fmt.Sprintf("uci set network.%s=route", name),
		fmt.Sprintf("uci set network.%s.target=%s", name, d.New),
		fmt.Sprintf("uci set network.%s.gateway=%s", name, gw),
		fmt.Sprintf("uci set network.%s.device=%s", name, iface),
		"uci commit network",
	}, nil
}

func (r *OpenWrtRenderer) routeDelCommands(d configparser.ConfigChange) ([]string, error) {
	name := "route_" + sanitizeUCIName(d.Path)
	return []string{
		fmt.Sprintf("uci del network.%s", name),
		"uci commit network",
	}, nil
}

func sanitizeUCIName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(name, "/", "_"), ":", "_"), ".", "_")
}

func patchField(patch []string, prefix string) string {
	for _, p := range patch {
		if strings.HasPrefix(p, prefix) {
			return strings.TrimPrefix(p, prefix)
		}
	}
	return ""
}

// diffVLANs computes add/delete VLAN changes for OpenWrt.
func diffVLANs(intended, observed *configparser.ConfigData) []configparser.ConfigChange {
	if observed == nil {
		return nil
	}
	var out []configparser.ConfigChange
	obsMap := map[string]map[string]bool{}
	for _, ci := range observed.Interfaces {
		set := map[string]bool{}
		for _, v := range ci.VLANs {
			set[v.ID] = true
		}
		obsMap[ci.Name] = set
	}
	for _, ci := range intended.Interfaces {
		obsSet := obsMap[ci.Name]
		for _, v := range ci.VLANs {
			if !obsSet[v.ID] {
				out = append(out, configparser.ConfigChange{
					Kind:  "add",
					Path:  fmt.Sprintf("%s VLAN %s tagged=%v", ci.Name, v.ID, v.Tagged),
					New:   v.ID,
					Patch: uciSetLine(ci.Name, v.ID, v.Tagged),
				})
			}
		}
	}
	return out
}

func uciSetLine(ifaceName, vlanID string, tagged bool) []string {
	taggedStr := "0"
	if tagged {
		taggedStr = "t"
	}
	return []string{
		fmt.Sprintf("uci add_list network.%s.vlan_members='%s %s'", ifaceName, vlanID, taggedStr),
	}
}
