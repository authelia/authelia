package authentication

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
)

func TestNewCredentialCacheHMAC(t *testing.T) {
	testCases := []struct {
		name     string
		lifespan time.Duration
	}{
		{
			"ShouldCreateWithDefaultLifespan",
			5 * time.Minute,
		},
		{
			"ShouldCreateWithZeroLifespan",
			0,
		},
		{
			"ShouldCreateWithLargeLifespan",
			24 * time.Hour,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewCredentialCacheHMAC(sha256.New, tc.lifespan)

			require.NotNil(t, cache)
			assert.NotNil(t, cache.hash)
			assert.NotNil(t, cache.secret)
			assert.Equal(t, tc.lifespan, cache.lifespan)
			assert.NotNil(t, cache.values)
			assert.Len(t, cache.values, 0)
			assert.Len(t, cache.secret, sha256.BlockSize)
		})
	}
}

func TestCredentialCacheHMAC_Check(t *testing.T) {
	testCases := []struct {
		name           string
		username       string
		password       string
		checkValid     bool
		checkErr       error
		expectedValid  bool
		expectedCached bool
		expectedErr    bool
	}{
		{
			"ShouldReturnValidOnCorrectPassword",
			"john",
			"password123",
			true,
			nil,
			true,
			false,
			false,
		},
		{
			"ShouldReturnInvalidOnWrongPassword",
			"john",
			"wrongpassword",
			false,
			nil,
			false,
			false,
			false,
		},
		{
			"ShouldReturnErrorOnProviderError",
			"john",
			"password123",
			false,
			fmt.Errorf("connection error"),
			false,
			false,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

			mock := &mockUserProvider{
				valid: tc.checkValid,
				err:   tc.checkErr,
			}

			ctx := &mockContext{
				provider: mock,
				clk:      clock.NewFixed(time.Now()),
			}

			valid, cached, err := cache.Check(ctx, tc.username, tc.password)

			assert.Equal(t, tc.expectedValid, valid)
			assert.Equal(t, tc.expectedCached, cached)

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCredentialCacheHMAC_CheckCached(t *testing.T) {
	testCases := []struct {
		name           string
		lifespan       time.Duration
		timeBetween    time.Duration
		expectedCached bool
	}{
		{
			"ShouldReturnCachedOnSecondCheck",
			5 * time.Minute,
			1 * time.Minute,
			true,
		},
		{
			"ShouldNotReturnCachedWhenExpired",
			5 * time.Minute,
			10 * time.Minute,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewCredentialCacheHMAC(sha256.New, tc.lifespan)

			now := time.Now()
			clk := clock.NewFixed(now)

			mock := &mockUserProvider{valid: true}

			ctx := &mockContext{
				provider: mock,
				clk:      clk,
			}

			valid, cached, err := cache.Check(ctx, "john", "password123")
			require.NoError(t, err)
			assert.True(t, valid)
			assert.False(t, cached)
			assert.Equal(t, 1, mock.calls)

			clk.Set(now.Add(tc.timeBetween))

			valid, cached, err = cache.Check(ctx, "john", "password123")
			require.NoError(t, err)
			assert.True(t, valid)
			assert.Equal(t, tc.expectedCached, cached)

			if tc.expectedCached {
				assert.Equal(t, 1, mock.calls)
			} else {
				assert.Equal(t, 2, mock.calls)
			}
		})
	}
}

func TestCredentialCacheHMAC_CheckDifferentCredentials(t *testing.T) {
	testCases := []struct {
		name           string
		firstUsername  string
		firstPassword  string
		secondUsername string
		secondPassword string
		expectedCached bool
	}{
		{
			"ShouldNotReturnCachedForDifferentUser",
			"john",
			"password123",
			"jane",
			"password123",
			false,
		},
		{
			"ShouldNotReturnCachedForDifferentPassword",
			"john",
			"password123",
			"john",
			"differentpassword",
			false,
		},
		{
			"ShouldReturnCachedForSameCredentials",
			"john",
			"password123",
			"john",
			"password123",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

			mock := &mockUserProvider{valid: true}

			ctx := &mockContext{
				provider: mock,
				clk:      clock.NewFixed(time.Now()),
			}

			valid, cached, err := cache.Check(ctx, tc.firstUsername, tc.firstPassword)
			require.NoError(t, err)
			assert.True(t, valid)
			assert.False(t, cached)

			valid, cached, err = cache.Check(ctx, tc.secondUsername, tc.secondPassword)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCached, cached)
			assert.Equal(t, valid, true)
		})
	}
}

