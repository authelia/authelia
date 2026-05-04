package storage

import (
	"context"
	"database/sql"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestNewProvider(t *testing.T) {
	testCases := []struct {
		name     string
		config   *schema.Configuration
		expectNl bool
	}{
		{
			"ShouldReturnSQLiteProvider",
			&schema.Configuration{
				Storage: schema.Storage{
					Local: &schema.StorageLocal{Path: filepath.Join(t.TempDir(), "db.sqlite3")},
				},
			},
			false,
		},
		{
			"ShouldReturnNilForNoStorage",
			&schema.Configuration{},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewProvider(tc.config, nil)

			if tc.expectNl {
				assert.Nil(t, provider)
			} else {
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestSQLProviderStartupAndClose(t *testing.T) {
	t.Run("ShouldStartupAndClose", func(t *testing.T) {
		provider := newTestSQLiteProvider(t)

		require.NoError(t, provider.StartupCheck())
		require.NoError(t, provider.Close())
	})
}

func TestSQLProviderPreferred2FAMethod(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	testCases := []struct {
		name           string
		username       string
		method         string
		expectedMethod string
	}{
		{"ShouldSaveAndLoadTOTP", "john", "totp", "totp"},
		{"ShouldSaveAndLoadWebauthn", "jane", "webauthn", "webauthn"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, provider.SavePreferred2FAMethod(ctx, tc.username, tc.method))

			method, err := provider.LoadPreferred2FAMethod(ctx, tc.username)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedMethod, method)
		})
	}

	t.Run("ShouldErrForUnknownUser", func(t *testing.T) {
		_, err := provider.LoadPreferred2FAMethod(ctx, "unknown")

		assert.Error(t, err)
	})
}

func TestSQLProviderUserOpaqueIdentifier(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldSaveAndLoadBySignature", func(t *testing.T) {
		id, err := uuid.NewRandom()
		require.NoError(t, err)

		opaqueID := model.UserOpaqueIdentifier{
			Service:    "openid",
			SectorID:   "example.com",
			Username:   "john",
			Identifier: id,
		}

		require.NoError(t, provider.SaveUserOpaqueIdentifier(ctx, opaqueID))

		loaded, err := provider.LoadUserOpaqueIdentifierBySignature(ctx, "openid", "example.com", "john")

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "john", loaded.Username)
		assert.Equal(t, id, loaded.Identifier)
	})

	t.Run("ShouldLoadByIdentifier", func(t *testing.T) {
		identifiers, err := provider.LoadUserOpaqueIdentifiers(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, identifiers)

		loaded, err := provider.LoadUserOpaqueIdentifier(ctx, identifiers[0].Identifier)

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, identifiers[0].Username, loaded.Username)
	})

	t.Run("ShouldReturnNilForUnknownIdentifier", func(t *testing.T) {
		loaded, err := provider.LoadUserOpaqueIdentifierBySignature(ctx, "openid", "unknown.com", "nobody")

		assert.NoError(t, err)
		assert.Nil(t, loaded)
	})
}

