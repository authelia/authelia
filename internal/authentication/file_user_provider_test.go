package authentication

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/internal/configuration/schema"
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
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		provider := NewFileUserProvider(&config)
		ok, err := provider.CheckUserPassword("john", "password")

		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestShouldCheckUserSHA512PasswordIsCorrect(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		provider := NewFileUserProvider(&config)
		ok, err := provider.CheckUserPassword("harry", "password")

		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestShouldCheckUserPasswordIsWrong(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		provider := NewFileUserProvider(&config)
		ok, err := provider.CheckUserPassword("john", "wrong_password")

		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestShouldCheckUserPasswordOfUnexistingUser(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		provider := NewFileUserProvider(&config)
		_, err := provider.CheckUserPassword("fake", "password")
		assert.Error(t, err)
		assert.Equal(t, "User 'fake' does not exist in database", err.Error())
	})
}

func TestShouldRetrieveUserDetails(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		provider := NewFileUserProvider(&config)
		details, err := provider.GetDetails("john")
		assert.NoError(t, err)
		assert.Equal(t, details.Username, "john")
		assert.Equal(t, details.Emails, []string{"john.doe@authelia.com"})
		assert.Equal(t, details.Groups, []string{"admins", "dev"})
	})
}

func TestShouldUpdatePassword(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		provider := NewFileUserProvider(&config)
		err := provider.UpdatePassword("harry", "newpassword")
		assert.NoError(t, err)

		// Reset the provider to force a read from disk.
		provider = NewFileUserProvider(&config)
		ok, err := provider.CheckUserPassword("harry", "newpassword")
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

// Checks both that the hashing algo changes and that it removes {CRYPT} from the start.
func TestShouldUpdatePasswordHashingAlgorithmToArgon2id(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		provider := NewFileUserProvider(&config)
		assert.True(t, strings.HasPrefix(provider.database.Users["harry"].HashedPassword, "{CRYPT}$6$"))
		err := provider.UpdatePassword("harry", "newpassword")
		assert.NoError(t, err)

		// Reset the provider to force a read from disk.
		provider = NewFileUserProvider(&config)
		ok, err := provider.CheckUserPassword("harry", "newpassword")
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.True(t, strings.HasPrefix(provider.database.Users["harry"].HashedPassword, "$argon2id$"))
	})
}

func TestShouldUpdatePasswordHashingAlgorithmToSHA512(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		config.PasswordHashing.Algorithm = "sha512"
		config.PasswordHashing.Iterations = 50000

		provider := NewFileUserProvider(&config)
		assert.True(t, strings.HasPrefix(provider.database.Users["john"].HashedPassword, "{CRYPT}$argon2id$"))
		err := provider.UpdatePassword("john", "newpassword")
		assert.NoError(t, err)

		// Reset the provider to force a read from disk.
		provider = NewFileUserProvider(&config)
		ok, err := provider.CheckUserPassword("john", "newpassword")
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.True(t, strings.HasPrefix(provider.database.Users["john"].HashedPassword, "$6$"))
	})
}

func TestShouldRaiseWhenLoadingMalformedDatabaseForFirstTime(t *testing.T) {
	WithDatabase(MalformedUserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		assert.PanicsWithValue(t, "Unable to parse database: yaml: line 4: mapping values are not allowed in this context", func() {
			NewFileUserProvider(&config)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadSchemaForFirstTime(t *testing.T) {
	WithDatabase(BadSchemaUserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		assert.PanicsWithValue(t, "Invalid schema of database: Users: non zero value required", func() {
			NewFileUserProvider(&config)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadSHA512HashesForTheFirstTime(t *testing.T) {
	WithDatabase(BadSHA512HashContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		assert.PanicsWithValue(t, "Unable to parse hash of user john: Hash key is not the last parameter, the hash is likely malformed ($6$rounds00000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/).", func() {
			NewFileUserProvider(&config)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashSettingsForTheFirstTime(t *testing.T) {
	WithDatabase(BadArgon2idHashSettingsContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		assert.PanicsWithValue(t, "Unable to parse hash of user john: Hash key is not the last parameter, the hash is likely malformed ($argon2id$v=19$m65536,t3,p2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM).", func() {
			NewFileUserProvider(&config)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashKeyForTheFirstTime(t *testing.T) {
	WithDatabase(BadArgon2idHashKeyContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		assert.PanicsWithValue(t, "Unable to parse hash of user john: Hash key contains invalid base64 characters.", func() {
			NewFileUserProvider(&config)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashSaltForTheFirstTime(t *testing.T) {
	WithDatabase(BadArgon2idHashSaltContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		assert.PanicsWithValue(t, "Unable to parse hash of user john: Salt contains invalid base64 characters.", func() {
			NewFileUserProvider(&config)
		})
	})
}

func TestShouldSupportHashPasswordWithoutCRYPT(t *testing.T) {
	WithDatabase(UserDatabaseWithoutCryptContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		provider := NewFileUserProvider(&config)
		ok, err := provider.CheckUserPassword("john", "password")

		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

var (
	DefaultFileAuthenticationBackendConfiguration = schema.FileAuthenticationBackendConfiguration{
		Path: "",
		PasswordHashing: &schema.PasswordHashingConfiguration{
			Iterations:  schema.DefaultCIPasswordOptionsConfiguration.Iterations,
			KeyLength:   schema.DefaultCIPasswordOptionsConfiguration.KeyLength,
			SaltLength:  schema.DefaultCIPasswordOptionsConfiguration.SaltLength,
			Algorithm:   schema.DefaultCIPasswordOptionsConfiguration.Algorithm,
			Memory:      schema.DefaultCIPasswordOptionsConfiguration.Memory,
			Parallelism: schema.DefaultCIPasswordOptionsConfiguration.Parallelism,
		},
	}
)

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

var BadArgon2idHashSettingsContent = []byte(`
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

var BadArgon2idHashKeyContent = []byte(`
users:
  john:
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$^^vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
`)
var BadArgon2idHashSaltContent = []byte(`
users:
  john:
    password: "$argon2id$v=19$m=65536,t=3,p=2$^^LnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
`)
