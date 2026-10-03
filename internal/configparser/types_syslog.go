// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package configparser

// ConfigSyslogTarget represents a single remote syslog target.
type ConfigSyslogTarget struct {
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"` // udp or tcp
	Facility    string `json:"facility"`
	Program     string `json:"program"`
	LogOnly     bool   `json:"logonly"`
}

// ConfigSyslogConfig holds the top-level syslog configuration.
type ConfigSyslogConfig struct {
	Enabled       bool                  `json:"enabled"`
	Targets       []ConfigSyslogTarget  `json:"targets"`
	PreserveFQDN  bool                  `json:"preservefqdn"`
}