func TestSQLProviderTOTPConfiguration(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldSaveAndLoad", func(t *testing.T) {
		config := model.TOTPConfiguration{
			CreatedAt: time.Now().Truncate(time.Second),
			Username:  "john",
			Issuer:    "Authelia",
			Algorithm: "SHA1",
			Digits:    6,
			Period:    30,
			Secret:    []byte("JBSWY3DPEHPK3PXP"),
		}

		require.NoError(t, provider.SaveTOTPConfiguration(ctx, config))

		loaded, err := provider.LoadTOTPConfiguration(ctx, "john")

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "john", loaded.Username)
		assert.Equal(t, "Authelia", loaded.Issuer)
		assert.Equal(t, uint32(6), loaded.Digits)
	})

	t.Run("ShouldReturnErrForUnknownUser", func(t *testing.T) {
		loaded, err := provider.LoadTOTPConfiguration(ctx, "nobody")

		assert.Error(t, err)
		assert.Nil(t, loaded)
	})

	t.Run("ShouldDeleteConfiguration", func(t *testing.T) {
		require.NoError(t, provider.DeleteTOTPConfiguration(ctx, "john"))

		loaded, err := provider.LoadTOTPConfiguration(ctx, "john")

		assert.Error(t, err)
		assert.Nil(t, loaded)
	})

	t.Run("ShouldLoadConfigurations", func(t *testing.T) {
		require.NoError(t, provider.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
			CreatedAt: time.Now().Truncate(time.Second),
			Username:  "alice",
			Issuer:    "Authelia",
			Algorithm: "SHA1",
			Digits:    6,
			Period:    30,
			Secret:    []byte("SECRET1"),
		}))

		configs, err := provider.LoadTOTPConfigurations(ctx, 10, 0)

		assert.NoError(t, err)
		assert.NotEmpty(t, configs)
	})

	t.Run("ShouldUpdateSignIn", func(t *testing.T) {
		loaded, err := provider.LoadTOTPConfiguration(ctx, "alice")
		require.NoError(t, err)

		now := sql.NullTime{Valid: true, Time: time.Now().Truncate(time.Second)}

		require.NoError(t, provider.UpdateTOTPConfigurationSignIn(ctx, loaded.ID, now))
	})

	t.Run("ShouldSaveAndCheckHistory", func(t *testing.T) {
		require.NoError(t, provider.SaveTOTPHistory(ctx, "alice", 12345))

		exists, err := provider.ExistsTOTPHistory(ctx, "alice", 12345)

		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = provider.ExistsTOTPHistory(ctx, "alice", 99999)

		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestSQLProviderTransactions(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	t.Run("ShouldBeginAndCommit", func(t *testing.T) {
		ctx, err := provider.BeginTX(context.Background())
		require.NoError(t, err)

		require.NoError(t, provider.SavePreferred2FAMethod(ctx, "tx-user", "totp"))
		require.NoError(t, provider.Commit(ctx))

		method, err := provider.LoadPreferred2FAMethod(context.Background(), "tx-user")

		assert.NoError(t, err)
		assert.Equal(t, "totp", method)
	})

	t.Run("ShouldBeginAndRollback", func(t *testing.T) {
		ctx, err := provider.BeginTX(context.Background())
		require.NoError(t, err)

		require.NoError(t, provider.Rollback(ctx))
	})
}

func TestSQLProviderWebAuthn(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldSaveAndLoadUser", func(t *testing.T) {
		user := model.WebAuthnUser{
			RPID:     "example.com",
			Username: "john",
			UserID:   "user-john-123",
		}

		require.NoError(t, provider.SaveWebAuthnUser(ctx, user))

		loaded, err := provider.LoadWebAuthnUser(ctx, "example.com", "john")

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "john", loaded.Username)
	})

	t.Run("ShouldLoadUserByUserID", func(t *testing.T) {
		loaded, err := provider.LoadWebAuthnUserByUserID(ctx, "example.com", "user-john-123")

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "john", loaded.Username)
	})

	t.Run("ShouldSaveAndLoadCredential", func(t *testing.T) {
		cred := model.WebAuthnCredential{
			CreatedAt:       time.Now().Truncate(time.Second),
			RPID:            "example.com",
			Username:        "john",
			Description:     "my-key",
			KID:             model.NewBase64([]byte("kid-1")),
			AttestationType: "none",
			Attachment:      "cross-platform",
			PublicKey:       []byte("fake-public-key"),
		}

		require.NoError(t, provider.SaveWebAuthnCredential(ctx, cred))

		creds, err := provider.LoadWebAuthnCredentialsByUsername(ctx, "example.com", "john")

		require.NoError(t, err)
		require.NotEmpty(t, creds)
		assert.Equal(t, "my-key", creds[0].Description)
	})

	t.Run("ShouldLoadCredentials", func(t *testing.T) {
		creds, err := provider.LoadWebAuthnCredentials(ctx, 10, 0)

		require.NoError(t, err)
		assert.NotEmpty(t, creds)
	})

	t.Run("ShouldLoadCredentialByID", func(t *testing.T) {
		creds, err := provider.LoadWebAuthnCredentials(ctx, 1, 0)
		require.NoError(t, err)
		require.NotEmpty(t, creds)

		cred, err := provider.LoadWebAuthnCredentialByID(ctx, creds[0].ID)

		require.NoError(t, err)
		require.NotNil(t, cred)
		assert.Equal(t, creds[0].Description, cred.Description)
	})

	t.Run("ShouldUpdateCredentialDescription", func(t *testing.T) {
		creds, err := provider.LoadWebAuthnCredentialsByUsername(ctx, "example.com", "john")
		require.NoError(t, err)
		require.NotEmpty(t, creds)

		require.NoError(t, provider.UpdateWebAuthnCredentialDescription(ctx, "john", creds[0].ID, "updated-key"))

		updated, err := provider.LoadWebAuthnCredentialByID(ctx, creds[0].ID)

		require.NoError(t, err)
		assert.Equal(t, "updated-key", updated.Description)
	})

	t.Run("ShouldUpdateCredentialSignIn", func(t *testing.T) {
		creds, err := provider.LoadWebAuthnCredentialsByUsername(ctx, "example.com", "john")
		require.NoError(t, err)
		require.NotEmpty(t, creds)

		creds[0].SignCount = 5

		require.NoError(t, provider.UpdateWebAuthnCredentialSignIn(ctx, creds[0]))
	})

	t.Run("ShouldDeleteCredentialByKID", func(t *testing.T) {
		cred := model.WebAuthnCredential{
			CreatedAt:       time.Now().Truncate(time.Second),
			RPID:            "example.com",
			Username:        "john",
			Description:     "delete-me",
			KID:             model.NewBase64([]byte("kid-delete")),
			AttestationType: "none",
			Attachment:      "cross-platform",
			PublicKey:       []byte("fake-key"),
		}

		require.NoError(t, provider.SaveWebAuthnCredential(ctx, cred))
		require.NoError(t, provider.DeleteWebAuthnCredential(ctx, model.NewBase64([]byte("kid-delete")).String()))
	})

	t.Run("ShouldDeleteCredentialByUsername", func(t *testing.T) {
		require.NoError(t, provider.DeleteWebAuthnCredentialByUsername(ctx, "john", ""))
	})
}

