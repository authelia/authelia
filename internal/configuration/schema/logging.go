package schema

// LogConfiguration represents the logging configuration.
type LogConfiguration struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	FilePath   string `mapstructure:"file_path"`
	KeepStdout bool   `mapstructure:"keep_stdout"`
}

// DefaultLoggingConfiguration is the default logging configuration.
var DefaultLoggingConfiguration = LogConfiguration{
	Level:  "info",
	Format: "text",
}
