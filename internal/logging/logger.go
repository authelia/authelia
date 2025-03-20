package logging

import (
	"fmt"
	"io"
	"os"

	logrus_stack "github.com/Gurpartap/logrus-stack"
	"github.com/sirupsen/logrus"
)

// Logger returns the standard logrus logger.
func Logger() *logrus.Logger {
	return logrus.StandardLogger()
}

// InitializeLogger configures the default logger similar to ConfigureLogger but also configures the stack levels hook.
func InitializeLogger(level, filePath, format string, keepStdout, log bool) (err error) {
	initializeStackTracer(level)

	return ConfigureLogger(level, filePath, format, keepStdout, log)
}

func initializeStackTracer(level string) {
	// Ensure the stack trace hook is only initialized once.
	stacktrace.Do(func() {
		var (
			callerLevels, stackLevels []logrus.Level
		)

		switch LogLevel(level).Level() {
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
func ConfigureLogger(level, filePath, format string, keepStdout, log bool) (err error) {
	setLevelStr(level, log)

	switch format {
	case FormatJSON:
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	var writers []io.Writer

	switch {
	case filePath != "":
		lf = NewFile(filePath)

		if err = lf.Open(); err != nil {
			return err
		}

		if format != FormatJSON {
			logrus.SetFormatter(&logrus.TextFormatter{
				DisableColors: true,
				FullTimestamp: true,
			})
		}

		writers = []io.Writer{lf}

		if keepStdout {
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
