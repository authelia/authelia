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

func resetTestEnv() {
	_ = os.Unsetenv("AUTHELIA_JWT_SECRET_FILE")
	_ = os.Unsetenv("AUTHELIA_DUO_API_SECRET_KEY_FILE")
	_ = os.Unsetenv("AUTHELIA_SESSION_SECRET_FILE")
	_ = os.Unsetenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE")
	_ = os.Unsetenv("AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE")
	_ = os.Unsetenv("AUTHELIA_SESSION_REDIS_PASSWORD_FILE")
	_ = os.Unsetenv("AUTHELIA_SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE")
	_ = os.Unsetenv("AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE")
	_ = os.Unsetenv("AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE")
	_ = os.Unsetenv("AUTHELIA_TLS_KEY_FILE")
	_ = os.Unsetenv("AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE")
	_ = os.Unsetenv("AUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE")
	_ = os.Unsetenv("AUTHELIA_JWT_SECRET")
	_ = os.Unsetenv("AUTHELIA_DUO_API_SECRET_KEY")
	_ = os.Unsetenv("AUTHELIA_SESSION_SECRET")
	_ = os.Unsetenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD")
	_ = os.Unsetenv("AUTHELIA_NOTIFIER_SMTP_PASSWORD")
	_ = os.Unsetenv("AUTHELIA_SESSION_REDIS_PASSWORD")
	_ = os.Unsetenv("AUTHELIA_SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD")
	_ = os.Unsetenv("AUTHELIA_STORAGE_MYSQL_PASSWORD")
	_ = os.Unsetenv("AUTHELIA_STORAGE_POSTGRES_PASSWORD")
	_ = os.Unsetenv("AUTHELIA_TLS_KEY")
	_ = os.Unsetenv("AUTHELIA_PORT")
	_ = os.Unsetenv("AUTHELIA_PORT_K8S_EXAMPLE")
	_ = os.Unsetenv("AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY")
	_ = os.Unsetenv("AUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET")
}

