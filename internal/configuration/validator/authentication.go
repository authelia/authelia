package validator

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
)

var ldapProtocolPrefix = "ldap://"
var ldapsProtocolPrefix = "ldaps://"

func validateFileAuthenticationBackend(configuration *schema.FileAuthenticationBackendConfiguration, validator *schema.StructValidator) {
	if configuration.Path == "" {
		validator.Push(errors.New("Please provide a `path` for the users database in `authentication_backend`"))
	}

	if configuration.Algorithm == "" {
		configuration.Algorithm = schema.DefaultFileAuthenticationBackendConfiguration.Algorithm
	} else {
		configuration.Algorithm = strings.ToLower(configuration.Algorithm)
		if configuration.Algorithm != "argon2id" && configuration.Algorithm != "sha512" {
			validator.Push(fmt.Errorf("Unknown hashing algorithm supplied, valid values are argon2id and sha512, you provided '%s'", configuration.Algorithm))
		}
	}

	if configuration.Rounds == 0 {
		if configuration.Algorithm == "argon2id" {
			configuration.Rounds = schema.DefaultFileAuthenticationBackendConfiguration.Rounds
		} else {
			configuration.Rounds = schema.DefaultFileAuthenticationBackendSHA512Configuration.Rounds
		}
	} else if configuration.Rounds < 0 {
		validator.Push(fmt.Errorf("The number of rounds specified is invalid, must be more than 0 but you specified %d", configuration.Rounds))
	}

	if configuration.SaltLength == 0 {
		configuration.SaltLength = schema.DefaultFileAuthenticationBackendConfiguration.SaltLength
	} else if configuration.SaltLength < 0 {
		validator.Push(fmt.Errorf("The salt length must 1 or more, you set it to %d", configuration.SaltLength))
	} else if configuration.SaltLength > 16 {
		validator.Push(fmt.Errorf("The salt length must be 16 or less, you set it to %d", configuration.SaltLength))
	}

	if configuration.Algorithm == "argon2id" {
		if configuration.Parallelism == 0 {
			configuration.Parallelism = schema.DefaultFileAuthenticationBackendConfiguration.Parallelism
		} else if configuration.Parallelism < 1 {
			validator.Push(fmt.Errorf("Parallelism for argon2id must be 0 or more, you set %d", configuration.Parallelism))
		}
		if configuration.Memory == 0 {
			configuration.Memory = schema.DefaultFileAuthenticationBackendConfiguration.Memory
		} else if configuration.Memory < configuration.Parallelism*8 {
			validator.Push(fmt.Errorf("Memory for argon2id must be %d or more (parallelism * 8), you set memory to %d and parallelism to %d", configuration.Parallelism*8, configuration.Memory, configuration.Parallelism))
		}
	}
}

func validateLdapURL(ldapURL string, validator *schema.StructValidator) string {
	u, err := url.Parse(ldapURL)

	if err != nil {
		validator.Push(errors.New("Unable to parse URL to ldap server. The scheme is probably missing: ldap:// or ldaps://"))
		return ""
	}

	if !(u.Scheme == "ldap" || u.Scheme == "ldaps") {
		validator.Push(errors.New("Unknown scheme for ldap url, should be ldap:// or ldaps://"))
		return ""
	}

	if u.Scheme == "ldap" && u.Port() == "" {
		u.Host += ":389"
	} else if u.Scheme == "ldaps" && u.Port() == "" {
		u.Host += ":636"
	}

	if !u.IsAbs() {
		validator.Push(fmt.Errorf("URL to LDAP %s is still not absolute, it should be something like ldap://127.0.0.1:389", u.String()))
	}

	return u.String()
}

func validateLdapAuthenticationBackend(configuration *schema.LDAPAuthenticationBackendConfiguration, validator *schema.StructValidator) {
	if configuration.URL == "" {
		validator.Push(errors.New("Please provide a URL to the LDAP server"))
	} else {
		configuration.URL = validateLdapURL(configuration.URL, validator)
	}

	if configuration.User == "" {
		validator.Push(errors.New("Please provide a user name to connect to the LDAP server"))
	}

	if configuration.Password == "" {
		validator.Push(errors.New("Please provide a password to connect to the LDAP server"))
	}

	if configuration.BaseDN == "" {
		validator.Push(errors.New("Please provide a base DN to connect to the LDAP server"))
	}

	if configuration.UsersFilter == "" {
		configuration.UsersFilter = "(cn={0})"
	}

	if !strings.HasPrefix(configuration.UsersFilter, "(") || !strings.HasSuffix(configuration.UsersFilter, ")") {
		validator.Push(errors.New("The users filter should contain enclosing parenthesis. For instance cn={0} should be (cn={0})"))
	}

	if configuration.GroupsFilter == "" {
		configuration.GroupsFilter = "(member={dn})"
	}

	if !strings.HasPrefix(configuration.GroupsFilter, "(") || !strings.HasSuffix(configuration.GroupsFilter, ")") {
		validator.Push(errors.New("The groups filter should contain enclosing parenthesis. For instance cn={0} should be (cn={0})"))
	}

	if configuration.GroupNameAttribute == "" {
		configuration.GroupNameAttribute = "cn"
	}

	if configuration.MailAttribute == "" {
		configuration.MailAttribute = "mail"
	}
}

// ValidateAuthenticationBackend validates and update authentication backend configuration.
func ValidateAuthenticationBackend(configuration *schema.AuthenticationBackendConfiguration, validator *schema.StructValidator) {
	if configuration.Ldap == nil && configuration.File == nil {
		validator.Push(errors.New("Please provide `ldap` or `file` object in `authentication_backend`"))
	}

	if configuration.Ldap != nil && configuration.File != nil {
		validator.Push(errors.New("You cannot provide both `ldap` and `file` objects in `authentication_backend`"))
	}

	if configuration.File != nil {
		validateFileAuthenticationBackend(configuration.File, validator)
	} else if configuration.Ldap != nil {
		validateLdapAuthenticationBackend(configuration.Ldap, validator)
	}
}
