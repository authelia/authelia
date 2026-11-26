// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultStrategy_SaveShouldLogTheFailureToSaveTheAuthzIndex(t *testing.T) {
	testCases := []struct {
		name    string
		err     error
		level   logrus.Level
		message string
	}{
		{
			"ShouldLogAnErrorAtTheErrorLevel",
			errTestFailure,
			logrus.ErrorLevel,
			"Error occurred saving the session authz index, the authz cookie will not be issued",
		},
		{
			"ShouldLogASupersededSessionAtTheDebugLevel",
			ErrSessionSuperseded,
			logrus.DebugLevel,
			"Session authz index was not saved as the session has been destroyed or moved by another request, the authz cookie will not be issued",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := &failingIndexSaveRepository{testRepository: newTestRepository(), err: tc.err}
			strategy := newTestStrategyWithRepository(t, repository, nil)

			saved := newTestSavedContext(t, strategy, testUsername)

			repository.fail = true

			logger, hook := test.NewNullLogger()
			logger.SetLevel(logrus.DebugLevel)

			ctx := &testLoggingContext{testContext: newTestContext(), logger: logrus.NewEntry(logger)}
			ctx.cookies[testNameSession] = saved.cookies[testNameSession]

			userSession, err := strategy.Get(ctx)
			require.NoError(t, err)

			assert.NotContains(t, ctx.cookies, testNameAuthz)

			entry := hook.LastEntry()

			require.NotNil(t, entry)
			assert.Equal(t, tc.level, entry.Level)
			assert.Equal(t, tc.message, entry.Message)
			assert.ErrorIs(t, entry.Data[logrus.ErrorKey].(error), tc.err)
			assert.Equal(t, testUsername, userSession.Username)
		})
	}
}

func TestDefaultStrategy_ShouldLogToTheGlobalLoggerWithoutALoggingContext(t *testing.T) {
	assert.NotNil(t, logContext(newTestContext()))
}

type testLoggingContext struct {
	*testContext

	logger *logrus.Entry
}

func (c *testLoggingContext) GetLogger() *logrus.Entry {
	return c.logger
}
