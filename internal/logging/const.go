package logging

import (
	"regexp"
	"sync"

	"github.com/sirupsen/logrus"
)

// Log Format values.
const (
	FormatText = "text"
	FormatJSON = "json"
)

type LogLevel string

// Log Level values.
const (
	LevelTrace = "trace"
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

func (l LogLevel) Level() logrus.Level {
	switch l {
	case LevelError:
		return logrus.ErrorLevel
	case LevelWarn:
		return logrus.WarnLevel
	case LevelInfo:
		return logrus.InfoLevel
	case LevelDebug:
		return logrus.DebugLevel
	case LevelTrace:
		return logrus.TraceLevel
	default:
		return logrus.InfoLevel
	}
}

// Field names.
const (
	FieldRemoteIP   = "remote_ip"
	FieldMethod     = "method"
	FieldPath       = "path"
	FieldPathRaw    = "path_raw"
	FieldStatusCode = "status_code"
)

var (
	stacktrace       sync.Once
	reFormatFilePath = regexp.MustCompile(`(%d|\{datetime(:([^}]+))?})`)
)
