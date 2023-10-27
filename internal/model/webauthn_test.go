package model

import (
	"crypto/rand"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWebAuthnCredentialImportExport(t *testing.T) {
	have := WebAuthnCredentialExport{
		WebAuthnCredentials: []WebAuthnCredential{
			{
				ID:              0,
				CreatedAt:       time.Now(),
				LastUsedAt:      sql.NullTime{Time: time.Now(), Valid: true},
				RPID:            "example",
				Username:        "john",
				Description:     "akey",
				KID:             NewBase64(MustRead(20)),
				PublicKey:       MustRead(128),
				AttestationType: "fido-u2f",
				Transport:       "",
				AAGUID:          MustNullUUID(NewRandomNullUUID()),
				SignCount:       20,
				CloneWarning:    false,
			},
			{
				ID:              0,
				CreatedAt:       time.Now(),
				LastUsedAt:      sql.NullTime{Valid: false},
				RPID:            "example2",
				Username:        "john2",
				Description:     "bkey",
				KID:             NewBase64(MustRead(60)),
				PublicKey:       MustRead(64),
				AttestationType: "packed",
				Transport:       "",
				AAGUID:          uuid.NullUUID{Valid: false},
				SignCount:       30,
				CloneWarning:    true,
			},
		},
	}

	out, err := yaml.Marshal(&have)
	require.NoError(t, err)

	imported := WebAuthnCredentialExport{}

	require.NoError(t, yaml.Unmarshal(out, &imported))
	require.Equal(t, len(have.WebAuthnCredentials), len(imported.WebAuthnCredentials))

	for i, actual := range imported.WebAuthnCredentials {
		t.Run(actual.Description, func(t *testing.T) {
			expected := have.WebAuthnCredentials[i]

			assert.Equal(t, expected.KID, actual.KID)
			assert.Equal(t, expected.PublicKey, actual.PublicKey)
			assert.Equal(t, expected.SignCount, actual.SignCount)
			assert.Equal(t, expected.AttestationType, actual.AttestationType)
			assert.Equal(t, expected.RPID, actual.RPID)
			assert.Equal(t, expected.AAGUID.Valid, actual.AAGUID.Valid)
			assert.Equal(t, expected.AAGUID.UUID, actual.AAGUID.UUID)
			assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Second)
			assert.WithinDuration(t, expected.LastUsedAt.Time, actual.LastUsedAt.Time, time.Second)
			assert.Equal(t, expected.LastUsedAt.Valid, actual.LastUsedAt.Valid)
			assert.Equal(t, expected.CloneWarning, actual.CloneWarning)
			assert.Equal(t, expected.Description, actual.Description)
			assert.Equal(t, expected.Username, actual.Username)
		})
	}
}

func MustRead(n int) []byte {
	data := make([]byte, n)

	if _, err := rand.Read(data); err != nil {
		panic(err)
	}

	return data
}
