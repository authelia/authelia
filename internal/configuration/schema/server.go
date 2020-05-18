package schema

// ServerConfiguration represents the configuration of the http server.
type ServerConfiguration struct {
	Path            string `mapstructure:"path"`
	ReadBufferSize  int    `mapstructure:"read_buffer_size"`
	WriteBufferSize int    `mapstructure:"write_buffer_size"`
}

// DefaultServerConfiguration represents the default values of the ServerConfiguration.
var DefaultServerConfiguration = ServerConfiguration{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}
