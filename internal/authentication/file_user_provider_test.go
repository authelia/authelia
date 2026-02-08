package authentication

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-crypt/crypt/algorithm/bcrypt"
	"github.com/go-crypt/crypt/algorithm/pbkdf2"
	"github.com/go-crypt/crypt/algorithm/scrypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestShouldErrorPermissionsOnLocalFS(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test due to being on windows")
	}

	dir := t.TempDir()

	_ = os.Mkdir(filepath.Join(dir, "noperms"), 0000)

	f := filepath.Join(dir, "noperms", "users_database.yml")
	require.EqualError(t, checkDatabase(f), fmt.Sprintf("error checking user authentication database file: stat %s: permission denied", f))
}

func TestShouldErrorAndGenerateUserDB(t *testing.T) {
	dir := t.TempDir()

	f := filepath.Join(dir, "users_database.yml")

	require.EqualError(t, checkDatabase(f), fmt.Sprintf("user authentication database file doesn't exist at path '%s' and has been generated", f))
}

func TestShouldErrorFailCreateDB(t *testing.T) {
	dir := t.TempDir()

	assert.NoError(t, os.Mkdir(filepath.Join(dir, "x"), 0000))

	f := filepath.Join(dir, "x", "users.yml")

	provider := NewFileUserProvider(&schema.AuthenticationBackendFile{Path: f, Password: schema.DefaultPasswordConfig})

	require.NotNil(t, provider)

	assert.EqualError(t, provider.StartupCheck(), "one or more errors occurred checking the authentication database")

	assert.NotNil(t, provider.database)

	reloaded, err := provider.Reload()

	assert.False(t, reloaded)
	assert.EqualError(t, err, fmt.Sprintf("failed to reload: error reading the authentication database: failed to read the '%s' file: open %s: permission denied", f, f))
}

func TestShouldErrorBadPasswordConfig(t *testing.T) {
	dir := t.TempDir()

	f := filepath.Join(dir, "users.yml")

	require.NoError(t, os.WriteFile(f, UserDatabaseContent, 0600))

	provider := NewFileUserProvider(&schema.AuthenticationBackendFile{Path: f})

	require.NotNil(t, provider)

	assert.EqualError(t, provider.StartupCheck(), "failed to initialize hash settings: argon2 validation error: parameter is invalid: parameter 't' must be between 1 and 2147483647 but is set to '0'")
}

func TestShouldNotPanicOnNilDB(t *testing.T) {
	dir := t.TempDir()

	f := filepath.Join(dir, "users.yml")

	assert.NoError(t, os.WriteFile(f, UserDatabaseContent, 0600))

	provider := &FileUserProvider{
		config:        &schema.AuthenticationBackendFile{Path: f, Password: schema.DefaultPasswordConfig},
		timeoutReload: time.Now().Add(-1 * time.Second),
	}

	assert.NoError(t, provider.StartupCheck())
}

func TestShouldHandleBadConfig(t *testing.T) {
	dir := t.TempDir()

	f := filepath.Join(dir, "users.yml")

	assert.NoError(t, os.WriteFile(f, UserDatabaseContentExtra, 0600))

	provider := &FileUserProvider{
		config:        &schema.AuthenticationBackendFile{Path: f, Password: schema.DefaultPasswordConfig, ExtraAttributes: map[string]schema.AuthenticationBackendExtraAttribute{"example": {ValueType: "integer"}}},
		mutex:         sync.Mutex{},
		timeoutReload: time.Now().Add(-1 * time.Second),
	}

	assert.EqualError(t, provider.StartupCheck(), "error decoding the authentication database: error occurred validating extra attributes for user 'john': attribute 'example' has the known type 'string' but 'integer' is the expected type")
}

