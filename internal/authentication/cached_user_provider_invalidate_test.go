// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authentication

import (
	"errors"
	"sync/atomic"
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

func TestCachedUserProviderShouldNotCacheReadWhichOverlapsARefresh(t *testing.T) {
	type read func(provider *CachedUserProvider) ([]string, error)

	readExtended := func(provider *CachedUserProvider) ([]string, error) {
		details, err := provider.GetDetailsExtendedCached("john")
		if err != nil {
			return nil, err
		}

		return details.Groups, nil
	}

	readDetails := func(provider *CachedUserProvider) ([]string, error) {
		details, err := provider.GetDetailsCached("john")
		if err != nil {
			return nil, err
		}

		return details.Groups, nil
	}

	refreshDetails := func(provider *CachedUserProvider) error {
		_, err := provider.GetDetails("john")

		return err
	}

	refreshExtended := func(provider *CachedUserProvider) error {
		_, err := provider.GetDetailsExtended("john")

		return err
	}

	testCases := []struct {
		name    string
		read    read
		refresh func(provider *CachedUserProvider) error
	}{
		{"ShouldHandleExtendedReadAndDetailsRefresh", readExtended, refreshDetails},
		{"ShouldHandleExtendedReadAndExtendedRefresh", readExtended, refreshExtended},
		{"ShouldHandleDetailsReadAndDetailsRefresh", readDetails, refreshDetails},
		{"ShouldHandleDetailsReadAndExtendedRefresh", readDetails, refreshExtended},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			backend := newBlockingUserProvider()

			provider := NewCachedUserProvider(backend, schema.NewRefreshIntervalDuration(5*time.Minute)).(*CachedUserProvider)

			done := make(chan []string, 1)

			go func() {
				groups, _ := tc.read(provider)

				done <- groups
			}()

			<-backend.started

			require.NoError(t, tc.refresh(provider))

			groups, err := tc.read(provider)
			require.NoError(t, err)
			assert.Equal(t, []string{"fresh"}, groups)

			close(backend.release)

			assert.Equal(t, []string{"stale"}, <-done)

			groups, err = tc.read(provider)
			require.NoError(t, err)
			assert.Equal(t, []string{"fresh"}, groups)
		})
	}
}

func TestCachedUserProviderShouldNotCacheReadWhichOverlapsAnInvalidation(t *testing.T) {
	backend := newBlockingUserProvider()

	provider := NewCachedUserProvider(backend, schema.NewRefreshIntervalDuration(5*time.Minute)).(*CachedUserProvider)

	done := make(chan struct{})

	go func() {
		defer close(done)

		_, _ = provider.GetDetailsExtendedCached("john")
	}()

	<-backend.started

	provider.invalidate("john")

	close(backend.release)

	<-done

	assert.Empty(t, provider.extended.values)
	assert.Empty(t, provider.details.values)
}

type blockingUserProvider struct {
	mockCachedUserProvider

	blocked atomic.Bool
	started chan struct{}
	release chan struct{}
}

func newBlockingUserProvider() *blockingUserProvider {
	return &blockingUserProvider{
		mockCachedUserProvider: mockCachedUserProvider{
			details:         &UserDetails{Username: "john", Groups: []string{"fresh"}},
			detailsExtended: &UserDetailsExtended{UserDetails: &UserDetails{Username: "john", Groups: []string{"fresh"}}},
		},
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (p *blockingUserProvider) block() (blocked bool) {
	if !p.blocked.CompareAndSwap(false, true) {
		return false
	}

	close(p.started)

	<-p.release

	return true
}

func (p *blockingUserProvider) GetDetails(username string) (details *UserDetails, err error) {
	if p.block() {
		return &UserDetails{Username: "john", Groups: []string{"stale"}}, nil
	}

	return &UserDetails{Username: "john", Groups: []string{"fresh"}}, nil
}

func (p *blockingUserProvider) GetDetailsExtended(username string) (details *UserDetailsExtended, err error) {
	if p.block() {
		return &UserDetailsExtended{UserDetails: &UserDetails{Username: "john", Groups: []string{"stale"}}}, nil
	}

	return &UserDetailsExtended{UserDetails: &UserDetails{Username: "john", Groups: []string{"fresh"}}}, nil
}
