package cmd

import (
	log "github.com/sirupsen/logrus"
)

var logLevel string

func levelStringToLevel(level string) log.Level {
	switch level {
	case "debug":
		return log.DebugLevel
	case "warning":
		return log.WarnLevel
	}

	return log.InfoLevel
}
