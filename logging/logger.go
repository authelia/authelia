package logging

import (
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

// Logger return the standard logrues logger.
func Logger() *logrus.Logger {
	return logrus.StandardLogger()
}

// NewRequestLogger create a new request logger for the given request.
func NewRequestLogger(ctx *fasthttp.RequestCtx) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"method":    string(ctx.Method()),
		"path":      string(ctx.Path()),
		"remote_ip": ctx.RemoteIP().String(),
	})
}

// SetLevel set the level of the logger.
func SetLevel(level logrus.Level) {
	logrus.SetLevel(level)
}