func TestSQLProviderIdentityVerification(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldSaveAndFind", func(t *testing.T) {
		jti, err := uuid.NewRandom()
		require.NoError(t, err)

		verification := model.IdentityVerification{
			JTI:       jti,
			IssuedAt:  time.Now().Truncate(time.Second),
			IssuedIP:  model.NewIP(net.ParseIP("127.0.0.1")),
			ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
			Action:    "reset_password",
			Username:  "john",
		}

		require.NoError(t, provider.SaveIdentityVerification(ctx, verification))

		found, err := provider.FindIdentityVerification(ctx, jti.String())

		require.NoError(t, err)
		assert.True(t, found)
	})

	t.Run("ShouldNotFindNonExistent", func(t *testing.T) {
		id, _ := uuid.NewRandom()

		found, err := provider.FindIdentityVerification(ctx, id.String())

		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("ShouldConsumeVerification", func(t *testing.T) {
		jti, err := uuid.NewRandom()
		require.NoError(t, err)

		verification := model.IdentityVerification{
			JTI:       jti,
			IssuedAt:  time.Now().Truncate(time.Second),
			IssuedIP:  model.NewIP(net.ParseIP("127.0.0.1")),
			ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
			Action:    "reset_password",
			Username:  "john",
		}

		require.NoError(t, provider.SaveIdentityVerification(ctx, verification))
		require.NoError(t, provider.ConsumeIdentityVerification(ctx, jti.String(), model.NullIP{}))
	})

	t.Run("ShouldRevokeVerification", func(t *testing.T) {
		jti, err := uuid.NewRandom()
		require.NoError(t, err)

		verification := model.IdentityVerification{
			JTI:       jti,
			IssuedAt:  time.Now().Truncate(time.Second),
			IssuedIP:  model.NewIP(net.ParseIP("127.0.0.1")),
			ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
			Action:    "reset_password",
			Username:  "john",
		}

		require.NoError(t, provider.SaveIdentityVerification(ctx, verification))
		require.NoError(t, provider.RevokeIdentityVerification(ctx, jti.String(), model.NullIP{}))
	})
}

func TestSQLProviderDuoDevice(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldSaveAndLoadDuoDevice", func(t *testing.T) {
		require.NoError(t, provider.SavePreferredDuoDevice(ctx, model.DuoDevice{
			Username: "john",
			Device:   "DXYZ123",
			Method:   "push",
		}))

		device, err := provider.LoadPreferredDuoDevice(ctx, "john")

		require.NoError(t, err)
		assert.Equal(t, "DXYZ123", device.Device)
		assert.Equal(t, "push", device.Method)
	})

	t.Run("ShouldErrForUnknownUser", func(t *testing.T) {
		_, err := provider.LoadPreferredDuoDevice(ctx, "nobody")

		assert.Error(t, err)
	})

	t.Run("ShouldDeleteDuoDevice", func(t *testing.T) {
		require.NoError(t, provider.DeletePreferredDuoDevice(ctx, "john"))

		_, err := provider.LoadPreferredDuoDevice(ctx, "john")

		assert.Error(t, err)
	})
}

func TestSQLProviderAuthenticationLog(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldAppendLog", func(t *testing.T) {
		attempt := model.AuthenticationAttempt{
			Time:       time.Now().Truncate(time.Second),
			Successful: false,
			Banned:     false,
			Username:   "john",
			Type:       "1fa",
			RemoteIP:   model.NewNullIP(net.ParseIP("127.0.0.1")),
		}

		require.NoError(t, provider.AppendAuthenticationLog(ctx, attempt))
	})

	t.Run("ShouldLoadRegulationRecords", func(t *testing.T) {
		records, err := provider.LoadRegulationRecordsByUser(ctx, "john", time.Now().Add(-time.Hour), 10)

		assert.NoError(t, err)
		assert.NotNil(t, records)
	})
}

