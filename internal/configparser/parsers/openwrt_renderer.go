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
	out = append(out, diffNTP(intended, observed)...)
	out = append(out, diffBanner(intended, observed)...)
	out = append(out, diffLLDP(intended, observed)...)
	out = append(out, diffSyslog(intended, observed)...)
	out = append(out, diffSNMP(intended, observed)...)
	return out
}

func (r *OpenWrtRenderer) Render(safety configparser.SafetyLevel, intended *configparser.ConfigData, sess configparser.Session, creds configparser.SSHCredentials) error {
	if safety == configparser.SafetyDryRun {
		return nil
	}
	diffs := r.Diff(intended, nil)
	var ntpDiffs, bannerDiffs, lldpDiffs, syslogDiffs, snmpDiffs []configparser.ConfigChange
	for _, d := range diffs {
		switch d.Kind {
		case "ntp-set":
			ntpDiffs = append(ntpDiffs, d)
		case "banner-set":
			bannerDiffs = append(bannerDiffs, d)
		case "lldp-set":
			lldpDiffs = append(lldpDiffs, d)
		case "syslog-set":
			syslogDiffs = append(syslogDiffs, d)
		case "snmp-set":
			snmpDiffs = append(snmpDiffs, d)
		default:
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
	}
	if len(ntpDiffs) > 0 && intended.NTP != nil {
		for _, cmd := range r.ntpCommands(intended.NTP) {
			if _, err := sess.Execute(cmd); err != nil {
				return fmt.Errorf("openwrt-renderer: %s: %w", cmd, err)
			}
		}
	}
	if len(bannerDiffs) > 0 && intended.Banner != nil {
		for _, cmd := range r.bannerCommands(intended.Banner) {
			if _, err := sess.Execute(cmd); err != nil {
				return fmt.Errorf("openwrt-renderer: %s: %w", cmd, err)
			}
		}
	}
	if len(lldpDiffs) > 0 && intended.LLDP != nil {
		for _, cmd := range r.lldpCommands(intended.LLDP) {
			if _, err := sess.Execute(cmd); err != nil {
				return fmt.Errorf("openwrt-renderer: %s: %w", cmd, err)
			}
		}
	}
	if len(syslogDiffs) > 0 && intended.Syslog != nil {
		for _, cmd := range r.syslogCommands(intended.Syslog) {
			if _, err := sess.Execute(cmd); err != nil {
				return fmt.Errorf("openwrt-renderer: %s: %w", cmd, err)
			}
		}
	}
	if len(snmpDiffs) > 0 && intended.SNMP != nil {
		for _, cmd := range r.snmpCommands(intended.SNMP) {
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
		return r.routeAddCommands(d), nil
	case "route-del":
		return r.routeDelCommands(d), nil
	default:
		return nil, nil
	}
}

func (r *OpenWrtRenderer) routeAddCommands(d configparser.ConfigChange) []string {
	name := "route_" + sanitizeUCIName(d.New)
	gw := patchField(d.Patch, "gateway=")
	iface := patchField(d.Patch, "interface=")
	return []string{
		fmt.Sprintf("uci set network.%s=route", name),
		fmt.Sprintf("uci set network.%s.target=%s", name, d.New),
		fmt.Sprintf("uci set network.%s.gateway=%s", name, gw),
		fmt.Sprintf("uci set network.%s.device=%s", name, iface),
		"uci commit network",
	}
}

func (r *OpenWrtRenderer) routeDelCommands(d configparser.ConfigChange) []string {
	name := "route_" + sanitizeUCIName(d.Path)
	return []string{
		fmt.Sprintf("uci del network.%s", name),
		"uci commit network",
	}
}

// ntpCommands returns UCI commands to set NTP servers and timezone.
func (r *OpenWrtRenderer) ntpCommands(cfg *configparser.ConfigNTPConfig) []string {
	if cfg == nil {
		return nil
	}
	var cmds []string
	enabled := "0"
	if cfg.Enabled {
		enabled = "1"
	}
	cmds = append(cmds, "uci set system.ntp=timeserver")
	cmds = append(cmds, fmt.Sprintf("uci set system.ntp.enabled=%s", enabled))
	cmds = append(cmds, "uci del system.ntp.server")
	for _, srv := range cfg.Servers {
		cmds = append(cmds, fmt.Sprintf("uci add_list system.ntp.server=%s", srv.Address))
	}
	if cfg.Timezone != "" {
		cmds = append(cmds, fmt.Sprintf("uci set system.system.timezone=%s", cfg.Timezone))
	}
	if cfg.LocalClock {
		cmds = append(cmds, "uci set system.ntp.use_local_clock=1")
	}
	cmds = append(cmds, "uci commit system")
	cmds = append(cmds, "/etc/init.d/sysntpd reload")
	return cmds
}

// bannerCommands returns shell commands to write the login and post-login banners.
func (r *OpenWrtRenderer) bannerCommands(cfg *configparser.ConfigBanner) []string {
	if cfg == nil {
		return nil
	}
	var cmds []string
	if cfg.LoginBanner != "" {
		cmds = append(cmds, fmt.Sprintf("cat > /etc/issue.net << 'NSLBANNER'\n%s\nNSLBANNER", cfg.LoginBanner))
	}
	if cfg.PostLogin != "" {
		cmds = append(cmds, fmt.Sprintf("cat > /etc/motd << 'NSLBANNER'\n%s\nNSLBANNER", cfg.PostLogin))
	}
	return cmds
}

// lldpCommands returns UCI commands to configure LLDP tx.
func (r *OpenWrtRenderer) lldpCommands(cfg *configparser.ConfigLLDPSettings) []string {
	if cfg == nil {
		return nil
	}
	var cmds []string
	cmds = append(cmds, "uci set lldpd.config=lldpd")
	enabled := "0"
	if cfg.Enabled {
		enabled = "1"
	}
	cmds = append(cmds, fmt.Sprintf("uci set lldpd.config.enabled=%s", enabled))
	if cfg.SystemName != "" {
		cmds = append(cmds, fmt.Sprintf("uci set lldpd.config.lldp_neigh=%s", cfg.SystemName))
	}
	if cfg.InterfacePattern != "" {
		cmds = append(cmds, fmt.Sprintf("uci set lldpd.config.interface=%s", cfg.InterfacePattern))
	}
	cmds = append(cmds, "uci commit lldpd")
	cmds = append(cmds, "/etc/init.d/lldpd reload")
	return cmds
}

// syslogCommands returns UCI commands to configure remote syslog targets.
func (r *OpenWrtRenderer) syslogCommands(cfg *configparser.ConfigSyslogConfig) []string {
	if cfg == nil {
		return nil
	}
	var cmds []string
	cmds = append(cmds, "uci del system.cfg001")
	cmds = append(cmds, "uci commit system")
	enabled := "0"
	if cfg.Enabled {
		enabled = "1"
	}
	cmds = append(cmds, fmt.Sprintf("uci add system log_remote"))
	cmds = append(cmds, fmt.Sprintf("uci set system.cfg001.enabled=%s", enabled))
	cmds = append(cmds, fmt.Sprintf("uci set system.cfg001.preserve_fqdn=%d", boolToInt(cfg.PreserveFQDN)))
	for i, t := range cfg.Targets {
		cmds = append(cmds, fmt.Sprintf("uci add system log_remote"))
		cmds = append(cmds, fmt.Sprintf("uci set system.@log_remote[-1].enabled=1"))
		cmds = append(cmds, fmt.Sprintf("uci set system.@log_remote[-1].ip=%s", t.Address))
		if t.Port != 0 {
			cmds = append(cmds, fmt.Sprintf("uci set system.@log_remote[-1].port=%d", t.Port))
		}
		if t.Protocol != "" {
			cmds = append(cmds, fmt.Sprintf("uci set system.@log_remote[-1].proto=%s", t.Protocol))
		}
		_ = i // index unused, UCI add_list appends
	}
	cmds = append(cmds, "uci commit system")
	cmds = append(cmds, "/etc/init.d/log reload")
	return cmds
}

// snmpCommands returns UCI commands to configure SNMP daemon.
func (r *OpenWrtRenderer) snmpCommands(cfg *configparser.ConfigSNMPConfig) []string {
	if cfg == nil {
		return nil
	}
	var cmds []string
	cmds = append(cmds, "uci set snmpd.config=snmpd")
	enabled := "0"
	if cfg.Enabled {
		enabled = "1"
	}
	cmds = append(cmds, fmt.Sprintf("uci set snmpd.config.enabled=%s", enabled))
	if cfg.Location != "" {
		cmds = append(cmds, fmt.Sprintf("uci set snmpd.config.syslocation=%s", cfg.Location))
	}
	if cfg.Contact != "" {
		cmds = append(cmds, fmt.Sprintf("uci set snmpd.config.syscontact=%s", cfg.Contact))
	}
	cmds = append(cmds, "uci del snmpd.config.community")
	for _, c := range cfg.Communities {
		cmds = append(cmds, fmt.Sprintf("uci add_list snmpd.config.community=%s", c.Name))
	}
	cmds = append(cmds, "uci commit snmpd")
	cmds = append(cmds, "/etc/init.d/snmpd reload")
	return cmds
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
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
					Kind:  "vlan-add",
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
