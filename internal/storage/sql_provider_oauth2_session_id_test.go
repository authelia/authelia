// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestSQLProviderOAuth2SessionID(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldBeIdempotent", func(t *testing.T) {
		first, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "", "public-id")
		require.NoError(t, err)
		require.NotNil(t, first)

		second, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "", "public-id")
		require.NoError(t, err)
		require.NotNil(t, second)

		assert.Equal(t, first.SessionID, second.SessionID)
		assert.Equal(t, uuid.Version(4), first.SessionID.Version())
	})

	t.Run("ShouldSeparateSectors", func(t *testing.T) {
		a, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "https://a.example.com", "sector-public-id")
		require.NoError(t, err)

		b, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "https://b.example.com", "sector-public-id")
		require.NoError(t, err)

		assert.NotEqual(t, a.SessionID, b.SessionID)
	})

	t.Run("ShouldRoundTripBySessionID", func(t *testing.T) {
		created, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "", "roundtrip-public-id")
		require.NoError(t, err)

		loaded, err := provider.LoadOAuth2SessionIDBySessionID(ctx, "issuer", created.SessionID.String())
		require.NoError(t, err)
		require.NotNil(t, loaded)

		assert.Equal(t, "roundtrip-public-id", loaded.PublicID)
	})

	t.Run("ShouldReturnNilWhenAbsent", func(t *testing.T) {
		loaded, err := provider.LoadOAuth2SessionIDBySessionID(ctx, "issuer", uuid.Must(uuid.NewRandom()).String())

		assert.NoError(t, err)
		assert.Nil(t, loaded)
	})

	t.Run("ShouldSeparateIssuers", func(t *testing.T) {
		a, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer-a", "", "shared-public-id")
		require.NoError(t, err)

		loaded, err := provider.LoadOAuth2SessionIDBySessionID(ctx, "issuer-b", a.SessionID.String())

		assert.NoError(t, err)
		assert.Nil(t, loaded)
	})

	t.Run("ShouldEndEverySectorByPublicID", func(t *testing.T) {
		a, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "https://a.example.com", "end-public-id")
		require.NoError(t, err)

		b, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "https://b.example.com", "end-public-id")
		require.NoError(t, err)

		unrelated, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "https://a.example.com", "end-unrelated-public-id")
		require.NoError(t, err)

		sids, err := provider.EndOAuth2Session(ctx, "issuer", "end-public-id", testLogoutSchedule())
		require.NoError(t, err)
		assert.Equal(t, []string{a.SessionID.String(), b.SessionID.String()}, sids)

		for _, record := range []*model.OAuth2SessionID{a, b} {
			loaded, err := provider.LoadOAuth2SessionIDBySessionID(ctx, "issuer", record.SessionID.String())
			require.NoError(t, err)
			assert.Nil(t, loaded)

			ended, err := provider.IsOAuth2SessionEnded(ctx, record.SessionID.String())
			require.NoError(t, err)
			assert.True(t, ended)
		}

		loaded, err := provider.LoadOAuth2SessionIDBySessionID(ctx, "issuer", unrelated.SessionID.String())
		require.NoError(t, err)
		assert.NotNil(t, loaded)

		ended, err := provider.IsOAuth2SessionEnded(ctx, unrelated.SessionID.String())
		require.NoError(t, err)
		assert.False(t, ended)

		sids, err = provider.EndOAuth2Session(ctx, "issuer", "end-public-id", testLogoutSchedule())
		require.NoError(t, err)
		assert.Empty(t, sids)
	})

	t.Run("ShouldEndOnlyMatchingIssuer", func(t *testing.T) {
		created, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer-end-single", "", "end-single-public-id")
		require.NoError(t, err)

		sids, err := provider.EndOAuth2Session(ctx, "issuer-end-single-other", "end-single-public-id", testLogoutSchedule())
		require.NoError(t, err)
		assert.Empty(t, sids)

		loaded, err := provider.LoadOAuth2SessionIDBySessionID(ctx, "issuer-end-single", created.SessionID.String())
		require.NoError(t, err)
		assert.NotNil(t, loaded)
	})

	t.Run("ShouldPageOldestByAfterAndLimit", func(t *testing.T) {
		paging := newTestSQLiteProvider(t)
		require.NoError(t, paging.StartupCheck())

		first, err := paging.GetOrCreateOAuth2SessionID(ctx, "issuer-paging", "", "paging-public-id-1")
		require.NoError(t, err)

		_, err = paging.GetOrCreateOAuth2SessionID(ctx, "issuer-paging", "", "paging-public-id-2")
		require.NoError(t, err)

		_, err = paging.GetOrCreateOAuth2SessionID(ctx, "issuer-paging", "", "paging-public-id-3")
		require.NoError(t, err)

		firstLoaded, err := paging.LoadOAuth2SessionIDBySessionID(ctx, "issuer-paging", first.SessionID.String())
		require.NoError(t, err)
		require.NotNil(t, firstLoaded)

		page, err := paging.LoadOAuth2SessionIDsOldest(ctx, 0, 1)
		require.NoError(t, err)
		require.Len(t, page, 1)
		assert.Equal(t, firstLoaded.ID, page[0].ID)
		assert.Equal(t, "paging-public-id-1", page[0].PublicID)

		next, err := paging.LoadOAuth2SessionIDsOldest(ctx, page[0].ID, 1)
		require.NoError(t, err)
		require.Len(t, next, 1)
		assert.NotEqual(t, page[0].ID, next[0].ID)
		assert.Equal(t, "paging-public-id-2", next[0].PublicID)
	})
}

