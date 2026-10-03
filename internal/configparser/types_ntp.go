// SPDX-License-Identifier: AGPL-3.0-or-later
// types_ntp.go: NTP configuration model types.
package configparser

// ConfigNTPServer is one upstream NTP server.
type ConfigNTPServer struct {
	Address string `json:"address"` // hostname or IP
	Prefer  bool   `json:"prefer,omitempty"`
	IBurst  bool   `json:"iburst,omitempty"`
	Enabled bool   `json:"enabled"`
}

// ConfigNTPConfig is the top-level NTP configuration for a device.
type ConfigNTPConfig struct {
	Enabled    bool              `json:"enabled"`
	Servers    []ConfigNTPServer `json:"servers"`
	LocalClock bool              `json:"local_clock,omitempty"` // OpenWrt: use local clock as fallback
	Timezone   string            `json:"timezone,omitempty"`     // IANA tz, e.g. "Europe/Madrid"
}
