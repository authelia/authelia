// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestErrorMessagesShouldHaveUniqueCodes(t *testing.T) {
	require.Len(t, middlewares.ErrorMessages, 20)

	pattern := regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*$`)
	seen := map[string]bool{}

	for _, message := range middlewares.ErrorMessages {
		assert.Regexp(t, pattern, message.Code)
		assert.NotEmpty(t, message.Message, message.Code)
		assert.False(t, seen[message.Code], "duplicate code %s", message.Code)

		seen[message.Code] = true
	}
}

func TestSetJSONErrorShouldIncludeCode(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.SetJSONError(middlewares.ErrorMessagePasswordPolicy)

	assert.Equal(t, `{"status":"KO","code":"password_policy","message":"Your supplied password does not meet the password policy requirements."}`, string(mock.Ctx.Response.Body()))
}
