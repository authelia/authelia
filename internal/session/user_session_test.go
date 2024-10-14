package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/authorization"
)

func TestUserSession_SetTwoFactorWebAuthn(t *testing.T) {
	testCases := []struct {
		name                         string
		at                           time.Time
		hardware, presence, verified bool
		expected                     authorization.AuthenticationMethodsReferences
	}{
		{
			"ShouldHandleHardware",
			time.Unix(1000, 0),
			true,
			true,
			true,
			authorization.AuthenticationMethodsReferences{
				WebAuthn:             true,
				WebAuthnHardware:     true,
				WebAuthnUserPresence: true,
				WebAuthnUserVerified: true,
			},
		},
		{
			"ShouldHandleSoftware",
			time.Unix(1000, 0),
			false,
			true,
			true,
			authorization.AuthenticationMethodsReferences{
				WebAuthn:             true,
				WebAuthnSoftware:     true,
				WebAuthnUserPresence: true,
				WebAuthnUserVerified: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := &UserSession{}

			actual.SetTwoFactorWebAuthn(tc.at, tc.hardware, tc.presence, tc.verified)

			assert.Equal(t, tc.expected, actual.AuthenticationMethodRefs)
		})
	}
}

func TestUserSession_Misc(t *testing.T) {
	session := &UserSession{}

	assert.Equal(t, Identity{}, session.Identity())
	assert.Equal(t, "", session.GetUsername())
	assert.Equal(t, "", session.GetDisplayName())
	assert.Nil(t, session.GetGroups())
	assert.Nil(t, session.GetEmails())
	assert.True(t, session.IsAnonymous())

	session.Username = "abc"

	assert.Equal(t, Identity{Username: "abc"}, session.Identity())
	assert.Equal(t, "abc", session.GetUsername())

	session.DisplayName = "A B C"

	assert.Equal(t, Identity{Username: "abc", DisplayName: "A B C"}, session.Identity())
	assert.Equal(t, "abc", session.GetUsername())
	assert.Equal(t, "A B C", session.GetDisplayName())

	session.Emails = []string{"abc@example.com", "xyz@example.com"}

	assert.Equal(t, Identity{Username: "abc", DisplayName: "A B C", Email: "abc@example.com"}, session.Identity())
	assert.Equal(t, "abc", session.GetUsername())
	assert.Equal(t, "A B C", session.GetDisplayName())
	assert.Equal(t, []string{"abc@example.com", "xyz@example.com"}, session.GetEmails())

	session.Groups = []string{"agroup", "bgroup"}
	assert.Equal(t, "abc", session.GetUsername())
	assert.Equal(t, "A B C", session.GetDisplayName())
	assert.Equal(t, []string{"abc@example.com", "xyz@example.com"}, session.GetEmails())
	assert.Equal(t, []string{"agroup", "bgroup"}, session.GetGroups())
}
