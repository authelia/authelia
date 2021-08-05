package schema

// NtpConfiguration represents the configuration related to ntp server.
type NtpConfiguration struct {
	Address string 				`koanf:"address"`
  	Version int 				`koanf:"version"`
  	MaximumDesync string 		`koanf:"max_desync"`
  	DisableStartupCheck bool 	`koanf:"disable_startup_check"`
}
var DefaultVersion = 4

// DefaultNtpConfiguration represents default configuration parameters for the Ntp server.
var DefaultNtpConfiguration = NtpConfiguration {
	Address: "time.cloudflare.com:123",
  	Version: 4,
  	MaximumDesync: "3s",
	DisableStartupCheck: false,
}
