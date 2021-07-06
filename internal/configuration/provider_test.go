package configuration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/authelia/authelia/internal/utils"
)

func TestShouldErrorSecretNotExist(t *testing.T) {
	testReset()

	dir, err := ioutil.TempDir("", "authelia-test-secret-not-exist")
	assert.NoError(t, err)

	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET_FILE", filepath.Join(dir, "jwt")))
	assert.NoError(t, os.Setenv(constEnvPrefix+"DUO_API_SECRET_KEY_FILE", filepath.Join(dir, "duo")))
	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET_FILE", filepath.Join(dir, "session")))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", filepath.Join(dir, "authentication")))
	assert.NoError(t, os.Setenv(constEnvPrefix+"NOTIFIER_SMTP_PASSWORD_FILE", filepath.Join(dir, "notifier")))
	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_REDIS_PASSWORD_FILE", filepath.Join(dir, "redis")))
	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE", filepath.Join(dir, "redis-sentinel")))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD_FILE", filepath.Join(dir, "mysql")))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_POSTGRES_PASSWORD_FILE", filepath.Join(dir, "postgres")))
	assert.NoError(t, os.Setenv(constEnvPrefix+"TLS_KEY_FILE", filepath.Join(dir, "tls")))
	assert.NoError(t, os.Setenv(constEnvPrefix+"IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE", filepath.Join(dir, "oidc-key")))
	assert.NoError(t, os.Setenv(constEnvPrefix+"IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE", filepath.Join(dir, "oidc-hmac")))

	val := schema.NewStructValidator()
	_, _ = Load(val, NewEnvironmentSource(), NewSecretsSource())

	assert.Len(t, val.Warnings(), 0)

	errs := val.Errors()
	require.Len(t, errs, 12)

	sort.Sort(utils.ErrSliceSortAlphabetical(errs))

	errFmt := utils.GetExpectedErrTxt("filenotfound")

	// ignore the errors before this as they are checked by the valdator.
	assert.EqualError(t, errs[0], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "authentication"), "authentication_backend.ldap.password", fmt.Sprintf(errFmt, filepath.Join(dir, "authentication"))))
	assert.EqualError(t, errs[1], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "duo"), "duo_api.secret_key", fmt.Sprintf(errFmt, filepath.Join(dir, "duo"))))
	assert.EqualError(t, errs[2], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "jwt"), "jwt_secret", fmt.Sprintf(errFmt, filepath.Join(dir, "jwt"))))
	assert.EqualError(t, errs[3], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "mysql"), "storage.mysql.password", fmt.Sprintf(errFmt, filepath.Join(dir, "mysql"))))
	assert.EqualError(t, errs[4], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "notifier"), "notifier.smtp.password", fmt.Sprintf(errFmt, filepath.Join(dir, "notifier"))))
	assert.EqualError(t, errs[5], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "oidc-hmac"), "identity_providers.oidc.hmac_secret", fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-hmac"))))
	assert.EqualError(t, errs[6], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "oidc-key"), "identity_providers.oidc.issuer_private_key", fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-key"))))
	assert.EqualError(t, errs[7], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "postgres"), "storage.postgres.password", fmt.Sprintf(errFmt, filepath.Join(dir, "postgres"))))
	assert.EqualError(t, errs[8], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "redis"), "session.redis.password", fmt.Sprintf(errFmt, filepath.Join(dir, "redis"))))
	assert.EqualError(t, errs[9], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "redis-sentinel"), "session.redis.high_availability.sentinel_password", fmt.Sprintf(errFmt, filepath.Join(dir, "redis-sentinel"))))
	assert.EqualError(t, errs[10], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "session"), "session.secret", fmt.Sprintf(errFmt, filepath.Join(dir, "session"))))
	assert.EqualError(t, errs[11], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "tls"), "tls_key", fmt.Sprintf(errFmt, filepath.Join(dir, "tls"))))
}

func TestShouldHaveNotifier(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	val := schema.NewStructValidator()
	_, config := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"})...)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
	assert.NotNil(t, config.Notifier)
}

func TestShouldValidateConfigurationWithEnv(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	val := schema.NewStructValidator()
	_, _ = Load(val, NewDefaultSources([]string{"./test_resources/config.yml"})...)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
}