func TestSQLProviderLoadUserInfo(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldReturnDefaultForUnknownUser", func(t *testing.T) {
		info, err := provider.LoadUserInfo(ctx, "unknown")

		assert.NoError(t, err)
		assert.Equal(t, "", info.Method)
	})

	t.Run("ShouldReturnInfoForUserWith2FA", func(t *testing.T) {
		require.NoError(t, provider.SavePreferred2FAMethod(ctx, "john", "totp"))
		require.NoError(t, provider.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
			CreatedAt: time.Now().Truncate(time.Second),
			Username:  "john",
			Issuer:    "Authelia",
			Algorithm: "SHA1",
			Digits:    6,
			Period:    30,
			Secret:    []byte("SECRET"),
		}))

		info, err := provider.LoadUserInfo(ctx, "john")

		assert.NoError(t, err)
		assert.Equal(t, "totp", info.Method)
		assert.True(t, info.HasTOTP)
	})
}

func TestSQLProviderCachedData(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldSaveAndLoadCachedData", func(t *testing.T) {
		require.NoError(t, provider.SaveCachedData(ctx, model.CachedData{Name: "test-key", Value: []byte("test-data")}))

		data, err := provider.LoadCachedData(ctx, "test-key")

		require.NoError(t, err)
		require.NotNil(t, data)
		assert.Equal(t, []byte("test-data"), data.Value)
	})

	t.Run("ShouldReturnNilForUnknownKey", func(t *testing.T) {
		data, err := provider.LoadCachedData(ctx, "unknown-key")

		assert.NoError(t, err)
		assert.Nil(t, data)
	})

	t.Run("ShouldDeleteCachedData", func(t *testing.T) {
		require.NoError(t, provider.SaveCachedData(ctx, model.CachedData{Name: "delete-me", Value: []byte("data")}))
		require.NoError(t, provider.DeleteCachedData(ctx, "delete-me"))

		data, err := provider.LoadCachedData(ctx, "delete-me")

		assert.NoError(t, err)
		assert.Nil(t, data)
	})
}

func TestSQLProviderOAuth2ConsentSession(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	challengeID, err := uuid.NewRandom()
	require.NoError(t, err)

	t.Run("ShouldSaveAndLoadConsentSession", func(t *testing.T) {
		consent := &model.OAuth2ConsentSession{
			ChallengeID:     challengeID,
			ClientID:        "test-client",
			RequestedAt:     time.Now().Truncate(time.Second),
			ExpiresAt:       time.Now().Add(time.Hour).Truncate(time.Second),
			RequestedScopes: model.StringSlicePipeDelimited{"openid", "profile"},
		}

		require.NoError(t, provider.SaveOAuth2ConsentSession(ctx, consent))

		loaded, err := provider.LoadOAuth2ConsentSessionByChallengeID(ctx, challengeID)

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "test-client", loaded.ClientID)
	})

	t.Run("ShouldSaveConsentResponse", func(t *testing.T) {
		loaded, err := provider.LoadOAuth2ConsentSessionByChallengeID(ctx, challengeID)
		require.NoError(t, err)

		require.NoError(t, provider.SaveOAuth2ConsentSessionResponse(ctx, loaded, true))
	})

	t.Run("ShouldSaveConsentGranted", func(t *testing.T) {
		loaded, err := provider.LoadOAuth2ConsentSessionByChallengeID(ctx, challengeID)
		require.NoError(t, err)

		require.NoError(t, provider.SaveOAuth2ConsentSessionGranted(ctx, loaded.ID))
	})

	t.Run("ShouldErrLoadNonExistentChallenge", func(t *testing.T) {
		badID, _ := uuid.NewRandom()

		_, err := provider.LoadOAuth2ConsentSessionByChallengeID(ctx, badID)

		assert.Error(t, err)
	})
}

