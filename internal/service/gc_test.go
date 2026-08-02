package service

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestProvisionOAuth2DPoPGarbageCollector(t *testing.T) {
	testCases := []struct {
		Name     string
		Config   *schema.Configuration
		Expected bool
	}{
		{
			"ShouldNotProvisionWithoutConfiguration",
			nil,
			false,
		},
		{
			"ShouldNotProvisionWithoutOpenIDConnect",
			&schema.Configuration{},
			false,
		},
		{
			"ShouldNotProvisionWithDPoPDisabled",
			&schema.Configuration{IdentityProviders: schema.IdentityProviders{OIDC: &schema.IdentityProvidersOpenIDConnect{}}},
			false,
		},
		{
			"ShouldProvisionWithDPoPEnabled",
			&schema.Configuration{IdentityProviders: schema.IdentityProviders{OIDC: &schema.IdentityProvidersOpenIDConnect{DPoP: schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true}}}},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := newMockServiceCtx()
			ctx.config = tc.Config

			service, err := ProvisionOAuth2DPoPGarbageCollector(ctx)

			require.NoError(t, err)

			if !tc.Expected {
				assert.Nil(t, service)

				return
			}

			require.NotNil(t, service)
			assert.Equal(t, "oauth2-dpop", service.ServiceName())
			assert.Equal(t, serviceTypeGarbageCollector, service.ServiceType())
			assert.NotNil(t, service.Log())
		})
	}
}

func TestGarbageCollectorRun(t *testing.T) {
	testCases := []struct {
		Name string
		Err  error
	}{
		{
			"ShouldCollectRepeatedly",
			nil,
		},
		{
			"ShouldContinueAfterCollectionError",
			errors.New("collect error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var collections atomic.Int64

			ctx := newMockServiceCtx()

			service := NewGarbageCollector("test", time.Millisecond*10, ctx, func(ctx Context) (err error) {
				collections.Add(1)

				return tc.Err
			})

			done := make(chan struct{})

			go func() {
				assert.NoError(t, service.Run())

				close(done)
			}()

			assert.Eventually(t, func() bool { return collections.Load() > 1 }, time.Second*5, time.Millisecond*10)

			service.Shutdown()

			select {
			case <-done:
			case <-time.After(time.Second * 5):
				t.Fatal("service did not shut down within timeout")
			}
		})
	}
}

func TestGarbageCollectorRunShouldRecoverFromPanic(t *testing.T) {
	var collections atomic.Int64

	service := NewGarbageCollector("test", time.Millisecond*10, newMockServiceCtx(), func(ctx Context) (err error) {
		collections.Add(1)

		panic("boom")
	})

	done := make(chan struct{})

	go func() {
		assert.NoError(t, service.Run())

		close(done)
	}()

	assert.Eventually(t, func() bool { return collections.Load() > 1 }, time.Second*5, time.Millisecond*10)

	service.Shutdown()

	select {
	case <-done:
	case <-time.After(time.Second * 5):
		t.Fatal("service did not shut down within timeout")
	}
}

func TestCollectOAuth2DPoP(t *testing.T) {
	testCases := []struct {
		Name     string
		Setup    func(store *mocks.MockStorage)
		Expected string
	}{
		{
			"ShouldCollectBothTables",
			func(store *mocks.MockStorage) {
				store.EXPECT().DeleteExpiredOAuth2DPoPProofs(gomock.Any(), gomock.Any()).Return(nil)
				store.EXPECT().DeleteExpiredOAuth2DPoPNonces(gomock.Any(), gomock.Any()).Return(nil)
			},
			"",
		},
		{
			"ShouldReturnProofError",
			func(store *mocks.MockStorage) {
				store.EXPECT().DeleteExpiredOAuth2DPoPProofs(gomock.Any(), gomock.Any()).Return(errors.New("proof error"))
			},
			"proof error",
		},
		{
			"ShouldReturnNonceError",
			func(store *mocks.MockStorage) {
				store.EXPECT().DeleteExpiredOAuth2DPoPProofs(gomock.Any(), gomock.Any()).Return(nil)
				store.EXPECT().DeleteExpiredOAuth2DPoPNonces(gomock.Any(), gomock.Any()).Return(errors.New("nonce error"))
			},
			"nonce error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			defer ctrl.Finish()

			store := mocks.NewMockStorage(ctrl)

			if tc.Setup != nil {
				tc.Setup(store)
			}

			ctx := newMockServiceCtx()
			ctx.providers.StorageProvider = store

			err := collectOAuth2DPoP(ctx)

			if tc.Expected == "" {
				assert.NoError(t, err)

				return
			}

			assert.EqualError(t, err, tc.Expected)
		})
	}
}
