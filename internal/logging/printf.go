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
func (l *PrintfLogger) Printf(format string, args ...interface{}) {
	l.logrus.Logf(l.level, format, args...)
}

// CtxPrintfLogger is a logger that implements a common Printf logger with a ctx.
type CtxPrintfLogger struct {
	level  logrus.Level
	logrus *logrus.Logger
}

// Printf is the implementation of the interface.
func (l *CtxPrintfLogger) Printf(_ context.Context, format string, args ...interface{}) {
	l.logrus.Logf(l.level, format, args...)
}