func TestSQLProviderOAuth2Session(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	challengeID := model.MustNullUUID(model.NewRandomNullUUID())

	session := model.OAuth2Session{
		ChallengeID:     challengeID,
		RequestID:       "req-123",
		ClientID:        "test-client",
		Signature:       "sig-123",
		Subject:         sql.NullString{Valid: true, String: "john"},
		Active:          true,
		RequestedScopes: model.StringSlicePipeDelimited{"openid"},
		GrantedScopes:   model.StringSlicePipeDelimited{"openid"},
		Session:         []byte("{}"),
	}

	t.Run("ShouldSaveAndLoadSession", func(t *testing.T) {
		require.NoError(t, provider.SaveOAuth2Session(ctx, OAuth2SessionTypeAccessToken, session))

		loaded, err := provider.LoadOAuth2Session(ctx, OAuth2SessionTypeAccessToken, "sig-123")

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "req-123", loaded.RequestID)
		assert.True(t, loaded.Active)
	})

	t.Run("ShouldDeactivateSession", func(t *testing.T) {
		require.NoError(t, provider.DeactivateOAuth2Session(ctx, OAuth2SessionTypeAccessToken, "sig-123"))

		loaded, err := provider.LoadOAuth2Session(ctx, OAuth2SessionTypeAccessToken, "sig-123")

		require.NoError(t, err)
		assert.False(t, loaded.Active)
	})

	t.Run("ShouldSaveAndRevokeSession", func(t *testing.T) {
		s := session
		s.Signature = "sig-revoke"
		s.ChallengeID = model.MustNullUUID(model.NewRandomNullUUID())

		require.NoError(t, provider.SaveOAuth2Session(ctx, OAuth2SessionTypeRefreshToken, s))
		require.NoError(t, provider.RevokeOAuth2Session(ctx, OAuth2SessionTypeRefreshToken, "sig-revoke"))
	})

	t.Run("ShouldDeactivateByRequestID", func(t *testing.T) {
		s := session
		s.Signature = "sig-deact-req"
		s.RequestID = "req-deact"
		s.ChallengeID = model.MustNullUUID(model.NewRandomNullUUID())

		require.NoError(t, provider.SaveOAuth2Session(ctx, OAuth2SessionTypeAuthorizeCode, s))
		require.NoError(t, provider.DeactivateOAuth2SessionByRequestID(ctx, OAuth2SessionTypeAuthorizeCode, "req-deact"))
	})

	t.Run("ShouldRevokeByRequestID", func(t *testing.T) {
		s := session
		s.Signature = "sig-rev-req"
		s.RequestID = "req-rev"
		s.ChallengeID = model.MustNullUUID(model.NewRandomNullUUID())

		require.NoError(t, provider.SaveOAuth2Session(ctx, OAuth2SessionTypeOpenIDConnect, s))
		require.NoError(t, provider.RevokeOAuth2SessionByRequestID(ctx, OAuth2SessionTypeOpenIDConnect, "req-rev"))
	})
}

func TestSQLProviderOAuth2DeviceCodeSession(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	deviceSession := &model.OAuth2DeviceCodeSession{
		Signature:         "dev-sig-123",
		RequestID:         "dev-req-123",
		ClientID:          "test-client",
		UserCodeSignature: "user-code-123",
		Active:            true,
		RequestedScopes:   model.StringSlicePipeDelimited{"openid"},
		GrantedScopes:     model.StringSlicePipeDelimited{"openid"},
		Session:           []byte("{}"),
		RequestedAt:       time.Now().Truncate(time.Second),
	}

	t.Run("ShouldSaveAndLoadBySignature", func(t *testing.T) {
		require.NoError(t, provider.SaveOAuth2DeviceCodeSession(ctx, deviceSession))

		loaded, err := provider.LoadOAuth2DeviceCodeSession(ctx, "dev-sig-123")

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "dev-req-123", loaded.RequestID)
		assert.True(t, loaded.Active)
	})

	t.Run("ShouldLoadByUserCode", func(t *testing.T) {
		loaded, err := provider.LoadOAuth2DeviceCodeSessionByUserCode(ctx, "user-code-123")

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "dev-req-123", loaded.RequestID)
	})

	t.Run("ShouldUpdateDeviceCodeSessionData", func(t *testing.T) {
		deviceSession.GrantedScopes = model.StringSlicePipeDelimited{"openid", "profile"}
		deviceSession.Session = []byte(`{"updated":true}`)

		require.NoError(t, provider.UpdateOAuth2DeviceCodeSessionData(ctx, deviceSession))
	})

	t.Run("ShouldDeactivateDeviceCodeSession", func(t *testing.T) {
		require.NoError(t, provider.DeactivateOAuth2DeviceCodeSession(ctx, "dev-sig-123"))

		loaded, err := provider.LoadOAuth2DeviceCodeSession(ctx, "dev-sig-123")

		require.NoError(t, err)
		assert.False(t, loaded.Active)
	})
}

func TestSQLProviderOAuth2PARContext(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	par := model.OAuth2PARContext{
		Signature:   "par-sig-123",
		RequestID:   "par-req-123",
		ClientID:    "test-client",
		RequestedAt: time.Now().Truncate(time.Second),
		Session:     []byte("{}"),
	}

	t.Run("ShouldSaveAndLoad", func(t *testing.T) {
		require.NoError(t, provider.SaveOAuth2PARContext(ctx, par))

		loaded, err := provider.LoadOAuth2PARContext(ctx, "par-sig-123")

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "par-req-123", loaded.RequestID)
	})

	t.Run("ShouldRevoke", func(t *testing.T) {
		require.NoError(t, provider.RevokeOAuth2PARContext(ctx, "par-sig-123"))

		loaded, err := provider.LoadOAuth2PARContext(ctx, "par-sig-123")

		require.NoError(t, err)
		assert.True(t, loaded.Revoked)
	})
}

