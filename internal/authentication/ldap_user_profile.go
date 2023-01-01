package authentication

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func (profile *ldapUserProfile) validateRequiredAttrs(config *schema.LDAPAuthenticationBackend, username string) error {
	if profile.NTHash == nil && config.AuthenticationMethod == schema.LDAPAuthenticationMethodNTPassword {
		return fmt.Errorf("user '%s' must have the attribute '%s'",
			username, config.NTPasswordAttribute)
	}

	if profile.Username == "" {
		return fmt.Errorf("user '%s' must have value for attribute '%s'",
			username, config.UsernameAttribute)
	}

	if profile.DN == "" {
		return fmt.Errorf("user '%s' must have a distinguished name but the result returned an empty distinguished name", username)
	}

	return nil
}
