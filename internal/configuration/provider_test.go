package configuration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/utils"
)

func TestShouldErrorSecretNotExist(t *testing.T) {
	testResetEnv()

	dir, err := ioutil.TempDir("", "authelia-test-secret-not-exist")
	assert.NoError(t, err)

	assert.NoError(t, os.Setenv(envPrefix+"JWT_SECRET_FILE", filepath.Join(dir, "jwt")))
	assert.NoError(t, os.Setenv(envPrefix+"DUO_API_SECRET_KEY_FILE", filepath.Join(dir, "duo")))
	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET_FILE", filepath.Join(dir, "session")))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", filepath.Join(dir, "authentication")))
	assert.NoError(t, os.Setenv(envPrefix+"NOTIFIER_SMTP_PASSWORD_FILE", filepath.Join(dir, "notifier")))
	assert.NoError(t, os.Setenv(envPrefix+"SESSION_REDIS_PASSWORD_FILE", filepath.Join(dir, "redis")))
	assert.NoError(t, os.Setenv(envPrefix+"SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE", filepath.Join(dir, "redis-sentinel")))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD_FILE", filepath.Join(dir, "mysql")))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_POSTGRES_PASSWORD_FILE", filepath.Join(dir, "postgres")))
	assert.NoError(t, os.Setenv(envPrefix+"TLS_KEY_FILE", filepath.Join(dir, "tls")))
	assert.NoError(t, os.Setenv(envPrefix+"IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE", filepath.Join(dir, "oidc-key")))
	assert.NoError(t, os.Setenv(envPrefix+"IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE", filepath.Join(dir, "oidc-hmac")))

	p := NewProvider()

	assert.NoError(t, p.LoadEnvironment())
	assert.NoError(t, p.LoadSecrets())

	errs := p.Errors()

	require.Len(t, errs, 12)

	errFmt := utils.GetExpectedErrTxt("filenotfound")

	assert.EqualError(t, errs[0], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "authentication"), "authentication_backend.ldap.password", fmt.Sprintf(errFmt, filepath.Join(dir, "authentication"))))
	assert.EqualError(t, errs[1], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "duo"), "duo_api.secret_key", fmt.Sprintf(errFmt, filepath.Join(dir, "duo"))))
	assert.EqualError(t, errs[2], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "oidc-hmac"), "identity_providers.oidc.hmac_secret", fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-hmac"))))
	assert.EqualError(t, errs[3], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "oidc-key"), "identity_providers.oidc.issuer_private_key", fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-key"))))
	assert.EqualError(t, errs[4], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "jwt"), "jwt_secret", fmt.Sprintf(errFmt, filepath.Join(dir, "jwt"))))
	assert.EqualError(t, errs[5], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "notifier"), "notifier.smtp.password", fmt.Sprintf(errFmt, filepath.Join(dir, "notifier"))))
	assert.EqualError(t, errs[6], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "redis-sentinel"), "session.redis.high_availability.sentinel_password", fmt.Sprintf(errFmt, filepath.Join(dir, "redis-sentinel"))))
	assert.EqualError(t, errs[7], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "redis"), "session.redis.password", fmt.Sprintf(errFmt, filepath.Join(dir, "redis"))))
	assert.EqualError(t, errs[8], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "session"), "session.secret", fmt.Sprintf(errFmt, filepath.Join(dir, "session"))))
	assert.EqualError(t, errs[9], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "mysql"), "storage.mysql.password", fmt.Sprintf(errFmt, filepath.Join(dir, "mysql"))))
	assert.EqualError(t, errs[10], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "postgres"), "storage.postgres.password", fmt.Sprintf(errFmt, filepath.Join(dir, "postgres"))))
	assert.EqualError(t, errs[11], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "tls"), "tls_key", fmt.Sprintf(errFmt, filepath.Join(dir, "tls"))))
}

func TestShouldValidateConfigurationWithEnv(t *testing.T) {
	testResetEnv()

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	p := NewProvider()

	assert.NoError(t, p.LoadPaths([]string{"./test_resources/config.yml"}))
	assert.NoError(t, p.LoadEnvironment())
	assert.NoError(t, p.LoadSecrets())

	assert.NoError(t, p.UnmarshalToStruct())
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.ValidateConfiguration()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)
}

func TestShouldIgnoreInvalidEnvs(t *testing.T) {
	testResetEnv()

	p := NewProvider()

	assert.NoError(t, p.LoadPaths([]string{"./test_resources/config.yml"}))

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL", "a bad env"))
	assert.NoError(t, os.Setenv(envPrefix+"JWT_SECRET", "an env jwt secret"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))
	assert.NoError(t, os.Setenv(envPrefixAlt+"AUTHENTICATION_BACKEND_LDAP_URL", "an env authentication backend ldap password"))

	require.NoError(t, p.LoadEnvironment())
	require.NoError(t, p.LoadSecrets())
	require.NoError(t, p.UnmarshalToStruct())

	p.ValidateConfiguration()

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)
}

