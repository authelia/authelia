package handlers

import (
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

func AssertLogEntryMessageAndError(t *testing.T, entry *logrus.Entry, message, err any) {
	t.Helper()

	require.NotNil(t, entry)

	switch value := message.(type) {
	case *regexp.Regexp:
		assert.Regexp(t, value, entry.Message)
	case string:
		assert.Equal(t, value, entry.Message)
	case nil:
		break
	default:
		t.Fatal("Message should be a string, nil, or *regexp.Regex")
	}

	v, ok := entry.Data["error"]

	switch value := err.(type) {
	case *regexp.Regexp:
		assert.True(t, ok)

		theErr, ok := v.(error)
		assert.True(t, ok)
		require.NotNil(t, theErr)

		assert.Regexp(t, value, theErr.Error())
	case string:
		if value == "" {
			assert.False(t, ok)
			assert.Nil(t, v)

			break
		}

		assert.True(t, ok)

		theErr, ok := v.(error)
		assert.True(t, ok)
		require.NotNil(t, theErr)

		assert.EqualError(t, theErr, value)
	case nil:
		assert.False(t, ok)
		assert.Nil(t, v)
	default:
		t.Fatal("Err should be a string, nil, or *regexp.Regex")
	}
}

func MustGetLogLastSeq(t *testing.T, hook *test.Hook, seq int) *logrus.Entry {
	t.Helper()

	require.Greater(t, len(hook.Entries), seq)

	return &hook.Entries[len(hook.Entries)-1-seq]
}

//nolint:unparam
func getStepTOTP(ctx *middlewares.AutheliaCtx, period int) uint64 {
	step := ctx.Clock.Now().Unix()

	if period < 0 {
		period = ctx.Configuration.TOTP.DefaultPeriod
	}

	return uint64(int(step) / period)
}
