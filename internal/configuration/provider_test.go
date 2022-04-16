package configuration

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShouldErrorSecretNotExist(t *testing.T) {
	testReset()

	dir, err := os.MkdirTemp("", "authelia-test-secret-not-exist")
	assert.NoError(t, err)

	testSetEnv(t, "JWT_SECRET_FILE", filepath.Join(dir, "jwt"))
	testSetEnv(t, "DUO_API_SECRET_KEY_FILE", filepath.Join(dir, "duo"))
	testSetEnv(t, "SESSION_SECRET_FILE", filepath.Join(dir, "session"))
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", filepath.Join(dir, "authentication"))
	testSetEnv(t, "NOTIFIER_SMTP_PASSWORD_FILE", filepath.Join(dir, "notifier"))
	testSetEnv(t, "SESSION_REDIS_PASSWORD_FILE", filepath.Join(dir, "redis"))
	testSetEnv(t, "SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE", filepath.Join(dir, "redis-sentinel"))
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD_FILE", filepath.Join(dir, "mysql"))
	testSetEnv(t, "STORAGE_POSTGRES_PASSWORD_FILE", filepath.Join(dir, "postgres"))
	testSetEnv(t, "SERVER_TLS_KEY_FILE", filepath.Join(dir, "tls"))
	testSetEnv(t, "IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE", filepath.Join(dir, "oidc-key"))
	testSetEnv(t, "IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE", filepath.Join(dir, "oidc-hmac"))

	val := schema.NewStructValidator()
	_, _, err = Load(val, NewEnvironmentSource(DefaultEnvPrefix, DefaultEnvDelimiter), NewSecretsSource(DefaultEnvPrefix, DefaultEnvDelimiter))

	assert.NoError(t, err)
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
	assert.EqualError(t, errs[11], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "tls"), "server.tls.key", fmt.Sprintf(errFmt, filepath.Join(dir, "tls"))))
}

func TestLoadShouldReturnErrWithoutValidator(t *testing.T) {
	_, _, err := Load(nil, NewEnvironmentSource(DefaultEnvPrefix, DefaultEnvDelimiter))
	assert.EqualError(t, err, "no validator provided")
}

func TestLoadShouldReturnErrWithoutSources(t *testing.T) {
	_, _, err := Load(schema.NewStructValidator())
	assert.EqualError(t, err, "no sources provided")
}

func TestShouldHaveNotifier(t *testing.T) {
	testReset()

	testSetEnv(t, "SESSION_SECRET", "abc")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "abc")
	testSetEnv(t, "JWT_SECRET", "abc")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
	assert.NotNil(t, config.Notifier)
}

func TestShouldValidateConfigurationWithEnv(t *testing.T) {
	testReset()

	testSetEnv(t, "SESSION_SECRET", "abc")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "abc")
	testSetEnv(t, "JWT_SECRET", "abc")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

	val := schema.NewStructValidator()
	_, _, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
}

func TestShouldNotIgnoreInvalidEnvs(t *testing.T) {
	testReset()

	testSetEnv(t, "SESSION_SECRET", "an env session secret")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "an env storage mysql password")
	testSetEnv(t, "STORAGE_MYSQL", "a bad env")
	testSetEnv(t, "JWT_SECRET", "an env jwt secret")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_URL", "an env authentication backend ldap password")

	val := schema.NewStructValidator()
	keys, _, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	require.Len(t, val.Warnings(), 1)
	assert.Len(t, val.Errors(), 0)

	assert.EqualError(t, val.Warnings()[0], fmt.Sprintf("configuration environment variable not expected: %sSTORAGE_MYSQL", DefaultEnvPrefix))
}

func TestShouldValidateAndRaiseErrorsOnNormalConfigurationAndSecret(t *testing.T) {
	testReset()

	testSetEnv(t, "SESSION_SECRET", "an env session secret")
	testSetEnv(t, "SESSION_SECRET_FILE", "./test_resources/example_secret")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "an env storage mysql password")
	testSetEnv(t, "JWT_SECRET_FILE", "./test_resources/example_secret")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password")
	testSetEnv(t, "STORAGE_ENCRYPTION_KEY", "a_very_bad_encryption_key")

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	assert.EqualError(t, val.Errors()[0], "secrets: error loading secret into key 'session.secret': it's already defined in other configuration sources")

	assert.Equal(t, "example_secret value", config.JWTSecret)
	assert.Equal(t, "example_secret value", config.Session.Secret)
	assert.Equal(t, "an env storage mysql password", config.Storage.MySQL.Password)
	assert.Equal(t, "an env authentication backend ldap password", config.AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "a_very_bad_encryption_key", config.Storage.EncryptionKey)
}

