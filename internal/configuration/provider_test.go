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

	p := GetProvider()

	err = p.LoadSources(NewEnvironmentSource(), NewSecretsSource(p))
	assert.EqualError(t, err, "errors occurred during loading configuration sources")

	require.Len(t, p.Errors(), 12)

	errs := make([]string, 0, 12)
	for _, err := range p.Errors() {
		errs = append(errs, err.Error())
	}

	sort.Strings(errs)

	errFmt := utils.GetExpectedErrTxt("filenotfound")

	assert.Equal(t, fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "authentication"), "authentication_backend.ldap.password", fmt.Sprintf(errFmt, filepath.Join(dir, "authentication"))), errs[0])
	assert.Equal(t, fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "duo"), "duo_api.secret_key", fmt.Sprintf(errFmt, filepath.Join(dir, "duo"))), errs[1])
	assert.Equal(t, fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "jwt"), "jwt_secret", fmt.Sprintf(errFmt, filepath.Join(dir, "jwt"))), errs[2])
	assert.Equal(t, fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "mysql"), "storage.mysql.password", fmt.Sprintf(errFmt, filepath.Join(dir, "mysql"))), errs[3])
	assert.Equal(t, fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "notifier"), "notifier.smtp.password", fmt.Sprintf(errFmt, filepath.Join(dir, "notifier"))), errs[4])
	assert.Equal(t, fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "oidc-hmac"), "identity_providers.oidc.hmac_secret", fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-hmac"))), errs[5])
	assert.Equal(t, fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "oidc-key"), "identity_providers.oidc.issuer_private_key", fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-key"))), errs[6])
	assert.Equal(t, fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "postgres"), "storage.postgres.password", fmt.Sprintf(errFmt, filepath.Join(dir, "postgres"))), errs[7])
	assert.Equal(t, fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "redis"), "session.redis.password", fmt.Sprintf(errFmt, filepath.Join(dir, "redis"))), errs[8])
	assert.Equal(t, fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "redis-sentinel"), "session.redis.high_availability.sentinel_password", fmt.Sprintf(errFmt, filepath.Join(dir, "redis-sentinel"))), errs[9])
	assert.Equal(t, fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "session"), "session.secret", fmt.Sprintf(errFmt, filepath.Join(dir, "session"))), errs[10])
	assert.Equal(t, fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "tls"), "tls_key", fmt.Sprintf(errFmt, filepath.Join(dir, "tls"))), errs[11])
}

func TestShouldHaveNotifier(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	p := GetProvider()

	assert.NoError(t, p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"}, p)...))

	err := p.UnmarshalToConfiguration()
	assert.NoError(t, err)

	assert.NotNil(t, GetProvider().Configuration().Notifier)
}

func TestShouldValidateConfigurationWithEnv(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	p := GetProvider()

	assert.NoError(t, p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"}, p)...))

	err := p.UnmarshalToConfiguration()
	assert.NoError(t, err)

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.Validate()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)
}

func TestShouldNotIgnoreInvalidEnvs(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL", "a bad env"))
	assert.NoError(t, os.Setenv(envPrefix+"JWT_SECRET", "an env jwt secret"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))
	assert.NoError(t, os.Setenv(envPrefixAlt+"AUTHENTICATION_BACKEND_LDAP_URL", "an env authentication backend ldap password"))

	p := GetProvider()

	assert.NoError(t, p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"}, p)...))

	err := p.UnmarshalToConfiguration()
	assert.NoError(t, err)

	p.Validate()

	assert.Len(t, p.Warnings(), 0)
	require.Len(t, p.Errors(), 1)

	assert.EqualError(t, p.Errors()[0], "configuration environment variable not expected: AUTHELIA__STORAGE_MYSQL")
}

func TestShouldIgnoreSingleUnderscoreNonSecretEnvs(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(envPrefix+"JWT_SECRET", "an env jwt secret"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))
	assert.NoError(t, os.Setenv(envPrefixAlt+"AUTHENTICATION_BACKEND_LDAP_URL", "an env authentication backend ldap password"))

	p := GetProvider()

	assert.NoError(t, p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"}, p)...))

	err := p.UnmarshalToConfiguration()
	assert.NoError(t, err)

	p.Validate()

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.Equal(t, "an env jwt secret", p.Configuration().JWTSecret)
	assert.Equal(t, "an env session secret", p.Configuration().Session.Secret)
	assert.Equal(t, "an env storage mysql password", p.Configuration().Storage.MySQL.Password)
	assert.Equal(t, "an env authentication backend ldap password", p.Configuration().AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "ldap://127.0.0.1", p.Configuration().AuthenticationBackend.LDAP.URL)
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

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET_FILE", sessionSecret))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD_FILE", storageSecret))
	assert.NoError(t, os.Setenv(envPrefixAlt+"JWT_SECRET_FILE", jwtSecret))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", ldapSecret))

	p := GetProvider()

	assert.NoError(t, p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"}, p)...))

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	err = p.UnmarshalToConfiguration()
	assert.NoError(t, err)

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.Validate()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.Equal(t, "a secret jwt_secret value", p.Configuration().JWTSecret)
	assert.Equal(t, "a secret session.secret value", p.Configuration().Session.Secret)
	assert.Equal(t, "a secret storage.mysql.password value", p.Configuration().Storage.MySQL.Password)
	assert.Equal(t, "a secret authentication_backend.ldap.password value", p.Configuration().AuthenticationBackend.LDAP.Password)
}

func TestShouldValidateAndRaiseErrorsOnNormalConfigurationAndSecret(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(envPrefixAlt+"JWT_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))

	p := GetProvider()

	assert.EqualError(t, p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"}, p)...), "errors occurred during loading configuration sources")

	require.Len(t, p.Errors(), 1)
	assert.Len(t, p.Warnings(), 0)
	assert.EqualError(t, p.Errors()[0], "error loading secret into key 'session.secret': it's already defined in the config files")

	p.Clear()

	err := p.UnmarshalToConfiguration()
	assert.NoError(t, err)

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.Validate()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.Equal(t, "example_secret value", p.Configuration().JWTSecret)
	assert.Equal(t, "an env session secret", p.Configuration().Session.Secret)
	assert.Equal(t, "an env storage mysql password", p.Configuration().Storage.MySQL.Password)
	assert.Equal(t, "an env authentication backend ldap password", p.Configuration().AuthenticationBackend.LDAP.Password)
}

func TestShouldRaiseIOErrOnUnreadableFile(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping test due to being on windows")
	}

	testReset()

	dir, err := ioutil.TempDir("", "authelia-conf")
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "myconf.yml"), []byte("server:\n  port: 9091\n"), 0000))

	p := GetProvider()

	cfg := filepath.Join(dir, "myconf.yml")
	assert.EqualError(t, p.LoadSources(NewYAMLFileSource(cfg)), fmt.Sprintf("open %s: permission denied", cfg))
}

