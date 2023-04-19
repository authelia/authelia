package validator

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Test constants.
const (
	testInvalid       = "invalid"
	testJWTSecret     = "a_secret"
	testLDAPBaseDN    = "base_dn"
	testLDAPPassword  = "password"
	testLDAPURL       = "ldap://ldap"
	testLDAPUser      = "user"
	testEncryptionKey = "a_not_so_secure_encryption_key"
)

const (
	exampleDotCom = "example.com"
	rs256         = "rs256"
)

const (
	local25 = "127.0.0.25"
)

var (
	testLDAPAddress = MustParseAddressPtr(testLDAPURL)
)

func MustParseAddressPtr(input string) *schema.Address {
	address, err := schema.NewAddress(input)
	if err != nil {
		panic(err)
	}

	return address
}

func MustParseAddress(input string) schema.Address {
	address, err := schema.NewAddress(input)
	if err != nil {
		panic(err)
	}

	return *address
}
