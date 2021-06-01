package schema

// LoggingConfiguration represents the logging configuration.
type LoggingConfiguration struct {
	Level      string `koanf:"level"`
	Format     string `koanf:"format"`
	FilePath   string `koanf:"file_path"`
	KeepStdout bool   `koanf:"keep_stdout"`
}

// DefaultLoggingConfiguration is the default logging configuration.
var DefaultLoggingConfiguration = LoggingConfiguration{
	Level:  "info",
	Format: "text",
}