func TestShouldIgnoreSingleUnderscoreNonSecretEnvs(t *testing.T) {
	testResetEnv()

	p := NewProvider()

	assert.NoError(t, p.LoadPaths([]string{"./test_resources/config.yml"}))

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(envPrefix+"JWT_SECRET", "an env jwt secret"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))
	assert.NoError(t, os.Setenv(envPrefixAlt+"AUTHENTICATION_BACKEND_LDAP_URL", "an env authentication backend ldap password"))

	require.NoError(t, p.LoadEnvironment())
	require.NoError(t, p.LoadSecrets())
	require.NoError(t, p.UnmarshalToStruct())

	p.ValidateConfiguration()

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.Equal(t, "an env jwt secret", p.Configuration.JWTSecret)
	assert.Equal(t, "an env session secret", p.Configuration.Session.Secret)
	assert.Equal(t, "an env storage mysql password", p.Configuration.Storage.MySQL.Password)
	assert.Equal(t, "an env authentication backend ldap password", p.Configuration.AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "ldap://127.0.0.1", p.Configuration.AuthenticationBackend.LDAP.URL)
}

func TestShouldAllowBothLegacyEnvSecretFilesAndNewOnes(t *testing.T) {
	testResetEnv()

	dir, err := ioutil.TempDir("", "authelia-test-secrets")
	assert.NoError(t, err)

	sessionSecret := filepath.Join(dir, "session")
	jwtSecret := filepath.Join(dir, "jwt")
	ldapSecret := filepath.Join(dir, "ldap")
	storageSecret := filepath.Join(dir, "storage")

	assert.NoError(t, testCreateFile(sessionSecret, "a secret session.secret value", 0600))
	assert.NoError(t, testCreateFile(jwtSecret, "a secret jwt_secret value", 0600))
	assert.NoError(t, testCreateFile(ldapSecret, "a secret authentication_backend.ldap.password value", 0600))
	assert.NoError(t, testCreateFile(storageSecret, "a secret storage.mysql.password value", 0600))

	p := NewProvider()

	assert.NoError(t, p.LoadPaths([]string{"./test_resources/config.yml"}))

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET_FILE", sessionSecret))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD_FILE", storageSecret))
	assert.NoError(t, os.Setenv(envPrefixAlt+"JWT_SECRET_FILE", jwtSecret))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", ldapSecret))

	assert.NoError(t, p.LoadEnvironment())
	assert.NoError(t, p.LoadSecrets())

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.NoError(t, p.UnmarshalToStruct())
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.ValidateConfiguration()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.Equal(t, "a secret jwt_secret value", p.Configuration.JWTSecret)
	assert.Equal(t, "a secret session.secret value", p.Configuration.Session.Secret)
	assert.Equal(t, "a secret storage.mysql.password value", p.Configuration.Storage.MySQL.Password)
	assert.Equal(t, "a secret authentication_backend.ldap.password value", p.Configuration.AuthenticationBackend.LDAP.Password)
}

func TestShouldValidateAndRaiseErrorsOnNormalConfigurationAndSecret(t *testing.T) {
	testResetEnv()

	p := NewProvider()

	assert.NoError(t, p.LoadPaths([]string{"./test_resources/config.yml"}))

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(envPrefixAlt+"JWT_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))

	assert.NoError(t, p.LoadEnvironment())
	assert.NoError(t, p.LoadSecrets())

	require.Len(t, p.Errors(), 1)
	assert.Len(t, p.Warnings(), 0)
	assert.EqualError(t, p.Errors()[0], "error loading secret into key 'session.secret': it's already defined in the config files")

	p.Clear()

	assert.NoError(t, p.UnmarshalToStruct())
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.ValidateConfiguration()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.Equal(t, "example_secret value", p.Configuration.JWTSecret)
	assert.Equal(t, "an env session secret", p.Configuration.Session.Secret)
	assert.Equal(t, "an env storage mysql password", p.Configuration.Storage.MySQL.Password)
	assert.Equal(t, "an env authentication backend ldap password", p.Configuration.AuthenticationBackend.LDAP.Password)
}

func TestShouldRaiseIOErrOnUnreadableFile(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping test due to being on windows")
	}

	dir, err := ioutil.TempDir("", "authelia-conf")
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "myconf.yml"), []byte("server:\n  port: 9091\n"), 0000))

	p := NewProvider()

	assert.EqualError(t, p.LoadPaths([]string{filepath.Join(dir, "myconf.yml")}), "one or more errors occurred while loading configuration files")

	require.Len(t, p.Errors(), 1)
	assert.Len(t, p.Warnings(), 0)
	assert.EqualError(t, p.Errors()[0], fmt.Sprintf("configuration file could not be loaded due to an error: open %s: permission denied", filepath.Join(dir, "myconf.yml")))
}

