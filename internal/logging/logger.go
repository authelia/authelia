package logging

import (
	"io"
	"os"
	"regexp"
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

// InitializeLogger configures the default logger similar to ConfigureLogger but also configures the stack levels hook.
func InitializeLogger(config schema.Log, log bool) (err error) {
	var callerLevels []logrus.Level

	stackLevels := []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel}

	logrus.AddHook(logrus_stack.NewHook(callerLevels, stackLevels))

	return ConfigureLogger(config, log)
}

var reFormatFilePath = regexp.MustCompile(`(%d|\{datetime(:([^}]+))?})`)

func FormatFilePath(in string, now time.Time) (out string) {
	matches := reFormatFilePath.FindStringSubmatch(in)

	if len(matches) == 0 {
		return in
	}

	layout := time.RFC3339

	if len(matches[3]) != 0 {
		layout = matches[3]
	}

	return strings.Replace(in, matches[0], now.Format(layout), 1)
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
	switch level {
	case LevelError:
		logrus.SetLevel(logrus.ErrorLevel)
	case LevelWarn:
		logrus.SetLevel(logrus.WarnLevel)
	case LevelInfo:
		logrus.SetLevel(logrus.InfoLevel)
	case LevelDebug:
		logrus.SetLevel(logrus.DebugLevel)
	case LevelTrace:
		logrus.SetLevel(logrus.TraceLevel)
	default:
		level = "info (default)"

		logrus.SetLevel(logrus.InfoLevel)
	}

	if log {
		logrus.Infof("Log severity set to %s", level)
	}
}
