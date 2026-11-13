// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultStrategy_Exists(t *testing.T) {
	t.Run("ShouldReturnTrueForPersistedSession", func(t *testing.T) {
		strategy := newTestStrategy(t, nil)
		ctx := newTestContext()

		userSession := strategy.New(testUsername)

		require.NoError(t, strategy.Save(ctx, &userSession))
		require.NotEmpty(t, userSession.PublicID)

		exists, err := strategy.Exists(ctx, userSession.PublicID)

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("ShouldReturnFalseForEmptyPublicID", func(t *testing.T) {
		strategy := newTestStrategy(t, nil)

		exists, err := strategy.Exists(newTestContext(), "")

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("ShouldReturnFalseForUnknownPublicID", func(t *testing.T) {
		strategy := newTestStrategy(t, nil)

		exists, err := strategy.Exists(newTestContext(), "unknown")

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("ShouldReturnFalseForDestroyedSession", func(t *testing.T) {
		strategy := newTestStrategy(t, nil)
		ctx := newTestContext()

		userSession := strategy.New(testUsername)

		require.NoError(t, strategy.Save(ctx, &userSession))
		require.NoError(t, strategy.Destroy(ctx))

		exists, err := strategy.Exists(ctx, userSession.PublicID)

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("ShouldReturnTrueForRegeneratedSession", func(t *testing.T) {
		strategy := newTestStrategy(t, nil)
		ctx := newTestContext()

		userSession := strategy.New(testUsername)

		require.NoError(t, strategy.Save(ctx, &userSession))
		require.NoError(t, strategy.Regenerate(ctx))

		exists, err := strategy.Exists(ctx, userSession.PublicID)

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("ShouldReturnFalseForSessionDestroyedByAnotherRequest", func(t *testing.T) {
		repository := newTestRepository()
		strategy := newTestStrategyWithRepository(t, repository, nil)
		ctx := newTestCachingContext()

		userSession := strategy.New(testUsername)

		require.NoError(t, strategy.Save(ctx, &userSession))

		current, err := strategy.Get(ctx)
		require.NoError(t, err)
		require.Equal(t, testUsername, current.Username)

		other := newTestContext()
		other.cookies[testName] = ctx.cookies[testName]

		require.NoError(t, strategy.Destroy(other))

		current, err = strategy.Get(ctx)
		require.NoError(t, err)
		require.Equal(t, testUsername, current.Username)

		exists, err := strategy.Exists(ctx, userSession.PublicID)

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("ShouldReturnFalseForSessionWithoutData", func(t *testing.T) {
		repository := &publicIDRecordRepository{testRepository: newTestRepository(), record: NewRecord("id", nil)}
		strategy := newTestStrategyWithRepository(t, repository, nil)

		exists, err := strategy.Exists(newTestContext(), "pid")

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("ShouldReturnErrorWhenRepositoryFails", func(t *testing.T) {
		repository := &publicIDRecordRepository{testRepository: newTestRepository(), err: errTestFailure}
		strategy := newTestStrategyWithRepository(t, repository, nil)

		exists, err := strategy.Exists(newTestContext(), "pid")

		assert.ErrorIs(t, err, ErrRepositoryGet)
		assert.ErrorIs(t, err, errTestFailure)
		assert.False(t, exists)
	})
}

type publicIDRecordRepository struct {
	*testRepository

	record Record
	err    error
}

func (r *publicIDRecordRepository) GetByPublicID(_ context.Context, _, _ string) (record Record, err error) {
	return r.record, r.err
}