func TestShouldValidateConfigurationWithEnvSecrets(t *testing.T) {
	testResetEnv()

	p := NewProvider()

	assert.NoError(t, p.LoadPaths([]string{"./test_resources/config.yml"}))

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(envPrefix+"JWT_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", "./test_resources/example_secret"))

	assert.NoError(t, p.LoadEnvironment())
	assert.NoError(t, p.LoadSecrets())
	assert.NoError(t, p.UnmarshalToStruct())
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.Equal(t, "example_secret value", p.Configuration.JWTSecret)
	assert.Equal(t, "example_secret value", p.Configuration.Session.Secret)
	assert.Equal(t, "example_secret value", p.Configuration.AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "example_secret value", p.Configuration.Storage.MySQL.Password)

	p.ValidateConfiguration()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)
}

func TestShouldValidateAndRaiseErrorsOnBadConfiguration(t *testing.T) {
	testResetEnv()

	p := NewProvider()

	assert.NoError(t, p.LoadPaths([]string{"./test_resources/config_bad_keys.yml"}))
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	assert.NoError(t, p.LoadEnvironment())
	assert.NoError(t, p.LoadSecrets())
	assert.NoError(t, p.UnmarshalToStruct())

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.ValidateConfiguration()
	assert.Len(t, p.Errors(), 2)
	assert.Len(t, p.Warnings(), 0)

	assert.EqualError(t, p.Errors()[0], "config key not expected: loggy_file")
	assert.EqualError(t, p.Errors()[1], "invalid configuration key 'logs_level' was replaced by 'log.level'")
}

func TestShouldGenerateConfiguration(t *testing.T) {
	testResetEnv()

	p := NewProvider()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	cfg := filepath.Join(dir, "config.yml")

	err = p.LoadPaths([]string{cfg})
	assert.EqualError(t, err, "one or more errors occurred while loading configuration files")

	_, err = os.Stat(cfg)
	assert.NoError(t, err)

	require.Len(t, p.Errors(), 1)
	assert.Len(t, p.Warnings(), 0)
	assert.EqualError(t, p.Errors()[0], fmt.Sprintf("configuration file did not exist at %s and generated with defaults but you will need to configure it", cfg))
}

func TestShouldNotGenerateConfigurationOnFSAccessDenied(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping test due to being on windows")
	}

	testResetEnv()

	p := NewProvider()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	assert.NoError(t, os.Mkdir(filepath.Join(dir, "zero"), 0000))

	cfg := filepath.Join(dir, "zero", "config.yml")

	err = p.LoadPaths([]string{cfg})
	assert.EqualError(t, err, "one or more errors occurred while loading configuration files")

	_, err = os.Stat(cfg)
	assert.EqualError(t, err, fmt.Sprintf("stat %s: permission denied", cfg))

	require.Len(t, p.Errors(), 1)
	assert.Len(t, p.Warnings(), 0)
	assert.EqualError(t, p.Errors()[0], fmt.Sprintf("configuration file could not be loaded due to an error: stat %s: permission denied", cfg))
}

func TestShouldNotReadConfigurationOnFSAccessDenied(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping test due to being on windows")
	}

	testResetEnv()

	p := NewProvider()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	assert.NoError(t, testCreateFile(filepath.Join(dir, "config.yml"), "port: 9091\n", 0000))

	cfg := filepath.Join(dir, "config.yml")

	err = p.LoadPaths([]string{cfg})
	assert.EqualError(t, err, "one or more errors occurred while loading configuration files")

	require.Len(t, p.Errors(), 1)
	assert.Len(t, p.Warnings(), 0)
	assert.EqualError(t, p.Errors()[0], fmt.Sprintf("configuration file could not be loaded due to an error: open %s: permission denied", cfg))
}

func TestShouldNotGenerateMultipleConfigurations(t *testing.T) {
	testResetEnv()

	p := NewProvider()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	cfgOne := filepath.Join(dir, "config.yml")
	cfgTwo := filepath.Join(dir, "config-acl.yml")

	err = p.LoadPaths([]string{cfgOne, cfgTwo})
	assert.EqualError(t, err, "one or more errors occurred while loading configuration files")

	require.Len(t, p.Errors(), 2)
	assert.Len(t, p.Warnings(), 0)
	assert.EqualError(t, p.Errors()[0], fmt.Sprintf("configuration file does not exist at %s", cfgOne))
	assert.EqualError(t, p.Errors()[1], fmt.Sprintf("configuration file does not exist at %s", cfgTwo))
}

