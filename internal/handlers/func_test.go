package handlers

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

func AssertLogEntryMessageAndError(t *testing.T, entry *logrus.Entry, message, err string) {
	require.NotNil(t, entry)

	assert.Equal(t, message, entry.Message)

	v, ok := entry.Data["error"]

	if err == "" {
		assert.False(t, ok)
		assert.Nil(t, v)
	} else {
		assert.True(t, ok)
		require.NotNil(t, v)

		theErr, ok := v.(error)
		assert.True(t, ok)
		require.NotNil(t, theErr)

		assert.EqualError(t, theErr, err)
	}
}

func MustGetLogLastSeq(t *testing.T, hook *test.Hook, seq int) *logrus.Entry {
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
