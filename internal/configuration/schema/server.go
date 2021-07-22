package schema

// ServerConfiguration represents the configuration of the http server.
type ServerConfiguration struct {
	Path            string `mapstructure:"path"`
	ReadBufferSize  int    `mapstructure:"read_buffer_size"`
	WriteBufferSize int    `mapstructure:"write_buffer_size"`
	EnablePprof     bool   `mapstructure:"enable_endpoint_pprof"`
	EnableExpvars   bool   `mapstructure:"enable_endpoint_expvars"`

	CORS CORSConfiguration `mapstructure:"cors"`
}

// CORSConfiguration represents the configuration of the http server CORS configuration.
type CORSConfiguration struct {
	Enable           bool     `mapstructure:"enable"`
	IncludeProtected bool     `mapstructure:"include_protected"`
	Origins          []string `mapstructure:"origins,weak"`
	Headers          []string `mapstructure:"headers,weak"`
	Methods          []string `mapstructure:"methods,weak"`
	Vary             []string `mapstructure:"vary,weak"`
	MaxAge           int      `mapstructure:"max_age,weak"`
}

// DefaultServerConfiguration represents the default values of the ServerConfiguration.
var DefaultServerConfiguration = ServerConfiguration{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CORS: CORSConfiguration{
		Vary:   []string{"Accept-Encoding", "Origin"},
		MaxAge: 100,
	},
}