func TestShouldNotIgnoreInvalidEnvs(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL", "a bad env"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET", "an env jwt secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))
	assert.NoError(t, os.Setenv(constEnvPrefixAlt+"AUTHENTICATION_BACKEND_LDAP_URL", "an env authentication backend ldap password"))

	val := schema.NewStructValidator()
	keys, _ := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"})...)

	validator.ValidateKeys(keys, val)

	assert.Len(t, val.Warnings(), 0)
	require.Len(t, val.Errors(), 1)

	assert.EqualError(t, val.Errors()[0], "configuration environment variable not expected: AUTHELIA__STORAGE_MYSQL")
}

func TestShouldIgnoreSingleUnderscoreNonSecretEnvs(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET", "an env jwt secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))
	assert.NoError(t, os.Setenv(constEnvPrefixAlt+"AUTHENTICATION_BACKEND_LDAP_URL", "an env authentication backend ldap password"))

	val := schema.NewStructValidator()
	_, config := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"})...)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "an env jwt secret", config.JWTSecret)
	assert.Equal(t, "an env session secret", config.Session.Secret)
	assert.Equal(t, "an env storage mysql password", config.Storage.MySQL.Password)
	assert.Equal(t, "an env authentication backend ldap password", config.AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "ldap://127.0.0.1", config.AuthenticationBackend.LDAP.URL)
}

func TestShouldAllowBothLegacyEnvSecretFilesAndNewOnes(t *testing.T) {
	testReset()

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

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET_FILE", sessionSecret))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD_FILE", storageSecret))
	assert.NoError(t, os.Setenv(constEnvPrefixAlt+"JWT_SECRET_FILE", jwtSecret))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", ldapSecret))

	val := schema.NewStructValidator()
	_, config := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"})...)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "a secret jwt_secret value", config.JWTSecret)
	assert.Equal(t, "a secret session.secret value", config.Session.Secret)
	assert.Equal(t, "a secret storage.mysql.password value", config.Storage.MySQL.Password)
	assert.Equal(t, "a secret authentication_backend.ldap.password value", config.AuthenticationBackend.LDAP.Password)
}

func TestShouldValidateAndRaiseErrorsOnNormalConfigurationAndSecret(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(constEnvPrefixAlt+"JWT_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))

	val := schema.NewStructValidator()
	_, config := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"})...)

	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	assert.EqualError(t, val.Errors()[0], "secrets: error loading secret into key 'session.secret': it's already defined in other configuration sources")

	assert.Equal(t, "example_secret value", config.JWTSecret)
	assert.Equal(t, "example_secret value", config.Session.Secret)
	assert.Equal(t, "an env storage mysql password", config.Storage.MySQL.Password)
	assert.Equal(t, "an env authentication backend ldap password", config.AuthenticationBackend.LDAP.Password)
}

func TestShouldRaiseIOErrOnUnreadableFile(t *testing.T) {
	if runtime.GOOS == constWindows {
		t.Skip("skipping test due to being on windows")
	}

	testReset()

	dir, err := ioutil.TempDir("", "authelia-conf")
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "myconf.yml"), []byte("server:\n  port: 9091\n"), 0000))

	cfg := filepath.Join(dir, "myconf.yml")

	val := schema.NewStructValidator()
	_, _ = Load(val, NewYAMLFileSource(cfg))

	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)
	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("failed to load configuration from yaml file(%s) source: open %s: permission denied", cfg, cfg))
}

func TestShouldValidateConfigurationWithEnvSecrets(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", "./test_resources/example_secret"))

	val := schema.NewStructValidator()
	_, config := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"})...)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "example_secret value", config.JWTSecret)
	assert.Equal(t, "example_secret value", config.Session.Secret)
	assert.Equal(t, "example_secret value", config.AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "example_secret value", config.Storage.MySQL.Password)
}

func TestShouldValidateAndRaiseErrorsOnBadConfiguration(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	val := schema.NewStructValidator()
	keys, _ := Load(val, NewDefaultSources([]string{"./test_resources/config_bad_keys.yml"})...)

	validator.ValidateKeys(keys, val)

	require.Len(t, val.Errors(), 2)
	assert.Len(t, val.Warnings(), 0)

	assert.EqualError(t, val.Errors()[0], "configuration key not expected: loggy_file")
	assert.EqualError(t, val.Errors()[1], "invalid configuration key 'logs_level' was replaced by 'log.level'")
}

