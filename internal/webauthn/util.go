package webauthn

import (
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/sirupsen/logrus"
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
