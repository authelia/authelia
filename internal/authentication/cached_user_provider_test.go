package authentication

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewCachedUserProvider(t *testing.T) {
	testCases := []struct {
		name       string
		lifespan   schema.RefreshIntervalDuration
		expectWrap bool
	}{
		{
			"ShouldReturnProviderDirectlyWhenAlways",
			schema.NewRefreshIntervalDurationAlways(),
			false,
		},
		{
			"ShouldReturnCachedProviderWhenNormal",
			schema.NewRefreshIntervalDuration(5 * time.Minute),
			true,
		},
		{
			"ShouldReturnCachedProviderWhenNever",
			schema.NewRefreshIntervalDurationNever(),
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockCachedUserProvider{
				details: &UserDetails{Username: "john"},
			}

			provider := NewCachedUserProvider(mock, tc.lifespan)

			if tc.expectWrap {
				cached, ok := provider.(*CachedUserProvider)

				require.True(t, ok)
				assert.Equal(t, mock, cached.UserProvider)
			} else {
				assert.Equal(t, mock, provider)
			}
		})
	}
}

func TestCachedUserProvider_GetDetails(t *testing.T) {
	testCases := []struct {
		name          string
		lifespan      schema.RefreshIntervalDuration
		details       *UserDetails
		detailsErr    error
		setup         func(provider *CachedUserProvider)
		expectedCalls int
		expectedErr   string
		expected      *UserDetails
	}{
		{
			"ShouldReturnDetailsFromProvider",
			schema.NewRefreshIntervalDuration(5 * time.Minute),
			&UserDetails{
				Username:    "john",
				DisplayName: "John Doe",
				Emails:      []string{"john@example.com"},
				Groups:      []string{"admin"},
			},
			nil,
			nil,
			1,
			"",
			nil,
		},
		{
			"ShouldReturnFreshDetailsWhenExpired",
			schema.NewRefreshIntervalDuration(5 * time.Minute),
			&UserDetails{Username: "john"},
			nil,
			func(provider *CachedUserProvider) {
				provider.details.Lock()
				provider.details.values["john"] = CachedUserDetailsItem{
					UserDetails: &UserDetails{Username: "john"},
					expires:     time.Now().Add(-1 * time.Minute),
				}
				provider.details.Unlock()
			},
			1,
			"",
			nil,
		},
		{
			"ShouldReturnErrorAndDeleteCacheOnProviderError",
			schema.NewRefreshIntervalDuration(5 * time.Minute),
			nil,
			fmt.Errorf("provider error"),
			func(provider *CachedUserProvider) {
				provider.details.Lock()
				provider.details.values["john"] = CachedUserDetailsItem{
					UserDetails: &UserDetails{Username: "john"},
					expires:     time.Now().Add(-1 * time.Minute),
				}
				provider.details.Unlock()
			},
			1,
			"provider error",
			nil,
		},
		{
			"ShouldNeverExpireWithNeverLifespan",
			schema.NewRefreshIntervalDurationNever(),
			&UserDetails{Username: "john", DisplayName: "FromProvider"},
			nil,
			func(provider *CachedUserProvider) {
				provider.details.Lock()
				provider.details.values["john"] = CachedUserDetailsItem{
					UserDetails: &UserDetails{Username: "john", DisplayName: "Cached"},
					expires:     time.Now().Add(-1 * time.Hour),
				}
				provider.details.Unlock()
			},
			0,
			"",
			&UserDetails{Username: "john", DisplayName: "Cached"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockCachedUserProvider{
				details:    tc.details,
				detailsErr: tc.detailsErr,
			}

			provider := NewCachedUserProvider(mock, tc.lifespan).(*CachedUserProvider)

			if tc.setup != nil {
				tc.setup(provider)
			}

			details, err := provider.GetDetailsCached("john")

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)

				provider.details.Lock()
				_, exists := provider.details.values["john"]
				provider.details.Unlock()

				assert.False(t, exists)
			} else {
				expected := tc.expected
				if expected == nil {
					expected = tc.details
				}

				require.NoError(t, err)
				assert.Equal(t, expected, details)
				assert.Equal(t, tc.expectedCalls, mock.detailsCalls)
			}
		})
	}

	t.Run("ShouldReturnCachedDetailsOnSecondCall", func(t *testing.T) {
		mock := &mockCachedUserProvider{
			details: &UserDetails{Username: "john"},
		}

		provider := NewCachedUserProvider(mock, schema.NewRefreshIntervalDuration(5*time.Minute)).(*CachedUserProvider)

		_, err := provider.GetDetailsCached("john")
		require.NoError(t, err)

		_, err = provider.GetDetailsCached("john")
		require.NoError(t, err)

		assert.Equal(t, 1, mock.detailsCalls)
	})

	t.Run("ShouldCacheDifferentUsers", func(t *testing.T) {
		mock := &mockCachedUserProvider{}

		provider := NewCachedUserProvider(mock, schema.NewRefreshIntervalDuration(5*time.Minute)).(*CachedUserProvider)

		mock.details = &UserDetails{Username: "john"}

		_, err := provider.GetDetailsCached("john")
		require.NoError(t, err)

		callCount := mock.detailsCalls

		mock.details = &UserDetails{Username: "jane"}

		_, err = provider.GetDetailsCached("jane")
		require.NoError(t, err)

		assert.Equal(t, callCount+1, mock.detailsCalls)

		// Both should now be cached.
		_, err = provider.GetDetailsCached("john")
		require.NoError(t, err)

		_, err = provider.GetDetailsCached("jane")
		require.NoError(t, err)

		assert.Equal(t, callCount+1, mock.detailsCalls)
	})
}

