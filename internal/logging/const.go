// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"regexp"
	"sync"

	"github.com/sirupsen/logrus"
)

// Log Format values.
const (
	FormatText = "text"
	FormatJSON = "json"
)

// LogLevel represents a log level in the configuration.
type LogLevel string

// Log Level values.
const (
	LevelTrace = "trace"
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// Level returns the logrus.Level for this LogLevel.
func (l LogLevel) Level() logrus.Level {
	switch l {
	case LevelError:
		return logrus.ErrorLevel
	case LevelWarn:
		return logrus.WarnLevel
	case LevelInfo:
		return logrus.InfoLevel
	case LevelDebug:
		return logrus.DebugLevel
	case LevelTrace:
		return logrus.TraceLevel
	default:
		return logrus.InfoLevel
	}
}

// Field names.
const (
	FieldRemoteIP            = "remote_ip"
	FieldMethod              = "method"
	FieldPath                = "path"
	FieldPathRaw             = "path_raw"
	FieldStatusCode          = "status_code"
	FieldFlowID              = "flow_id"
	FieldFlow                = "flow"
	FieldSubflow             = "subflow"
	FieldUsername            = "username"
	FieldSignature           = "signature"
	FieldClientID            = "client_id"
	FieldScope               = "scope"
	FieldGroups              = "groups"
	FieldExpiration          = "expiration"
	FieldSessionID           = "session_id"
	FieldRequestID           = "request_id"
	FieldAuthenticationLevel = "authentication_level"
	FieldAuthorizationPolicy = "authorization_policy"
	FieldSubject             = "subject"
	FieldResponded           = "responded"
	FieldGranted             = "granted"
	FieldStatus              = "status"
	FieldProvider            = "provider"
	FieldCaller              = "caller"
	FieldStack               = "stack"
)

// Stack trace hook values.
const (
	stackSkipFrames       = 8
	stackSkipFramesFields = 6
	stackMaxDepth         = 32

	pathPackageLogrus = "github.com/sirupsen/logrus"
)

var (
	stacktrace       sync.Once
	reFormatFilePath = regexp.MustCompile(`(%d|\{datetime(:([^}]+))?})`)
	logFile          *File
)