func TestSQLProviderOAuth2SessionsShouldPersistSessionID(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	sid := model.NewNullString(uuid.Must(uuid.NewRandom()).String())

	for _, sessionType := range []OAuth2SessionType{
		OAuth2SessionTypeAccessToken,
		OAuth2SessionTypeAuthorizeCode,
		OAuth2SessionTypeOpenIDConnect,
		OAuth2SessionTypePKCEChallenge,
		OAuth2SessionTypeRefreshToken,
	} {
		t.Run(sessionType.String(), func(t *testing.T) {
			signature := "sig-" + sessionType.String()

			require.NoError(t, provider.SaveOAuth2Session(ctx, sessionType, model.OAuth2Session{
				ChallengeID: model.MustNullUUID(model.NewRandomNullUUID()),
				Subject:     model.NewNullString("john"),
				RequestID:   "req-" + sessionType.String(),
				ClientID:    "client-id",
				SessionID:   sid,
				Signature:   signature,
				Active:      true,
				Session:     []byte(`{}`),
			}))

			session, err := provider.LoadOAuth2Session(ctx, sessionType, signature)
			require.NoError(t, err)

			assert.Equal(t, "client-id", session.ClientID)
			assert.Equal(t, sid, session.SessionID)
		})
	}

	t.Run("ShouldPersistAbsentSessionID", func(t *testing.T) {
		require.NoError(t, provider.SaveOAuth2Session(ctx, OAuth2SessionTypeAccessToken, model.OAuth2Session{
			RequestID: "req-no-sid",
			ClientID:  "client-id",
			Signature: "sig-no-sid",
			Active:    true,
			Session:   []byte(`{}`),
		}))

		session, err := provider.LoadOAuth2Session(ctx, OAuth2SessionTypeAccessToken, "sig-no-sid")
		require.NoError(t, err)

		assert.False(t, session.SessionID.Valid)
	})

	t.Run("DeviceCode", func(t *testing.T) {
		device := &model.OAuth2DeviceCodeSession{
			RequestID:         "req-device",
			ClientID:          "client-id",
			Signature:         "sig-device",
			UserCodeSignature: "sig-user-code",
			Active:            true,
			Session:           []byte(`{}`),
		}

		require.NoError(t, provider.SaveOAuth2DeviceCodeSession(ctx, device))

		loaded, err := provider.LoadOAuth2DeviceCodeSession(ctx, "sig-device")
		require.NoError(t, err)

		assert.Equal(t, "client-id", loaded.ClientID)
		assert.False(t, loaded.SessionID.Valid, "the session identifier is not known until the End-User authorizes the device")

		loaded.SessionID = sid

		require.NoError(t, provider.UpdateOAuth2DeviceCodeSessionData(ctx, loaded))

		loaded, err = provider.LoadOAuth2DeviceCodeSession(ctx, "sig-device")
		require.NoError(t, err)

		assert.Equal(t, sid, loaded.SessionID)

		loaded.SessionID = model.NewNullString("")

		require.NoError(t, provider.UpdateOAuth2DeviceCodeSession(ctx, loaded))

		loaded, err = provider.LoadOAuth2DeviceCodeSession(ctx, "sig-device")
		require.NoError(t, err)

		assert.False(t, loaded.SessionID.Valid)
	})

	t.Run("PushedAuthorization", func(t *testing.T) {
		require.NoError(t, provider.SaveOAuth2PushedAuthorizationSession(ctx, model.OAuth2PushedAuthorizationSession{
			RequestID: "req-par",
			ClientID:  "client-id",
			SessionID: sid,
			Signature: "sig-par",
			Session:   []byte(`{}`),
		}))

		par, err := provider.LoadOAuth2PushedAuthorizationSession(ctx, "sig-par")
		require.NoError(t, err)

		assert.Equal(t, "client-id", par.ClientID)
		assert.Equal(t, sid, par.SessionID)

		par.SessionID = model.NewNullString("")

		require.NoError(t, provider.UpdateOAuth2PushedAuthorizationSession(ctx, *par))

		par, err = provider.LoadOAuth2PushedAuthorizationSession(ctx, "sig-par")
		require.NoError(t, err)

		assert.False(t, par.SessionID.Valid)
	})

	t.Run("ShouldPersistTheSessionIDOnDeviceCodeDecision", func(t *testing.T) {
		sid := uuid.Must(uuid.NewRandom()).String()

		require.NoError(t, provider.SaveOAuth2DeviceCodeSession(ctx, &model.OAuth2DeviceCodeSession{
			Signature:         "sig-device-decide",
			RequestID:         "req-device-decide",
			ClientID:          "client",
			UserCodeSignature: "sig-user-device-decide",
			Status:            model.OAuth2DeviceCodeStatusNew,
			Active:            true,
			Session:           []byte(`{}`),
		}))

		require.NoError(t, provider.DecideOAuth2DeviceCodeSession(ctx, &model.OAuth2DeviceCodeSession{
			Signature: "sig-device-decide",
			ClientID:  "client",
			SessionID: model.NewNullString(sid),
			Status:    int(oauthelia2.DeviceAuthorizeStatusApproved),
			Session:   []byte(`{}`),
		}))

		device, err := provider.LoadOAuth2DeviceCodeSession(ctx, "sig-device-decide")
		require.NoError(t, err)

		assert.Equal(t, model.NewNullString(sid), device.SessionID)
	})
}

