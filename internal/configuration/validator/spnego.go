package validator

import (
	"fmt"
	"strings"

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

	if config.AuthenticationBackend.LDAP.PrincipalsFilter == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "principals_filter"))
	} else {

		if !strings.HasPrefix(config.AuthenticationBackend.LDAP.PrincipalsFilter, "(") || !strings.HasSuffix(config.AuthenticationBackend.LDAP.PrincipalsFilter, ")") {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterEnclosingParenthesis, "principals_filter", config.AuthenticationBackend.LDAP.PrincipalsFilter, config.AuthenticationBackend.LDAP.PrincipalsFilter))
		}

		if !strings.Contains(config.AuthenticationBackend.LDAP.PrincipalsFilter, "{principal_attribute}") {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterMissingPlaceholder, "principals_filter", "principal_attribute"))
		}

		// This test helps the user know that principals_filter is broken after the breaking change induced by this commit.
		if !strings.Contains(config.AuthenticationBackend.LDAP.PrincipalsFilter, "{input}") {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterMissingPlaceholder, "principals_filter", "input"))
		}
	}
}