func TestShouldErrorSecretNotExist(t *testing.T) {
	resetTestEnv()

	dir, err := ioutil.TempDir("", "authelia-test-secret-not-exist")
	assert.NoError(t, err)

	assert.NoError(t, os.Setenv("AUTHELIA_JWT_SECRET_FILE", filepath.Join(dir, "jwt")))
	assert.NoError(t, os.Setenv("AUTHELIA_DUO_API_SECRET_KEY_FILE", filepath.Join(dir, "duo")))
	assert.NoError(t, os.Setenv("AUTHELIA_SESSION_SECRET_FILE", filepath.Join(dir, "session")))
	assert.NoError(t, os.Setenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", filepath.Join(dir, "authentication")))
	assert.NoError(t, os.Setenv("AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE", filepath.Join(dir, "notifier")))
	assert.NoError(t, os.Setenv("AUTHELIA_SESSION_REDIS_PASSWORD_FILE", filepath.Join(dir, "redis")))
	assert.NoError(t, os.Setenv("AUTHELIA_SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE", filepath.Join(dir, "redis-sentinel")))
	assert.NoError(t, os.Setenv("AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE", filepath.Join(dir, "mysql")))
	assert.NoError(t, os.Setenv("AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE", filepath.Join(dir, "postgres")))
	assert.NoError(t, os.Setenv("AUTHELIA_TLS_KEY_FILE", filepath.Join(dir, "tls")))
	assert.NoError(t, os.Setenv("AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE", filepath.Join(dir, "oidc-key")))
	assert.NoError(t, os.Setenv("AUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE", filepath.Join(dir, "oidc-hmac")))

	p := NewProvider()

	assert.NoError(t, p.LoadEnvironment())
	assert.Error(t, p.LoadSecrets(), "one or more errors occurred during loading secrets")

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
	resetTestEnv()

	assert.NoError(t, os.Setenv("AUTHELIA_SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv("AUTHELIA_STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv("AUTHELIA_JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	p := NewProvider()

	assert.NoError(t, p.LoadPaths([]string{"./test_resources/config.yml"}))
	assert.NoError(t, p.LoadEnvironment())
	assert.NoError(t, p.LoadSecrets())

	assert.NoError(t, p.UnmarshalToStruct())
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.ValidateKeys()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.ValidateConfiguration()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)
}

func TestShouldNotLoadIgnoredKeys(t *testing.T) {
	resetTestEnv()

	assert.NoError(t, os.Setenv("AUTHELIA_PORT", "abc"))
	assert.NoError(t, os.Setenv("AUTHELIA_PORT_K8S_EXAMPLE", "abc"))

	p := NewProvider()

	assert.NoError(t, p.LoadEnvironment())
	assert.NoError(t, p.UnmarshalToStruct())
	assert.Len(t, p.Keys(), 0)
}

func TestShouldValidateAndRaiseErrorsOnNormalConfigurationAndSecret(t *testing.T) {
	resetTestEnv()

	p := NewProvider()

	assert.NoError(t, p.LoadPaths([]string{"./test_resources/config.yml"}))

	assert.NoError(t, os.Setenv("AUTHELIA_SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv("AUTHELIA_SESSION_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv("AUTHELIA_STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv("AUTHELIA_JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	assert.NoError(t, p.LoadEnvironment())
	assert.EqualError(t, p.LoadSecrets(), "one or more errors occurred during loading secrets")

	require.Len(t, p.Errors(), 1)
	assert.Len(t, p.Warnings(), 0)
	assert.EqualError(t, p.Errors()[0], "error loading secret into key 'session.secret': it's already defined in the config file")

	p.Clear()

	assert.NoError(t, p.UnmarshalToStruct())
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.ValidateKeys()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.ValidateConfiguration()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)
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
	assert.EqualError(t, p.Errors()[0], fmt.Sprintf("open %s: permission denied", filepath.Join(dir, "myconf.yml")))
}

func TestShouldValidateConfigurationWithEnvSecrets(t *testing.T) {
	resetTestEnv()

	p := NewProvider()

	assert.NoError(t, p.LoadPaths([]string{"./test_resources/config.yml"}))

	assert.NoError(t, os.Setenv("AUTHELIA_SESSION_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv("AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv("AUTHELIA_JWT_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", "./test_resources/example_secret"))

	assert.NoError(t, p.LoadEnvironment())
	assert.NoError(t, p.LoadSecrets())
	assert.NoError(t, p.UnmarshalToStruct())
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.Equal(t, "example_secret value", p.Configuration.JWTSecret)
	assert.Equal(t, "example_secret value", p.Configuration.Session.Secret)
	assert.Equal(t, "example_secret value", p.Configuration.AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "example_secret value", p.Configuration.Storage.MySQL.Password)

	p.ValidateKeys()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.ValidateConfiguration()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)
}

func TestShouldValidateAndRaiseErrorsOnBadConfiguration(t *testing.T) {
	resetTestEnv()

	p := NewProvider()

	assert.NoError(t, p.LoadPaths([]string{"./test_resources/config_bad_keys.yml"}))
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.NoError(t, os.Setenv("AUTHELIA_SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv("AUTHELIA_STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv("AUTHELIA_JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	assert.NoError(t, p.LoadEnvironment())
	assert.NoError(t, p.LoadSecrets())
	assert.NoError(t, p.UnmarshalToStruct())

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.ValidateKeys()
	require.Len(t, p.Errors(), 2)
	assert.Len(t, p.Warnings(), 0)

	assert.EqualError(t, p.Errors()[0], "config key not expected: loggy_file")
	assert.EqualError(t, p.Errors()[1], "invalid configuration key 'logs_level' was replaced by 'log.level'")

	p.ValidateConfiguration()
	assert.Len(t, p.Errors(), 2)
	assert.Len(t, p.Warnings(), 0)
}

func TestShouldGenerateConfiguration(t *testing.T) {
	resetTestEnv()

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
	assert.EqualError(t, p.Errors()[0], fmt.Sprintf("configuration file did not exist a default one has been generated at %s", cfg))
}

func TestShouldNotGenerateMultipleConfigurations(t *testing.T) {
	resetTestEnv()

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
	resetTestEnv()

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
	resetTestEnv()

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
	resetTestEnv()

	p := GetProvider()

	assert.NoError(t, os.Setenv("AUTHELIA_SESSION_SECRET", "xyz"))

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
