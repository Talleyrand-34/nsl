// SPDX-License-Identifier: MIT
// opnsense_renderer.go: OPNsense renderer speaking the REST API.
package parsers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/scanner"
	"github.com/t34/opnsense-api/modules/interfaces"
	"github.com/t34/opnsense-api/modules/lldp"
	"github.com/t34/opnsense-api/modules/ntp"
	"github.com/t34/opnsense-api/modules/routes"
	"github.com/t34/opnsense-api/modules/routing"
	"github.com/t34/opnsense-api/modules/syslog"
	"github.com/t34/opnsense-api/modules/system"
	"github.com/t34/opnsense-api/opnsense"
)

type opnsenseRenderer struct {
	c        *opnsense.Client
	ifaces   *interfaces.Module
	routes   *routes.Module
	gateways *routing.Module
	ntp      *ntp.Module
	system   *system.Module
	lldp     *lldp.Module
	syslog   *syslog.Module
}

func init() { configparser.DefaultRendererRegistry.RegisterRenderer(newOpnsenseRenderer()) }

func newOpnsenseRenderer() configparser.ConfigRenderer {
	return &opnsenseLazyRenderer{}
}

// NewOpnsenseRendererForURL is the engine-facing constructor.
func NewOpnsenseRendererForURL(baseURL, key, secret string) configparser.ConfigRenderer {
	opts := []opnsense.Option{opnsense.WithInsecureTLS()}
	c := opnsense.NewClient(baseURL, key, secret, opts...)
	return &opnsenseRenderer{
		c:        c,
		ifaces:   interfaces.New(c),
		routes:   routes.New(c),
		gateways: routing.New(c),
		ntp:      ntp.New(c),
		system:   system.New(c),
		lldp:     lldp.New(c),
		syslog:   syslog.New(c),
	}
}

// opnsenseLazyRenderer is the registry entry when no client is wired.
type opnsenseLazyRenderer struct{}

func (r *opnsenseLazyRenderer) GetOsType() string                              { return "opnsense" }
func (r *opnsenseLazyRenderer) SupportsDevice(d any) bool {
	if dev, ok := d.(scanner.SNMPDevice); ok {
		s := dev.SysDescr + " " + dev.SysName
		return containsFold(s, "opnsense") || containsFold(s, "freebsd")
	}
	return false
}
func (r *opnsenseLazyRenderer) Diff(intended *configparser.ConfigData, observed *configparser.ConfigData) []configparser.ConfigChange {
	return diffRoutes(intended, observed)
}
func (r *opnsenseLazyRenderer) Render(_ configparser.SafetyLevel, _ *configparser.ConfigData, _ configparser.Session, _ configparser.SSHCredentials) error {
	return configparser.ErrUnsupported{OS: "opnsense", Reason: "render requires typed session; engine must call opnsenseRenderer directly"}
}

// --- real renderer ---

func (r *opnsenseRenderer) GetOsType() string { return "opnsense" }
func (r *opnsenseRenderer) SupportsDevice(d any) bool {
	if dev, ok := d.(scanner.SNMPDevice); ok {
		s := dev.SysDescr + " " + dev.SysName
		return containsFold(s, "opnsense") || containsFold(s, "freebsd")
	}
	return false
}

func (r *opnsenseRenderer) Diff(intended *configparser.ConfigData, observed *configparser.ConfigData) []configparser.ConfigChange {
	var out []configparser.ConfigChange
	out = append(out, diffVLANs(intended, observed)...)
	out = append(out, diffRoutes(intended, observed)...)
	out = append(out, diffNTP(intended, observed)...)
	out = append(out, diffBanner(intended, observed)...)
	out = append(out, diffLLDP(intended, observed)...)
	out = append(out, diffSyslog(intended, observed)...)
	return out
}

