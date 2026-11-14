// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestSQLProviderClaimOAuth2DeviceCodeSession(t *testing.T) {
	challenge := uuid.NullUUID{UUID: uuid.MustParse("6d7e6b0e-5d3a-4c8e-9b1e-1f2a3b4c5d6e"), Valid: true}
	other := uuid.NullUUID{UUID: uuid.MustParse("0f1e2d3c-4b5a-4968-8776-655443322110"), Valid: true}

	testCases := []struct {
		name     string
		existing func(session *model.OAuth2DeviceCodeSession)
		setup    func(t *testing.T, provider *SQLiteProvider, session *model.OAuth2DeviceCodeSession)
		claim    func(session *model.OAuth2DeviceCodeSession)
		err      string
		expected func(t *testing.T, loaded *model.OAuth2DeviceCodeSession)
	}{
		{
			name: "ShouldClaimNewSession",
			claim: func(session *model.OAuth2DeviceCodeSession) {
				session.ChallengeID, session.Status = challenge, model.OAuth2DeviceCodeStatusClaimed
			},
			expected: func(t *testing.T, loaded *model.OAuth2DeviceCodeSession) {
				assert.Equal(t, challenge, loaded.ChallengeID)
				assert.Equal(t, model.OAuth2DeviceCodeStatusClaimed, loaded.Status)
				assert.True(t, loaded.Active)
			},
		},
		{
			name: "ShouldRecordDenial",
			claim: func(session *model.OAuth2DeviceCodeSession) {
				session.ChallengeID, session.Status, session.Active = challenge, int(oauthelia2.DeviceAuthorizeStatusDenied), false
			},
			expected: func(t *testing.T, loaded *model.OAuth2DeviceCodeSession) {
				assert.Equal(t, challenge, loaded.ChallengeID)
				assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusDenied), loaded.Status)
				assert.False(t, loaded.Active)
			},
		},
		{
			name: "ShouldNotClaimSessionTwice",
			setup: func(t *testing.T, provider *SQLiteProvider, session *model.OAuth2DeviceCodeSession) {
				claimed := *session
				claimed.ChallengeID, claimed.Status = other, model.OAuth2DeviceCodeStatusClaimed

				require.NoError(t, provider.ClaimOAuth2DeviceCodeSession(context.Background(), &claimed))
			},
			claim: func(session *model.OAuth2DeviceCodeSession) {
				session.ChallengeID, session.Status = challenge, model.OAuth2DeviceCodeStatusClaimed
			},
			err: "error claiming oauth2 device code session with signature 'dev-claim-sig': no rows affected",
			expected: func(t *testing.T, loaded *model.OAuth2DeviceCodeSession) {
				assert.Equal(t, other, loaded.ChallengeID)
			},
		},
		{
			name: "ShouldNotClaimSessionWithChallenge",
			existing: func(session *model.OAuth2DeviceCodeSession) {
				session.ChallengeID = other
			},
			claim: func(session *model.OAuth2DeviceCodeSession) {
				session.ChallengeID, session.Status = challenge, model.OAuth2DeviceCodeStatusClaimed
			},
			err: "error claiming oauth2 device code session with signature 'dev-claim-sig': no rows affected",
		},
		{
			name: "ShouldNotClaimDecidedSession",
			existing: func(session *model.OAuth2DeviceCodeSession) {
				session.Status = int(oauthelia2.DeviceAuthorizeStatusApproved)
			},
			claim: func(session *model.OAuth2DeviceCodeSession) {
				session.ChallengeID, session.Status = challenge, model.OAuth2DeviceCodeStatusClaimed
			},
			err: "error claiming oauth2 device code session with signature 'dev-claim-sig': no rows affected",
		},
		{
			name: "ShouldNotClaimInactiveSession",
			existing: func(session *model.OAuth2DeviceCodeSession) {
				session.Active = false
			},
			claim: func(session *model.OAuth2DeviceCodeSession) {
				session.ChallengeID, session.Status, session.Active = challenge, model.OAuth2DeviceCodeStatusClaimed, true
			},
			err: "error claiming oauth2 device code session with signature 'dev-claim-sig': no rows affected",
		},
		{
			name: "ShouldNotClaimRevokedSession",
			existing: func(session *model.OAuth2DeviceCodeSession) {
				session.Revoked = true
			},
			claim: func(session *model.OAuth2DeviceCodeSession) {
				session.ChallengeID, session.Status = challenge, model.OAuth2DeviceCodeStatusClaimed
			},
			err: "error claiming oauth2 device code session with signature 'dev-claim-sig': no rows affected",
		},
		{
			name: "ShouldNotClaimUnknownSession",
			claim: func(session *model.OAuth2DeviceCodeSession) {
				session.Signature, session.ChallengeID, session.Status = "unknown", challenge, model.OAuth2DeviceCodeStatusClaimed
			},
			err: "error claiming oauth2 device code session with signature 'unknown': no rows affected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			session := newTestOAuth2DeviceCodeSession()

			if tc.existing != nil {
				tc.existing(session)
			}

			require.NoError(t, provider.SaveOAuth2DeviceCodeSession(ctx, session))

			if tc.setup != nil {
				tc.setup(t, provider, session)
			}

			claim := *session

			tc.claim(&claim)

			err := provider.ClaimOAuth2DeviceCodeSession(ctx, &claim)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}

			if tc.expected != nil {
				loaded, err := provider.LoadOAuth2DeviceCodeSession(ctx, session.Signature)
				require.NoError(t, err)

				tc.expected(t, loaded)
			}
		})
	}
}