func TestSQLProviderOAuth2Logout(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	save := func(t *testing.T, sessionType OAuth2SessionType, signature, clientID, subject, sid, mode string) {
		t.Helper()

		require.NoError(t, provider.SaveOAuth2Session(ctx, sessionType, model.OAuth2Session{
			ChallengeID:      model.MustNullUUID(model.NewRandomNullUUID()),
			RequestID:        "req-" + signature,
			ClientID:         clientID,
			SessionID:        model.NewNullString(sid),
			Signature:        signature,
			Subject:          model.NewNullString(subject),
			RefreshTokenMode: mode,
			Active:           true,
			Session:          []byte(`{}`),
		}))
	}

	revoked := func(t *testing.T, sessionType OAuth2SessionType, signature string) bool {
		t.Helper()

		session, err := provider.LoadOAuth2Session(ctx, sessionType, signature)
		if errors.Is(err, sql.ErrNoRows) {
			return true
		}

		require.NoError(t, err)

		return session.Revoked
	}

	t.Run("ShouldRevokeOnlyOnlineRefreshTokensBySessionID", func(t *testing.T) {
		sid, other := uuid.Must(uuid.NewRandom()).String(), uuid.Must(uuid.NewRandom()).String()

		save(t, OAuth2SessionTypeAccessToken, "sid-access", "client", "john", sid, "")
		save(t, OAuth2SessionTypeAuthorizeCode, "sid-code", "client", "john", sid, "")
		save(t, OAuth2SessionTypeOpenIDConnect, "sid-openid", "client", "john", sid, "")
		save(t, OAuth2SessionTypePKCEChallenge, "sid-pkce", "client", "john", sid, "")
		save(t, OAuth2SessionTypeRefreshToken, "sid-refresh", "client", "john", sid, "online")
		save(t, OAuth2SessionTypeRefreshToken, "sid-refresh-unrecorded", "client", "john", sid, "")
		save(t, OAuth2SessionTypeRefreshToken, "sid-refresh-offline", "client", "john", sid, "offline")
		save(t, OAuth2SessionTypeRefreshToken, "other-refresh", "client", "john", other, "online")

		require.NoError(t, provider.RevokeOAuth2RefreshTokenSessionsBySessionID(ctx, sid))

		assert.False(t, revoked(t, OAuth2SessionTypeAccessToken, "sid-access"))
		assert.False(t, revoked(t, OAuth2SessionTypeAuthorizeCode, "sid-code"))
		assert.False(t, revoked(t, OAuth2SessionTypeOpenIDConnect, "sid-openid"))
		assert.False(t, revoked(t, OAuth2SessionTypePKCEChallenge, "sid-pkce"))
		assert.True(t, revoked(t, OAuth2SessionTypeRefreshToken, "sid-refresh"))
		assert.True(t, revoked(t, OAuth2SessionTypeRefreshToken, "sid-refresh-unrecorded"))
		assert.False(t, revoked(t, OAuth2SessionTypeRefreshToken, "sid-refresh-offline"))
		assert.False(t, revoked(t, OAuth2SessionTypeRefreshToken, "other-refresh"))

		require.NoError(t, provider.RevokeOAuth2RefreshTokenSessionsBySessionID(ctx, sid))
	})

	t.Run("ShouldNotConsumeRefreshTokensRevokedBySessionID", func(t *testing.T) {
		sid := uuid.Must(uuid.NewRandom()).String()

		save(t, OAuth2SessionTypeRefreshToken, "consume-refresh", "client", "john", sid, "online")

		require.NoError(t, provider.RevokeOAuth2RefreshTokenSessionsBySessionID(ctx, sid))

		assert.EqualError(t, provider.DeactivateOAuth2Session(ctx, OAuth2SessionTypeRefreshToken, "consume-refresh"), "error deactivating oauth2 refresh token session with signature 'consume-refresh': no rows affected")
		assert.EqualError(t, provider.UpdateOAuth2RefreshTokenSessionRequest(ctx, model.OAuth2Session{Signature: "consume-refresh", Session: []byte(`{}`)}), "error updating oauth2 refresh_token session with signature 'consume-refresh': no rows affected")
	})

	t.Run("ShouldRoundTripRefreshTokenMode", func(t *testing.T) {
		sid := uuid.Must(uuid.NewRandom()).String()

		save(t, OAuth2SessionTypeRefreshToken, "mode-refresh", "client", "john", sid, "offline")

		session, err := provider.LoadOAuth2Session(ctx, OAuth2SessionTypeRefreshToken, "mode-refresh")
		require.NoError(t, err)
		assert.Equal(t, "offline", session.RefreshTokenMode)

		require.NoError(t, provider.UpdateOAuth2RefreshTokenSessionRequest(ctx, model.OAuth2Session{Signature: "mode-refresh", RefreshTokenMode: "online", Session: []byte(`{}`)}))

		session, err = provider.LoadOAuth2Session(ctx, OAuth2SessionTypeRefreshToken, "mode-refresh")
		require.NoError(t, err)
		assert.Equal(t, "online", session.RefreshTokenMode)
	})

	t.Run("ShouldConsumeSessionsWhichAreNotRevoked", func(t *testing.T) {
		sid := uuid.Must(uuid.NewRandom()).String()

		save(t, OAuth2SessionTypeAuthorizeCode, "active-code", "client", "john", sid, "online")
		save(t, OAuth2SessionTypeRefreshToken, "active-refresh", "client", "john", sid, "online")

		require.NoError(t, provider.UpdateOAuth2RefreshTokenSessionRequest(ctx, model.OAuth2Session{Signature: "active-refresh", Session: []byte(`{}`)}))
		require.NoError(t, provider.DeactivateOAuth2Session(ctx, OAuth2SessionTypeAuthorizeCode, "active-code"))
		require.NoError(t, provider.DeactivateOAuth2Session(ctx, OAuth2SessionTypeRefreshToken, "active-refresh"))
	})
}

