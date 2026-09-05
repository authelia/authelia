package oidc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestOpenIDConnectProviderGarbageCollectionFrequency(t *testing.T) {
	testCases := []struct {
		name     string
		enabled  bool
		expected time.Duration
	}{
		{
			"ShouldReturnZeroWhenDisabled",
			false,
			0,
		},
		{
			"ShouldReturnFrequencyWhenEnabled",
			true,
			time.Minute * 30,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestGarbageCollectorProvider(t, tc.enabled, nil)

			assert.Equal(t, tc.expected, provider.GarbageCollectionFrequency(context.TODO()))
		})
	}
}

func TestOpenIDConnectProviderGarbageCollectionFrequencyShouldReturnZeroWhenNotConfigured(t *testing.T) {
	provider := oidc.NewOpenIDConnectProvider(nil, nil, nil)

	assert.Equal(t, time.Duration(0), provider.GarbageCollectionFrequency(context.TODO()))
}

func TestOpenIDConnectProviderGarbageCollection(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(store *mocks.MockStorage)
		err   string
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
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			defer ctrl.Finish()

			store := mocks.NewMockStorage(ctrl)

			if tc.setup != nil {
				tc.setup(store)
			}

			provider := newTestGarbageCollectorProvider(t, true, store)

			err := provider.GarbageCollection(context.TODO())

			if tc.err == "" {
				assert.NoError(t, err)

				return
			}

			assert.EqualError(t, err, tc.err)
		})
	}
}

func TestOpenIDConnectProviderGarbageCollectionShouldNotCollectWhenNotConfigured(t *testing.T) {
	provider := oidc.NewOpenIDConnectProvider(nil, nil, nil)

	assert.NoError(t, provider.GarbageCollection(context.TODO()))
}

func newTestGarbageCollectorProvider(t *testing.T, enabled bool, store storage.Provider) (provider *oidc.OpenIDConnectProvider) {
	t.Helper()

	provider = oidc.NewOpenIDConnectProvider(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       badhmac,
				IssuerPrivateKey: x509PrivateKeyRSA2048,
				DPoP:             schema.IdentityProvidersOpenIDConnectDPoP{Enabled: enabled},
			},
		},
	}, store, nil)

	assert.NotNil(t, provider)

	return provider
}