func TestSQLProviderOAuth2BlacklistedJTI(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldSaveAndLoad", func(t *testing.T) {
		jti := model.OAuth2BlacklistedJTI{
			Signature: "jti-sig-123",
			ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
		}

		require.NoError(t, provider.SaveOAuth2BlacklistedJTI(ctx, jti))

		loaded, err := provider.LoadOAuth2BlacklistedJTI(ctx, "jti-sig-123")

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "jti-sig-123", loaded.Signature)
	})

	t.Run("ShouldReturnErrForUnknown", func(t *testing.T) {
		_, err := provider.LoadOAuth2BlacklistedJTI(ctx, "unknown")

		assert.Error(t, err)
	})
}

func TestSQLProviderBannedUser(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldSaveAndLoadByUsername", func(t *testing.T) {
		ban := &model.BannedUser{
			Username: "baduser",
			Source:   "cli",
			Expires:  sql.NullTime{Valid: true, Time: time.Now().Add(time.Hour).Truncate(time.Second)},
		}

		require.NoError(t, provider.SaveBannedUser(ctx, ban))

		bans, err := provider.LoadBannedUser(ctx, "baduser")

		require.NoError(t, err)
		assert.NotEmpty(t, bans)
		assert.Equal(t, "baduser", bans[0].Username)
	})

	t.Run("ShouldLoadByID", func(t *testing.T) {
		bans, err := provider.LoadBannedUser(ctx, "baduser")
		require.NoError(t, err)
		require.NotEmpty(t, bans)

		ban, err := provider.LoadBannedUserByID(ctx, bans[0].ID)

		require.NoError(t, err)
		assert.Equal(t, "baduser", ban.Username)
	})

	t.Run("ShouldLoadBannedUsers", func(t *testing.T) {
		bans, err := provider.LoadBannedUsers(ctx, 10, 0)

		require.NoError(t, err)
		assert.NotEmpty(t, bans)
	})

	t.Run("ShouldRevokeBan", func(t *testing.T) {
		bans, err := provider.LoadBannedUser(ctx, "baduser")
		require.NoError(t, err)
		require.NotEmpty(t, bans)

		require.NoError(t, provider.RevokeBannedUser(ctx, bans[0].ID, time.Now()))
	})
}

func TestSQLProviderBannedIP(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldSaveAndLoadByIP", func(t *testing.T) {
		ban := &model.BannedIP{
			IP:      model.NewIP(net.ParseIP("192.168.1.100")),
			Source:  "cli",
			Expires: sql.NullTime{Valid: true, Time: time.Now().Add(time.Hour).Truncate(time.Second)},
		}

		require.NoError(t, provider.SaveBannedIP(ctx, ban))

		bans, err := provider.LoadBannedIP(ctx, model.NewIP(net.ParseIP("192.168.1.100")))

		require.NoError(t, err)
		assert.NotEmpty(t, bans)
	})

	t.Run("ShouldLoadByID", func(t *testing.T) {
		bans, err := provider.LoadBannedIP(ctx, model.NewIP(net.ParseIP("192.168.1.100")))
		require.NoError(t, err)
		require.NotEmpty(t, bans)

		ban, err := provider.LoadBannedIPByID(ctx, bans[0].ID)

		require.NoError(t, err)
		assert.Equal(t, "192.168.1.100", ban.IP.String())
	})

	t.Run("ShouldLoadBannedIPs", func(t *testing.T) {
		bans, err := provider.LoadBannedIPs(ctx, 10, 0)

		require.NoError(t, err)
		assert.NotEmpty(t, bans)
	})

	t.Run("ShouldRevokeBan", func(t *testing.T) {
		bans, err := provider.LoadBannedIP(ctx, model.NewIP(net.ParseIP("192.168.1.100")))
		require.NoError(t, err)
		require.NotEmpty(t, bans)

		require.NoError(t, provider.RevokeBannedIP(ctx, bans[0].ID, time.Now()))
	})
}

func TestSQLProviderLoadRegulationRecordsByIP(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldReturnEmptyForUnknownIP", func(t *testing.T) {
		records, err := provider.LoadRegulationRecordsByIP(ctx, model.NewIP(net.ParseIP("10.0.0.1")), time.Now().Add(-time.Hour), 10)

		assert.NoError(t, err)
		assert.Empty(t, records)
	})
}

