package authentication

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
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

func TestShouldCheckUserArgon2idPasswordIsCorrect(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		provider := NewFileUserProvider(path)
		ok, err := provider.CheckUserPassword("john", "password")

		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestShouldCheckUserSHA512PasswordIsCorrect(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		provider := NewFileUserProvider(path)
		ok, err := provider.CheckUserPassword("harry", "password")

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
		assert.True(t, strings.HasPrefix(provider.database.Users["harry"].HashedPassword, "{CRYPT}$6$"))
		err := provider.UpdatePassword("harry", "newpassword")
		assert.NoError(t, err)

		// Reset the provider to force a read from disk.
		provider = NewFileUserProvider(path)
		ok, err := provider.CheckUserPassword("harry", "newpassword")
		assert.NoError(t, err)
		assert.True(t, ok)
		fmt.Println(provider.database.Users["harry"].HashedPassword)
		assert.True(t, strings.HasPrefix(provider.database.Users["harry"].HashedPassword, "{CRYPT}$argon2id$"))
	})
}

func TestShouldRaiseWhenLoadingMalformedDatabaseForFirstTime(t *testing.T) {
	WithDatabase(MalformedUserDatabaseContent, func(path string) {
		assert.PanicsWithValue(t, "Unable to parse database: yaml: line 4: mapping values are not allowed in this context", func() {
			NewFileUserProvider(path)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadSchemaForFirstTime(t *testing.T) {
	WithDatabase(BadSchemaUserDatabaseContent, func(path string) {
		assert.PanicsWithValue(t, "Invalid schema of database: Users: non zero value required", func() {
			NewFileUserProvider(path)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadSHA512HashesForTheFirstTime(t *testing.T) {
	WithDatabase(BadSHA512HashContent, func(path string) {
		assert.PanicsWithValue(t, "Unable to parse hash of user john: Cannot match pattern 'rounds=<int>' to find the number of rounds. Cause: input does not match format", func() {
			NewFileUserProvider(path)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashesForTheFirstTime(t *testing.T) {
	WithDatabase(BADArgon2idHashContent, func(path string) {
		assert.PanicsWithValue(t, "Unable to parse hash of user john: Cannot match pattern 'm=<int>,t=<int>,p=<int>' to find the argon2id params. Cause: input does not match format", func() {
			NewFileUserProvider(path)
		})
	})
}

func TestShouldSupportHashPasswordWithoutCRYPT(t *testing.T) {
	WithDatabase(UserDatabaseWithoutCryptContent, func(path string) {
		provider := NewFileUserProvider(path)
		ok, err := provider.CheckUserPassword("john", "password")

		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

var UserDatabaseContent = []byte(`
users:
  john:
    password: "{CRYPT}$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
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
- admins
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

var UserDatabaseWithoutCryptContent = []byte(`
users:
  john:
    password: "$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
  james:
    password: "$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: james.dean@authelia.com
`)

var BadSHA512HashContent = []byte(`
users:
  john:
    password: "$6$rounds00000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
  james:
    password: "$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: james.dean@authelia.com
`)

var BADArgon2idHashContent = []byte(`
users:
  john:
    password: "$argon2id$v=19$m65536,t3,p2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
  james:
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: james.dean@authelia.com
`)
