package schema

type LoggingConfiguration struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	FilePath   string `mapstructure:"file_path"`
	KeepStdout bool   `mapstructure:"keep_stdout"`
}

var DefaultLoggingConfiguration = LoggingConfiguration{
	Level:  "info",
	Format: "text",
}
