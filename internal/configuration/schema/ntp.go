package schema

// NTPConfiguration represents the configuration related to ntp server.
type NTPConfiguration struct {
	Address             string `koanf:"address"`
	Version             int    `koanf:"version"`
	MaximumDesync       string `koanf:"max_desync"`
	DisableStartupCheck bool   `koanf:"disable_startup_check"`
}

// DefaultNTPConfiguration represents default configuration parameters for the NTP server.
var DefaultNTPConfiguration = NTPConfiguration{
	Address:             "time.cloudflare.com:123",
	Version:             4,
	MaximumDesync:       "3s",
	DisableStartupCheck: false,
}
