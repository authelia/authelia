package model_test

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestOpenIDConnectLinkMarshalJSON(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	lastUsedAt := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		link     model.OpenIDConnectLink
		expected string
	}{
		{
			"ShouldOmitLastUsedAtAndRemoteUsernameWhenUnset",
			model.OpenIDConnectLink{
				ID:        1,
				CreatedAt: createdAt,
				Provider:  "example",
				Issuer:    "https://op.example.com",
				Subject:   "abc123",
				Username:  "confidential-john",
			},
			`{"id":1,"created_at":"2026-01-01T00:00:00Z","provider":"example","issuer":"https://op.example.com","subject":"abc123"}`,
		},
		{
			"ShouldIncludeLastUsedAtAndRemoteUsernameWhenSet",
			model.OpenIDConnectLink{
				ID:             1,
				CreatedAt:      createdAt,
				LastUsedAt:     sql.NullTime{Valid: true, Time: lastUsedAt},
				Provider:       "example",
				Issuer:         "https://op.example.com",
				Subject:        "abc123",
				Username:       "confidential-john",
				RemoteUsername: sql.NullString{Valid: true, String: "john@idp.example"},
			},
			`{"id":1,"created_at":"2026-01-01T00:00:00Z","last_used_at":"2026-02-01T00:00:00Z","provider":"example","issuer":"https://op.example.com","subject":"abc123","remote_username":"john@idp.example"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.link)

			require.NoError(t, err)
			assert.JSONEq(t, tc.expected, string(data))
			assert.NotContains(t, string(data), tc.link.Username)
		})
	}
}
