// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authentication

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestCachedUserProviderShouldInvalidateWhenTheUserIsNotFound(t *testing.T) {
	testCases := []struct {
		name        string
		call        func(provider *CachedUserProvider) error
		err         error
		invalidated bool
	}{
		{
			"ShouldInvalidateOnGetDetailsNotFound",
			func(provider *CachedUserProvider) error {
				_, err := provider.GetDetails("john")

				return err
			},
			ErrUserNotFound,
			true,
		},
		{
			"ShouldInvalidateOnGetDetailsExtendedNotFound",
			func(provider *CachedUserProvider) error {
				_, err := provider.GetDetailsExtended("john")

				return err
			},
			ErrUserNotFound,
			true,
		},
		{
			"ShouldNotInvalidateOnGetDetailsExtendedError",
			func(provider *CachedUserProvider) error {
				_, err := provider.GetDetailsExtended("john")

				return err
			},
			errors.New("an error"),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockCachedUserProvider{detailsErr: tc.err, detailsExtendedErr: tc.err}

			provider, ok := NewCachedUserProvider(mock, schema.NewRefreshIntervalDuration(5*time.Minute)).(*CachedUserProvider)

			require.True(t, ok)

			expires := time.Now().Add(time.Hour)

			provider.details.values["john"] = CachedUserDetailsItem{UserDetails: &UserDetails{Username: "john"}, expires: expires}
			provider.extended.values["john"] = CachedUserDetailsExtendedItem{UserDetailsExtended: &UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}}, expires: expires}

			assert.ErrorIs(t, tc.call(provider), tc.err)

			_, details := provider.details.values["john"]
			_, extended := provider.extended.values["john"]

			assert.Equal(t, !tc.invalidated, details)
			assert.Equal(t, !tc.invalidated, extended)
		})
	}
}