func TestCredentialCacheHMAC_Sum(t *testing.T) {
	testCases := []struct {
		name            string
		username        string
		password        string
		secondUsername  string
		secondPassword  string
		expectedSameSum bool
	}{
		{
			"ShouldProduceSameSumForSameInput",
			"john",
			"password",
			"john",
			"password",
			true,
		},
		{
			"ShouldProduceDifferentSumForDifferentPassword",
			"john",
			"password",
			"john",
			"other",
			false,
		},
		{
			"ShouldProduceDifferentSumForDifferentUsername",
			"john",
			"password",
			"jane",
			"password",
			false,
		},
		{
			"ShouldNotCollideOnConcatenationBoundary",
			"john",
			"pass",
			"ohn",
			"passj",
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

			hex1, sum1, err := cache.sum(tc.username, tc.password)
			require.NoError(t, err)
			assert.NotEmpty(t, hex1)
			assert.NotNil(t, sum1)

			hex2, sum2, err := cache.sum(tc.secondUsername, tc.secondPassword)
			require.NoError(t, err)

			if tc.expectedSameSum {
				assert.Equal(t, hex1, hex2)
				assert.Equal(t, sum1, sum2)
			} else {
				assert.NotEqual(t, hex1, hex2)
				assert.NotEqual(t, sum1, sum2)
			}
		})
	}
}

