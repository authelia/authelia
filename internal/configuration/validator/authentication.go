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

	if configuration.PasswordHashing == nil {
		configuration.PasswordHashing = &schema.DefaultPasswordOptionsConfiguration
	} else {
		if configuration.PasswordHashing.Algorithm == "" {
			configuration.PasswordHashing.Algorithm = schema.DefaultPasswordOptionsConfiguration.Algorithm
		} else {
			configuration.PasswordHashing.Algorithm = strings.ToLower(configuration.PasswordHashing.Algorithm)
			if configuration.PasswordHashing.Algorithm != "argon2id" && configuration.PasswordHashing.Algorithm != "sha512" {
				validator.Push(fmt.Errorf("Unknown hashing algorithm supplied, valid values are argon2id and sha512, you configured '%s'", configuration.PasswordHashing.Algorithm))
			}
		}

		// Iterations (time)
		if configuration.PasswordHashing.Iterations == 0 {
			if configuration.PasswordHashing.Algorithm == "argon2id" {
				configuration.PasswordHashing.Iterations = schema.DefaultPasswordOptionsConfiguration.Iterations
			} else {
				configuration.PasswordHashing.Iterations = schema.DefaultPasswordOptionsSHA512Configuration.Iterations
			}
		} else if configuration.PasswordHashing.Iterations < 1 {
			validator.Push(fmt.Errorf("The number of iterations specified is invalid, must be 1 or more, you configured %d", configuration.PasswordHashing.Iterations))
		}

		//Salt Length
		if configuration.PasswordHashing.SaltLength == 0 {
			configuration.PasswordHashing.SaltLength = schema.DefaultPasswordOptionsConfiguration.SaltLength
		} else if configuration.PasswordHashing.SaltLength < 2 {
			validator.Push(fmt.Errorf("The salt length must be 2 or more, you configured %d", configuration.PasswordHashing.SaltLength))
		} else if configuration.PasswordHashing.SaltLength > 16 {
			validator.Push(fmt.Errorf("The salt length must be 16 or less, you configured %d", configuration.PasswordHashing.SaltLength))
		}

		if configuration.PasswordHashing.Algorithm == "argon2id" {

			// Parallelism
			if configuration.PasswordHashing.Parallelism == 0 {
				configuration.PasswordHashing.Parallelism = schema.DefaultPasswordOptionsConfiguration.Parallelism
			} else if configuration.PasswordHashing.Parallelism < 1 {
				validator.Push(fmt.Errorf("Parallelism for argon2id must be 1 or more, you configured %d", configuration.PasswordHashing.Parallelism))
			}

			// Memory
			if configuration.PasswordHashing.Memory == 0 {
				configuration.PasswordHashing.Memory = schema.DefaultPasswordOptionsConfiguration.Memory
			} else if configuration.PasswordHashing.Memory < configuration.PasswordHashing.Parallelism*8 {
				validator.Push(fmt.Errorf("Memory for argon2id must be %d or more (parallelism * 8), you configured memory as %d and parallelism as %d", configuration.PasswordHashing.Parallelism*8, configuration.PasswordHashing.Memory, configuration.PasswordHashing.Parallelism))
			}

			// Key Length
			if configuration.PasswordHashing.KeyLength == 0 {
				configuration.PasswordHashing.KeyLength = schema.DefaultPasswordOptionsConfiguration.KeyLength
			} else if configuration.PasswordHashing.KeyLength < 16 {
				validator.Push(fmt.Errorf("Key length for argon2id must be 16, you configured %d", configuration.PasswordHashing.KeyLength))
			}
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

	// TODO: see if it's possible to disable this check if disable_reset_password is set and when anonymous/user binding is supported (#101 and #387)
	if configuration.User == "" {
		validator.Push(errors.New("Please provide a user name to connect to the LDAP server"))
	}

	// TODO: see if it's possible to disable this check if disable_reset_password is set and when anonymous/user binding is supported (#101 and #387)
	if configuration.Password == "" {
		validator.Push(errors.New("Please provide a password to connect to the LDAP server"))
	}

	if configuration.BaseDN == "" {
		validator.Push(errors.New("Please provide a base DN to connect to the LDAP server"))
	}

	if configuration.UsersFilter == "" {
		validator.Push(errors.New("Please provide a users filter with `users_filter` attribute"))
	} else {
		if !strings.HasPrefix(configuration.UsersFilter, "(") || !strings.HasSuffix(configuration.UsersFilter, ")") {
			validator.Push(errors.New("The users filter should contain enclosing parenthesis. For instance uid={input} should be (uid={input})"))
		}

		// This test helps the user know that users_filter is broken after the breaking change induced by this commit.
		if !strings.Contains(configuration.UsersFilter, "{0}") && !strings.Contains(configuration.UsersFilter, "{input}") {
			validator.Push(errors.New("Unable to detect {input} placeholder in users_filter, your configuration might be broken. " +
				"Please review configuration options listed at https://docs.authelia.com/configuration/authentication/ldap.html"))
		}
	}

	if configuration.GroupsFilter == "" {
		validator.Push(errors.New("Please provide a groups filter with `groups_filter` attribute"))
	} else {
		if !strings.HasPrefix(configuration.GroupsFilter, "(") || !strings.HasSuffix(configuration.GroupsFilter, ")") {
			validator.Push(errors.New("The groups filter should contain enclosing parenthesis. For instance cn={input} should be (cn={input})"))
		}
	}

	if configuration.UsernameAttribute == "" {
		validator.Push(errors.New("Please provide a username attribute with `username_attribute`"))
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
