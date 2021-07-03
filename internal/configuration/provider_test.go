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

	p := GetProvider()

	loadErrs := p.LoadSources(NewEnvironmentSource(), NewSecretsSource())

	require.Len(t, loadErrs, 12)

	errs := make([]string, 0, 12)

	for _, err := range loadErrs {
		errs = append(errs, err.Error())
	}

	sort.Strings(errs)

	errFmt := utils.GetExpectedErrTxt("filenotfound")

	assert.Equal(t, "secrets: "+fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "authentication"), "authentication_backend.ldap.password", fmt.Sprintf(errFmt, filepath.Join(dir, "authentication"))), errs[0])
	assert.Equal(t, "secrets: "+fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "duo"), "duo_api.secret_key", fmt.Sprintf(errFmt, filepath.Join(dir, "duo"))), errs[1])
	assert.Equal(t, "secrets: "+fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "jwt"), "jwt_secret", fmt.Sprintf(errFmt, filepath.Join(dir, "jwt"))), errs[2])
	assert.Equal(t, "secrets: "+fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "mysql"), "storage.mysql.password", fmt.Sprintf(errFmt, filepath.Join(dir, "mysql"))), errs[3])
	assert.Equal(t, "secrets: "+fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "notifier"), "notifier.smtp.password", fmt.Sprintf(errFmt, filepath.Join(dir, "notifier"))), errs[4])
	assert.Equal(t, "secrets: "+fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "oidc-hmac"), "identity_providers.oidc.hmac_secret", fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-hmac"))), errs[5])
	assert.Equal(t, "secrets: "+fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "oidc-key"), "identity_providers.oidc.issuer_private_key", fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-key"))), errs[6])
	assert.Equal(t, "secrets: "+fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "postgres"), "storage.postgres.password", fmt.Sprintf(errFmt, filepath.Join(dir, "postgres"))), errs[7])
	assert.Equal(t, "secrets: "+fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "redis"), "session.redis.password", fmt.Sprintf(errFmt, filepath.Join(dir, "redis"))), errs[8])
	assert.Equal(t, "secrets: "+fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "redis-sentinel"), "session.redis.high_availability.sentinel_password", fmt.Sprintf(errFmt, filepath.Join(dir, "redis-sentinel"))), errs[9])
	assert.Equal(t, "secrets: "+fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "session"), "session.secret", fmt.Sprintf(errFmt, filepath.Join(dir, "session"))), errs[10])
	assert.Equal(t, "secrets: "+fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "tls"), "tls_key", fmt.Sprintf(errFmt, filepath.Join(dir, "tls"))), errs[11])
}

func TestShouldHaveNotifier(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	p := GetProvider()

	errs := p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"})...)

	assert.Len(t, errs, 0)

	_, _ = p.Unmarshal()

	assert.NotNil(t, GetProvider().Configuration.Notifier)
}

func TestShouldValidateConfigurationWithEnv(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	p := GetProvider()

	errs := p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"})...)
	assert.Len(t, errs, 0)

	warns, errs := p.Unmarshal()

	assert.Len(t, warns, 0)
	assert.Len(t, errs, 0)
}

func TestShouldNotIgnoreInvalidEnvs(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL", "a bad env"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET", "an env jwt secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))
	assert.NoError(t, os.Setenv(constEnvPrefixAlt+"AUTHENTICATION_BACKEND_LDAP_URL", "an env authentication backend ldap password"))

	p := GetProvider()

	errs := p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"})...)
	assert.Len(t, errs, 0)

	warns, errs := p.Unmarshal()

	assert.Len(t, warns, 0)
	require.Len(t, errs, 1)

	assert.EqualError(t, p.Validator.Errors()[0], "configuration environment variable not expected: AUTHELIA__STORAGE_MYSQL")
}

func TestShouldIgnoreSingleUnderscoreNonSecretEnvs(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET", "an env jwt secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))
	assert.NoError(t, os.Setenv(constEnvPrefixAlt+"AUTHENTICATION_BACKEND_LDAP_URL", "an env authentication backend ldap password"))

	p := GetProvider()

	errs := p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"})...)
	assert.Len(t, errs, 0)

	warns, errs := p.Unmarshal()

	assert.Len(t, errs, 0)
	assert.Len(t, warns, 0)

	assert.Equal(t, "an env jwt secret", p.Configuration.JWTSecret)
	assert.Equal(t, "an env session secret", p.Configuration.Session.Secret)
	assert.Equal(t, "an env storage mysql password", p.Configuration.Storage.MySQL.Password)
	assert.Equal(t, "an env authentication backend ldap password", p.Configuration.AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "ldap://127.0.0.1", p.Configuration.AuthenticationBackend.LDAP.URL)
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

	p := GetProvider()

	errs := p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"})...)
	assert.Len(t, errs, 0)

	assert.Len(t, p.Validator.Errors(), 0)
	assert.Len(t, p.Validator.Warnings(), 0)

	warns, errs := p.Unmarshal()

	assert.Len(t, errs, 0)
	assert.Len(t, warns, 0)

	assert.Equal(t, "a secret jwt_secret value", p.Configuration.JWTSecret)
	assert.Equal(t, "a secret session.secret value", p.Configuration.Session.Secret)
	assert.Equal(t, "a secret storage.mysql.password value", p.Configuration.Storage.MySQL.Password)
	assert.Equal(t, "a secret authentication_backend.ldap.password value", p.Configuration.AuthenticationBackend.LDAP.Password)
}

