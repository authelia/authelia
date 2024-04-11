package logging

import (
	"context"

	"github.com/sirupsen/logrus"
)

// PrintfLogger is a logger that implements a common Printf logger.
type PrintfLogger struct {
	level  logrus.Level
	logrus *logrus.Logger
}

// Printf is the implementation of the interface.
func (l *PrintfLogger) Printf(format string, args ...any) {
	l.logrus.Logf(l.level, format, args...)
}

// CtxPrintfLogger is a logger that implements a common Printf logger with a ctx.
type CtxPrintfLogger struct {
	level  logrus.Level
	logrus *logrus.Logger
}

// Printf is the implementation of the interface.
func (l *CtxPrintfLogger) Printf(_ context.Context, format string, args ...any) {
	l.logrus.Logf(l.level, format, args...)
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