func TestShouldReloadDatabase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.yml")

	testCases := []struct {
		name     string
		setup    func(t *testing.T, provider *FileUserProvider)
		expected bool
		err      string
	}{
		{
			"ShouldSkipReloadRecentlyReloaded",
			func(t *testing.T, provider *FileUserProvider) {
				provider.timeoutReload = time.Now().Add(time.Minute)
			},
			false,
			"",
		},
		{
			"ShouldReloadWithoutError",
			func(t *testing.T, provider *FileUserProvider) {
				provider.timeoutReload = time.Now().Add(time.Minute * -1)
			},
			true,
			"",
		},
		{
			"ShouldNotReloadWithNoContent",
			func(t *testing.T, provider *FileUserProvider) {
				p := filepath.Join(dir, "empty.yml")

				_, _ = os.Create(p)

				provider.timeoutReload = time.Now().Add(time.Minute * -1)

				provider.config.Path = p

				provider.database = NewFileUserDatabase(p, provider.config.Search.Email, provider.config.Search.CaseInsensitive, nil)
			},
			false,
			"",
		},
	}

	require.NoError(t, os.WriteFile(path, UserDatabaseContent, 0600))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewFileUserProvider(&schema.AuthenticationBackendFile{
				Path:     path,
				Password: schema.DefaultPasswordConfig,
				ExtraAttributes: map[string]schema.AuthenticationBackendExtraAttribute{
					"example": {
						ValueType: "string",
					},
				},
			})

			tc.setup(t, provider)

			actual, theError := provider.Reload()

			assert.Equal(t, tc.expected, actual)

			if tc.err == "" {
				assert.NoError(t, theError)
			} else {
				assert.EqualError(t, theError, tc.err)
			}
		})
	}
}

