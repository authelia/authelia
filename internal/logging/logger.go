package logging

import (
	"io"
	"os"

	logrus_stack "github.com/Gurpartap/logrus-stack"
	"github.com/sirupsen/logrus"
)

// Logger returns the standard logrus logger.
func Logger() *logrus.Logger {
	return logrus.StandardLogger()
}

// SetLevelStr sets the logrus.Level of the default logger when provided a valid string.
func SetLevelStr(level string) {
	switch level {
	case "error":
		logrus.Info("Log severity set to error")
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn":
		logrus.Info("Log severity set to warn")
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.Info("Log severity set to info")
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.Info("Log severity set to debug")
		logrus.SetLevel(logrus.DebugLevel)
	case "trace":
		logrus.Info("Log severity set to trace")
		logrus.SetLevel(logrus.TraceLevel)
	}
}

// InitializeLogger configures the default loggers stack levels, formatting, and the output destinations.
func InitializeLogger(format, filename string, stdout bool) error {
	callerLevels := []logrus.Level{}
	stackLevels := []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel}
	logrus.AddHook(logrus_stack.NewHook(callerLevels, stackLevels))

	if format == logFormatJSON {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	if filename != "" {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

		if err != nil {
			return err
		}

		if format != logFormatJSON {
			logrus.SetFormatter(&logrus.TextFormatter{
				DisableColors: true,
				FullTimestamp: true,
			})
		}

		if stdout {
			logLocations := io.MultiWriter(os.Stdout, f)
			logrus.SetOutput(logLocations)
		} else {
			logrus.SetOutput(f)
		}
	}

	return nil
}
