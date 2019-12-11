package logging

import (
	logrus_stack "github.com/Gurpartap/logrus-stack"
	"github.com/sirupsen/logrus"
)

func init() {
	callerLevels := []logrus.Level{}
	stackLevels := []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel}
	logrus.AddHook(logrus_stack.NewHook(callerLevels, stackLevels))
}

// Logger return the standard logrues logger.
func Logger() *logrus.Logger {
	return logrus.StandardLogger()
}

// SetLevel set the level of the logger.
func SetLevel(level logrus.Level) {
	logrus.SetLevel(level)
}
