// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// DynamicReferences holds the data from outside the dynamic configuration which items inside it are allowed to
// reference. It exists so the same validation can be performed at startup and on every reload, where the references
// are rebuilt from the current state of the static configuration.
type DynamicReferences struct {
	// OAuth2ClientIDs is the set of registered OpenID Connect 1.0 client identifiers.
	OAuth2ClientIDs map[string]struct{}
}

// NewDynamicReferences creates the DynamicReferences from a static configuration.
func NewDynamicReferences(config *schema.Configuration) (refs DynamicReferences) {
	refs.OAuth2ClientIDs = map[string]struct{}{}

	if config == nil || config.IdentityProviders.OIDC == nil {
		return refs
	}

	for _, client := range config.IdentityProviders.OIDC.Clients {
		refs.OAuth2ClientIDs[client.ID] = struct{}{}
	}

	return refs
}

// IsOAuth2Client returns true if the given client id is a registered OAuth 2.0 client.
func (r DynamicReferences) IsOAuth2Client(id string) (ok bool) {
	_, ok = r.OAuth2ClientIDs[id]

	return ok
}

// ValidateDynamicConfiguration validates the dynamic configuration. It is safe to call on a configuration which will
// not be used if validation fails; like the static validators it may normalize values in place on success.
func ValidateDynamicConfiguration(config *schema.DynamicConfiguration, refs DynamicReferences, validator *schema.StructValidator) {
	validateAccessControlDefaultPolicy(&config.AccessControl, validator)
	validateAccessControlRules(&config.AccessControl, refs.IsOAuth2Client, validator)
}

// ValidateDynamicKeys validates that the keys loaded from the dynamic configuration files are valid, and are for
// sections which actually support being dynamic. Keys for other sections are rejected rather than silently ignored,
// as otherwise editing them would appear to work without doing anything.
func ValidateDynamicKeys(keys, remapped []string, prefix string, validator *schema.StructValidator) {
	ValidateKeys(keys, remapped, prefix, validator)

	for _, key := range keys {
		if isDynamicKey(key) {
			continue
		}

		validator.Push(fmt.Errorf(errFmtDynamicConfigurationUnsupportedKey, key))
	}
}

func isDynamicKey(key string) bool {
	for _, section := range dynamicSections {
		if key == section || strings.HasPrefix(key, section+".") || strings.HasPrefix(key, section+"[") {
			return true
		}
	}

	return false
}

var dynamicSections = []string{"access_control", "definitions.network"}

// ValidateDynamicConfigurationSource validates the static configuration which describes the dynamic configuration. When
// it is in use the sections which it provides must not also be present in the static configuration, as it would be
// ambiguous which one applies.
func ValidateDynamicConfigurationSource(config *schema.Configuration, validator *schema.StructValidator) {
	for i, path := range config.DynamicConfiguration.Paths {
		if path == "" {
			validator.Push(fmt.Errorf(errFmtDynamicConfigurationPathEmpty, i+1))
		}
	}

	ac := config.AccessControl

	if ac.DefaultPolicy != "" || len(ac.Networks) != 0 || len(ac.Rules) != 0 {
		validator.Push(fmt.Errorf(errFmtDynamicConfigurationStaticConflict, "access_control"))
	}
}
