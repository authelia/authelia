package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func ValidateSpnego(config *schema.Configuration, validator *schema.StructValidator) {
	if config.SPNEGO.Disable == true {
		return
	}
	if config.AuthenticationBackend.LDAP == nil {
		validator.Push(fmt.Errorf("SPNEGO Kerberos authentication is enabled but no LDAP authentication backend is enabled"))
	}

	if config.SPNEGO.Principal == "" {
		validator.Push(fmt.Errorf("SPNEGO Kerberos authentication is enabled but no service principal is configured"))
	}

	if config.SPNEGO.Keytab == "" {
		validator.Push(fmt.Errorf("SPNEGO Kerberos authentication is enabled but no keytab file path is configured"))
	}

	if config.SPNEGO.Realm == "" {
		validator.Push(fmt.Errorf("SPNEGO Kerberos authentication is enabled but no Kerberos realm is configured"))
	}
}
