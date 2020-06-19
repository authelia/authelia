package authentication

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"aletheia.icu/broccoli/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestShouldErrorNoUserDBInEmbeddedFS(t *testing.T) {
	oldCfg := cfg
	cfg = fs.New(false, []byte("\x1b~\x00\x80\x8d\x94n\xc2|\x84J\xf7\xbfn\xfd\xf7w;.\x8d m\xb2&\xd1Z\xec\xb2\x05\xb9\xc00\x8a\xf7(\x80^78\t(\f\f\xc3p\xc2\xc1\x06[a\xa2\xb3\xa4P\xe5\xa14\xfb\x19\xb2cp\xf6\x90-Z\xb2\x11\xe0l\xa1\x80\\\x95Vh\t\xc5\x06\x16\xfa\x8c\xc0\"!\xa5\xcf\xf7$\x9a\xb2\a`\xc6\x18\xc8~\xce8\r\x16Z\x9d\xc3\xe3\xff\x00"))
	errors := checkDatabase("./nonexistent.yml")
	cfg = oldCfg

	require.Len(t, errors, 3)

	require.EqualError(t, errors[0], "Unable to find database file: ./nonexistent.yml")
	require.EqualError(t, errors[1], "Generating database file: ./nonexistent.yml")
	require.EqualError(t, errors[2], "Unable to open users_database.template.yml: file does not exist")
}

func TestShouldErrorPermissionsOnLocalFS(t *testing.T) {
	_ = os.Mkdir("/tmp/noperms/", 0000)
	errors := checkDatabase("/tmp/noperms/users_database.yml")

	require.Len(t, errors, 3)

	require.EqualError(t, errors[0], "Unable to find database file: /tmp/noperms/users_database.yml")
	require.EqualError(t, errors[1], "Generating database file: /tmp/noperms/users_database.yml")
	require.EqualError(t, errors[2], "Unable to generate /tmp/noperms/users_database.yml: open /tmp/noperms/users_database.yml: permission denied")
}

func TestShouldErrorAndGenerateUserDB(t *testing.T) {
	errors := checkDatabase("./nonexistent.yml")
	_ = os.Remove("./nonexistent.yml")

	require.Len(t, errors, 3)

	require.EqualError(t, errors[0], "Unable to find database file: ./nonexistent.yml")
	require.EqualError(t, errors[1], "Generating database file: ./nonexistent.yml")
	require.EqualError(t, errors[2], "Generated database at: ./nonexistent.yml")
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

func TestShouldCheckUserPasswordIsWrongForEnumerationCompare(t *testing.T) {
	WithDatabase(UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		provider := NewFileUserProvider(&config)

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
		assert.True(t, strings.HasPrefix(provider.database.Users["harry"].HashedPassword, "$6$"))
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
		config.Password.Algorithm = "sha512"
		config.Password.Iterations = 50000

		provider := NewFileUserProvider(&config)
		assert.True(t, strings.HasPrefix(provider.database.Users["john"].HashedPassword, "$argon2id$"))
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
		assert.PanicsWithError(t, "Unable to parse database: yaml: line 4: mapping values are not allowed in this context", func() {
			NewFileUserProvider(&config)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadSchemaForFirstTime(t *testing.T) {
	WithDatabase(BadSchemaUserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		assert.PanicsWithError(t, "Invalid schema of database: Users: non zero value required", func() {
			NewFileUserProvider(&config)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadSHA512HashesForTheFirstTime(t *testing.T) {
	WithDatabase(BadSHA512HashContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		assert.PanicsWithError(t, "Unable to parse hash of user john: Hash key is not the last parameter, the hash is likely malformed ($6$rounds00000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/)", func() {
			NewFileUserProvider(&config)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashSettingsForTheFirstTime(t *testing.T) {
	WithDatabase(BadArgon2idHashSettingsContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		assert.PanicsWithError(t, "Unable to parse hash of user john: Hash key is not the last parameter, the hash is likely malformed ($argon2id$v=19$m65536,t3,p2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM)", func() {
			NewFileUserProvider(&config)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashKeyForTheFirstTime(t *testing.T) {
	WithDatabase(BadArgon2idHashKeyContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		assert.PanicsWithError(t, "Unable to parse hash of user john: Hash key contains invalid base64 characters", func() {
			NewFileUserProvider(&config)
		})
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashSaltForTheFirstTime(t *testing.T) {
	WithDatabase(BadArgon2idHashSaltContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		assert.PanicsWithError(t, "Unable to parse hash of user john: Salt contains invalid base64 characters", func() {
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
		Password: &schema.PasswordConfiguration{
			Iterations:  schema.DefaultCIPasswordConfiguration.Iterations,
			KeyLength:   schema.DefaultCIPasswordConfiguration.KeyLength,
			SaltLength:  schema.DefaultCIPasswordConfiguration.SaltLength,
			Algorithm:   schema.DefaultCIPasswordConfiguration.Algorithm,
			Memory:      schema.DefaultCIPasswordConfiguration.Memory,
			Parallelism: schema.DefaultCIPasswordConfiguration.Parallelism,
		},
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