func TestCachedUserProvider_GetDetailsExtended(t *testing.T) {
	testCases := []struct {
		name               string
		lifespan           schema.RefreshIntervalDuration
		detailsExtended    *UserDetailsExtended
		detailsExtendedErr error
		setup              func(provider *CachedUserProvider)
		expectedCalls      int
		expectedErr        string
		expected           *UserDetailsExtended
	}{
		{
			"ShouldReturnDetailsFromProvider",
			schema.NewRefreshIntervalDuration(5 * time.Minute),
			&UserDetailsExtended{
				GivenName:  "John",
				FamilyName: "Doe",
				UserDetails: &UserDetails{
					Username: "john",
				},
			},
			nil,
			nil,
			1,
			"",
			nil,
		},
		{
			"ShouldReturnFreshDetailsWhenExpired",
			schema.NewRefreshIntervalDuration(5 * time.Minute),
			&UserDetailsExtended{
				GivenName:   "John",
				UserDetails: &UserDetails{Username: "john"},
			},
			nil,
			func(provider *CachedUserProvider) {
				provider.extended.Lock()
				provider.extended.values["john"] = CachedUserDetailsExtendedItem{
					UserDetailsExtended: &UserDetailsExtended{
						GivenName:   "Old",
						UserDetails: &UserDetails{Username: "john"},
					},
					expires: time.Now().Add(-1 * time.Minute),
				}
				provider.extended.Unlock()
			},
			1,
			"",
			nil,
		},
		{
			"ShouldReturnErrorAndDeleteCacheOnProviderError",
			schema.NewRefreshIntervalDuration(5 * time.Minute),
			nil,
			fmt.Errorf("provider error"),
			func(provider *CachedUserProvider) {
				provider.extended.Lock()
				provider.extended.values["john"] = CachedUserDetailsExtendedItem{
					UserDetailsExtended: &UserDetailsExtended{
						UserDetails: &UserDetails{Username: "john"},
					},
					expires: time.Now().Add(-1 * time.Minute),
				}
				provider.extended.Unlock()
			},
			1,
			"provider error",
			nil,
		},
		{
			"ShouldNeverExpireWithNeverLifespan",
			schema.NewRefreshIntervalDurationNever(),
			&UserDetailsExtended{
				GivenName:   "FromProvider",
				UserDetails: &UserDetails{Username: "john"},
			},
			nil,
			func(provider *CachedUserProvider) {
				// Seed cache with an entry that has an expired timestamp.
				// With Never() lifespan, it should still be valid because the code checks p.lifespan.Never() first.
				provider.extended.Lock()
				provider.extended.values["john"] = CachedUserDetailsExtendedItem{
					UserDetailsExtended: &UserDetailsExtended{
						GivenName:   "Cached",
						UserDetails: &UserDetails{Username: "john"},
					},
					expires: time.Now().Add(-1 * time.Hour),
				}
				provider.extended.Unlock()
			},
			0,
			"",
			&UserDetailsExtended{
				GivenName:   "Cached",
				UserDetails: &UserDetails{Username: "john"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockCachedUserProvider{
				detailsExtended:    tc.detailsExtended,
				detailsExtendedErr: tc.detailsExtendedErr,
			}

			provider := NewCachedUserProvider(mock, tc.lifespan).(*CachedUserProvider)

			if tc.setup != nil {
				tc.setup(provider)
			}

			details, err := provider.GetDetailsExtendedCached("john")

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)

				provider.extended.Lock()
				_, exists := provider.extended.values["john"]
				provider.extended.Unlock()

				assert.False(t, exists)
			} else {
				expected := tc.expected
				if expected == nil {
					expected = tc.detailsExtended
				}

				require.NoError(t, err)
				assert.Equal(t, expected, details)
				assert.Equal(t, tc.expectedCalls, mock.detailsExtendedCalls)
			}
		})
	}

	t.Run("ShouldReturnCachedDetailsOnSecondCall", func(t *testing.T) {
		mock := &mockCachedUserProvider{
			detailsExtended: &UserDetailsExtended{
				UserDetails: &UserDetails{Username: "john"},
			},
		}

		provider := NewCachedUserProvider(mock, schema.NewRefreshIntervalDuration(5*time.Minute)).(*CachedUserProvider)

		_, err := provider.GetDetailsExtendedCached("john")
		require.NoError(t, err)

		_, err = provider.GetDetailsExtendedCached("john")
		require.NoError(t, err)

		assert.Equal(t, 1, mock.detailsExtendedCalls)
	})

	t.Run("ShouldCacheDifferentUsers", func(t *testing.T) {
		mock := &mockCachedUserProvider{}

		provider := NewCachedUserProvider(mock, schema.NewRefreshIntervalDuration(5*time.Minute)).(*CachedUserProvider)

		mock.detailsExtended = &UserDetailsExtended{
			GivenName:   "John",
			UserDetails: &UserDetails{Username: "john"},
		}

		_, err := provider.GetDetailsExtendedCached("john")
		require.NoError(t, err)

		callCount := mock.detailsExtendedCalls

		mock.detailsExtended = &UserDetailsExtended{
			GivenName:   "Jane",
			UserDetails: &UserDetails{Username: "jane"},
		}

		_, err = provider.GetDetailsExtendedCached("jane")
		require.NoError(t, err)

		assert.Equal(t, callCount+1, mock.detailsExtendedCalls)

		// Both should now be cached.
		_, err = provider.GetDetailsExtendedCached("john")
		require.NoError(t, err)

		_, err = provider.GetDetailsExtendedCached("jane")
		require.NoError(t, err)

		assert.Equal(t, callCount+1, mock.detailsExtendedCalls)
	})
}

