package model

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOAuth2SessionFromRequest(t *testing.T) {
	challenge := NullUUID(uuid.Must(uuid.Parse("a9e4638d-e273-4636-a43e-3b34cc9a76ee")))
	session := &OpenIDSession{
		ChallengeID: challenge,
		DefaultSession: &openid.DefaultSession{
			Subject: "sub",
		},
	}

	sessionBytes, _ := json.Marshal(session)

	testCaases := []struct {
		name      string
		signature string
		have      fosite.Requester
		expected  *OAuth2Session
		err       string
	}{
		{
			"ShouldNewUpStandard",
			"abc",
			&fosite.Request{
				ID: "example",
				Client: &fosite.DefaultClient{
					ID: "client_id",
				},
				Session:        session,
				RequestedScope: fosite.Arguments{"openid"},
				GrantedScope:   fosite.Arguments{"openid"},
			},
			&OAuth2Session{
				ChallengeID:     challenge,
				RequestID:       "example",
				ClientID:        "client_id",
				Signature:       "abc",
				Subject:         sql.NullString{String: "sub", Valid: true},
				RequestedScopes: StringSlicePipeDelimited{"openid"},
				GrantedScopes:   StringSlicePipeDelimited{"openid"},
				Active:          true,
				Session:         sessionBytes,
			},
			"",
		},
		{
			"ShouldNewUpWithoutScopes",
			"abc",
			&fosite.Request{
				ID: "example",
				Client: &fosite.DefaultClient{
					ID: "client_id",
				},
				Session:        session,
				RequestedScope: nil,
				GrantedScope:   nil,
			},
			&OAuth2Session{
				ChallengeID:     challenge,
				RequestID:       "example",
				ClientID:        "client_id",
				Signature:       "abc",
				Subject:         sql.NullString{String: "sub", Valid: true},
				RequestedScopes: StringSlicePipeDelimited{},
				GrantedScopes:   StringSlicePipeDelimited{},
				Active:          true,
				Session:         sessionBytes,
			},
			"",
		},
		{
			"ShouldRaiseErrorOnInvalidSessionType",
			"abc",
			&fosite.Request{
				ID: "example",
				Client: &fosite.DefaultClient{
					ID: "client_id",
				},
				Session:        &openid.DefaultSession{},
				RequestedScope: nil,
				GrantedScope:   nil,
			},
			nil,
			"failed to create new *model.OAuth2Session: the session type *model.OpenIDSession was expected but the type '*openid.DefaultSession' was used",
		},
		{
			"ShouldRaiseErrorOnNilRequester",
			"abc",
			nil,
			nil,
			"failed to create new *model.OAuth2Session: the fosite.Requester was nil",
		},
	}

	for _, tc := range testCaases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewOAuth2SessionFromRequest(tc.signature, tc.have)

			if len(tc.err) > 0 {
				assert.Nil(t, actual)
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, actual)

				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestOAuth2Session_SetSubject(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected sql.NullString
	}{
		{
			"ShouldParseValidNullString",
			"example",
			sql.NullString{String: "example", Valid: true},
		},
		{
			"ShouldParseEmptyNullString",
			"",
			sql.NullString{String: "", Valid: false},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			x := &OAuth2Session{}

			assert.Equal(t, x.Subject, sql.NullString{})

			x.SetSubject(tc.have)

			assert.Equal(t, tc.expected, x.Subject)
		})
	}
}