func (r *opnsenseRenderer) Render(safety configparser.SafetyLevel, intended *configparser.ConfigData, _ configparser.Session, _ configparser.SSHCredentials) error {
	ctx := context.Background()
	switch safety {
	case configparser.SafetyDryRun:
		return nil
	case configparser.SafetyStaged, configparser.SafetyApply:
		observed, err := r.fetchObserved(ctx)
		if err != nil {
			return fmt.Errorf("opnsense-renderer: fetch observed: %w", err)
		}
		diffs := r.Diff(intended, observed)
		var touchedIface, touchedRoutes, touchedNTP, touchedBanner, touchedLLDP, touchedSyslog bool
		for _, d := range diffs {
			if err := r.applyChange(ctx, d); err != nil {
				return fmt.Errorf("opnsense-renderer: %w", err)
			}
			switch d.Kind {
			case "route-add", "route-del":
				touchedRoutes = true
			case "vlan-add", "vlan-del":
				touchedIface = true
			case "ntp-set":
				touchedNTP = true
			case "banner-set":
				touchedBanner = true
			case "lldp-set":
				touchedLLDP = true
			case "syslog-set":
				touchedSyslog = true
			}
		}
		if touchedRoutes {
			if err := r.routes.RouteApply(ctx); err != nil {
				return fmt.Errorf("opnsense-renderer: routes reconfigure: %w", err)
			}
		}
		if touchedIface {
			if err := r.ifaces.OverviewCommit(ctx); err != nil {
				return fmt.Errorf("opnsense-renderer: interfaces commit: %w", err)
			}
		}
		if touchedNTP {
			if err := r.applyNTP(ctx, intended.NTP); err != nil {
				return fmt.Errorf("opnsense-renderer: ntp: %w", err)
			}
		}
		if touchedBanner {
			if err := r.applyBanner(ctx, intended.Banner); err != nil {
				return fmt.Errorf("opnsense-renderer: banner: %w", err)
			}
		}
		if touchedLLDP {
			if err := r.applyLLDP(ctx, intended.LLDP); err != nil {
				return fmt.Errorf("opnsense-renderer: lldp: %w", err)
			}
		}
		if touchedSyslog {
			if err := r.applySyslog(ctx, intended.Syslog); err != nil {
				return fmt.Errorf("opnsense-renderer: syslog: %w", err)
			}
		}
		return nil
	default:
		return fmt.Errorf("opnsense-renderer: unknown safety %d", int(safety))
	}
}

// applyChange dispatches one ConfigChange to the right OPNsense API call.
func (r *opnsenseRenderer) applyChange(ctx context.Context, change configparser.ConfigChange) error {
	switch change.Kind {
	case "route-add":
		return r.routes.RouteAdd(ctx, routes.RouteAdd{
			Network: change.New,
			Gateway: change.Old,
			Descr:   "nsl-graph push",
		})
	case "vlan-add":
		iface, tag := splitPathVLAN(change.Path)
		if iface == "" {
			return fmt.Errorf("opnsense-renderer: vlan-add path must include interface: %q", change.Path)
		}
		_, err := r.ifaces.VLANAdd(ctx, interfaces.VLANAdd{
			If:    iface,
			Tag:   tag,
			Descr: "nsl-graph push",
		})
		return err
	case "ntp-set", "banner-set", "lldp-set", "syslog-set":
		return nil
	default:
		return fmt.Errorf("opnsense-renderer: unhandled change kind %q", change.Kind)
	}
}

// applyNTP pushes the full NTP config via POST /api/ntp/settings/set.
func (r *opnsenseRenderer) applyNTP(ctx context.Context, cfg *configparser.ConfigNTPConfig) error {
	if cfg == nil {
		return nil
	}
	enabled := "0"
	if cfg.Enabled {
		enabled = "1"
	}
	servers := make([]string, len(cfg.Servers))
	for i, s := range cfg.Servers {
		servers[i] = s.Address
	}
	return r.ntp.NTPSet(ctx, ntp.NTPSettings{
		Enable:      enabled,
		Timeservers: servers,
		Timezone:   cfg.Timezone,
	})
}

// applyBanner pushes the login banner via POST /api/system/general/set.
func (r *opnsenseRenderer) applyBanner(ctx context.Context, cfg *configparser.ConfigBanner) error {
	if cfg == nil {
		return nil
	}
	return r.system.GeneralSet(ctx, cfg.LoginBanner)
}

// applyLLDP enables/disables LLDP tx and reconfigures.
func (r *opnsenseRenderer) applyLLDP(ctx context.Context, cfg *configparser.ConfigLLDPSettings) error {
	if cfg == nil {
		return nil
	}
	enabled := false
	if cfg.Enabled {
		enabled = true
	}
	if err := r.lldp.ServiceSet(ctx, lldp.LLDPServiceSettings{Enabled: enabled}); err != nil {
		return err
	}
	return r.lldp.ServiceReconfigure(ctx)
}

