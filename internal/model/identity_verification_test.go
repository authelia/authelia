package model

import (
	"net"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIdentityVerification(t *testing.T) {
	testCases := []struct {
		name          string
		jti           uuid.UUID
		username      string
		action        string
		ip            net.IP
		expiration    time.Duration
		expected      IdentityVerification
		expectedClaim IdentityVerificationClaim
	}{
		{
			"ShouldHandleNormal",
			uuid.MustParse("bdc765ef-e1a4-4bf7-a1ef-89102cde635c"),
			"example",
			"an_action",
			net.ParseIP("127.0.0.1"),
			time.Hour,
			IdentityVerification{
				JTI:      uuid.MustParse("bdc765ef-e1a4-4bf7-a1ef-89102cde635c"),
				Action:   "an_action",
				Username: "example",
				IssuedIP: IP{IP: net.ParseIP("127.0.0.1")},
			},
			IdentityVerificationClaim{
				RegisteredClaims: jwt.RegisteredClaims{
					ID:     "bdc765ef-e1a4-4bf7-a1ef-89102cde635c",
					Issuer: "Authelia",
				},
				Action:   "an_action",
				Username: "example",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NewIdentityVerification(tc.jti, tc.username, tc.action, tc.ip, tc.expiration)
			assert.Equal(t, tc.expected.Username, result.Username)
			assert.Equal(t, tc.expected.Action, result.Action)
			assert.Equal(t, tc.expected.IssuedIP, result.IssuedIP)
			assert.Equal(t, tc.expected.JTI, result.JTI)

			assert.WithinDuration(t, time.Now().Add(tc.expiration), result.ExpiresAt, time.Second)

			claim := result.ToIdentityVerificationClaim()

			assert.Equal(t, tc.expectedClaim.Username, claim.Username)
			assert.Equal(t, tc.expectedClaim.Action, claim.Action)
			assert.Equal(t, tc.expectedClaim.Issuer, claim.Issuer)
			assert.Equal(t, tc.expectedClaim.ID, claim.ID)

			reverse, err := claim.ToIdentityVerification()
			require.NoError(t, err)

			assert.Equal(t, result.JTI, reverse.JTI)
			assert.Equal(t, result.Username, reverse.Username)
			assert.Equal(t, result.Action, reverse.Action)
			assert.WithinDuration(t, result.ExpiresAt, reverse.ExpiresAt, time.Second)
		})
	}
}

func TestIdentityVerificationClaim_ToIdentityVerification(t *testing.T) {
	have := IdentityVerificationClaim{
		RegisteredClaims: jwt.RegisteredClaims{
			ID: "example",
		},
	}

	result, err := have.ToIdentityVerification()
	assert.EqualError(t, err, "invalid UUID length: 7")
	assert.Nil(t, result)
}
