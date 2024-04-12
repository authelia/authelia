package logging

import (
	"io"
	"os"
	"time"

	logrus_stack "github.com/Gurpartap/logrus-stack"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Logger returns the standard logrus logger.
func Logger() *logrus.Logger {
	return logrus.StandardLogger()
}

// InitializeLogger configures the default logger similar to ConfigureLogger but also configures the stack levels hook.
func InitializeLogger(config schema.Log, log bool) (err error) {
	initializeStackTracer(config)

	return ConfigureLogger(config, log)
}

func initializeStackTracer(config schema.Log) {
	// Ensure the stack trace hook is only initialized once.
	stacktrace.Do(func() {
		var (
			callerLevels, stackLevels []logrus.Level
		)

		switch LogLevel(config.Level).Level() {
		case logrus.DebugLevel:
			stackLevels = []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel}
		case logrus.TraceLevel:
			stackLevels = []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel}
			callerLevels = logrus.AllLevels
		default:
			stackLevels = []logrus.Level{logrus.PanicLevel, logrus.FatalLevel}
		}

		logrus.AddHook(logrus_stack.NewHook(callerLevels, stackLevels))
	})
}

// ConfigureLogger configures the default loggers level, formatting, and the output destinations.
func ConfigureLogger(config schema.Log, log bool) (err error) {
	setLevelStr(config.Level, log)

	switch config.Format {
	case FormatJSON:
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	var writers []io.Writer

	switch {
	case config.FilePath != "":
		var file *os.File

		if file, err = os.OpenFile(FormatFilePath(config.FilePath, time.Now()), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600); err != nil {
			return err
		}

		if config.Format != FormatJSON {
			logrus.SetFormatter(&logrus.TextFormatter{
				DisableColors: true,
				FullTimestamp: true,
			})
		}

		writers = []io.Writer{file}

		if config.KeepStdout {
			writers = append(writers, os.Stdout)
		}
	default:
		writers = []io.Writer{os.Stdout}
	}

	logrus.SetOutput(io.MultiWriter(writers...))

	return nil
}

func setLevelStr(level string, log bool) {
	logrus.SetLevel(LogLevel(level).Level())

	if log {
		logrus.Infof("Log severity set to %s", level)
	}
}
