package schema

// ServerConfiguration represents the configuration of the http server.
type ServerConfiguration struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Path            string `mapstructure:"path"`
	ReadBufferSize  int    `mapstructure:"read_buffer_size"`
	WriteBufferSize int    `mapstructure:"write_buffer_size"`
	EnablePprof     bool   `mapstructure:"enable_endpoint_pprof"`
	EnableExpvars   bool   `mapstructure:"enable_endpoint_expvars"`

	TLS ServerTLSConfiguration `mapstructure:"tls"`
}

// ServerTLSConfiguration represents the configuration of the http servers TLS options.
type ServerTLSConfiguration struct {
	Certificate string `mapstructure:"certificate"`
	Key         string `mapstructure:"key"`
}

// DefaultServerConfiguration represents the default values of the ServerConfiguration.
var DefaultServerConfiguration = ServerConfiguration{
	Host:            "0.0.0.0",
	Port:            9091,
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}
