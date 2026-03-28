package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

const (
	testLDAPURL  = "ldap://127.0.0.1:389"
	testLDAPSURL = "ldaps://127.0.0.1:389"
)

var (
	testLDAPAddress  = MustParseAddress(testLDAPURL)
	testLDAPSAddress = MustParseAddress(testLDAPSURL)
)

func MustParseAddress(input string) *schema.AddressLDAP {
	address, err := schema.NewAddress(input)
	if err != nil {
		panic(err)
	}

	return &schema.AddressLDAP{Address: *address}
}

func TestErrReload(t *testing.T) {
	have := &errReload{err: ErrWatcherCooldown}

	assert.Equal(t, have.Unwrap(), ErrWatcherCooldown)
	assert.Equal(t, have.Error(), "watcher on cooldown")
	assert.Equal(t, have.WatcherReloadErrorCritical(), false)

	have.critical = true

	assert.Equal(t, have.WatcherReloadErrorCritical(), true)
}
