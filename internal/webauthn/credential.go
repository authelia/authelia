package webauthn

import (
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

func VerifyCredential(config *schema.WebAuthn, credential *model.WebAuthnCredential, mds MetaDataProvider) (result VerifyCredentialResult) {
	var (
		c   *webauthn.Credential
		err error
	)
	if c, err = credential.ToCredential(); err != nil {
		result.Malformed = true
	}

	if len(credential.Attestation) == 0 {
		result.MissingStatement = true
	} else if c != nil && mds != nil {
		if err = c.Verify(mds); err != nil {
			result.MetaDataValidationError = true
		}
	}

	if config.Filtering.ProhibitBackupEligibility && credential.BackupEligible {
		result.IsProhibitedBackupEligibility = true
	}

	if len(config.Filtering.PermittedAAGUIDs) != 0 {
		found := false

		for _, aaguid := range config.Filtering.PermittedAAGUIDs {
			if credential.AAGUID.UUID == aaguid {
				found = true

				break
			}
		}

		if !found {
			result.IsProhibitedAAGUID = true
		}
	}

	if !result.IsProhibitedAAGUID {
		for _, aaguid := range config.Filtering.ProhibitedAAGUIDs {
			if credential.AAGUID.UUID == aaguid {
				result.IsProhibitedAAGUID = true

				break
			}
		}
	}

	return result
}

type VerifyCredentialResult struct {
	Malformed                     bool
	MissingStatement              bool
	IsProhibitedBackupEligibility bool
	IsProhibitedAAGUID            bool
	MetaDataValidationError       bool
}