func TestCredentialCacheHMAC_Valid(t *testing.T) {
	testCases := []struct {
		name          string
		setup         func(cache *CredentialCacheHMAC, clk *clock.Fixed, sum []byte)
		expectedValid bool
		expectedOK    bool
	}{
		{
			"ShouldReturnNotFoundWhenEmpty",
			func(cache *CredentialCacheHMAC, clk *clock.Fixed, sum []byte) {},
			false,
			false,
		},
		{
			"ShouldReturnValidWhenCachedAndNotExpired",
			func(cache *CredentialCacheHMAC, clk *clock.Fixed, sum []byte) {
				cache.values["john"] = CachedCredential{
					expires: clk.Now().Add(5 * time.Minute),
					value:   sum,
				}
			},
			true,
			true,
		},
		{
			"ShouldReturnNotFoundWhenExpired",
			func(cache *CredentialCacheHMAC, clk *clock.Fixed, sum []byte) {
				cache.values["john"] = CachedCredential{
					expires: clk.Now().Add(-1 * time.Minute),
					value:   sum,
				}
			},
			false,
			false,
		},
		{
			"ShouldReturnInvalidWhenSumMismatch",
			func(cache *CredentialCacheHMAC, clk *clock.Fixed, sum []byte) {
				cache.values["john"] = CachedCredential{
					expires: clk.Now().Add(5 * time.Minute),
					value:   []byte("different-sum"),
				}
			},
			false,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

			clk := clock.NewFixed(time.Now())

			_, sum, err := cache.sum("john", "password123")
			require.NoError(t, err)

			ctx := &mockContext{clk: clk}

			tc.setup(cache, clk, sum)

			valid, ok := cache.valid(ctx, "john", sum)

			assert.Equal(t, tc.expectedValid, valid)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestCredentialCacheHMAC_ValidDeletesExpired(t *testing.T) {
	cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

	clk := clock.NewFixed(time.Now())

	_, sum, err := cache.sum("john", "password123")
	require.NoError(t, err)

	cache.values["john"] = CachedCredential{
		expires: clk.Now().Add(-1 * time.Minute),
		value:   sum,
	}

	ctx := &mockContext{clk: clk}

	_, ok := cache.valid(ctx, "john", sum)
	assert.False(t, ok)

	_, exists := cache.values["john"]
	assert.False(t, exists)
}

func TestCredentialCacheHMAC_Put(t *testing.T) {
	cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

	now := time.Now()
	clk := clock.NewFixed(now)
	ctx := &mockContext{clk: clk}

	_, sum, err := cache.sum("john", "password123")
	require.NoError(t, err)

	err = cache.put(ctx, "john", sum)
	require.NoError(t, err)

	entry, exists := cache.values["john"]
	require.True(t, exists)
	assert.Equal(t, sum, entry.value)
	assert.Equal(t, now.Add(5*time.Minute), entry.expires)
}

func TestCredentialCacheHMAC_CheckCacheInvalidatedOnPasswordChange(t *testing.T) {
	cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

	mock := &mockUserProvider{valid: true}
	clk := clock.NewFixed(time.Now())
	ctx := &mockContext{provider: mock, clk: clk}

	valid, cached, err := cache.Check(ctx, "john", "password123")
	require.NoError(t, err)
	assert.True(t, valid)
	assert.False(t, cached)
	assert.Equal(t, 1, mock.calls)

	valid, cached, err = cache.Check(ctx, "john", "differentpassword")
	require.NoError(t, err)
	assert.True(t, valid)
	assert.False(t, cached)
	assert.Equal(t, 2, mock.calls)
}

func TestCredentialCacheHMAC_CheckMultipleUsers(t *testing.T) {
	cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

	mock := &mockUserProvider{valid: true}
	clk := clock.NewFixed(time.Now())
	ctx := &mockContext{provider: mock, clk: clk}

	valid, cached, err := cache.Check(ctx, "john", "password1")
	require.NoError(t, err)
	assert.True(t, valid)
	assert.False(t, cached)

	valid, cached, err = cache.Check(ctx, "jane", "password2")
	require.NoError(t, err)
	assert.True(t, valid)
	assert.False(t, cached)
	assert.Equal(t, 2, mock.calls)

	valid, cached, err = cache.Check(ctx, "john", "password1")
	require.NoError(t, err)
	assert.True(t, valid)
	assert.True(t, cached)
	assert.Equal(t, 2, mock.calls)

	valid, cached, err = cache.Check(ctx, "jane", "password2")
	require.NoError(t, err)
	assert.True(t, valid)
	assert.True(t, cached)
	assert.Equal(t, 2, mock.calls)
}

func TestCredentialCacheHMAC_CheckConcurrent(t *testing.T) {
	cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

	mock := &mockUserProviderConcurrent{valid: true}
	clk := clock.NewFixed(time.Now())

	done := make(chan struct{}, 20)

	for i := 0; i < 20; i++ {
		go func(idx int) {
			defer func() { done <- struct{}{} }()

			ctx := &mockContext{provider: mock, clk: clk}
			username := fmt.Sprintf("user%d", idx%5)

			valid, _, err := cache.Check(ctx, username, "password")

			assert.NoError(t, err)
			assert.True(t, valid)
		}(i)
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestCredentialCacheHMAC_PutOverwrite(t *testing.T) {
	cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

	now := time.Now()
	clk := clock.NewFixed(now)
	ctx := &mockContext{clk: clk}

	_, sum1, err := cache.sum("john", "password1")
	require.NoError(t, err)

	require.NoError(t, cache.put(ctx, "john", sum1))

	entry1 := cache.values["john"]

	_, sum2, err := cache.sum("john", "password2")
	require.NoError(t, err)

	clk.Set(now.Add(time.Minute))

	require.NoError(t, cache.put(ctx, "john", sum2))

	entry2 := cache.values["john"]

	assert.NotEqual(t, entry1.value, entry2.value)
	assert.True(t, entry2.expires.After(entry1.expires))
}

func TestCredentialCacheHMAC_SumDeterministic(t *testing.T) {
	cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

	hex1, sum1, err := cache.sum("user", "pass")
	require.NoError(t, err)

	hex2, sum2, err := cache.sum("user", "pass")
	require.NoError(t, err)

	assert.Equal(t, hex1, hex2)
	assert.Equal(t, sum1, sum2)
}

func TestCredentialCacheHMAC_SumDifferentSecrets(t *testing.T) {
	cache1 := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)
	cache2 := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

	hex1, _, err := cache1.sum("john", "password")
	require.NoError(t, err)

	hex2, _, err := cache2.sum("john", "password")
	require.NoError(t, err)

	assert.NotEqual(t, hex1, hex2)
}

func TestCredentialCacheHMAC_SumEmptyInputs(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"ShouldHandleEmptyUsername", "", "password"},
		{"ShouldHandleEmptyPassword", "username", ""},
		{"ShouldHandleBothEmpty", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

			hex, sum, err := cache.sum(tc.username, tc.password)

			assert.NoError(t, err)
			assert.NotEmpty(t, hex)
			assert.NotNil(t, sum)
		})
	}
}

func TestCredentialCacheHMAC_ValidMultipleEntries(t *testing.T) {
	cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)
	clk := clock.NewFixed(time.Now())
	ctx := &mockContext{clk: clk}

	_, sumJohn, err := cache.sum("john", "password1")
	require.NoError(t, err)

	_, sumJane, err := cache.sum("jane", "password2")
	require.NoError(t, err)

	cache.values["john"] = CachedCredential{expires: clk.Now().Add(5 * time.Minute), value: sumJohn}
	cache.values["jane"] = CachedCredential{expires: clk.Now().Add(5 * time.Minute), value: sumJane}

	valid, ok := cache.valid(ctx, "john", sumJohn)
	assert.True(t, valid)
	assert.True(t, ok)

	valid, ok = cache.valid(ctx, "jane", sumJane)
	assert.True(t, valid)
	assert.True(t, ok)

	valid, ok = cache.valid(ctx, "john", sumJane)
	assert.False(t, valid)
	assert.True(t, ok)
}