func TestSQLProviderOAuth2LogoutEndedSession(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	save := func(t *testing.T, sessionType OAuth2SessionType, signature, sid string) {
		t.Helper()

		require.NoError(t, provider.SaveOAuth2Session(ctx, sessionType, model.OAuth2Session{
			ChallengeID: model.MustNullUUID(model.NewRandomNullUUID()),
			RequestID:   "req-" + signature,
			ClientID:    "client",
			SessionID:   model.NewNullString(sid),
			Signature:   signature,
			Subject:     model.NewNullString("john"),
			Active:      true,
			Session:     []byte(`{}`),
		}))
	}

	saveDevice := func(t *testing.T, signature, sid string) {
		t.Helper()

		require.NoError(t, provider.SaveOAuth2DeviceCodeSession(ctx, &model.OAuth2DeviceCodeSession{
			Signature:         signature,
			RequestID:         "req-" + signature,
			ClientID:          "client",
			UserCodeSignature: "user-" + signature,
			SessionID:         model.NewNullString(sid),
			Active:            true,
			Session:           []byte(`{}`),
		}))
	}

	ended := uuid.Must(uuid.NewRandom()).String()
	live := uuid.Must(uuid.NewRandom()).String()

	require.NoError(t, provider.SaveOAuth2SessionLogoutTask(ctx, testLogoutSchedule().RevokeTask(ended)))

	testCases := []struct {
		name        string
		sessionType OAuth2SessionType
		sid         string
		expected    string
	}{
		{"ShouldNotConsumeAuthorizeCodeOfEndedSession", OAuth2SessionTypeAuthorizeCode, ended, "error deactivating oauth2 authorization code session with signature 'ShouldNotConsumeAuthorizeCodeOfEndedSession': no rows affected"},
		{"ShouldNotConsumeDeviceCodeOfEndedSession", OAuth2SessionTypeDeviceAuthorizeCode, ended, "error deactivating oauth2 device code session with signature 'ShouldNotConsumeDeviceCodeOfEndedSession': no rows affected"},
		{"ShouldConsumeAuthorizeCodeOfLiveSession", OAuth2SessionTypeAuthorizeCode, live, ""},
		{"ShouldConsumeDeviceCodeOfLiveSession", OAuth2SessionTypeDeviceAuthorizeCode, live, ""},
		{"ShouldConsumeAuthorizeCodeWithoutSession", OAuth2SessionTypeAuthorizeCode, "", ""},
		{"ShouldConsumeDeviceCodeWithoutSession", OAuth2SessionTypeDeviceAuthorizeCode, "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.sessionType == OAuth2SessionTypeDeviceAuthorizeCode {
				saveDevice(t, tc.name, tc.sid)
			} else {
				save(t, tc.sessionType, tc.name, tc.sid)
			}

			err := provider.DeactivateOAuth2Session(ctx, tc.sessionType, tc.name)

			if tc.expected == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expected)
			}
		})
	}

	t.Run("ShouldReadEndedWithinTransaction", func(t *testing.T) {
		tx, err := provider.BeginTX(ctx)
		require.NoError(t, err)

		value, err := provider.IsOAuth2SessionEnded(tx, ended)
		require.NoError(t, err)
		assert.True(t, value)

		value, err = provider.IsOAuth2SessionEnded(tx, live)
		require.NoError(t, err)
		assert.False(t, value)

		require.NoError(t, provider.Rollback(tx))
	})

	t.Run("ShouldNotConsumeCodesOfEndedSessionWithinTransaction", func(t *testing.T) {
		save(t, OAuth2SessionTypeAuthorizeCode, "tx-code-ended", ended)
		save(t, OAuth2SessionTypeAuthorizeCode, "tx-code-live", live)
		saveDevice(t, "tx-device-ended", ended)

		tx, err := provider.BeginTX(ctx)
		require.NoError(t, err)

		assert.EqualError(t, provider.DeactivateOAuth2Session(tx, OAuth2SessionTypeAuthorizeCode, "tx-code-ended"), "error deactivating oauth2 authorization code session with signature 'tx-code-ended': no rows affected")
		assert.EqualError(t, provider.DeactivateOAuth2Session(tx, OAuth2SessionTypeDeviceAuthorizeCode, "tx-device-ended"), "error deactivating oauth2 device code session with signature 'tx-device-ended': no rows affected")
		assert.NoError(t, provider.DeactivateOAuth2Session(tx, OAuth2SessionTypeAuthorizeCode, "tx-code-live"))

		require.NoError(t, provider.Commit(tx))
	})
}

