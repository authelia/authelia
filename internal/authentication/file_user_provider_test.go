package authentication

import (
	"log"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func WithDatabase(content []byte, f func(path string)) {
	tmpfile, err := os.CreateTemp("", "users_database.*.yaml")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmpfile.Name()) // Clean up.

	if _, err := tmpfile.Write(content); err != nil {
		tmpfile.Close()
		log.Panic(err)
	}

	f(tmpfile.Name())

	if err := tmpfile.Close(); err != nil {
		log.Panic(err)
	}
}

func TestShouldErrorPermissionsOnLocalFS(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test due to being on windows")
	}

	_ = os.Mkdir("/tmp/noperms/", 0000)
	err := checkDatabase("/tmp/noperms/users_database.yml")

	require.EqualError(t, err, "error checking user authentication database file: stat /tmp/noperms/users_database.yml: permission denied")
}

func TestShouldErrorAndGenerateUserDB(t *testing.T) {
	err := checkDatabase("./nonexistent.yml")
	_ = os.Remove("./nonexistent.yml")

	require.EqualError(t, err, "user authentication database file doesn't exist at path './nonexistent.yml' and has been generated")
}

func TestShouldCheckUserArgon2idPasswordIsCorrect(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

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

		assert.NoError(t, provider.StartupCheck())

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

		assert.NoError(t, provider.StartupCheck())

		ok, err := provider.CheckUserPassword("john", "wrong_password")

		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestShouldCheckUserPasswordIsWrongForEnumerationCompare(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ok, err := provider.CheckUserPassword("enumeration", "wrong_password")
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestShouldCheckUserPasswordOfUserThatDoesNotExist(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ok, err := provider.CheckUserPassword("fake", "password")
		assert.Error(t, err)
		assert.Equal(t, false, ok)
		assert.EqualError(t, err, "user not found")
	})
}

func TestShouldRetrieveUserDetails(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		details, err := provider.GetDetails("john")
		assert.NoError(t, err)
		assert.Equal(t, "john", details.Username)
		assert.Equal(t, []string{"john.doe@authelia.com"}, details.Emails)
		assert.Equal(t, []string{"admins", "dev"}, details.Groups)
	})
}

func TestShouldUpdatePassword(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		err := provider.UpdatePassword("harry", "newpassword")
		assert.NoError(t, err)

		// Reset the provider to force a read from disk.
		provider = NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

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

		assert.NoError(t, provider.StartupCheck())

		assert.True(t, strings.HasPrefix(provider.database.Users["harry"].Digest.Encode(), "$6$"))
		err := provider.UpdatePassword("harry", "newpassword")
		assert.NoError(t, err)

		// Reset the provider to force a read from disk.
		provider = NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ok, err := provider.CheckUserPassword("harry", "newpassword")
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.True(t, strings.HasPrefix(provider.database.Users["harry"].Digest.Encode(), "$argon2id$"))
	})
}

func TestShouldUpdatePasswordHashingAlgorithmToSHA512(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		config.Password.Algorithm = "sha2crypt"
		config.Password.SHA2Crypt.Iterations = 50000

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		assert.True(t, strings.HasPrefix(provider.database.Users["john"].Digest.Encode(), "$argon2id$"))
		err := provider.UpdatePassword("john", "newpassword")
		assert.NoError(t, err)

		// Reset the provider to force a read from disk.
		provider = NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ok, err := provider.CheckUserPassword("john", "newpassword")
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.True(t, strings.HasPrefix(provider.database.Users["john"].Digest.Encode(), "$6$"))
	})
}

func TestShouldRaiseWhenLoadingMalformedDatabaseForFirstTime(t *testing.T) {
	WithDatabase(MalformedUserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error reading the authentication database: could not parse the YAML database: yaml: line 4: mapping values are not allowed in this context")
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadSchemaForFirstTime(t *testing.T) {
	WithDatabase(BadSchemaUserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error reading the authentication database: could not validate the schema: Users: non zero value required")
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadSHA512HashesForTheFirstTime(t *testing.T) {
	WithDatabase(BadSHA512HashContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error decoding the authentication database: failed to parse hash for user 'john': sha2crypt decode error: provided encoded hash has an invalid option: option 'rounds00000' is invalid")
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashSettingsForTheFirstTime(t *testing.T) {
	WithDatabase(BadArgon2idHashSettingsContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error decoding the authentication database: failed to parse hash for user 'john': argon2 decode error: provided encoded hash has an invalid option: option 'm65536' is invalid")
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashKeyForTheFirstTime(t *testing.T) {
	WithDatabase(BadArgon2idHashKeyContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error decoding the authentication database: failed to parse hash for user 'john': argon2 decode error: provided encoded hash has a key value that can't be decoded: illegal base64 data at input byte 0")
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashSaltForTheFirstTime(t *testing.T) {
	WithDatabase(BadArgon2idHashSaltContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error decoding the authentication database: failed to parse hash for user 'john': argon2 decode error: provided encoded hash has a salt value that can't be decoded: illegal base64 data at input byte 0")
	})
}

func TestShouldSupportHashPasswordWithoutCRYPT(t *testing.T) {
	WithDatabase(UserDatabaseWithoutCryptContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ok, err := provider.CheckUserPassword("john", "password")

		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

var (
	DefaultFileAuthenticationBackendConfiguration = schema.FileAuthenticationBackend{
		Path:     "",
		Password: schema.DefaultCIPasswordConfig,
	}
)

var UserDatabaseContent = []byte(`
users:
  john:
    displayname: "John Doe"
    password: "{CRYPT}$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev

  harry:
    displayname: "Harry Potter"
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: harry.potter@authelia.com
    groups: []

  bob:
    displayname: "Bob Dylan"
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: bob.dylan@authelia.com
    groups:
      - dev

  james:
    displayname: "James Dean"
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: james.dean@authelia.com


  enumeration:
    displayname: "Enumeration"
    password: "$argon2id$v=19$m=131072,p=8$BpLnfgDsc2WD8F2q$O126GHPeZ5fwj7OLSs7PndXsTbje76R+QW9/EGfhkJg"
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

// The YAML is valid but the root key is user instead of users.
var BadSchemaUserDatabaseContent = []byte(`
user:
  john:
    displayname: "John Doe"
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
`)

var UserDatabaseWithoutCryptContent = []byte(`
users:
  john:
    displayname: "John Doe"
    password: "$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
  james:
    displayname: "James Dean"
    password: "$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: james.dean@authelia.com
`)

var BadSHA512HashContent = []byte(`
users:
  john:
    displayname: "John Doe"
    password: "$6$rounds00000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
  james:
    displayname: "James Dean"
    password: "$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: james.dean@authelia.com
`)

var BadArgon2idHashSettingsContent = []byte(`
users:
  john:
    displayname: "John Doe"
    password: "$argon2id$v=19$m65536,t3,p2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
  james:
    displayname: "James Dean"
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: james.dean@authelia.com
`)

var BadArgon2idHashKeyContent = []byte(`
users:
  john:
    displayname: "John Doe"
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$^^vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
`)

var BadArgon2idHashSaltContent = []byte(`
users:
  john:
    displayname: "John Doe"
    password: "$argon2id$v=19$m=65536,t=3,p=2$^^LnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
`)
