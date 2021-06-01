package schema

// ServerConfiguration represents the configuration of the http server.
type ServerConfiguration struct {
	Path            string `koanf:"path"`
	ReadBufferSize  int    `koanf:"read_buffer_size"`
	WriteBufferSize int    `koanf:"write_buffer_size"`
	EnablePprof     bool   `koanf:"enable_endpoint_pprof"`
	EnableExpvars   bool   `koanf:"enable_endpoint_expvars"`
}

// DefaultServerConfiguration represents the default values of the ServerConfiguration.
var DefaultServerConfiguration = ServerConfiguration{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}