type mockUserProviderConcurrent struct {
	UserProvider

	valid bool
	err   error
}

func (m *mockUserProviderConcurrent) GetDetailsExtendedCached(username string) (*UserDetailsExtended, error) {
	// TODO implement me.
	panic("implement me")
}

func (m *mockUserProviderConcurrent) CheckUserPassword(username, password string) (bool, error) {
	return m.valid, m.err
}

type mockUserProvider struct {
	UserProvider

	valid bool
	err   error
	calls int
}

func (m *mockUserProvider) GetDetailsExtendedCached(username string) (*UserDetailsExtended, error) {
	// TODO implement me.
	panic("implement me")
}

func (m *mockUserProvider) CheckUserPassword(username, password string) (bool, error) {
	m.calls++

	return m.valid, m.err
}

type mockContext struct {
	context.Context

	provider UserProvider
	clk      clock.Provider
	logger   *logrus.Entry
}

func (m *mockContext) GetUserProvider() UserProvider {
	return m.provider
}

func (m *mockContext) GetClock() clock.Provider {
	return m.clk
}

func (m *mockContext) GetLogger() *logrus.Entry {
	if m.logger != nil {
		return m.logger
	}

	l, _ := test.NewNullLogger()

	return logrus.NewEntry(l)
}

type mockCachedUserProvider struct {
	UserProvider

	detailsCalls         int
	detailsExtendedCalls int
	details              *UserDetails
	detailsExtended      *UserDetailsExtended
	detailsErr           error
	detailsExtendedErr   error
}

