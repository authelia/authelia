package storage_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestSQLProviderEncryptErrors(t *testing.T) {
	testCases := []struct {
		name      string
		invoke    func(p *storage.SQLProvider) error
		expectErr string
	}{
		{
			name: "ShouldReturnErrSaveTOTPConfiguration",
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveTOTPConfiguration(context.Background(), model.TOTPConfiguration{
					Username: "john",
					Secret:   []byte("JBSWY3DPEHPK3PXP"),
				})
			},
			expectErr: "error encrypting TOTP configuration secret for user 'john': crypto/aes: invalid key size 0",
		},
		{
			name: "ShouldReturnErrSaveWebAuthnCredential",
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveWebAuthnCredential(context.Background(), model.WebAuthnCredential{
					RPID:      "example.com",
					Username:  "john",
					PublicKey: []byte("fake-public-key"),
				})
			},
			expectErr: "error encrypting WebAuthn credential public key for user 'john' kid '': crypto/aes: invalid key size 0",
		},
		{
			name: "ShouldReturnErrSaveOneTimeCode",
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.SaveOneTimeCode(context.Background(), model.OneTimeCode{
					Username: "john",
					Intent:   "reset_password",
					Code:     []byte("123456"),
				})

				return err
			},
			expectErr: "error encrypting the one-time code value for user 'john' with signature '1eadd86dd4fd13d44aeb406bb7104534c306051c6504e18afc98be5d6201d7cf9c2422c3a7457d21a21c33d145b3e4ec3d1993a79bd06f935cd2365d03621290': crypto/aes: invalid key size 0",
		},
		{
			name: "ShouldReturnErrSaveOAuth2Session",
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveOAuth2Session(context.Background(), storage.OAuth2SessionTypeAccessToken, model.OAuth2Session{
					RequestID: "req-123",
					Signature: "sig-123",
					Session:   []byte(`{"access":"token"}`),
				})
			},
			expectErr: "error encrypting oauth2 access token session data for subject '' and request id 'req-123' and challenge id '00000000-0000-0000-0000-000000000000': crypto/aes: invalid key size 0",
		},
		{
			name: "ShouldReturnErrSaveOAuth2SessionUnknownType",
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveOAuth2Session(context.Background(), storage.OAuth2SessionType(255), model.OAuth2Session{
					RequestID: "req-123",
				})
			},
			expectErr: "error inserting oauth2 session for subject '' and request id 'req-123': unknown oauth2 session type 'invalid'",
		},
		{
			name: "ShouldReturnErrSaveOAuth2DeviceCodeSession",
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveOAuth2DeviceCodeSession(context.Background(), &model.OAuth2DeviceCodeSession{
					Signature: "dev-sig-123",
					RequestID: "dev-req-123",
					Session:   []byte(`{"device":"code"}`),
				})
			},
			expectErr: "error encrypting oauth2 device code session data for session with signature '' for subject 'dev-sig-123' and request id 'dev-req-123': crypto/aes: invalid key size 0",
		},
		{
			name: "ShouldReturnErrUpdateOAuth2DeviceCodeSession",
			invoke: func(p *storage.SQLProvider) error {
				return p.UpdateOAuth2DeviceCodeSession(context.Background(), &model.OAuth2DeviceCodeSession{
					Signature: "dev-sig-123",
					RequestID: "dev-req-123",
					Session:   []byte(`{"device":"code"}`),
				})
			},
			expectErr: "error encrypting oauth2 device code session data for session with signature '' for subject 'dev-sig-123' and request id 'dev-req-123': crypto/aes: invalid key size 0",
		},
		{
			name: "ShouldReturnErrUpdateOAuth2DeviceCodeSessionData",
			invoke: func(p *storage.SQLProvider) error {
				return p.UpdateOAuth2DeviceCodeSessionData(context.Background(), &model.OAuth2DeviceCodeSession{
					Signature: "dev-sig-123",
					RequestID: "dev-req-123",
					Session:   []byte(`{"device":"code"}`),
				})
			},
			expectErr: "error encrypting oauth2 device code session data for session with signature '' for subject 'dev-sig-123' and request id 'dev-req-123': crypto/aes: invalid key size 0",
		},
		{
			name: "ShouldReturnErrSaveOAuth2PushedAuthorizationSession",
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveOAuth2PushedAuthorizationSession(context.Background(), model.OAuth2PushedAuthorizationSession{
					Signature: "par-sig-123",
					RequestID: "par-req-123",
					Session:   []byte(`{"par":"session"}`),
				})
			},
			expectErr: "error encrypting oauth2 pushed authorization request session data for with signature 'par-sig-123' and request id 'par-req-123': crypto/aes: invalid key size 0",
		},
		{
			name: "ShouldReturnErrUpdateOAuth2PushedAuthorizationSessionZeroID",
			invoke: func(p *storage.SQLProvider) error {
				return p.UpdateOAuth2PushedAuthorizationSession(context.Background(), model.OAuth2PushedAuthorizationSession{
					Signature: "par-sig-123",
					RequestID: "par-req-123",
					Session:   []byte(`{"par":"session"}`),
				})
			},
			expectErr: "error updating oauth2 pushed authorization request session data with signature 'par-sig-123' and request id 'par-req-123': the id was a zero value",
		},
		{
			name: "ShouldReturnErrUpdateOAuth2PushedAuthorizationSession",
			invoke: func(p *storage.SQLProvider) error {
				return p.UpdateOAuth2PushedAuthorizationSession(context.Background(), model.OAuth2PushedAuthorizationSession{
					ID:        1,
					Signature: "par-sig-123",
					RequestID: "par-req-123",
					Session:   []byte(`{"par":"session"}`),
				})
			},
			expectErr: "error encrypting oauth2 pushed authorization request session data with id '1' and signature 'par-sig-123' and request id 'par-req-123': crypto/aes: invalid key size 0",
		},
		{
			name: "ShouldReturnErrSaveCachedData",
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveCachedData(context.Background(), model.CachedData{
					Name:      "cache-key",
					Value:     []byte("cache-value"),
					Encrypted: true,
				})
			},
			expectErr: "error encrypting cached data name 'cache-key': crypto/aes: invalid key size 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)
			p := storage.NewSQLProviderForTestingWithKey(db, []byte{})

			assert.EqualError(t, tc.invoke(p), tc.expectErr)
		})
	}
}
