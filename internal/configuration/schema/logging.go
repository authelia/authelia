package schema

// LoggingConfiguration represents the logging configuration.
type LoggingConfiguration struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	FilePath   string `mapstructure:"file_path"`
	KeepStdout bool   `mapstructure:"keep_stdout"`
}

// DefaultLoggingConfiguration is the default logging configuration.
var DefaultLoggingConfiguration = LoggingConfiguration{
	Level:  "info",
	Format: "text",
}
