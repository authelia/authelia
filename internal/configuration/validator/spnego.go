package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func ValidateSpnego(config *schema.Configuration, validator *schema.StructValidator) {
	if !config.SPNEGO.Enabled {
		return
	}

	if len(config.SPNEGO.Keytab) == 0 {
		validator.Push(fmt.Errorf("spnego: option 'keytab' must be provided"))
	}

	if config.SPNEGO.Principal == "" {
		validator.Push(fmt.Errorf("spnego: option 'principal' must be provided"))
	}

	if config.SPNEGO.Realm == "" {
		validator.Push(fmt.Errorf("spnego: option 'realm' must be provided"))
	}

	if config.AuthenticationBackend.LDAP == nil {
		validator.Push(fmt.Errorf("spnego: kerberos authentication requires LDAP backend to be configured"))
	}
}