func TestShouldCheckUserArgon2idPasswordIsCorrect(t *testing.T) {
	WithDatabase(t, UserDatabaseContent, func(path string) {
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
	WithDatabase(t, UserDatabaseContent, func(path string) {
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
	WithDatabase(t, UserDatabaseContent, func(path string) {
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
	WithDatabase(t, UserDatabaseContent, func(path string) {
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
	WithDatabase(t, UserDatabaseContent, func(path string) {
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
	WithDatabase(t, UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		details, err := provider.GetDetails("john")
		assert.NoError(t, err)
		assert.Equal(t, "john", details.Username)
		assert.Equal(t, []string{"john.doe@authelia.com"}, details.Emails)
		assert.Equal(t, []string{"admins", "dev"}, details.Groups)

		extended, err := provider.GetDetailsExtended("john")
		assert.NoError(t, err)
		assert.Equal(t, "john", extended.Username)
		assert.Equal(t, []string{"john.doe@authelia.com"}, extended.Emails)
		assert.Equal(t, []string{"admins", "dev"}, extended.Groups)
	})
}

func TestShouldErrOnUserDetailsNoUser(t *testing.T) {
	WithDatabase(t, UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		details, err := provider.GetDetails("nouser")
		assert.Nil(t, details)
		assert.Equal(t, err, ErrUserNotFound)

		details, err = provider.GetDetails("dis")
		assert.Nil(t, details)
		assert.Equal(t, err, ErrUserNotFound)

		extended, err := provider.GetDetailsExtended("dis")
		assert.Nil(t, extended)
		assert.Equal(t, err, ErrUserNotFound)
	})
}

func TestShouldUpdatePassword(t *testing.T) {
	WithDatabase(t, UserDatabaseContent, func(path string) {
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
	WithDatabase(t, UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		db, ok := provider.database.(*FileUserDatabase)
		require.True(t, ok)

		assert.True(t, strings.HasPrefix(db.Users["harry"].Password.Encode(), "$6$"))

		err := provider.UpdatePassword("harry", "newpassword")
		assert.NoError(t, err)

		// Reset the provider to force a read from disk.
		provider = NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ok, err = provider.CheckUserPassword("harry", "newpassword")
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.True(t, strings.HasPrefix(db.Users["harry"].Password.Encode(), "$argon2id$"))
	})
}

func TestShouldUpdatePasswordHashingAlgorithmToSHA512(t *testing.T) {
	WithDatabase(t, UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		config.Password.Algorithm = "sha2crypt"
		config.Password.SHA2Crypt.Iterations = 50000

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		db, ok := provider.database.(*FileUserDatabase)
		require.True(t, ok)

		assert.True(t, strings.HasPrefix(db.Users["john"].Password.Encode(), "$argon2id$"))

		err := provider.UpdatePassword("john", "newpassword")
		assert.NoError(t, err)

		// Reset the provider to force a read from disk.
		provider = NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ok, err = provider.CheckUserPassword("john", "newpassword")
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.True(t, strings.HasPrefix(db.Users["john"].Password.Encode(), "$6$"))
	})
}

func TestShouldErrOnUpdatePasswordNoUser(t *testing.T) {
	WithDatabase(t, UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		assert.Equal(t, provider.UpdatePassword("nousers", "newpassword"), ErrUserNotFound)
		assert.Equal(t, provider.UpdatePassword("dis", "example"), ErrUserNotFound)
	})
}

func TestShouldRaiseWhenLoadingMalformedDatabaseForFirstTime(t *testing.T) {
	WithDatabase(t, MalformedUserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error reading the authentication database: could not parse the YAML database: yaml: line 4, column 6: mapping values are not allowed in this context")
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadSchemaForFirstTime(t *testing.T) {
	WithDatabase(t, BadSchemaUserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error reading the authentication database: could not validate the schema: users: non zero value required")
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadSHA512HashesForTheFirstTime(t *testing.T) {
	WithDatabase(t, BadSHA512HashContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error decoding the authentication database: error occurred decoding the password hash for 'john': shacrypt decode error: parameter pair 'rounds00000' is not properly encoded: does not contain kv separator '='")
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashSettingsForTheFirstTime(t *testing.T) {
	WithDatabase(t, BadArgon2idHashSettingsContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error decoding the authentication database: error occurred decoding the password hash for 'john': argon2 decode error: parameter pair 'm65536' is not properly encoded: does not contain kv separator '='")
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashKeyForTheFirstTime(t *testing.T) {
	WithDatabase(t, BadArgon2idHashKeyContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error decoding the authentication database: error occurred decoding the password hash for 'john': argon2 decode error: provided encoded hash has a key value that can't be decoded: illegal base64 data at input byte 0")
	})
}

func TestShouldRaiseWhenLoadingDatabaseWithBadArgon2idHashSaltForTheFirstTime(t *testing.T) {
	WithDatabase(t, BadArgon2idHashSaltContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error decoding the authentication database: error occurred decoding the password hash for 'john': argon2 decode error: provided encoded hash has a salt value that can't be decoded: illegal base64 data at input byte 0")
	})
}

func TestShouldSupportHashPasswordWithoutCRYPT(t *testing.T) {
	WithDatabase(t, UserDatabaseWithoutCryptContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ok, err := provider.CheckUserPassword("john", "password")

		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestShouldNotAllowLoginOfDisabledUsers(t *testing.T) {
	WithDatabase(t, UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ok, err := provider.CheckUserPassword("dis", "password")

		assert.False(t, ok)
		assert.EqualError(t, err, "user not found")
	})
}

func TestShouldErrorOnInvalidCaseSensitiveFile(t *testing.T) {
	WithDatabase(t, UserDatabaseContentInvalidSearchCaseInsenstive, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		config.Search.Email = false
		config.Search.CaseInsensitive = true

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error loading authentication database: username 'JOHN' is not lowercase but this is required when case-insensitive search is enabled")
	})
}

func TestShouldErrorOnDuplicateEmail(t *testing.T) {
	WithDatabase(t, UserDatabaseContentInvalidSearchEmail, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		config.Search.Email = true
		config.Search.CaseInsensitive = false

		provider := NewFileUserProvider(&config)

		err := provider.StartupCheck()
		assert.Regexp(t, regexp.MustCompile(`^error loading authentication database: email 'john.doe@authelia.com' is configured for for more than one user \(users are '(harry|john)', '(harry|john)'\) which isn't allowed when email search is enabled$`), err.Error())
	})
}

func TestShouldNotErrorOnEmailAsUsername(t *testing.T) {
	WithDatabase(t, UserDatabaseContentSearchEmailAsUsername, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		config.Search.Email = true
		config.Search.CaseInsensitive = false

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())
	})
}

func TestShouldErrorOnEmailAsUsernameWithDuplicateEmail(t *testing.T) {
	WithDatabase(t, UserDatabaseContentInvalidSearchEmailAsUsername, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		config.Search.Email = true
		config.Search.CaseInsensitive = false

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error loading authentication database: email 'john.doe@authelia.com' is also a username which isn't allowed when email search is enabled")
	})
}

func TestShouldErrorOnEmailAsUsernameWithDuplicateEmailCase(t *testing.T) {
	WithDatabase(t, UserDatabaseContentInvalidSearchEmailAsUsernameCase, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		config.Search.Email = false
		config.Search.CaseInsensitive = true

		provider := NewFileUserProvider(&config)

		assert.EqualError(t, provider.StartupCheck(), "error loading authentication database: username 'john.doe@authelia.com' is configured as an email for user with username 'john' which isn't allowed when case-insensitive search is enabled")
	})
}

func TestShouldAllowLookupByEmail(t *testing.T) {
	WithDatabase(t, UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		config.Search.Email = true

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ok, err := provider.CheckUserPassword("john", "password")

		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = provider.CheckUserPassword("john.doe@authelia.com", "password")

		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = provider.CheckUserPassword("JOHN.doe@authelia.com", "password")

		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestShouldAllowLookupCI(t *testing.T) {
	WithDatabase(t, UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path
		config.Search.CaseInsensitive = true

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ok, err := provider.CheckUserPassword("john", "password")

		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = provider.CheckUserPassword("John", "password")

		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestNewFileCryptoHashFromConfig(t *testing.T) {
	testCases := []struct {
		name     string
		have     schema.AuthenticationBackendFilePassword
		expected any
		err      string
	}{
		{
			"ShouldCreatePBKDF2",
			schema.AuthenticationBackendFilePassword{
				Algorithm: "pbkdf2",
				PBKDF2: schema.AuthenticationBackendFilePasswordPBKDF2{
					Variant:    "sha256",
					Iterations: 100000,
					SaltLength: 16,
				},
			},
			&pbkdf2.Hasher{},
			"",
		},
		{
			"ShouldCreateScrypt",
			schema.AuthenticationBackendFilePassword{
				Algorithm: "scrypt",
				Scrypt: schema.AuthenticationBackendFilePasswordScrypt{
					Iterations:  12,
					SaltLength:  16,
					Parallelism: 1,
					BlockSize:   1,
					KeyLength:   32,
				},
			},
			&scrypt.Hasher{},
			"",
		},
		{
			"ShouldCreateBcrypt",
			schema.AuthenticationBackendFilePassword{
				Algorithm: "bcrypt",
				Bcrypt: schema.AuthenticationBackendFilePasswordBcrypt{
					Variant: "standard",
					Cost:    12,
				},
			},
			&bcrypt.Hasher{},
			"",
		},
		{
			"ShouldFailToCreateScryptInvalidParameter",
			schema.AuthenticationBackendFilePassword{
				Algorithm: "scrypt",
			},
			nil,
			"failed to initialize hash settings: scrypt validation error: parameter is invalid: parameter 'iterations' must be between 1 and 58 but is set to '0'",
		},
		{
			"ShouldFailUnknown",
			schema.AuthenticationBackendFilePassword{
				Algorithm: "unknown",
			},
			nil,
			"algorithm 'unknown' is unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, theError := NewFileCryptoHashFromConfig(tc.have)

			if tc.err == "" {
				assert.NoError(t, theError)
				require.NotNil(t, actual)
				assert.IsType(t, tc.expected, actual)
			} else {
				assert.EqualError(t, theError, tc.err)
				assert.Nil(t, actual)
			}
		})
	}
}

func TestHashError(t *testing.T) {
	WithDatabase(t, UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Search.CaseInsensitive = true
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mock := NewMockHash(ctrl)
		provider.hash = mock

		mock.EXPECT().Hash("apple123").Return(nil, fmt.Errorf("failed to mock hash"))

		assert.EqualError(t, provider.UpdatePassword("john", "apple123"), "failed to mock hash")
	})
}

func TestDatabaseError(t *testing.T) {
	WithDatabase(t, UserDatabaseContent, func(path string) {
		db := NewFileUserDatabase(path, false, false, nil)
		assert.NoError(t, db.Load())

		config := DefaultFileAuthenticationBackendConfiguration
		config.Search.CaseInsensitive = true
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mock := NewMockFileUserDatabase(ctrl)

		provider.database = mock

		gomock.InOrder(
			mock.EXPECT().GetUserDetails("john").Return(db.GetUserDetails("john")),
			mock.EXPECT().SetUserDetails("john", gomock.Any()),
			mock.EXPECT().Save().Return(fmt.Errorf("failed to mock save")),
		)

		assert.EqualError(t, provider.UpdatePassword("john", "apple123"), "failed to mock save")
	})
}

func TestDatabaseErrorExtended(t *testing.T) {
	WithDatabase(t, UserDatabaseContent, func(path string) {
		db := NewFileUserDatabase(path, false, false, nil)
		assert.NoError(t, db.Load())

		config := DefaultFileAuthenticationBackendConfiguration
		config.Search.CaseInsensitive = true
		config.Path = path

		provider := NewFileUserProvider(&config)

		assert.NoError(t, provider.StartupCheck())

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mock := NewMockFileUserDatabase(ctrl)

		provider.database = mock

		gomock.InOrder(
			mock.EXPECT().GetUserDetails("john").Return(FileUserDatabaseUserDetails{}, fmt.Errorf("bad")),
		)

		details, err := provider.GetDetailsExtended("john")
		assert.Nil(t, details)
		assert.EqualError(t, err, "bad")
	})
}

var (
	DefaultFileAuthenticationBackendConfiguration = schema.AuthenticationBackendFile{
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
    email: enumeration@authelia.com


  dis:
    displayname: "Enumeration"
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    disabled: true
    email: disabled@authelia.com
`)

var UserDatabaseContentExtra = []byte(`
users:
  john:
    displayname: "John Doe"
    password: "{CRYPT}$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
    extra:
      example: '123'
`)

var UserDatabaseContentInvalidSearchCaseInsenstive = []byte(`
users:
  john:
    displayname: "John Doe"
    password: "{CRYPT}$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev

  JOHN:
    displayname: "Harry Potter"
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: harry.potter@authelia.com
    groups: []
`)

var UserDatabaseContentInvalidSearchEmail = []byte(`
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
    email: john.doe@authelia.com
    groups: []
`)

var UserDatabaseContentSearchEmailAsUsername = []byte(`
users:
  john.doe@authelia.com:
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
`)

var UserDatabaseContentInvalidSearchEmailAsUsername = []byte(`
users:
  john.doe@authelia.com:
    displayname: "John Doe"
    password: "{CRYPT}$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john@authelia.com
    groups:
      - admins
      - dev

  harry:
    displayname: "Harry Potter"
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups: []
`)

var UserDatabaseContentInvalidSearchEmailAsUsernameCase = []byte(`
users:
  john.doe@authelia.com:
    displayname: "John Doe"
    password: "{CRYPT}$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: JOHN@authelia.com
    groups:
      - admins
      - dev

  john:
    displayname: "John Potter"
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups: []
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

func WithDatabase(t *testing.T, content []byte, f func(path string)) {
	t.Helper()

	dir := t.TempDir()

	db, err := os.CreateTemp(dir, "users_database.*.yml")
	require.NoError(t, err)

	_, err = db.Write(content)
	require.NoError(t, err)

	f(db.Name())

	require.NoError(t, db.Close())
}
