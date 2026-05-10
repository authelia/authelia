package server

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

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

		mock.Ctx.Request.SetRequestURI("/?flow_id=" + uuid.NewString())

		assert.Equal(t, "", resolveOIDCConsentLogoURI(mock.Ctx))
	})

	t.Run("ShouldReturnEmptyOnMissingFlowID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Request.SetRequestURI(oidc.FrontendEndpointPathConsentDecision)

		assert.Equal(t, "", resolveOIDCConsentLogoURI(mock.Ctx))
	})

	t.Run("ShouldReturnEmptyOnInvalidFlowID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Request.SetRequestURI(oidc.FrontendEndpointPathConsentDecision + "?flow_id=not-a-uuid")

		assert.Equal(t, "", resolveOIDCConsentLogoURI(mock.Ctx))
	})

	t.Run("ShouldReturnEmptyWhenStorageReturnsError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

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

		flowID := uuid.New()

		mock.Ctx.Request.SetRequestURI(oidc.FrontendEndpointPathConsentDecision + "?flow_id=" + flowID.String())

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), flowID).
			Return(&model.OAuth2ConsentSession{ClientID: "any"}, nil)

		assert.Nil(t, mock.Ctx.Providers.OpenIDConnect)
		assert.Equal(t, "", resolveOIDCConsentLogoURI(mock.Ctx))
	})
}