func TestSQLProviderOneTimeCode(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	publicID, err := uuid.NewRandom()
	require.NoError(t, err)

	code := model.OneTimeCode{
		PublicID:  publicID,
		IssuedAt:  time.Now().Truncate(time.Second),
		IssuedIP:  model.NewIP(net.ParseIP("127.0.0.1")),
		ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
		Username:  "john",
		Intent:    "reset_password",
		Code:      []byte("123456"),
	}

	var signature string

	t.Run("ShouldSaveOneTimeCode", func(t *testing.T) {
		var err error

		signature, err = provider.SaveOneTimeCode(ctx, code)

		require.NoError(t, err)
		assert.NotEmpty(t, signature)
	})

	t.Run("ShouldLoadOneTimeCode", func(t *testing.T) {
		loaded, err := provider.LoadOneTimeCode(ctx, "john", "reset_password", "123456")

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "john", loaded.Username)
	})

	t.Run("ShouldLoadOneTimeCodeBySignature", func(t *testing.T) {
		loaded, err := provider.LoadOneTimeCodeBySignature(ctx, signature)

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "john", loaded.Username)
	})

	t.Run("ShouldLoadOneTimeCodeByPublicID", func(t *testing.T) {
		loaded, err := provider.LoadOneTimeCodeByPublicID(ctx, publicID)

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "john", loaded.Username)
	})

	t.Run("ShouldConsumeOneTimeCode", func(t *testing.T) {
		loaded, err := provider.LoadOneTimeCodeBySignature(ctx, signature)
		require.NoError(t, err)

		loaded.ConsumedAt = sql.NullTime{Valid: true, Time: time.Now().Truncate(time.Second)}
		loaded.ConsumedIP = model.NewNullIP(net.ParseIP("127.0.0.1"))

		require.NoError(t, provider.ConsumeOneTimeCode(ctx, loaded))
	})

	t.Run("ShouldRevokeOneTimeCode", func(t *testing.T) {
		pubID2, _ := uuid.NewRandom()

		code2 := model.OneTimeCode{
			PublicID:  pubID2,
			IssuedAt:  time.Now().Truncate(time.Second),
			IssuedIP:  model.NewIP(net.ParseIP("127.0.0.1")),
			ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
			Username:  "jane",
			Intent:    "reset_password",
			Code:      []byte("654321"),
		}

		_, err := provider.SaveOneTimeCode(ctx, code2)
		require.NoError(t, err)

		require.NoError(t, provider.RevokeOneTimeCode(ctx, pubID2, model.NewIP(net.ParseIP("127.0.0.1"))))
	})

	t.Run("ShouldLoadOneTimeCodeByID", func(t *testing.T) {
		loaded, err := provider.LoadOneTimeCodeBySignature(ctx, signature)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		byID, err := provider.LoadOneTimeCodeByID(ctx, loaded.ID)

		require.NoError(t, err)
		require.NotNil(t, byID)
		assert.Equal(t, loaded.ID, byID.ID)
	})

	t.Run("ShouldReturnNilForUnknownCode", func(t *testing.T) {
		loaded, err := provider.LoadOneTimeCode(ctx, "nobody", "unknown", "000000")

		assert.NoError(t, err)
		assert.Nil(t, loaded)
	})
}

func TestSQLProviderLoadIdentityVerification(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldSaveAndLoad", func(t *testing.T) {
		jti, err := uuid.NewRandom()
		require.NoError(t, err)

		verification := model.IdentityVerification{
			JTI:       jti,
			IssuedAt:  time.Now().Truncate(time.Second),
			IssuedIP:  model.NewIP(net.ParseIP("127.0.0.1")),
			ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
			Action:    "reset_password",
			Username:  "john",
		}

		require.NoError(t, provider.SaveIdentityVerification(ctx, verification))

		loaded, err := provider.LoadIdentityVerification(ctx, jti.String())

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "john", loaded.Username)
		assert.Equal(t, "reset_password", loaded.Action)
	})

	t.Run("ShouldErrForUnknown", func(t *testing.T) {
		id, _ := uuid.NewRandom()

		_, err := provider.LoadIdentityVerification(ctx, id.String())

		assert.Error(t, err)
	})
}

func TestSQLProviderOAuth2ConsentPreConfiguration(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	subject, _ := uuid.NewRandom()

	t.Run("ShouldSaveAndLoadPreConfiguration", func(t *testing.T) {
		config := model.OAuth2ConsentPreConfig{
			ClientID:  "test-client",
			Subject:   subject,
			CreatedAt: time.Now().Truncate(time.Second),
			Scopes:    model.StringSlicePipeDelimited{"openid", "profile"},
			Audience:  model.StringSlicePipeDelimited{"https://example.com"},
		}

		id, err := provider.SaveOAuth2ConsentPreConfiguration(ctx, config)

		require.NoError(t, err)
		assert.Greater(t, id, int64(0))
	})

	t.Run("ShouldLoadPreConfigurations", func(t *testing.T) {
		rows, err := provider.LoadOAuth2ConsentPreConfigurations(ctx, "test-client", subject, time.Now())

		require.NoError(t, err)
		require.NotNil(t, rows)

		rows.Close()
	})

	t.Run("ShouldReturnEmptyForUnknownClient", func(t *testing.T) {
		unknownSubject, _ := uuid.NewRandom()

		rows, err := provider.LoadOAuth2ConsentPreConfigurations(ctx, "unknown-client", unknownSubject, time.Now())

		require.NoError(t, err)
		require.NotNil(t, rows)

		rows.Close()
	})
}

