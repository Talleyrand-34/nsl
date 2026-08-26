// SPDX-License-Identifier: MIT
// types_banner.go: login / MOTD banner model.
package configparser

// ConfigBanner holds the pre-login and post-login banner text.
type ConfigBanner struct {
	LoginBanner string `json:"login_banner,omitempty"` // shown before auth (issue.net / /etc/issue.net)
	PostLogin   string `json:"post_login,omitempty"`    // shown after auth (motd / /etc/motd)
}
