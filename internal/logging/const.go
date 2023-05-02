package logging

// Log Format values.
const (
	FormatText = "text"
	FormatJSON = "json"
)

// Log Level values.
const (
	LevelTrace = "trace"
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// Field names.
const (
	FieldRemoteIP   = "remote_ip"
	FieldMethod     = "method"
	FieldPath       = "path"
	FieldStatusCode = "status_code"
)
