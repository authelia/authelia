package logging

import (
	"io"
	"os"

	logrus_stack "github.com/Gurpartap/logrus-stack"
	"github.com/sirupsen/logrus"
)

// Logger return the standard logrus logger.
func Logger() *logrus.Logger {
	return logrus.StandardLogger()
}

// SetLevel set the level of the logger.
func SetLevel(level logrus.Level) {
	logrus.SetLevel(level)
}

// InitializeLogger initialize logger.
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