func TestShouldNotGenerateConfiguration(t *testing.T) {
	testResetEnv()

	p := NewProvider()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	cfg := filepath.Join(dir, "..", "not-a-dir", "config.yml")

	err = p.LoadPaths([]string{cfg})
	assert.EqualError(t, err, "one or more errors occurred while loading configuration files")

	require.Len(t, p.Errors(), 1)
	assert.Len(t, p.Warnings(), 0)
	assert.EqualError(t, p.Errors()[0], fmt.Sprintf("configuration file could not be generated at %s: %s", cfg, fmt.Sprintf(utils.GetExpectedErrTxt("pathnotfound"), cfg)))
}

func TestShouldNotLoadDirectoryConfiguration(t *testing.T) {
	testResetEnv()

	p := NewProvider()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	err = p.LoadPaths([]string{dir, dir})
	assert.EqualError(t, err, "one or more errors occurred while loading configuration files")

	require.Len(t, p.Errors(), 2)
	assert.Len(t, p.Warnings(), 0)
	assert.EqualError(t, p.Errors()[0], fmt.Sprintf("error loading path '%s': is not a file", dir))
	assert.EqualError(t, p.Errors()[1], fmt.Sprintf("error loading path '%s': is not a file", dir))
}

func TestShouldRetrieveGlobalConfiguration(t *testing.T) {
	testResetEnv()

	p := GetProvider()

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "xyz"))

	assert.NoError(t, p.LoadEnvironment())
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.NoError(t, p.UnmarshalToStruct())
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.Equal(t, "xyz", p.Configuration.Session.Secret)

	q := GetProvider()
	assert.Equal(t, "xyz", q.Configuration.Session.Secret)
}

func testResetEnv() {
	_ = os.Unsetenv(envPrefixAlt + "JWT_SECRET_FILE")
	_ = os.Unsetenv(envPrefixAlt + "JWT_SECRET")
	_ = os.Unsetenv(envPrefix + "JWT_SECRET_FILE")
	_ = os.Unsetenv(envPrefix + "DUO_API_SECRET_KEY_FILE")
	_ = os.Unsetenv(envPrefix + "SESSION_SECRET_FILE")
	_ = os.Unsetenv(envPrefix + "AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE")
	_ = os.Unsetenv(envPrefix + "NOTIFIER_SMTP_PASSWORD_FILE")
	_ = os.Unsetenv(envPrefix + "SESSION_REDIS_PASSWORD_FILE")
	_ = os.Unsetenv(envPrefix + "SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE")
	_ = os.Unsetenv(envPrefix + "STORAGE_MYSQL_PASSWORD_FILE")
	_ = os.Unsetenv(envPrefix + "STORAGE_POSTGRES_PASSWORD_FILE")
	_ = os.Unsetenv(envPrefix + "TLS_KEY_FILE")
	_ = os.Unsetenv(envPrefix + "IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE")
	_ = os.Unsetenv(envPrefix + "IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE")
	_ = os.Unsetenv(envPrefix + "JWT_SECRET")
	_ = os.Unsetenv(envPrefix + "DUO_API_SECRET_KEY")
	_ = os.Unsetenv(envPrefix + "SESSION_SECRET")
	_ = os.Unsetenv(envPrefix + "AUTHENTICATION_BACKEND_LDAP_PASSWORD")
	_ = os.Unsetenv(envPrefix + "NOTIFIER_SMTP_PASSWORD")
	_ = os.Unsetenv(envPrefix + "SESSION_REDIS_PASSWORD")
	_ = os.Unsetenv(envPrefix + "SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD")
	_ = os.Unsetenv(envPrefix + "STORAGE_MYSQL_PASSWORD")
	_ = os.Unsetenv(envPrefix + "STORAGE_POSTGRES_PASSWORD")
	_ = os.Unsetenv(envPrefix + "TLS_KEY")
	_ = os.Unsetenv(envPrefix + "PORT")
	_ = os.Unsetenv(envPrefix + "PORT_K8S_EXAMPLE")
	_ = os.Unsetenv(envPrefix + "IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY")
	_ = os.Unsetenv(envPrefix + "IDENTITY_PROVIDERS_OIDC_HMAC_SECRET")
}

func testCreateFile(path, value string, perm os.FileMode) (err error) {
	return os.WriteFile(path, []byte(value), perm)
}