func TestSQLProviderOAuth2SessionLogoutTask(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	sid := uuid.Must(uuid.NewRandom()).String()

	schedule := model.OAuth2SessionLogoutTaskSchedule{CreatedAt: now, NextAttemptAt: now.Add(time.Minute), ExpiresAt: now.Add(time.Hour)}

	require.NoError(t, provider.SaveOAuth2SessionLogoutTask(ctx, schedule.RevokeTask(sid)))
	require.NoError(t, provider.SaveOAuth2SessionLogoutTask(ctx, model.OAuth2SessionLogoutTaskSchedule{CreatedAt: now, NextAttemptAt: now, ExpiresAt: now}.RevokeTask(sid)))

	tasks, err := provider.LoadOAuth2SessionLogoutTasksDue(ctx, now, 10)
	require.NoError(t, err)
	assert.Empty(t, tasks)

	tasks, err = provider.LoadOAuth2SessionLogoutTasksDue(ctx, now.Add(time.Minute), 10)
	require.NoError(t, err)
	require.Len(t, tasks, 1)

	task := tasks[0]

	assert.Equal(t, model.OAuth2SessionLogoutTaskKindRevoke, task.Kind)
	assert.Equal(t, sid, task.SessionID)
	assert.Equal(t, 0, task.Attempts)
	assert.True(t, now.Add(time.Hour).Equal(task.ExpiresAt))

	claimed, err := provider.ClaimOAuth2SessionLogoutTask(ctx, task.ID, task.Attempts, now.Add(2*time.Minute))
	require.NoError(t, err)
	assert.True(t, claimed)

	claimed, err = provider.ClaimOAuth2SessionLogoutTask(ctx, task.ID, task.Attempts, now.Add(2*time.Minute))
	require.NoError(t, err)
	assert.False(t, claimed)

	rescheduled, err := provider.RescheduleOAuth2SessionLogoutTask(ctx, task.ID, task.Attempts, now.Add(time.Hour))
	require.NoError(t, err)
	assert.False(t, rescheduled)

	rescheduled, err = provider.RescheduleOAuth2SessionLogoutTask(ctx, task.ID, task.Attempts+1, now.Add(time.Hour))
	require.NoError(t, err)
	assert.True(t, rescheduled)

	tasks, err = provider.LoadOAuth2SessionLogoutTasksDue(ctx, now.Add(30*time.Minute), 10)
	require.NoError(t, err)
	assert.Empty(t, tasks)

	deleted, err := provider.DeleteOAuth2SessionLogoutTask(ctx, task.ID, task.Attempts)
	require.NoError(t, err)
	assert.False(t, deleted)

	deleted, err = provider.DeleteOAuth2SessionLogoutTask(ctx, task.ID, task.Attempts+1)
	require.NoError(t, err)
	assert.True(t, deleted)

	ended, err := provider.IsOAuth2SessionEnded(ctx, sid)
	require.NoError(t, err)
	assert.False(t, ended)
}

func testLogoutSchedule() model.OAuth2SessionLogoutTaskSchedule {
	now := time.Now()

	return model.OAuth2SessionLogoutTaskSchedule{CreatedAt: now, NextAttemptAt: now.Add(time.Minute), ExpiresAt: now.Add(time.Hour)}
}
