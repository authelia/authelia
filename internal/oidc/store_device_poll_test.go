// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/rfc8628"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/storage"
)

type testDevicePollStrategy struct{}

func (testDevicePollStrategy) RFC8628DeviceCodeSignature(_ context.Context, code string) (string, error) {
	return code, nil
}

func (testDevicePollStrategy) GenerateRFC8628DeviceCode(_ context.Context) (string, string, error) {
	return "", "", nil
}

func (testDevicePollStrategy) ValidateRFC8628DeviceCode(_ context.Context, _ oauthelia2.Requester, _ string) error {
	return nil
}

func (testDevicePollStrategy) RFC8628UserCodeSignature(_ context.Context, code string) (string, error) {
	return code, nil
}

func (testDevicePollStrategy) GenerateRFC8628UserCode(_ context.Context) (string, string, error) {
	return "", "", nil
}

func (testDevicePollStrategy) ValidateRFC8628UserCode(_ context.Context, _ oauthelia2.Requester, _ string) error {
	return nil
}

func TestDeviceCodePollAfterDenyShouldReturnAccessDenied(t *testing.T) {
	ctx := context.Background()

	provider, err := storage.NewSQLiteProvider(&schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "db.sqlite3"),
			},
		},
	})

	require.NoError(t, err)
	require.NoError(t, provider.StartupCheck())

	s := oidc.NewStore(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				IssuerCertificateChain: schema.X509CertificateChain{},
				IssuerPrivateKey:       x509PrivateKeyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  myclient,
						Name:                myclientname,
						AuthorizationPolicy: onefactor,
						Scopes:              []string{oidc.ScopeOpenID},
						Secret:              tOpenIDConnectPlainTextClientSecret,
					},
				},
			},
		},
	}, provider)

	client := &oidc.RegisteredClient{ID: myclient}

	requester := &oauthelia2.DeviceAuthorizeRequest{
		Request: oauthelia2.Request{
			ID:             "req-device-poll-deny",
			Client:         client,
			RequestedScope: oauthelia2.Arguments{oidc.ScopeOpenID},
			Session:        &oidc.Session{},
		},
		DeviceCodeSignature: "sig-device-poll-deny",
		UserCodeSignature:   "user-device-poll-deny",
	}

	require.NoError(t, s.CreateDeviceCodeSession(ctx, requester.DeviceCodeSignature, requester))

	challenge := uuid.MustParse("6d7e6b0e-5d3a-4c8e-9b1e-1f2a3b4c5d6e")

	require.NoError(t, provider.ClaimOAuth2DeviceCodeSession(ctx, &model.OAuth2DeviceCodeSession{
		Signature:   requester.DeviceCodeSignature,
		ChallengeID: uuid.NullUUID{UUID: challenge, Valid: true},
		Status:      model.OAuth2DeviceCodeStatusClaimed,
		Active:      true,
	}))

	require.NoError(t, provider.DenyOAuth2DeviceCodeSession(ctx, requester.DeviceCodeSignature, challenge))

	handler := &rfc8628.DeviceCodeTokenHandler{
		Storage:  s,
		Strategy: testDevicePollStrategy{},
		Config:   &oauthelia2.Config{},
	}

	request := &oauthelia2.AccessRequest{
		Request: oauthelia2.Request{
			Client:  client,
			Form:    url.Values{"device_code": []string{requester.DeviceCodeSignature}},
			Session: &oidc.Session{},
		},
	}

	_, _, _, err = handler.GetCodeAndSession(ctx, request)

	assert.ErrorIs(t, err, oauthelia2.ErrAccessDenied)
	assert.NotErrorIs(t, err, oauthelia2.ErrInvalidatedDeviceCode)
}

func TestDeviceCodePollWhileClaimedShouldKeepClaim(t *testing.T) {
	ctx := context.Background()

	provider, err := storage.NewSQLiteProvider(&schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "db.sqlite3"),
			},
		},
	})

	require.NoError(t, err)
	require.NoError(t, provider.StartupCheck())

	s := oidc.NewStore(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				IssuerCertificateChain: schema.X509CertificateChain{},
				IssuerPrivateKey:       x509PrivateKeyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  myclient,
						Name:                myclientname,
						AuthorizationPolicy: onefactor,
						Scopes:              []string{oidc.ScopeOpenID},
						Secret:              tOpenIDConnectPlainTextClientSecret,
					},
				},
			},
		},
	}, provider)

	client := &oidc.RegisteredClient{ID: myclient}

	requester := &oauthelia2.DeviceAuthorizeRequest{
		Request: oauthelia2.Request{
			ID:             "req-device-poll-claimed",
			Client:         client,
			RequestedScope: oauthelia2.Arguments{oidc.ScopeOpenID},
			Session:        &oidc.Session{},
		},
		DeviceCodeSignature: "sig-device-poll-claimed",
		UserCodeSignature:   "user-device-poll-claimed",
	}

	require.NoError(t, s.CreateDeviceCodeSession(ctx, requester.DeviceCodeSignature, requester))

	challenge := uuid.MustParse("0b9c2f4e-7a61-4d3b-8f25-6c1e9d0a4b7f")

	require.NoError(t, provider.ClaimOAuth2DeviceCodeSession(ctx, &model.OAuth2DeviceCodeSession{
		Signature:   requester.DeviceCodeSignature,
		ChallengeID: uuid.NullUUID{UUID: challenge, Valid: true},
		Status:      model.OAuth2DeviceCodeStatusClaimed,
		Active:      true,
	}))

	handler := &rfc8628.DeviceCodeTokenHandler{
		Storage:  s,
		Strategy: testDevicePollStrategy{},
		Config:   &oauthelia2.Config{},
	}

	request := &oauthelia2.AccessRequest{
		Request: oauthelia2.Request{
			Client:  client,
			Form:    url.Values{"device_code": []string{requester.DeviceCodeSignature}},
			Session: &oidc.Session{},
		},
	}

	_, _, _, err = handler.GetCodeAndSession(ctx, request)

	assert.ErrorIs(t, err, oauthelia2.ErrAuthorizationPending)

	device, err := provider.LoadOAuth2DeviceCodeSessionByChallengeID(ctx, challenge)

	require.NoError(t, err)
	assert.Equal(t, requester.DeviceCodeSignature, device.Signature)
	assert.Equal(t, model.OAuth2DeviceCodeStatusClaimed, device.Status)
	assert.False(t, device.CheckedAt.IsZero())
}
