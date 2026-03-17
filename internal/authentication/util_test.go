// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authentication

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustGetUserDetailsSafe(t *testing.T) {
	testCases := []struct {
		name        string
		username    string
		details     *UserDetails
		err         error
		expected    UserDetails
		expectedErr string
		calls       int
	}{
		{"ShouldNotConsultTheProviderWhenAnonymous", "", &UserDetails{Username: "john"}, nil, UserDetails{}, "", 0},
		{"ShouldReturnEmptyDetailsOnError", "john", nil, errors.New("an error"), UserDetails{}, "an error", 1},
		{"ShouldReturnDetails", "john", &UserDetails{Username: "john"}, nil, UserDetails{Username: "john"}, "", 1},
		{"ShouldReturnDetailsAlongsideAnError", "john", &UserDetails{Username: "john"}, errors.New("an error"), UserDetails{Username: "john"}, "an error", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := &mockCachedUserProvider{details: tc.details, detailsErr: tc.err}

			actual, err := MustGetUserDetailsSafe(tc.username, provider)

			if tc.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr)
			}

			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.calls, provider.detailsCalls)
		})
	}
}

func TestMustGetUserDetailsExtendedSafe(t *testing.T) {
	functions := []struct {
		name string
		fn   func(username string, provider UserProvider) (UserDetailsExtended, error)
	}{
		{"Extended", MustGetUserDetailsExtendedSafe},
		{"ExtendedCached", MustGetUserDetailsExtendedCachedSafe},
	}

	testCases := []struct {
		name        string
		username    string
		details     *UserDetailsExtended
		err         error
		expected    UserDetailsExtended
		expectedErr string
		calls       int
	}{
		{"ShouldNotConsultTheProviderWhenAnonymous", "", &UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}}, nil, UserDetailsExtended{UserDetails: &UserDetails{}}, "", 0},
		{"ShouldReturnEmptyDetailsOnError", "john", nil, errors.New("an error"), UserDetailsExtended{UserDetails: &UserDetails{}}, "an error", 1},
		{"ShouldReturnDetails", "john", &UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}}, nil, UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}}, "", 1},
		{"ShouldReturnDetailsAlongsideAnError", "john", &UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}}, errors.New("an error"), UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}}, "an error", 1},
	}

	for _, f := range functions {
		t.Run(f.name, func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					provider := &mockCachedUserProvider{detailsExtended: tc.details, detailsExtendedErr: tc.err}

					actual, err := f.fn(tc.username, provider)

					if tc.expectedErr == "" {
						assert.NoError(t, err)
					} else {
						assert.EqualError(t, err, tc.expectedErr)
					}

					assert.Equal(t, tc.expected, actual)
					assert.Equal(t, tc.calls, provider.detailsExtendedCalls)
				})
			}
		})
	}
}
