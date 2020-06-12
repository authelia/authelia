package validator

import (
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// ValidateOIDC validates and update OIDC configuration.
func ValidateOIDC(configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if configuration.OIDCIssuerPrivateKeyPath == "" {
		validator.Push(fmt.Errorf("Issuer private key path must be provided"))
	}
}
