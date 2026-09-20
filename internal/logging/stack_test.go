// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStackFrameString(t *testing.T) {
	frame := stackFrame{File: "/srv/authelia/internal/logging/stack.go", Line: 42, Name: "Fire"}

	assert.Equal(t, "/srv/authelia/internal/logging/stack.go:42 Fire", frame.String())
}

func TestStackTraceStringAlignsNames(t *testing.T) {
	testCases := []struct {
		name     string
		have     stackTrace
		expected string
	}{
		{
			"ShouldRenderEmpty",
			stackTrace{},
			"",
		},
		{
			"ShouldRenderSingleFrameWithSingleSpace",
			stackTrace{{File: "a.go", Line: 1, Name: "One"}},
			"a.go:1 One",
		},
		{
			"ShouldPadShorterFramesSoNamesAlign",
			stackTrace{
				{File: "short.go", Line: 1, Name: "One"},
				{File: "muchlonger.go", Line: 100, Name: "Two"},
			},
			"short.go:1        One\nmuchlonger.go:100 Two",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.String())
		})
	}
}

func TestStackStripPackage(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldStripModulePath", "github.com/authelia/authelia/v4/internal/logging.Fire", "Fire"},
		{"ShouldStripBuiltinPackage", "runtime.goexit", "goexit"},
		{"ShouldRetainMethodReceiver", "github.com/authelia/authelia/v4/internal/logging.stackHook.Fire", "stackHook.Fire"},
		{"ShouldReturnNameWithoutSeparator", "goexit", "goexit"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, stackStripPackage(tc.have))
		})
	}
}

func TestStackHookLevels(t *testing.T) {
	assert.Equal(t, logrus.AllLevels, stackHook{}.Levels())
}

func fireStackHook(t *testing.T, hook stackHook, fields logrus.Fields) logrus.Fields {
	t.Helper()

	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	logger.SetLevel(logrus.TraceLevel)
	logger.AddHook(hook)

	var captured logrus.Fields

	logger.AddHook(&stackCaptureHook{fields: &captured})

	if fields == nil {
		logger.Error("example")
	} else {
		logger.WithFields(fields).Error("example")
	}

	return captured
}

type stackCaptureHook struct {
	fields *logrus.Fields
}

func (h *stackCaptureHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *stackCaptureHook) Fire(entry *logrus.Entry) error {
	fields := logrus.Fields{}

	for k, v := range entry.Data {
		fields[k] = v
	}

	*h.fields = fields

	return nil
}

func TestStackHookFire(t *testing.T) {
	testCases := []struct {
		name      string
		hook      stackHook
		fields    logrus.Fields
		hasCaller bool
		hasStack  bool
	}{
		{
			"ShouldSetNeitherWhenLevelNotConfigured",
			stackHook{},
			nil,
			false,
			false,
		},
		{
			"ShouldSetCallerOnly",
			stackHook{CallerLevels: logrus.AllLevels},
			nil,
			true,
			false,
		},
		{
			"ShouldSetStackOnly",
			stackHook{StackLevels: []logrus.Level{logrus.ErrorLevel}},
			nil,
			false,
			true,
		},
		{
			"ShouldSetBoth",
			stackHook{CallerLevels: logrus.AllLevels, StackLevels: []logrus.Level{logrus.ErrorLevel}},
			nil,
			true,
			true,
		},
		{
			"ShouldSetBothWithFields",
			stackHook{CallerLevels: logrus.AllLevels, StackLevels: []logrus.Level{logrus.ErrorLevel}},
			logrus.Fields{"k": "v"},
			true,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fields := fireStackHook(t, tc.hook, tc.fields)

			caller, ok := fields[FieldCaller]
			assert.Equal(t, tc.hasCaller, ok)

			if tc.hasCaller {
				frame, ok := caller.(stackFrame)
				require.True(t, ok)

				assert.Equal(t, "stack_test.go", frame.File[strings.LastIndex(frame.File, "/")+1:])
				assert.NotEmpty(t, frame.Name)
			}

			trace, ok := fields[FieldStack]
			assert.Equal(t, tc.hasStack, ok)

			if tc.hasStack {
				frames, ok := trace.(stackTrace)
				require.True(t, ok)
				require.NotEmpty(t, frames)

				assert.Equal(t, "stack_test.go", frames[0].File[strings.LastIndex(frames[0].File, "/")+1:])

				for _, frame := range frames {
					assert.NotContains(t, frame.File, pathPackageLogrus)
				}
			}
		})
	}
}

func TestStackHookJSONShape(t *testing.T) {
	fields := fireStackHook(t, stackHook{CallerLevels: logrus.AllLevels, StackLevels: []logrus.Level{logrus.ErrorLevel}}, nil)

	data, err := json.Marshal(fields[FieldCaller])
	require.NoError(t, err)

	caller := map[string]any{}
	require.NoError(t, json.Unmarshal(data, &caller))

	assert.Contains(t, caller, "File")
	assert.Contains(t, caller, "Line")
	assert.Contains(t, caller, "Name")

	if data, err = json.Marshal(fields[FieldStack]); err != nil {
		require.NoError(t, err)
	}

	var trace []map[string]any

	require.NoError(t, json.Unmarshal(data, &trace))
	require.NotEmpty(t, trace)

	assert.Contains(t, trace[0], "File")
	assert.Contains(t, trace[0], "Line")
	assert.Contains(t, trace[0], "Name")
}
