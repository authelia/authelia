// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestExpandCSPTemplate(t *testing.T) {
	testCases := []struct {
		name               string
		tmpl               string
		nonce              string
		oidcClientLogoURIs string
		expected           string
	}{
		{
			"ShouldSubstituteNonceAndOIDCClientLogoURIs",
			"script-src 'nonce-${NONCE}'; img-src 'self' data:${OIDC_CLIENT_LOGO_URIS}",
			"abc",
			" https://example.com",
			"script-src 'nonce-abc'; img-src 'self' data: https://example.com",
		},
		{
			"ShouldStripOIDCClientLogoURIsWhenEmpty",
			"img-src 'self' data:${OIDC_CLIENT_LOGO_URIS}",
			"abc",
			"",
			"img-src 'self' data:",
		},
		{
			"ShouldLeaveTemplateAloneWithoutPlaceholders",
			"default-src 'self'",
			"abc",
			" https://example.com",
			"default-src 'self'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, expandCSPTemplate(tc.tmpl, tc.nonce, tc.oidcClientLogoURIs))
		})
	}
}

func TestIsOIDCConsentShellPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{oidc.FrontendEndpointPathConsentDecision, true},
		{oidc.FrontendEndpointPathConsentDeviceAuthorization, true},
		{oidc.FrontendEndpointPathConsentCompletion, false},
		{"/consent/openid/decision/extra", false},
		{"/", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			assert.Equal(t, tc.expected, isOIDCConsentShellPath(tc.path))
		})
	}
}

func TestResolveOIDCConsentLogoURI(t *testing.T) {
	t.Run("ShouldReturnEmptyOnNonConsentPath", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestCSPOIDCProvider(t, mock)

		mock.Ctx.Request.SetRequestURI("/?flow_id=" + uuid.NewString())

		assert.Equal(t, "", resolveOIDCConsentLogoURI(mock.Ctx))
	})

	t.Run("ShouldReturnEmptyOnMissingFlowIDAndUserCode", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestCSPOIDCProvider(t, mock)

		mock.Ctx.Request.SetRequestURI(oidc.FrontendEndpointPathConsentDecision)

		assert.Equal(t, "", resolveOIDCConsentLogoURI(mock.Ctx))
	})

	t.Run("ShouldReturnEmptyOnInvalidFlowID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestCSPOIDCProvider(t, mock)

		mock.Ctx.Request.SetRequestURI(oidc.FrontendEndpointPathConsentDecision + "?flow_id=not-a-uuid")

		assert.Equal(t, "", resolveOIDCConsentLogoURI(mock.Ctx))
	})

	t.Run("ShouldReturnEmptyWhenStorageReturnsError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestCSPOIDCProvider(t, mock)

		flowID := uuid.New()

		mock.Ctx.Request.SetRequestURI(oidc.FrontendEndpointPathConsentDecision + "?flow_id=" + flowID.String())

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), flowID).
			Return(nil, errors.New("not found"))

		assert.Equal(t, "", resolveOIDCConsentLogoURI(mock.Ctx))
	})

	t.Run("ShouldReturnEmptyWhenOIDCProviderNotConfigured", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Request.SetRequestURI(oidc.FrontendEndpointPathConsentDecision + "?flow_id=" + uuid.NewString())

		assert.Nil(t, mock.Ctx.Providers.OpenIDConnect)
		assert.Equal(t, "", resolveOIDCConsentLogoURI(mock.Ctx))
	})

	t.Run("ShouldReturnLogoHostForFlowID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestCSPOIDCProvider(t, mock)

		flowID := uuid.New()

		mock.Ctx.Request.SetRequestURI(oidc.FrontendEndpointPathConsentDecision + "?flow_id=" + flowID.String())

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), flowID).
			Return(&model.OAuth2ConsentSession{ClientID: testCSPClientIDLogo}, nil)

		assert.Equal(t, " https://logo.example.com", resolveOIDCConsentLogoURI(mock.Ctx))
	})

	t.Run("ShouldReturnEmptyForFlowIDWhenClientHasNoLogo", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestCSPOIDCProvider(t, mock)

		flowID := uuid.New()

		mock.Ctx.Request.SetRequestURI(oidc.FrontendEndpointPathConsentDecision + "?flow_id=" + flowID.String())

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), flowID).
			Return(&model.OAuth2ConsentSession{ClientID: testCSPClientIDNoLogo}, nil)

		assert.Equal(t, "", resolveOIDCConsentLogoURI(mock.Ctx))
	})

	t.Run("ShouldReturnLogoHostForUserCode", func(t *testing.T) {
		for _, path := range []string{oidc.FrontendEndpointPathConsentDecision, oidc.FrontendEndpointPathConsentDeviceAuthorization} {
			t.Run(path, func(t *testing.T) {
				mock := mocks.NewMockAutheliaCtx(t)
				defer mock.Close()

				setupTestCSPOIDCProvider(t, mock)

				signature, err := mock.Ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(mock.Ctx, "BGKMRTVX")
				require.NoError(t, err)

				mock.Ctx.Request.SetRequestURI(path + "?flow=openid_connect&subflow=device_authorization&user_code=BGKMRTVX")

				mock.StorageMock.EXPECT().
					LoadOAuth2DeviceCodeSessionByUserCode(gomock.Any(), signature).
					Return(&model.OAuth2DeviceCodeSession{ClientID: testCSPClientIDLogo}, nil)

				assert.Equal(t, " https://logo.example.com", resolveOIDCConsentLogoURI(mock.Ctx))
			})
		}
	})

	t.Run("ShouldReturnEmptyForUserCodeWhenStorageReturnsError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestCSPOIDCProvider(t, mock)

		mock.Ctx.Request.SetRequestURI(oidc.FrontendEndpointPathConsentDecision + "?user_code=BGKMRTVX")

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByUserCode(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("not found"))

		assert.Equal(t, "", resolveOIDCConsentLogoURI(mock.Ctx))
	})

	t.Run("ShouldReturnEmptyForUserCodeOnNonConsentPath", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestCSPOIDCProvider(t, mock)

		mock.Ctx.Request.SetRequestURI("/?user_code=BGKMRTVX")

		assert.Equal(t, "", resolveOIDCConsentLogoURI(mock.Ctx))
	})
}

const (
	testCSPClientIDLogo   = "logo"
	testCSPClientIDNoLogo = "no-logo"
)

func setupTestCSPOIDCProvider(t *testing.T, mock *mocks.MockAutheliaCtx) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	mock.Ctx.Configuration.IdentityProviders.OIDC = &schema.IdentityProvidersOpenIDConnect{
		HMACSecret: "abcdefghijklmnopqrstuvwxyz1234567890",
		JSONWebKeys: []schema.JWK{
			{
				KeyID:     "ecdsa-default",
				Use:       oidc.KeyUseSignature,
				Algorithm: oidc.SigningAlgECDSAUsingP256AndSHA256,
				Key:       key,
			},
		},
		Clients: []schema.IdentityProvidersOpenIDConnectClient{
			{ID: testCSPClientIDLogo, LogoURI: &url.URL{Scheme: "https", Host: "logo.example.com", Path: "/logo.png"}},
			{ID: testCSPClientIDNoLogo},
		},
	}

	mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(&mock.Ctx.Configuration, mock.StorageMock, mock.Ctx.Providers.Templates)

	require.NotNil(t, mock.Ctx.Providers.OpenIDConnect)
}
