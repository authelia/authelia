package schema

// Log represents the logging configuration.
type Log struct {
	Level      string `koanf:"level" yaml:"level,omitempty" toml:"level,omitempty" json:"level,omitempty" jsonschema:"enum=error,enum=warn,enum=info,enum=debug,enum=trace,title=Level" jsonschema_description:"The minimum Level a Log message must be before it's added to the log."`
	Format     string `koanf:"format" yaml:"format,omitempty" toml:"format,omitempty" json:"format,omitempty" jsonschema:"enum=json,enum=text,title=Format" jsonschema_description:"The Format of Log messages."`
	FilePath   string `koanf:"file_path" yaml:"file_path,omitempty" toml:"file_path,omitempty" json:"file_path,omitempty" jsonschema:"title=File Path" jsonschema_description:"The File Path to save the logs to instead of sending them to stdout, it's strongly recommended this option is only enabled with 'keep_stdout' also enabled."`
	KeepStdout bool   `koanf:"keep_stdout" yaml:"keep_stdout" toml:"keep_stdout" json:"keep_stdout" jsonschema:"default=false,title=Keep Stdout" jsonschema_description:"Enables keeping stdout when using the File Path option."`
}

// DefaultLoggingConfiguration is the default logging configuration.
var DefaultLoggingConfiguration = Log{
	Level:  "info",
	Format: "text",
}
