package authentication

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func WithDatabase(content []byte, f func(path string)) {
	tmpfile, err := ioutil.TempFile("", "users_database.*.yaml")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(content); err != nil {
		tmpfile.Close()
		log.Fatal(err)
	}

	f(tmpfile.Name())

	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}
}

func TestShouldCheckUserPasswordIsCorrect(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		provider := NewFileUserProvider(path)
		ok, err := provider.CheckUserPassword("john", "password")

		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestShouldCheckUserPasswordIsWrong(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		provider := NewFileUserProvider(path)
		ok, err := provider.CheckUserPassword("john", "wrong_password")

		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestShouldCheckUserPasswordOfUnexistingUser(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		provider := NewFileUserProvider(path)
		_, err := provider.CheckUserPassword("fake", "password")
		assert.Error(t, err)
		assert.Equal(t, "User 'fake' does not exist in database", err.Error())
	})
}

func TestShouldRetrieveUserDetails(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		provider := NewFileUserProvider(path)
		details, err := provider.GetDetails("john")
		assert.NoError(t, err)
		assert.Equal(t, details.Emails, []string{"john.doe@authelia.com"})
		assert.Equal(t, details.Groups, []string{"admins", "dev"})
	})
}

func TestShouldUpdatePassword(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		provider := NewFileUserProvider(path)
		err := provider.UpdatePassword("john", "newpassword")
		assert.NoError(t, err)

		// Reset the provider to force a read from disk.
		provider = NewFileUserProvider(path)
		ok, err := provider.CheckUserPassword("john", "newpassword")
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestShouldRaiseWhenLoadingMalformedDatabaseForFirstTime(t *testing.T) {
	WithDatabase(MalformedUserDatabaseContent, func(path string) {
		assert.Panics(t, func() {
			NewFileUserProvider(path)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadSchemaForFirstTime(t *testing.T) {
	WithDatabase(BadSchemaUserDatabaseContent, func(path string) {
		assert.Panics(t, func() {
			NewFileUserProvider(path)
		})
	})
}

var UserDatabaseContent = []byte(`
users:
  john:
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev

  harry:
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: harry.potter@authelia.com
    groups: []

  bob:
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: bob.dylan@authelia.com
    groups:
      - dev

  james:
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: james.dean@authelia.com
`)

var MalformedUserDatabaseContent = []byte(`
users
john
email: john.doe@authelia.com
groups:
- admin
- dev
`)

// The YAML is valid but the root key is user instead of users
var BadSchemaUserDatabaseContent = []byte(`
user:
  john:
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
`)