func TestShouldRaiseIOErrOnUnreadableFile(t *testing.T) {
	if runtime.GOOS == constWindows {
		t.Skip("skipping test due to being on windows")
	}

	testReset()

	dir, err := os.MkdirTemp("", "authelia-conf")
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "myconf.yml"), []byte("server:\n  port: 9091\n"), 0000))

	cfg := filepath.Join(dir, "myconf.yml")

	val := schema.NewStructValidator()
	_, _, err = Load(val, NewYAMLFileSource(cfg))

	assert.NoError(t, err)
	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)
	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("failed to load configuration from yaml file(%s) source: open %s: permission denied", cfg, cfg))
}

func TestShouldValidateConfigurationWithEnvSecrets(t *testing.T) {
	testReset()

	testSetEnv(t, "SESSION_SECRET_FILE", "./test_resources/example_secret")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD_FILE", "./test_resources/example_secret")
	testSetEnv(t, "JWT_SECRET_FILE", "./test_resources/example_secret")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", "./test_resources/example_secret")
	testSetEnv(t, "STORAGE_ENCRYPTION_KEY_FILE", "./test_resources/example_secret")

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "example_secret value", config.JWTSecret)
	assert.Equal(t, "example_secret value", config.Session.Secret)
	assert.Equal(t, "example_secret value", config.AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "example_secret value", config.Storage.MySQL.Password)
	assert.Equal(t, "example_secret value", config.Storage.EncryptionKey)
}

func TestShouldLoadURLList(t *testing.T) {
	testReset()

	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_oidc.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	require.Len(t, config.IdentityProviders.OIDC.CORS.AllowedOrigins, 2)
	assert.Equal(t, "https://google.com", config.IdentityProviders.OIDC.CORS.AllowedOrigins[0].String())
	assert.Equal(t, "https://example.com", config.IdentityProviders.OIDC.CORS.AllowedOrigins[1].String())
}