func TestCredentialCacheHMAC_CheckValidThenInvalid(t *testing.T) {
	cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

	mock := &mockUserProvider{valid: true}
	clk := clock.NewFixed(time.Now())
	ctx := &mockContext{provider: mock, clk: clk}

	valid, cached, err := cache.Check(ctx, "john", "goodpassword")
	require.NoError(t, err)
	assert.True(t, valid)
	assert.False(t, cached)

	mock.valid = false

	valid, cached, err = cache.Check(ctx, "john", "badpassword")
	require.NoError(t, err)
	assert.False(t, valid)
	assert.False(t, cached)

	valid, cached, err = cache.Check(ctx, "john", "goodpassword")
	require.NoError(t, err)
	assert.True(t, valid)
	assert.True(t, cached)
}

func FuzzCredentialCacheHMAC_Sum(f *testing.F) {
	f.Add("john", "password")
	f.Add("", "")
	f.Add("user", "")
	f.Add("", "pass")
	f.Add("a", "b")
	f.Add("john", "pass")
	f.Add("ohn", "passj")
	f.Add(string(make([]byte, 1000)), string(make([]byte, 1000)))

	cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

	f.Fuzz(func(t *testing.T, username, password string) {
		hex, sum, err := cache.sum(username, password)

		require.NoError(t, err)
		assert.NotEmpty(t, hex)
		assert.NotNil(t, sum)

		hex2, sum2, err := cache.sum(username, password)

		require.NoError(t, err)
		assert.Equal(t, hex, hex2)
		assert.Equal(t, sum, sum2)
	})
}

func FuzzCredentialCacheHMAC_SumNoCollision(f *testing.F) {
	f.Add("user", "pass", "use", "rpass")
	f.Add("a", "bc", "ab", "c")
	f.Add("john", "password", "johnpass", "word")
	f.Add("", "ab", "a", "b")

	cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

	f.Fuzz(func(t *testing.T, user1, pass1, user2, pass2 string) {
		if user1 == user2 && pass1 == pass2 {
			return
		}

		hex1, _, err := cache.sum(user1, pass1)
		require.NoError(t, err)

		hex2, _, err := cache.sum(user2, pass2)
		require.NoError(t, err)

		assert.NotEqual(t, hex1, hex2,
			"collision: sum(%q, %q) == sum(%q, %q)", user1, pass1, user2, pass2)
	})
}

type mockUserProviderConcurrent struct {
	UserProvider

	valid bool
	err   error
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
