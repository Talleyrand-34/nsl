// SPDX-License-Identifier: MIT
// fetchers.go: per-OS strategies for fetching the device's currently-running
// config (LiveConfig) and its native raw bytes (RawFetcher).
//
// One Fetcher impl per supported OS. The Service picks the right one
// from a map keyed by OS type. New vendors add a new map entry and a new
// fetcher type — the Service body stays unchanged.
//
// Note on raw bytes: OpenWrt implements both Fetcher and RawFetcher
// (UCI is already a parser pipeline). OPNsense only implements Fetcher
// because parsing /api/core/backup/download's config.xml would require
// a brand-new XML parser we haven't written. The snapshot pipeline
// tolerates a missing RawFetcher; the parsed snapshot is still
// persisted and the BackupRelpath is left empty.
package push

import (
	"context"
	"fmt"
	"net"
	"strings"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/configparser/parsers"
	s "nsl-graph/internal/scanner"
)

// Fetcher is the OS-agnostic read-only surface for "give me the device's
// currently-running *ConfigData". Implementations are owned by the push
// package and may live alongside the typed renderers.
type Fetcher interface {
	// Fetch opens a session to the device, pulls the current config,
	// parses it, and returns the result. May return ErrCredentialsUnavailable
	// (wrapped) when no creds are available, or any transport error.
	Fetch(ctx context.Context, deviceID string) (*configparser.ConfigData, error)
}

// RawFetcher is the optional companion to Fetcher. Vendors that can
// produce the device's native-format raw bytes (UCI for OpenWrt, etc.)
// implement this so the snapshot pipeline can write a backup file
// alongside the parsed snapshot. Vendors that can't yet (OPNsense's
// config.xml, VyOS, RouterOS, Infix, Fortinet) just implement Fetcher;
// the Service tolerates a missing RawFetcher and writes an empty
// BackupRelpath in the snapshot row.
type RawFetcher interface {
	// FetchRaw returns the device's native raw bytes plus the OS-specific
	// file extension ("conf", "xml", "boot", "rsc"). ext is used to name
	// the backup file on disk.
	FetchRaw(ctx context.Context, deviceID string) (raw []byte, ext string, err error)
}

// ---------------------------------------------------------------------------
// OpenWrt
// ---------------------------------------------------------------------------

// OpenWrtFetcher fetches via SSH using the read-side parser pipeline
// (OpenWrtParser.Fetch + OpenWrtParser.ParseConfig). Credentials come
// from the Service's CredentialResolver.SSH.
//
// ponytail: the host to SSH to is resolved by a host resolver injected
// at construction. Production wires this to the device's primary IP
// (Device.Ips[0]); tests pass a fixed host.
type OpenWrtFetcher struct {
	creds     CredentialResolver
	transport configparser.Transport
	hostFor   func(deviceID string) (string, error)
}

func NewOpenWrtFetcher(creds CredentialResolver, transport configparser.Transport, hostFor func(string) (string, error)) *OpenWrtFetcher {
	return &OpenWrtFetcher{creds: creds, transport: transport, hostFor: hostFor}
}

func (f *OpenWrtFetcher) Fetch(ctx context.Context, deviceID string) (*configparser.ConfigData, error) {
	cd, _, err := f.fetchAll(ctx, deviceID)
	return cd, err
}

// FetchRaw implements RawFetcher. OpenWrt's native format is the
// concatenated UCI block returned by OpenWrtParser.Fetch.
func (f *OpenWrtFetcher) FetchRaw(ctx context.Context, deviceID string) ([]byte, string, error) {
	_, raw, err := f.fetchAll(ctx, deviceID)
	return raw, "conf", err
}

// fetchAll is the shared SSH pipeline: open a session, run the parser
// fetch (returns raw UCI), parse the raw into *ConfigData. Both Fetch
// and FetchRaw share this so the session is opened exactly once per
// snapshot capture.
func (f *OpenWrtFetcher) fetchAll(ctx context.Context, deviceID string) (*configparser.ConfigData, []byte, error) {
	host, err := f.hostFor(deviceID)
	if err != nil {
		return nil, nil, fmt.Errorf("openwrt-fetcher: resolve host for %s: %w", deviceID, err)
	}
	ssh, err := f.creds.SSH(deviceID)
	if err != nil {
		return nil, nil, err
	}
	sess, err := f.transport.Open(host, ssh)
	if err != nil {
		return nil, nil, fmt.Errorf("openwrt-fetcher: ssh to %s: %w", host, err)
	}
	defer sess.Close()

	raw, err := parsers.NewOpenWrtParser().Fetch(sess)
	if err != nil {
		return nil, nil, fmt.Errorf("openwrt-fetcher: openwrt fetch: %w", err)
	}
	cd, err := parsers.NewOpenWrtParser().ParseConfig(raw, s.SNMPDevice{})
	if err != nil {
		return nil, nil, fmt.Errorf("openwrt-fetcher: openwrt parse: %w", err)
	}
	return cd, []byte(raw), nil
}

// ---------------------------------------------------------------------------
// OPNsense
// ---------------------------------------------------------------------------

// OpnSenseFetcher fetches via the typed REST client (no SSH, no parser).
// ponytail: the renderer is built per-call from the API credentials so a
// rotation of the API key takes effect on the next call without
// recreating the Service.
type OpnSenseFetcher struct {
	creds CredentialResolver
}

func NewOpnSenseFetcher(creds CredentialResolver) *OpnSenseFetcher {
	return &OpnSenseFetcher{creds: creds}
}

func (f *OpnSenseFetcher) Fetch(ctx context.Context, deviceID string) (*configparser.ConfigData, error) {
	base, key, secret, err := f.creds.API(deviceID, "opnsense")
	if err != nil {
		return nil, err
	}
	r := parsers.NewOpnsenseRendererForURL(base, key, secret)
	if fl, ok := r.(interface {
		FetchLiveConfig(ctx context.Context) (*configparser.ConfigData, error)
	}); ok {
		return fl.FetchLiveConfig(ctx)
	}
	return nil, fmt.Errorf("opnsense-fetcher: renderer %T does not implement FetchLiveConfig", r)
}

// deviceIndex is the minimal slice the host resolver needs from the
// device repository. Avoids importing the full repository here.
type deviceIndex struct {
	hosts map[string]string
}

func newDeviceIndex() *deviceIndex { return &deviceIndex{hosts: map[string]string{}} }

func (d *deviceIndex) Set(deviceID, host string) { d.hosts[deviceID] = host }
func (d *deviceIndex) Get(deviceID string) (string, bool) {
	h, ok := d.hosts[deviceID]
	return h, ok
}

// hostFromIP returns the first non-loopback IPv4/IPv6 from a comma-
// separated IP list. Used when the device row stores IPs as a list.
func hostFromIP(ips string) string {
	for _, ip := range strings.Split(ips, ",") {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		parsed := net.ParseIP(ip)
		if parsed == nil {
			continue
		}
		if parsed.IsLoopback() {
			continue
		}
		return ip
	}
	return ""
}