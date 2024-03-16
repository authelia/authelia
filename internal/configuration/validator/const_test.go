package validator

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Test constants.
const (
	id = "id"

	testInvalid       = "invalid"
	testJWTSecret     = "a_secret"
	testLDAPBaseDN    = "base_dn"
	testLDAPPassword  = "password"
	testLDAPURL       = "ldap://ldap"
	testLDAPUser      = "user"
	testEncryptionKey = "a_not_so_secure_encryption_key"

	member            = "member"
	memberof          = "memberof"
	memberOf          = "memberOf"
	filterMemberOfRDN = "(|({memberof:rdn}))"
)

const (
	authdot       = "auth."
	exampleDotCom = "example.com"
	rs256         = "rs256"
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
