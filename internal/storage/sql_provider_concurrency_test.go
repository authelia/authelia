package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestSQLProviderConcurrentTransactions(t *testing.T) {
	testCases := []struct {
		name        string
		concurrency int
		iterations  int
	}{
		{name: "ShouldRotateConcurrentlyWithSixteenWorkers", concurrency: 16, iterations: 10},
		{name: "ShouldRotateConcurrentlyWithThirtyTwoWorkers", concurrency: 32, iterations: 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)

			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			for i := 0; i < tc.concurrency; i++ {
				require.NoError(t, provider.SaveOAuth2Session(ctx, OAuth2SessionTypeRefreshToken, newTestConcurrentOAuth2Session(i, 0)))
			}

			var (
				wg   sync.WaitGroup
				mu   sync.Mutex
				errs []error
			)

			record := func(err error) {
				mu.Lock()

				errs = append(errs, err)

				mu.Unlock()
			}

			for i := 0; i < tc.concurrency; i++ {
				wg.Add(1)

				go func(i int) {
					defer wg.Done()

					for j := 0; j < tc.iterations; j++ {
						if err := rotateTestConcurrentOAuth2Session(provider, i, j); err != nil {
							record(err)

							return
						}
					}
				}(i)
			}

			wg.Wait()

			assert.Empty(t, errs)

			for i := 0; i < tc.concurrency; i++ {
				session, err := provider.LoadOAuth2Session(ctx, OAuth2SessionTypeRefreshToken, newTestConcurrentOAuth2Signature(i, tc.iterations))

				require.NoError(t, err)
				assert.True(t, session.Active)
			}
		})
	}
}

func rotateTestConcurrentOAuth2Session(provider *SQLiteProvider, worker, iteration int) (err error) {
	var ctx context.Context

	if ctx, err = provider.BeginTX(context.Background()); err != nil {
		return err
	}

	if _, err = provider.LoadOAuth2Session(ctx, OAuth2SessionTypeRefreshToken, newTestConcurrentOAuth2Signature(worker, iteration)); err != nil {
		_ = provider.Rollback(ctx)

		return err
	}

	if err = provider.RevokeOAuth2Session(ctx, OAuth2SessionTypeRefreshToken, newTestConcurrentOAuth2Signature(worker, iteration)); err != nil {
		_ = provider.Rollback(ctx)

		return err
	}

	if err = provider.SaveOAuth2Session(ctx, OAuth2SessionTypeAccessToken, newTestConcurrentOAuth2Session(worker, iteration+1)); err != nil {
		_ = provider.Rollback(ctx)

		return err
	}

	if err = provider.SaveOAuth2Session(ctx, OAuth2SessionTypeRefreshToken, newTestConcurrentOAuth2Session(worker, iteration+1)); err != nil {
		_ = provider.Rollback(ctx)

		return err
	}

	return provider.Commit(ctx)
}

func newTestConcurrentOAuth2Signature(worker, iteration int) (signature string) {
	return fmt.Sprintf("concurrent-sig-%d-%d", worker, iteration)
}

func newTestConcurrentOAuth2Session(worker, iteration int) (session model.OAuth2Session) {
	return model.OAuth2Session{
		ChallengeID:     model.MustNullUUID(model.NewRandomNullUUID()),
		RequestID:       fmt.Sprintf("concurrent-req-%d-%d", worker, iteration),
		ClientID:        "test-client",
		Signature:       newTestConcurrentOAuth2Signature(worker, iteration),
		Subject:         sql.NullString{Valid: true, String: "john"},
		Active:          true,
		RequestedScopes: model.StringSlicePipeDelimited{"openid", "offline_access"},
		GrantedScopes:   model.StringSlicePipeDelimited{"openid", "offline_access"},
		Session:         []byte("{}"),
	}
}
