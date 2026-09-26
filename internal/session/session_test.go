// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncapsulatedSession_ShouldDelegateToTheStrategy(t *testing.T) {
	ctx := newTestContext()
	manager := NewEncapsulatedSession(newTestStrategy(t, nil), ctx)

	config := manager.GetSessionConfig()

	assert.Equal(t, testDomain, config.Domain)
	assert.Equal(t, testName, config.Name)

	userSession := manager.NewDefaultUserSession()

	assert.True(t, userSession.IsAnonymous())
	assert.Equal(t, testDomain, userSession.CookieDomain)

	userSession = NewUserSession(testUsername)

	require.NoError(t, manager.SaveSession(&userSession))

	original := ctx.cookies[testName]
	require.NotEmpty(t, original)

	actual, err := manager.GetSession()

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)

	require.NoError(t, manager.RegenerateSession())

	regenerated := ctx.cookies[testName]

	assert.NotEqual(t, original, regenerated)

	actual, err = manager.GetSession()

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)

	require.NoError(t, manager.DestroySession())

	assert.NotContains(t, ctx.cookies, testName)

	actual, err = manager.GetSession()

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
}

func TestEncapsulatedSession_GetSessionShouldReturnDefaultSessionAndErrorWhenTheBackendFails(t *testing.T) {
	ctx := newTestContext()
	manager := NewEncapsulatedSession(newTestStrategyWithRepository(t, &failingRepository{testRepository: newTestRepository()}, nil), ctx)

	ctx.cookies[testName] = newTestCookie("an-identifier-which-cannot-be-retrieved")

	userSession, err := manager.GetSession()

	assert.ErrorIs(t, err, ErrRepositoryGet)
	assert.True(t, userSession.IsAnonymous())
	assert.Equal(t, testDomain, userSession.CookieDomain)
}