// applySyslog replaces the full syslog destination list.
func (r *opnsenseRenderer) applySyslog(ctx context.Context, cfg *configparser.ConfigSyslogConfig) error {
	if cfg == nil {
		return nil
	}
	dests := make([]syslog.Destination, len(cfg.Targets))
	for i, t := range cfg.Targets {
		port := "514"
		if t.Port != 0 {
			port = fmt.Sprintf("%d", t.Port)
		}
		transport := "udp"
		if t.Protocol != "" {
			transport = t.Protocol
		}
		level := ""
		if t.LogOnly {
			level = "info"
		}
		dests[i] = syslog.Destination{
			Address:   t.Address,
			Port:      port,
			Transport: transport,
			Facility:  t.Facility,
			Program:   t.Program,
			Level:     level,
			Enabled:   "1",
		}
	}
	return r.syslog.SetDestination(ctx, cfg.Enabled, cfg.PreserveFQDN, dests)
}

// fetchObserved pulls the live state from the device.
func (r *opnsenseRenderer) fetchObserved(ctx context.Context) (*configparser.ConfigData, error) {
	out := &configparser.ConfigData{}
	ifaceResp, err := r.ifaces.OverviewList(ctx)
	if err != nil {
		return nil, err
	}
	if err := decodeInterfacesInto(ifaceResp.RawBody, out); err != nil {
		return nil, fmt.Errorf("decode interfaces: %w", err)
	}
	routeResp, err := r.routes.SearchRoute(ctx)
	if err != nil {
		return nil, err
	}
	if err := decodeRoutesInto(routeResp.RawBody, out); err != nil {
		return nil, fmt.Errorf("decode routes: %w", err)
	}
	return out, nil
}

// decodeInterfacesInto parses the JSON returned by /api/interfaces/overview/list.
func decodeInterfacesInto(raw []byte, out *configparser.ConfigData) error {
	if len(raw) == 0 {
		return nil
	}
	var env struct {
		Rows []struct {
			Device     string `json:"device"`
			Identifier string `json:"identifier"`
			VlanTag    string `json:"vlan_tag"`
			Vlan       *struct {
				Tag    string `json:"tag"`
				Parent string `json:"parent"`
			} `json:"vlan"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return err
	}
	for _, row := range env.Rows {
		if row.Device == "" {
			continue
		}
		ci := configparser.ConfigInterface{
			Name:    row.Device,
			Type:    "physical",
			Enabled: true,
		}
		tag := row.VlanTag
		if tag == "" && row.Vlan != nil {
			tag = row.Vlan.Tag
		}
		if tag != "" {
			ci.VLANs = []configparser.ConfigVLAN{{ID: tag, Tagged: true}}
		}
		out.Interfaces = append(out.Interfaces, ci)
	}
	return nil
}

// decodeRoutesInto parses the JSON returned by /api/routes/routes/searchroute.
func decodeRoutesInto(raw []byte, out *configparser.ConfigData) error {
	if len(raw) == 0 {
		return nil
	}
	var env struct {
		Rows []struct {
			UUID    string `json:"uuid"`
			Network string `json:"network"`
			Gateway string `json:"gateway"`
			Descr   string `json:"descr"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return err
	}
	for _, row := range env.Rows {
		if row.Network == "" {
			continue
		}
		out.Routes = append(out.Routes, configparser.ConfigRoute{
			Network:     row.Network,
			Gateway:     row.Gateway,
			Description: row.Descr,
		})
	}
	return nil
}

// splitPathVLAN parses the Path field of a vlan-add ConfigChange.
func splitPathVLAN(path string) (string, int) {
	parts := strings.Fields(path)
	if len(parts) < 3 || parts[1] != "VLAN" {
		return "", 0
	}
	tag := 0
	for _, c := range parts[2] {
		if c < '0' || c > '9' {
			return parts[0], 0
		}
		tag = tag*10 + int(c-'0')
	}
	return parts[0], tag
}

// containsFold is a tiny case-insensitive substring test.
func containsFold(haystack, needle string) bool {
	if len(needle) > len(haystack) {
		return false
	}
	for i := range haystack {
		if i+len(needle) > len(haystack) {
			break
		}
		match := true
		for j := range needle {
			h := haystack[i+j]
			n := needle[j]
			if h >= 'A' && h <= 'Z' {
				h += 'a' - 'A'
			}
			if n >= 'A' && n <= 'Z' {
				n += 'a' - 'A'
			}
			if h != n {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// pin imports
var (
	_ = http.MethodGet
	_ = context.Background
	_ = fmt.Sprintf
)