func TestSQLProviderDecideOAuth2DeviceCodeSession(t *testing.T) {
	challenge := uuid.NullUUID{UUID: uuid.MustParse("6d7e6b0e-5d3a-4c8e-9b1e-1f2a3b4c5d6e"), Valid: true}
	other := uuid.NullUUID{UUID: uuid.MustParse("0f1e2d3c-4b5a-4968-8776-655443322110"), Valid: true}

	testCases := []struct {
		name    string
		claimed bool
		decided bool
		decide  uuid.NullUUID
		err     string
	}{
		{
			name:    "ShouldDecideSessionClaimedByTheConsentSession",
			claimed: true,
			decide:  challenge,
		},
		{
			name:    "ShouldNotDecideSessionClaimedByAnotherConsentSession",
			claimed: true,
			decide:  other,
			err:     "error deciding oauth2 device code session with signature 'dev-claim-sig': no rows affected",
		},
		{
			name:   "ShouldNotDecideUnclaimedSession",
			decide: challenge,
			err:    "error deciding oauth2 device code session with signature 'dev-claim-sig': no rows affected",
		},
		{
			name:    "ShouldNotDecideSessionTwice",
			claimed: true,
			decided: true,
			decide:  challenge,
			err:     "error deciding oauth2 device code session with signature 'dev-claim-sig': no rows affected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			session := newTestOAuth2DeviceCodeSession()

			require.NoError(t, provider.SaveOAuth2DeviceCodeSession(ctx, session))

			if tc.claimed {
				claimed := *session
				claimed.ChallengeID, claimed.Status = challenge, model.OAuth2DeviceCodeStatusClaimed

				require.NoError(t, provider.ClaimOAuth2DeviceCodeSession(ctx, &claimed))
			}

			decision := *session
			decision.ChallengeID, decision.Status, decision.Session = tc.decide, int(oauthelia2.DeviceAuthorizeStatusApproved), []byte("{}")

			if tc.decided {
				first := decision

				require.NoError(t, provider.DecideOAuth2DeviceCodeSession(ctx, &first))
			}

			err := provider.DecideOAuth2DeviceCodeSession(ctx, &decision)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)

				loaded, err := provider.LoadOAuth2DeviceCodeSession(ctx, session.Signature)
				require.NoError(t, err)

				assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusApproved), loaded.Status)
				assert.Equal(t, challenge, loaded.ChallengeID)
			}
		})
	}
}

func newTestOAuth2DeviceCodeSession() *model.OAuth2DeviceCodeSession {
	return &model.OAuth2DeviceCodeSession{
		Signature:         "dev-claim-sig",
		RequestID:         "dev-claim-req",
		ClientID:          "test-client",
		UserCodeSignature: "user-claim-code",
		Status:            model.OAuth2DeviceCodeStatusNew,
		Active:            true,
		RequestedScopes:   model.StringSlicePipeDelimited{"openid"},
		Session:           []byte("{}"),
		RequestedAt:       time.Now().Truncate(time.Second),
	}
}
