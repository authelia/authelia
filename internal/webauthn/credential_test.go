package webauthn

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestVerifyCredential(t *testing.T) {
	testCases := []struct {
		name         string
		config       *schema.WebAuthn
		credential   *model.WebAuthnCredential
		expectResult VerifyCredentialResult
	}{
		{
			name:         "ShouldVerifyMissingStatement",
			config:       &schema.WebAuthn{},
			credential:   &model.WebAuthnCredential{},
			expectResult: VerifyCredentialResult{MissingStatement: true},
		},
		{
			name:   "ShouldVerifyMalformedStatement",
			config: &schema.WebAuthn{},
			credential: &model.WebAuthnCredential{
				Attestation: []byte("abc"),
			},
			expectResult: VerifyCredentialResult{Malformed: true},
		},
		{
			name: "ShouldVerifyProhibitedAAGUID",
			config: &schema.WebAuthn{
				Filtering: schema.WebAuthnFiltering{
					ProhibitedAAGUIDs: []uuid.UUID{
						uuid.MustParse("e87c6826-9e40-4a69-a68a-523d45a10941"),
					},
				},
			},
			credential: &model.WebAuthnCredential{
				AAGUID: uuid.NullUUID{UUID: uuid.MustParse("e87c6826-9e40-4a69-a68a-523d45a10941"), Valid: true},
			},
			expectResult: VerifyCredentialResult{MissingStatement: true, IsProhibitedAAGUID: true},
		},
		{
			name: "ShouldVerifyNotPermittedAAGUID",
			config: &schema.WebAuthn{
				Filtering: schema.WebAuthnFiltering{
					PermittedAAGUIDs: []uuid.UUID{
						uuid.MustParse("e87c6826-9e40-4a69-a68a-523d45a10942"),
					},
				},
			},
			credential: &model.WebAuthnCredential{
				AAGUID: uuid.NullUUID{UUID: uuid.MustParse("e87c6826-9e40-4a69-a68a-523d45a10941"), Valid: true},
			},
			expectResult: VerifyCredentialResult{MissingStatement: true, IsProhibitedAAGUID: true},
		},
		{
			name: "ShouldVerifyBackupEligible",
			config: &schema.WebAuthn{
				Filtering: schema.WebAuthnFiltering{
					ProhibitBackupEligibility: true,
					PermittedAAGUIDs: []uuid.UUID{
						uuid.MustParse("e87c6826-9e40-4a69-a68a-523d45a10941"),
					},
				},
			},
			credential: &model.WebAuthnCredential{
				AAGUID:         uuid.NullUUID{UUID: uuid.MustParse("e87c6826-9e40-4a69-a68a-523d45a10941"), Valid: true},
				BackupEligible: true,
			},
			expectResult: VerifyCredentialResult{MissingStatement: true, IsProhibitedBackupEligibility: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectResult, VerifyCredential(tc.config, tc.credential, nil))
		})
	}
}
