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