func TestShouldValidateAndRaiseErrorsOnBadConfiguration(t *testing.T) {
	testReset()

	testSetEnv(t, "SESSION_SECRET", "abc")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "abc")
	testSetEnv(t, "JWT_SECRET", "abc")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

	val := schema.NewStructValidator()
	keys, c, err := Load(val, NewDefaultSources([]string{"./test_resources/config_bad_keys.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	require.Len(t, val.Errors(), 1)
	require.Len(t, val.Warnings(), 1)

	assert.EqualError(t, val.Errors()[0], "configuration key not expected: loggy_file")
	assert.EqualError(t, val.Warnings()[0], "configuration key 'logs_level' is deprecated in 4.7.0 and has been replaced by 'log.level': this has been automatically mapped for you but you will need to adjust your configuration to remove this message")

	assert.Equal(t, "debug", c.Log.Level)
}

func TestShouldRaiseErrOnInvalidNotifierSMTPSender(t *testing.T) {
	testReset()

	val := schema.NewStructValidator()
	keys, _, err := Load(val, NewDefaultSources([]string{"./test_resources/config_smtp_sender_invalid.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	assert.EqualError(t, val.Errors()[0], "error occurred during unmarshalling configuration: 1 error(s) decoding:\n\n* error decoding 'notifier.smtp.sender': could not decode 'admin' to a mail.Address (RFC5322): mail: missing '@' or angle-addr")
}

func TestShouldHandleErrInvalidatorWhenSMTPSenderBlank(t *testing.T) {
	testReset()

	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_smtp_sender_blank.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "", config.Notifier.SMTP.Sender.Name)
	assert.Equal(t, "", config.Notifier.SMTP.Sender.Address)

	validator.ValidateNotifier(&config.Notifier, val)

	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	assert.EqualError(t, val.Errors()[0], "notifier: smtp: option 'sender' is required")
}

func TestShouldDecodeSMTPSenderWithoutName(t *testing.T) {
	testReset()

	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "", config.Notifier.SMTP.Sender.Name)
	assert.Equal(t, "admin@example.com", config.Notifier.SMTP.Sender.Address)
}

func TestShouldDecodeSMTPSenderWithName(t *testing.T) {
	testReset()

	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_alt.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "Admin", config.Notifier.SMTP.Sender.Name)
	assert.Equal(t, "admin@example.com", config.Notifier.SMTP.Sender.Address)
	assert.Equal(t, schema.RememberMeDisabled, config.Session.RememberMeDuration)
}

func TestShouldParseRegex(t *testing.T) {
	testReset()

	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_domain_regex.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	validator.ValidateRules(config, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Len(t, config.AccessControl.Rules[0].DomainsRegex[0].SubexpNames(), 2)
	assert.Equal(t, "", config.AccessControl.Rules[0].DomainsRegex[0].SubexpNames()[0])
	assert.Equal(t, "", config.AccessControl.Rules[0].DomainsRegex[0].SubexpNames()[1])

	assert.Len(t, config.AccessControl.Rules[1].DomainsRegex[0].SubexpNames(), 2)
	assert.Equal(t, "", config.AccessControl.Rules[1].DomainsRegex[0].SubexpNames()[0])
	assert.Equal(t, "User", config.AccessControl.Rules[1].DomainsRegex[0].SubexpNames()[1])

	assert.Len(t, config.AccessControl.Rules[2].DomainsRegex[0].SubexpNames(), 3)
	assert.Equal(t, "", config.AccessControl.Rules[2].DomainsRegex[0].SubexpNames()[0])
	assert.Equal(t, "User", config.AccessControl.Rules[2].DomainsRegex[0].SubexpNames()[1])
	assert.Equal(t, "Group", config.AccessControl.Rules[2].DomainsRegex[0].SubexpNames()[2])
}

func TestShouldErrOnParseInvalidRegex(t *testing.T) {
	testReset()

	val := schema.NewStructValidator()
	keys, _, err := Load(val, NewDefaultSources([]string{"./test_resources/config_domain_bad_regex.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	assert.EqualError(t, val.Errors()[0], "error occurred during unmarshalling configuration: 1 error(s) decoding:\n\n* error decoding 'access_control.rules[0].domain_regex[0]': could not decode '^\\K(public|public2).example.com$' to a regexp.Regexp: error parsing regexp: invalid escape sequence: `\\K`")
}

func TestShouldNotReadConfigurationOnFSAccessDenied(t *testing.T) {
	if runtime.GOOS == constWindows {
		t.Skip("skipping test due to being on windows")
	}

	testReset()

	dir, err := os.MkdirTemp("", "authelia-config")
	assert.NoError(t, err)

	cfg := filepath.Join(dir, "config.yml")
	assert.NoError(t, testCreateFile(filepath.Join(dir, "config.yml"), "port: 9091\n", 0000))

	val := schema.NewStructValidator()
	_, _, err = Load(val, NewYAMLFileSource(cfg))

	assert.NoError(t, err)
	require.Len(t, val.Errors(), 1)

	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("failed to load configuration from yaml file(%s) source: open %s: permission denied", cfg, cfg))
}

func TestShouldNotLoadDirectoryConfiguration(t *testing.T) {
	testReset()

	dir, err := os.MkdirTemp("", "authelia-config")
	assert.NoError(t, err)

	val := schema.NewStructValidator()
	_, _, err = Load(val, NewYAMLFileSource(dir))

	assert.NoError(t, err)
	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	expectedErr := fmt.Sprintf(utils.GetExpectedErrTxt("yamlisdir"), dir)
	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("failed to load configuration from yaml file(%s) source: %s", dir, expectedErr))
}

func testSetEnv(t *testing.T, key, value string) {
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+key, value))
}

func testReset() {
	testUnsetEnvName("STORAGE_MYSQL")
	testUnsetEnvName("JWT_SECRET")
	testUnsetEnvName("DUO_API_SECRET_KEY")
	testUnsetEnvName("SESSION_SECRET")
	testUnsetEnvName("AUTHENTICATION_BACKEND_LDAP_PASSWORD")
	testUnsetEnvName("AUTHENTICATION_BACKEND_LDAP_URL")
	testUnsetEnvName("NOTIFIER_SMTP_PASSWORD")
	testUnsetEnvName("SESSION_REDIS_PASSWORD")
	testUnsetEnvName("SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD")
	testUnsetEnvName("STORAGE_MYSQL_PASSWORD")
	testUnsetEnvName("STORAGE_POSTGRES_PASSWORD")
	testUnsetEnvName("SERVER_TLS_KEY")
	testUnsetEnvName("SERVER_PORT")
	testUnsetEnvName("IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY")
	testUnsetEnvName("IDENTITY_PROVIDERS_OIDC_HMAC_SECRET")
	testUnsetEnvName("STORAGE_ENCRYPTION_KEY")
}

func testUnsetEnvName(name string) {
	_ = os.Unsetenv(DefaultEnvPrefix + name)
	_ = os.Unsetenv(DefaultEnvPrefix + name + constSecretSuffix)
}

func testCreateFile(path, value string, perm os.FileMode) (err error) {
	return os.WriteFile(path, []byte(value), perm)
}