func TestSQLProviderUpdateOAuth2PARContext(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldErrWhenIDIsZero", func(t *testing.T) {
		par := model.OAuth2PARContext{
			ID:        0,
			Signature: "par-sig",
			RequestID: "par-req",
		}

		err := provider.UpdateOAuth2PARContext(ctx, par)

		assert.ErrorContains(t, err, "the id was a zero value")
	})

	t.Run("ShouldUpdateExistingPAR", func(t *testing.T) {
		par := model.OAuth2PARContext{
			Signature:   "par-update-sig",
			RequestID:   "par-update-req",
			ClientID:    "test-client",
			RequestedAt: time.Now().Truncate(time.Second),
			Session:     []byte("{}"),
		}

		require.NoError(t, provider.SaveOAuth2PARContext(ctx, par))

		loaded, err := provider.LoadOAuth2PARContext(ctx, "par-update-sig")
		require.NoError(t, err)

		loaded.ClientID = "updated-client-id-par"

		require.NoError(t, provider.UpdateOAuth2PARContext(ctx, *loaded))
	})
}

func TestSQLProviderUpdateOAuth2DeviceCodeSession(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldUpdateFullSession", func(t *testing.T) {
		session := &model.OAuth2DeviceCodeSession{
			Signature:         "dev-update-sig",
			RequestID:         "dev-update-req",
			ClientID:          "test-client",
			UserCodeSignature: "user-update-code",
			Active:            true,
			RequestedScopes:   model.StringSlicePipeDelimited{"openid"},
			GrantedScopes:     model.StringSlicePipeDelimited{"openid"},
			Session:           []byte("{}"),
			RequestedAt:       time.Now().Truncate(time.Second),
		}

		require.NoError(t, provider.SaveOAuth2DeviceCodeSession(ctx, session))

		session.ClientID = "updated-client"
		session.Session = []byte(`{"updated":true}`)

		require.NoError(t, provider.UpdateOAuth2DeviceCodeSession(ctx, session))
	})
}

func TestSQLProviderOAuth2SessionAllTypes(t *testing.T) {
	testCases := []struct {
		name        string
		sessionType OAuth2SessionType
	}{
		{"ShouldHandleAccessToken", OAuth2SessionTypeAccessToken},
		{"ShouldHandleRefreshToken", OAuth2SessionTypeRefreshToken},
		{"ShouldHandleAuthorizeCode", OAuth2SessionTypeAuthorizeCode},
		{"ShouldHandleOpenIDConnect", OAuth2SessionTypeOpenIDConnect},
		{"ShouldHandlePKCEChallenge", OAuth2SessionTypePKCEChallenge},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			challengeID := model.MustNullUUID(model.NewRandomNullUUID())
			sig, _ := uuid.NewRandom()

			session := model.OAuth2Session{
				ChallengeID:     challengeID,
				RequestID:       "req-" + sig.String(),
				ClientID:        "test-client",
				Signature:       sig.String(),
				Subject:         sql.NullString{Valid: true, String: "john"},
				Active:          true,
				RequestedScopes: model.StringSlicePipeDelimited{"openid"},
				GrantedScopes:   model.StringSlicePipeDelimited{"openid"},
				Session:         []byte("{}"),
			}

			require.NoError(t, provider.SaveOAuth2Session(ctx, tc.sessionType, session))

			loaded, err := provider.LoadOAuth2Session(ctx, tc.sessionType, sig.String())
			require.NoError(t, err)
			require.NotNil(t, loaded)
			assert.True(t, loaded.Active)

			require.NoError(t, provider.DeactivateOAuth2Session(ctx, tc.sessionType, sig.String()))

			loaded, err = provider.LoadOAuth2Session(ctx, tc.sessionType, sig.String())
			require.NoError(t, err)
			assert.False(t, loaded.Active)
		})
	}
}

func newTestSQLiteProvider(t *testing.T) *SQLiteProvider {
	t.Helper()

	config := &schema.Configuration{
		Storage: schema.Storage{
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "db.sqlite3"),
			},
		},
	}

	provider := NewSQLiteProvider(config)

	require.NotNil(t, provider)

	return provider
}