func TestShouldValidateAndRaiseErrorsOnNormalConfigurationAndSecret(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(constEnvPrefixAlt+"JWT_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))

	p := GetProvider()

	errs := p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"})...)

	require.Len(t, errs, 1)
	assert.Len(t, p.Validator.Warnings(), 0)
	assert.EqualError(t, p.Validator.Errors()[0], "secrets: error loading secret into key 'session.secret': it's already defined in other configuration sources")

	p.Validator.Clear()

	warns, errs := p.Unmarshal()

	assert.Len(t, errs, 0)
	assert.Len(t, warns, 0)

	assert.Equal(t, "example_secret value", p.Configuration.JWTSecret)
	assert.Equal(t, "example_secret value", p.Configuration.Session.Secret)
	assert.Equal(t, "an env storage mysql password", p.Configuration.Storage.MySQL.Password)
	assert.Equal(t, "an env authentication backend ldap password", p.Configuration.AuthenticationBackend.LDAP.Password)
}

func TestShouldRaiseIOErrOnUnreadableFile(t *testing.T) {
	if runtime.GOOS == constWindows {
		t.Skip("skipping test due to being on windows")
	}

	testReset()

	dir, err := ioutil.TempDir("", "authelia-conf")
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "myconf.yml"), []byte("server:\n  port: 9091\n"), 0000))

	p := GetProvider()

	cfg := filepath.Join(dir, "myconf.yml")

	errs := p.LoadSources(NewYAMLFileSource(cfg))

	require.Len(t, errs, 1)

	assert.EqualError(t, errs[0], fmt.Sprintf("failed to load configuration from yaml file(%s) source: open %s: permission denied", cfg, cfg))
}

func TestShouldValidateConfigurationWithEnvSecrets(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", "./test_resources/example_secret"))

	p := GetProvider()

	errs := p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"})...)
	assert.Len(t, errs, 0)

	warns, errs := p.Unmarshal()

	assert.Len(t, warns, 0)
	assert.Len(t, errs, 0)

	assert.Equal(t, "example_secret value", p.Configuration.JWTSecret)
	assert.Equal(t, "example_secret value", p.Configuration.Session.Secret)
	assert.Equal(t, "example_secret value", p.Configuration.AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "example_secret value", p.Configuration.Storage.MySQL.Password)
}

func TestShouldValidateAndRaiseErrorsOnBadConfiguration(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(constEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	p := GetProvider()

	errs := p.LoadSources(NewDefaultSources([]string{"./test_resources/config_bad_keys.yml"})...)

	require.Len(t, errs, 0)

	warns, errs := p.Unmarshal()

	require.Len(t, errs, 2)
	assert.Len(t, warns, 0)

	assert.EqualError(t, p.Validator.Errors()[0], "configuration key not expected: loggy_file")
	assert.EqualError(t, p.Validator.Errors()[1], "invalid configuration key 'logs_level' was replaced by 'log.level'")
}

func TestShouldNotReadConfigurationOnFSAccessDenied(t *testing.T) {
	if runtime.GOOS == constWindows {
		t.Skip("skipping test due to being on windows")
	}

	testReset()

	p := GetProvider()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	assert.NoError(t, testCreateFile(filepath.Join(dir, "config.yml"), "port: 9091\n", 0000))

	cfg := filepath.Join(dir, "config.yml")

	errs := p.LoadSources(NewYAMLFileSource(cfg))

	require.Len(t, errs, 1)

	assert.EqualError(t, errs[0], fmt.Sprintf("failed to load configuration from yaml file(%s) source: open %s: permission denied", cfg, cfg))
}

func TestShouldNotLoadDirectoryConfiguration(t *testing.T) {
	testReset()

	p := GetProvider()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	errs := p.LoadSources(NewYAMLFileSource(dir))
	require.Len(t, errs, 1)

	expectedErr := fmt.Sprintf(utils.GetExpectedErrTxt("yamlisdir"), dir)
	assert.EqualError(t, errs[0], fmt.Sprintf("failed to load configuration from yaml file(%s) source: %s", dir, expectedErr))
}

func TestShouldRetrieveGlobalConfiguration(t *testing.T) {
	testReset()

	p := GetProvider()

	assert.NoError(t, os.Setenv(constEnvPrefix+"SESSION_SECRET", "xyz"))

	errs := p.LoadSources(NewEnvironmentSource())
	assert.Len(t, errs, 0)

	assert.Len(t, p.Validator.Errors(), 0)
	assert.Len(t, p.Validator.Warnings(), 0)

	_, _ = p.Unmarshal()

	assert.Equal(t, "xyz", p.Configuration.Session.Secret)

	q := GetProvider()
	assert.Equal(t, "xyz", q.Configuration.Session.Secret)

	assert.Equal(t, p, q)
}

func testReset() {
	provider = nil

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
