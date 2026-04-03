package logging

import (
	"fmt"
	"io"
	"os"

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

	// Ensure the stack trace hook is only initialized once.
	stacktrace.Do(func() {
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
		lf = NewFile(config.FilePath)

		if err = lf.Open(); err != nil {
			return err
		}

		if config.Format != FormatJSON {
			logrus.SetFormatter(&logrus.TextFormatter{
				DisableColors: true,
				FullTimestamp: true,
			})
		}

		writers = []io.Writer{lf}

		if config.KeepStdout {
			writers = append(writers, os.Stdout)
		}
	default:
		writers = []io.Writer{os.Stdout}
	}

	logrus.SetOutput(io.MultiWriter(writers...))

	return nil
}

// Reopen handles safely reopening the log file.
func Reopen() (err error) {
	if lf == nil {
		return fmt.Errorf("error reopening log file: file is not configured or open")
	}

	return lf.Reopen()
}

func setLevelStr(level string, log bool) {
	logrus.SetLevel(LogLevel(level).Level())

	if log {
		logrus.Infof("Log severity set to %s", level)
	}
}