func TestShouldNotReadConfigurationOnFSAccessDenied(t *testing.T) {
	if runtime.GOOS == constWindows {
		t.Skip("skipping test due to being on windows")
	}

	testReset()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	cfg := filepath.Join(dir, "config.yml")
	assert.NoError(t, testCreateFile(filepath.Join(dir, "config.yml"), "port: 9091\n", 0000))

	val := schema.NewStructValidator()
	_, _ = Load(val, NewYAMLFileSource(cfg))

	require.Len(t, val.Errors(), 1)

	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("failed to load configuration from yaml file(%s) source: open %s: permission denied", cfg, cfg))
}

func TestShouldNotLoadDirectoryConfiguration(t *testing.T) {
	testReset()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	val := schema.NewStructValidator()
	_, _ = Load(val, NewYAMLFileSource(dir))

	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	expectedErr := fmt.Sprintf(utils.GetExpectedErrTxt("yamlisdir"), dir)
	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("failed to load configuration from yaml file(%s) source: %s", dir, expectedErr))
}

func testReset() {
	_ = os.Unsetenv(constEnvPrefix + "STORAGE_MYSQL")
	_ = os.Unsetenv(constEnvPrefixAlt + "JWT_SECRET_FILE")
	_ = os.Unsetenv(constEnvPrefixAlt + "JWT_SECRET")
	_ = os.Unsetenv(constEnvPrefix + "JWT_SECRET_FILE")
	_ = os.Unsetenv(constEnvPrefix + "DUO_API_SECRET_KEY_FILE")
	_ = os.Unsetenv(constEnvPrefix + "SESSION_SECRET_FILE")
	_ = os.Unsetenv(constEnvPrefix + "AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE")
	_ = os.Unsetenv(constEnvPrefix + "NOTIFIER_SMTP_PASSWORD_FILE")
	_ = os.Unsetenv(constEnvPrefix + "SESSION_REDIS_PASSWORD_FILE")
	_ = os.Unsetenv(constEnvPrefix + "SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE")
	_ = os.Unsetenv(constEnvPrefix + "STORAGE_MYSQL_PASSWORD_FILE")
	_ = os.Unsetenv(constEnvPrefix + "STORAGE_POSTGRES_PASSWORD_FILE")
	_ = os.Unsetenv(constEnvPrefix + "TLS_KEY_FILE")
	_ = os.Unsetenv(constEnvPrefix + "IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE")
	_ = os.Unsetenv(constEnvPrefix + "IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE")
	_ = os.Unsetenv(constEnvPrefix + "JWT_SECRET")
	_ = os.Unsetenv(constEnvPrefix + "DUO_API_SECRET_KEY")
	_ = os.Unsetenv(constEnvPrefix + "SESSION_SECRET")
	_ = os.Unsetenv(constEnvPrefix + "AUTHENTICATION_BACKEND_LDAP_PASSWORD")
	_ = os.Unsetenv(constEnvPrefix + "NOTIFIER_SMTP_PASSWORD")
	_ = os.Unsetenv(constEnvPrefix + "SESSION_REDIS_PASSWORD")
	_ = os.Unsetenv(constEnvPrefix + "SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD")
	_ = os.Unsetenv(constEnvPrefix + "STORAGE_MYSQL_PASSWORD")
	_ = os.Unsetenv(constEnvPrefix + "STORAGE_POSTGRES_PASSWORD")
	_ = os.Unsetenv(constEnvPrefix + "TLS_KEY")
	_ = os.Unsetenv(constEnvPrefix + "PORT")
	_ = os.Unsetenv(constEnvPrefix + "PORT_K8S_EXAMPLE")
	_ = os.Unsetenv(constEnvPrefix + "IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY")
	_ = os.Unsetenv(constEnvPrefix + "IDENTITY_PROVIDERS_OIDC_HMAC_SECRET")
}

func testCreateFile(path, value string, perm os.FileMode) (err error) {
	return os.WriteFile(path, []byte(value), perm)
}