func (m *mockCachedUserProvider) GetDetails(username string) (*UserDetails, error) {
	m.detailsCalls++

	return m.details, m.detailsErr
}

func (m *mockCachedUserProvider) GetDetailsCached(username string) (*UserDetails, error) {
	return m.GetDetails(username)
}

func (m *mockCachedUserProvider) GetDetailsExtended(username string) (*UserDetailsExtended, error) {
	m.detailsExtendedCalls++

	return m.detailsExtended, m.detailsExtendedErr
}

func (m *mockCachedUserProvider) GetDetailsExtendedCached(username string) (*UserDetailsExtended, error) {
	return m.GetDetailsExtended(username)
}

func TestCachedUserProviderSecurityEventRefresh(t *testing.T) {
	testCases := []struct {
		name             string
		lifespan         schema.RefreshIntervalDuration
		event            func(provider *CachedUserProvider) error
		expectedErr      string
		expectedExtended int
	}{
		{
			"ShouldInvalidateExtendedOnGetDetails",
			schema.NewRefreshIntervalDuration(5 * time.Minute),
			func(provider *CachedUserProvider) error {
				_, err := provider.GetDetails("john")

				return err
			},
			"",
			2,
		},
		{
			"ShouldInvalidateExtendedOnGetDetailsWhenNever",
			schema.NewRefreshIntervalDurationNever(),
			func(provider *CachedUserProvider) error {
				_, err := provider.GetDetails("john")

				return err
			},
			"",
			2,
		},
		{
			"ShouldInvalidateExtendedOnUserNotFound",
			schema.NewRefreshIntervalDuration(5 * time.Minute),
			func(provider *CachedUserProvider) error {
				_, err := provider.GetDetails("john")

				return err
			},
			"user not found",
			2,
		},
		{
			"ShouldRetainExtendedOnTransientError",
			schema.NewRefreshIntervalDuration(5 * time.Minute),
			func(provider *CachedUserProvider) error {
				_, err := provider.GetDetails("john")

				return err
			},
			"backend unreachable",
			1,
		},
		{
			"ShouldRetainExtendedWithoutASecurityEvent",
			schema.NewRefreshIntervalDuration(5 * time.Minute),
			nil,
			"",
			1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockCachedUserProvider{
				details:         &UserDetails{Username: "john", Groups: []string{"admin"}},
				detailsExtended: &UserDetailsExtended{UserDetails: &UserDetails{Username: "john", Groups: []string{"admin"}}},
			}

			switch tc.expectedErr {
			case "":
				break
			case "user not found":
				mock.details, mock.detailsErr = nil, ErrUserNotFound
			default:
				mock.details, mock.detailsErr = nil, fmt.Errorf("%s", tc.expectedErr)
			}

			provider := NewCachedUserProvider(mock, tc.lifespan).(*CachedUserProvider)

			_, err := provider.GetDetailsExtendedCached("john")
			require.NoError(t, err)
			require.Equal(t, 1, mock.detailsExtendedCalls)

			if tc.event != nil {
				err = tc.event(provider)

				if tc.expectedErr == "" {
					require.NoError(t, err)
				} else {
					require.EqualError(t, err, tc.expectedErr)
				}
			}

			_, err = provider.GetDetailsExtendedCached("john")
			require.NoError(t, err)

			assert.Equal(t, tc.expectedExtended, mock.detailsExtendedCalls)
		})
	}
}

func TestCachedUserProviderGetDetailsExtendedPopulatesBothCaches(t *testing.T) {
	mock := &mockCachedUserProvider{
		detailsExtended: &UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}},
	}

	provider := NewCachedUserProvider(mock, schema.NewRefreshIntervalDuration(5*time.Minute)).(*CachedUserProvider)

	_, err := provider.GetDetailsExtended("john")
	require.NoError(t, err)

	provider.details.Lock()
	cached, ok := provider.details.values["john"]
	provider.details.Unlock()

	require.True(t, ok)
	assert.Equal(t, "john", cached.Username)
	assert.Equal(t, 0, mock.detailsCalls)
}
