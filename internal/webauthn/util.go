package webauthn

import (
	"errors"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

// IsCredentialCreationDiscoverable returns true if the *protocol.ParsedCredentialCreationData indicates a discoverable
// credential was generated.
func IsCredentialCreationDiscoverable(logger *logrus.Entry, response *protocol.ParsedCredentialCreationData) (discoverable bool) {
	if value, ok := response.ClientExtensionResults[ExtensionCredProps]; ok {
		switch credentialProperties := value.(type) {
		case map[string]any:
			var v any

			if v, ok = credentialProperties[ExtensionCredPropsResidentKey]; ok {
				if discoverable, ok = v.(bool); ok {
					logger.WithFields(map[string]any{LogFieldDiscoverable: discoverable}).Trace("Determined Credential Discoverability via Client Extension Results")

					return discoverable
				} else {
					logger.WithFields(map[string]any{LogFieldDiscoverable: false}).Trace("Assuming Credential Discoverability is false as the 'rk' field for the 'credProps' extension in the Client Extension Results was not a boolean")
				}
			} else {
				logger.WithFields(map[string]any{LogFieldDiscoverable: false}).Trace("Assuming Credential Discoverability is false as the 'rk' field for the 'credProps' extension was missing from the Client Extension Results")
			}

			return false
		default:
			logger.WithFields(map[string]any{LogFieldDiscoverable: false}).Trace("Assuming Credential Discoverability is false as the 'credProps' extension in the Client Extension Results does not appear to be a dictionary")

			return false
		}
	}

	logger.WithFields(map[string]any{LogFieldDiscoverable: false}).Trace("Assuming Credential Discoverability is false as the 'credProps' extension is missing from the Client Extension Results")

	return false
}

func ValidateCredentialAllowed(config *schema.WebAuthn, credential *model.WebAuthnCredential) (err error) {
	if config.Filtering.ProhibitBackupEligibility && credential.BackupEligible {
		return fmt.Errorf("error checking webauthn credential: filters have been configured which prohibit credentials that are backup eligible")
	}

	if len(config.Filtering.PermittedAAGUIDs) != 0 {
		for _, aaguid := range config.Filtering.PermittedAAGUIDs {
			if credential.AAGUID.UUID == aaguid {
				return nil
			}
		}

		return fmt.Errorf("error checking webauthn credential: filters have been configured which explicitly require only permitted AAGUID's be used and '%s' is not permitted", credential.AAGUID.UUID)
	}

	for _, aaguid := range config.Filtering.ProhibitedAAGUIDs {
		if credential.AAGUID.UUID == aaguid {
			return fmt.Errorf("error checking webauthn credential: filters have been configured which prohibit the AAGUID '%s' from registration", aaguid)
		}
	}

	return nil
}

func FormatError(err error) error {
	out := &protocol.Error{}

	if errors.As(err, &out) {
		if len(out.DevInfo) == 0 {
			return err
		}

		if len(out.Type) == 0 {
			return fmt.Errorf("%w: %s", err, out.DevInfo)
		}

		return fmt.Errorf("%w (%s): %s", err, out.Type, out.DevInfo)
	}

	return err
}
