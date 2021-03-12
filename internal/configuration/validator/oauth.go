package validator

import (
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateOAuth validates and update OAuth configuration.
func ValidateOAuth(configuration *schema.OAuthConfiguration, validator *schema.StructValidator) {
	validateOIDCServer(configuration.OIDCServer, validator)
}

func validateOIDCServer(configuration *schema.OpenIDConnectServerConfiguration, validator *schema.StructValidator) {
	if configuration != nil {
		if configuration.IssuerPrivateKeyPath == "" {
			validator.Push(fmt.Errorf("OIDC Server issuer private key path must be provided"))
		} else {
			exists, err := utils.FileExists(configuration.IssuerPrivateKeyPath)
			if !exists {
				validator.Push(fmt.Errorf("OIDC Server issuer private key path doesn't exist"))
			} else if err != nil {
				validator.Push(fmt.Errorf("OIDC Server issuer private key path is invalid: %v", err))
			}
		}

		if len(configuration.HMACSecret) != 32 {
			validator.Push(fmt.Errorf("OIDC Server HMAC secret must be exactly 32 chars long"))
		}
	}
}
