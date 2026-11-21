// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeveledLogger(t *testing.T) {
	testCases := []struct {
		name     string
		log      func(logger *LeveledLogger)
		expected logrus.Level
	}{
		{"ShouldLogErrorAsError", func(logger *LeveledLogger) { logger.Error("message", "url", "https://example.com") }, logrus.ErrorLevel},
		{"ShouldLogWarnAsWarn", func(logger *LeveledLogger) { logger.Warn("message", "url", "https://example.com") }, logrus.WarnLevel},
		{"ShouldLogInfoAsDebug", func(logger *LeveledLogger) { logger.Info("message", "url", "https://example.com") }, logrus.DebugLevel},
		{"ShouldLogDebugAsTrace", func(logger *LeveledLogger) { logger.Debug("message", "url", "https://example.com") }, logrus.TraceLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			base, hook := test.NewNullLogger()
			base.SetLevel(logrus.TraceLevel)

			tc.log(&LeveledLogger{logrus: base})

			entry := hook.LastEntry()

			require.NotNil(t, entry)
			assert.Equal(t, tc.expected, entry.Level)
			assert.Equal(t, "message", entry.Message)
			assert.Equal(t, logrus.Fields{"url": "https://example.com"}, entry.Data)
		})
	}
}

func TestLeveledLoggerFields(t *testing.T) {
	base, hook := test.NewNullLogger()

	logger := &LeveledLogger{logrus: base}

	logger.Error("message", "attempt", 2, 3, "odd")

	entry := hook.LastEntry()

	require.NotNil(t, entry)
	assert.Equal(t, logrus.Fields{"attempt": 2, "3": "odd"}, entry.Data)

	logger.Error("message", "dangling")

	entry = hook.LastEntry()

	require.NotNil(t, entry)
	assert.Equal(t, logrus.Fields{"dangling": nil}, entry.Data)
}

func TestNewLeveledLogger(t *testing.T) {
	assert.Equal(t, logrus.StandardLogger(), NewLeveledLogger().logrus)
}
