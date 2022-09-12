package logging

import (
	"io"
	"os"
	"strings"
	"time"

	logrus_stack "github.com/Gurpartap/logrus-stack"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Logger returns the standard logrus logger.
func Logger() *logrus.Logger {
	return logrus.StandardLogger()
}

// LoggerPrintf returns a new PrintfLogger given a level.
func LoggerPrintf(level logrus.Level) (logger *PrintfLogger) {
	return &PrintfLogger{
		level:  level,
		logrus: logrus.StandardLogger(),
	}
}

// LoggerCtxPrintf returns a new CtxPrintfLogger given a level.
func LoggerCtxPrintf(level logrus.Level) (logger *CtxPrintfLogger) {
	return &CtxPrintfLogger{
		level:  level,
		logrus: logrus.StandardLogger(),
	}
}

// InitializeLogger configures the default loggers stack levels, formatting, and the output destinations.
func InitializeLogger(config schema.LogConfiguration, log bool) error {
	setLevelStr(config.Level, log)

	callerLevels := []logrus.Level{}
	stackLevels := []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel}
	logrus.AddHook(logrus_stack.NewHook(callerLevels, stackLevels))

	if config.Format == logFormatJSON {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	if config.FilePath != "" {
		filePath := strings.ReplaceAll(config.FilePath, "%d", time.Now().Format("2006-02-01T150405Z"))

		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

		if err != nil {
			return err
		}

		if config.Format != logFormatJSON {
			logrus.SetFormatter(&logrus.TextFormatter{
				DisableColors: true,
				FullTimestamp: true,
			})
		}

		if config.KeepStdout {
			logLocations := io.MultiWriter(os.Stdout, f)
			logrus.SetOutput(logLocations)
		} else {
			logrus.SetOutput(f)
		}
	}

	return nil
}

func setLevelStr(level string, log bool) {
	switch level {
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	default:
		level = "info (default)"

		logrus.SetLevel(logrus.InfoLevel)
	}

	if log {
		logrus.Infof("Log severity set to %s", level)
	}
}
