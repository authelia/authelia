// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webauthn

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

// IsCredentialCreationDiscoverable returns true if the *protocol.ParsedCredentialCreationData indicates a discoverable
// credential was generated.
func IsCredentialCreationDiscoverable(logger *logrus.Entry, response *protocol.ParsedCredentialCreationData) (discoverable bool) {
	credentialProperties := response.ClientExtensionResults.CredProps

	if credentialProperties == nil {
		logger.WithFields(map[string]any{LogFieldDiscoverable: false}).Trace("Assuming Credential Discoverability is false as the 'credProps' extension is missing from the Client Extension Results")

		return false
	}

	if credentialProperties.RK == nil {
		logger.WithFields(map[string]any{LogFieldDiscoverable: false}).Trace("Assuming Credential Discoverability is false as the 'rk' field for the 'credProps' extension was missing from the Client Extension Results")

		return false
	}

	discoverable = *credentialProperties.RK

	logger.WithFields(map[string]any{LogFieldDiscoverable: discoverable}).Trace("Determined Credential Discoverability via Client Extension Results")

	return discoverable
}

// ValidateCredentialAllowed returns an error if the given credential is prohibited by the filters configured for the
// relying party the ceremony was performed against.
func ValidateCredentialAllowed(config *schema.WebAuthnBase, credential *model.WebAuthnCredential) (err error) {
	if config.Filtering.IsProhibitBackupEligibility() && credential.BackupEligible {
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

// FormatError returns the given error with the WebAuthn specific details included.
func FormatError(err error) error {
	out := &protocol.Error{}
	if errors.As(err, &out) {
		if len(out.DevInfo) == 0 {
			if len(out.Type) == 0 {
				return err
			}

			return fmt.Errorf("%w (%s)", err, out.Type)
		}

		if len(out.Type) == 0 {
			return fmt.Errorf("%w: %s", err, out.DevInfo)
		}

		return fmt.Errorf("%w (%s): %s", err, out.Type, out.DevInfo)
	}

	return err
}

// GetRelatedOriginConfigByRPID returns a *schema.WebAuthnRelyingParty provided it can match it to a rpid.
func GetRelatedOriginConfigByRPID(config schema.WebAuthn, rpid string) (ro *schema.WebAuthnRelyingParty) {
	if value, ok := config.RelyingParties[strings.ToLower(rpid)]; ok {
		return &value
	}

	return nil
}

// GetRelatedOriginConfigByOrigin returns a *schema.WebAuthnRelyingParty provided it can match it to an origin string.
func GetRelatedOriginConfigByOrigin(config schema.WebAuthn, origin *url.URL) (relyingPartyID string, ro *schema.WebAuthnRelyingParty) {
	if origin == nil {
		return "", nil
	}

	for rpid, r := range config.RelyingParties {
		ro = &r

		for _, o := range ro.Origins {
			if o == nil || !IsOriginEqual(o, origin) {
				continue
			}

			if o.Path != "" || origin.Path != "" {
				continue
			}

			return rpid, ro
		}
	}

	return "", nil
}

// IsOriginEqual returns true if the scheme, hostname, and effective port of both URLs are equal. The scheme and
// hostname are compared case-insensitively and a port which is the default for the scheme is treated as equal to no
// port.
func IsOriginEqual(a, b *url.URL) (equal bool) {
	if a == nil || b == nil {
		return false
	}

	if !strings.EqualFold(a.Scheme, b.Scheme) || !strings.EqualFold(a.Hostname(), b.Hostname()) {
		return false
	}

	return originEffectivePort(a) == originEffectivePort(b)
}

// OriginKey returns a canonical representation of the origin of a URL built from the same values IsOriginEqual
// compares, so two URLs have the same key if and only if IsOriginEqual considers them equal. The scheme and hostname
// are lower case and a port which is the default for the scheme is omitted.
func OriginKey(u *url.URL) (key string) {
	if u == nil {
		return ""
	}

	scheme, hostname := strings.ToLower(u.Scheme), strings.ToLower(u.Hostname())

	switch port := u.Port(); {
	case port == "", port == originEffectivePort(&url.URL{Scheme: scheme}):
		return scheme + "://" + hostname
	default:
		return scheme + "://" + net.JoinHostPort(hostname, port)
	}
}

func originEffectivePort(u *url.URL) (port string) {
	if port = u.Port(); port != "" {
		return port
	}

	switch strings.ToLower(u.Scheme) {
	case "https":
		return "443"
	case "http":
		return "80"
	default:
		return ""
	}
}
