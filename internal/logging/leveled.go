// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// LeveledLogger is a logger that implements a common leveled logger which accepts alternating key and value pairs,
// such as the one used by retryablehttp. The levels are lowered by one for Info and Debug as these loggers report each
// request at these levels.
type LeveledLogger struct {
	logrus *logrus.Logger
}

// Error is the implementation of the interface.
func (l *LeveledLogger) Error(msg string, keysAndValues ...any) {
	l.log(logrus.ErrorLevel, msg, keysAndValues)
}

// Warn is the implementation of the interface.
func (l *LeveledLogger) Warn(msg string, keysAndValues ...any) {
	l.log(logrus.WarnLevel, msg, keysAndValues)
}

// Info is the implementation of the interface.
func (l *LeveledLogger) Info(msg string, keysAndValues ...any) {
	l.log(logrus.DebugLevel, msg, keysAndValues)
}

// Debug is the implementation of the interface.
func (l *LeveledLogger) Debug(msg string, keysAndValues ...any) {
	l.log(logrus.TraceLevel, msg, keysAndValues)
}

func (l *LeveledLogger) log(level logrus.Level, msg string, keysAndValues []any) {
	if !l.logrus.IsLevelEnabled(level) {
		return
	}

	fields := make(logrus.Fields, (len(keysAndValues)+1)/2)

	for i := 0; i < len(keysAndValues); i += 2 {
		var value any

		if i+1 < len(keysAndValues) {
			value = keysAndValues[i+1]
		}

		fields[fmt.Sprint(keysAndValues[i])] = value
	}

	l.logrus.WithFields(fields).Log(level, msg)
}

// NewLeveledLogger returns a new LeveledLogger which logs to the standard logger.
func NewLeveledLogger() (logger *LeveledLogger) {
	return &LeveledLogger{
		logrus: logrus.StandardLogger(),
	}
}