func TestShouldValidateConfigurationWithEnvSecrets(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(envPrefix+"JWT_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", "./test_resources/example_secret"))

	p := GetProvider()

	assert.NoError(t, p.LoadSources(NewDefaultSources([]string{"./test_resources/config.yml"}, p)...))

	err := p.UnmarshalToConfiguration()
	assert.NoError(t, err)

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.Equal(t, "example_secret value", p.Configuration().JWTSecret)
	assert.Equal(t, "example_secret value", p.Configuration().Session.Secret)
	assert.Equal(t, "example_secret value", p.Configuration().AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "example_secret value", p.Configuration().Storage.MySQL.Password)

	p.Validate()
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)
}

func TestShouldValidateAndRaiseErrorsOnBadConfiguration(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(envPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	p := GetProvider()

	assert.NoError(t, p.LoadSources(NewDefaultSources([]string{"./test_resources/config_bad_keys.yml"}, p)...))

	err := p.UnmarshalToConfiguration()
	assert.NoError(t, err)

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	p.Validate()
	assert.Len(t, p.Errors(), 2)
	assert.Len(t, p.Warnings(), 0)

	assert.EqualError(t, p.Errors()[0], "configuration key not expected: loggy_file")
	assert.EqualError(t, p.Errors()[1], "invalid configuration key 'logs_level' was replaced by 'log.level'")
}

func TestShouldNotReadConfigurationOnFSAccessDenied(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping test due to being on windows")
	}

	testReset()

	p := GetProvider()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	assert.NoError(t, testCreateFile(filepath.Join(dir, "config.yml"), "port: 9091\n", 0000))

	cfg := filepath.Join(dir, "config.yml")

	assert.EqualError(t, p.LoadSources(NewYAMLFileSource(cfg)), fmt.Sprintf("open %s: permission denied", cfg))
}

func TestShouldNotLoadDirectoryConfiguration(t *testing.T) {
	testReset()

	p := GetProvider()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	expectedErr := fmt.Sprintf(utils.GetExpectedErrTxt("yamlisdir"), dir)
	assert.EqualError(t, p.LoadSources(NewYAMLFileSource(dir)), expectedErr)
}

func TestShouldRetrieveGlobalConfiguration(t *testing.T) {
	testReset()

	p := GetProvider()

	assert.NoError(t, os.Setenv(envPrefix+"SESSION_SECRET", "xyz"))

	assert.NoError(t, p.LoadSources(NewEnvironmentSource()))
	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	err := p.UnmarshalToConfiguration()
	assert.NoError(t, err)

	assert.Len(t, p.Errors(), 0)
	assert.Len(t, p.Warnings(), 0)

	assert.Equal(t, "xyz", p.Configuration().Session.Secret)

	q := GetProvider()
	assert.Equal(t, "xyz", q.Configuration().Session.Secret)
}

func testReset() {
	provider = nil

	_ = os.Unsetenv(envPrefix + "STORAGE_MYSQL")
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
