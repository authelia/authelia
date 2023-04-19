package configuration

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShouldErrorSecretNotExist(t *testing.T) {
	dir := t.TempDir()

	testSetEnv(t, "JWT_SECRET_FILE", filepath.Join(dir, "jwt"))
	testSetEnv(t, "DUO_API_SECRET_KEY_FILE", filepath.Join(dir, "duo"))
	testSetEnv(t, "SESSION_SECRET_FILE", filepath.Join(dir, "session"))
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", dir)
	testSetEnv(t, "NOTIFIER_SMTP_PASSWORD_FILE", filepath.Join(dir, "notifier"))
	testSetEnv(t, "SESSION_REDIS_PASSWORD_FILE", filepath.Join(dir, "redis"))
	testSetEnv(t, "SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE", filepath.Join(dir, "redis-sentinel"))
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD_FILE", filepath.Join(dir, "mysql"))
	testSetEnv(t, "STORAGE_POSTGRES_PASSWORD_FILE", filepath.Join(dir, "postgres"))
	testSetEnv(t, "SERVER_TLS_KEY_FILE", filepath.Join(dir, "tls"))
	testSetEnv(t, "IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE", filepath.Join(dir, "oidc-key"))
	testSetEnv(t, "IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE", filepath.Join(dir, "oidc-hmac"))

	val := schema.NewStructValidator()
	_, _, err := Load(val, NewEnvironmentSource(DefaultEnvPrefix, DefaultEnvDelimiter), NewSecretsSource(DefaultEnvPrefix, DefaultEnvDelimiter))

	assert.NoError(t, err)
	assert.Len(t, val.Warnings(), 0)

	errs := val.Errors()
	require.Len(t, errs, 12)

	sort.Sort(utils.ErrSliceSortAlphabetical(errs))

	errFmt := utils.GetExpectedErrTxt("filenotfound")
	errFmtDir := utils.GetExpectedErrTxt("isdir")

	// ignore the errors before this as they are checked by the validator.
	assert.EqualError(t, errs[0], fmt.Sprintf("secrets: error loading secret path %s into key 'authentication_backend.ldap.password': %s", dir, fmt.Sprintf(errFmtDir, dir)))
	assert.EqualError(t, errs[1], fmt.Sprintf("secrets: error loading secret path %s into key 'duo_api.secret_key': file does not exist error occurred: %s", filepath.Join(dir, "duo"), fmt.Sprintf(errFmt, filepath.Join(dir, "duo"))))
	assert.EqualError(t, errs[2], fmt.Sprintf("secrets: error loading secret path %s into key 'jwt_secret': file does not exist error occurred: %s", filepath.Join(dir, "jwt"), fmt.Sprintf(errFmt, filepath.Join(dir, "jwt"))))
	assert.EqualError(t, errs[3], fmt.Sprintf("secrets: error loading secret path %s into key 'storage.mysql.password': file does not exist error occurred: %s", filepath.Join(dir, "mysql"), fmt.Sprintf(errFmt, filepath.Join(dir, "mysql"))))
	assert.EqualError(t, errs[4], fmt.Sprintf("secrets: error loading secret path %s into key 'notifier.smtp.password': file does not exist error occurred: %s", filepath.Join(dir, "notifier"), fmt.Sprintf(errFmt, filepath.Join(dir, "notifier"))))
	assert.EqualError(t, errs[5], fmt.Sprintf("secrets: error loading secret path %s into key 'identity_providers.oidc.hmac_secret': file does not exist error occurred: %s", filepath.Join(dir, "oidc-hmac"), fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-hmac"))))
	assert.EqualError(t, errs[6], fmt.Sprintf("secrets: error loading secret path %s into key 'identity_providers.oidc.issuer_private_key': file does not exist error occurred: %s", filepath.Join(dir, "oidc-key"), fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-key"))))
	assert.EqualError(t, errs[7], fmt.Sprintf("secrets: error loading secret path %s into key 'storage.postgres.password': file does not exist error occurred: %s", filepath.Join(dir, "postgres"), fmt.Sprintf(errFmt, filepath.Join(dir, "postgres"))))
	assert.EqualError(t, errs[8], fmt.Sprintf("secrets: error loading secret path %s into key 'session.redis.password': file does not exist error occurred: %s", filepath.Join(dir, "redis"), fmt.Sprintf(errFmt, filepath.Join(dir, "redis"))))
	assert.EqualError(t, errs[9], fmt.Sprintf("secrets: error loading secret path %s into key 'session.redis.high_availability.sentinel_password': file does not exist error occurred: %s", filepath.Join(dir, "redis-sentinel"), fmt.Sprintf(errFmt, filepath.Join(dir, "redis-sentinel"))))
	assert.EqualError(t, errs[10], fmt.Sprintf("secrets: error loading secret path %s into key 'session.secret': file does not exist error occurred: %s", filepath.Join(dir, "session"), fmt.Sprintf(errFmt, filepath.Join(dir, "session"))))
	assert.EqualError(t, errs[11], fmt.Sprintf("secrets: error loading secret path %s into key 'server.tls.key': file does not exist error occurred: %s", filepath.Join(dir, "tls"), fmt.Sprintf(errFmt, filepath.Join(dir, "tls"))))
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

func TestShouldParseLargeIntegerDurations(t *testing.T) {
	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.durations.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, durationMax, config.Regulation.FindTime)
	assert.Equal(t, time.Second*1000, config.Regulation.BanTime)
}

func TestShouldValidateConfigurationWithEnv(t *testing.T) {
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

func TestShouldValidateConfigurationWithFilters(t *testing.T) {
	testSetEnv(t, "SESSION_SECRET", "abc")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "abc")
	testSetEnv(t, "JWT_SECRET", "abc")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

	t.Setenv("ABC_CLIENT_SECRET", "$plaintext$example-abc")
	t.Setenv("XYZ_CLIENT_SECRET", "$plaintext$example-xyz")
	t.Setenv("ANOTHER_CLIENT_SECRET", "$plaintext$example-123")
	t.Setenv("SERVICES_SERVER", "10.10.10.10")
	t.Setenv("ROOT_DOMAIN", "example.org")

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSourcesFiltered([]string{"./test_resources/config.filtered.yml"}, NewFileFiltersDefault(), DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	require.Len(t, val.Errors(), 0)
	require.Len(t, val.Warnings(), 0)

	assert.Equal(t, "api-123456789.example.org", config.DuoAPI.Hostname)
	assert.Equal(t, "smtp://10.10.10.10:1025", config.Notifier.SMTP.Address.String())
	assert.Equal(t, "10.10.10.10", config.Session.Redis.Host)

	require.Len(t, config.IdentityProviders.OIDC.Clients, 3)
	assert.Equal(t, "$plaintext$example-abc", config.IdentityProviders.OIDC.Clients[0].Secret.String())
	assert.Equal(t, "$plaintext$example-xyz", config.IdentityProviders.OIDC.Clients[1].Secret.String())
	assert.Equal(t, "$plaintext$example-123", config.IdentityProviders.OIDC.Clients[2].Secret.String())
}

func TestShouldNotIgnoreInvalidEnvs(t *testing.T) {
	testSetEnv(t, "SESSION_SECRET", "an env session secret")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "an env storage mysql password")
	testSetEnv(t, "STORAGE_MYSQL", "a bad env")
	testSetEnv(t, "JWT_SECRET", "an env jwt secret")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_ADDRESS", "an env authentication backend ldap password")

	val := schema.NewStructValidator()
	keys, _, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	require.Len(t, val.Warnings(), 1)
	assert.Len(t, val.Errors(), 1)

	assert.EqualError(t, val.Warnings()[0], fmt.Sprintf("configuration environment variable not expected: %sSTORAGE_MYSQL", DefaultEnvPrefix))
	assert.EqualError(t, val.Errors()[0], "error occurred during unmarshalling configuration: 1 error(s) decoding:\n\n* error decoding 'authentication_backend.ldap.address': could not decode 'an env authentication backend ldap password' to a *schema.AddressLDAP: could not parse string 'an env authentication backend ldap password' as address: expected format is [<scheme>://]<hostname>[:<port>]: parse \"ldaps://an env authentication backend ldap password\": invalid character \" \" in host name")
}

func TestShouldValidateAndRaiseErrorsOnNormalConfigurationAndSecret(t *testing.T) {
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

	dir := t.TempDir()

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "myconf.yml"), []byte("server:\n  port: 9091\n"), 0000))

	cfg := filepath.Join(dir, "myconf.yml")

	val := schema.NewStructValidator()
	_, _, err := Load(val, NewFileSource(cfg))

	assert.NoError(t, err)
	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)
	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("failed to load configuration from file path(%s) source: open %s: permission denied", cfg, cfg))
}

func TestShouldValidateConfigurationWithEnvSecrets(t *testing.T) {
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

/*
func TestShouldLoadNewOIDCConfig(t *testing.T) {
	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_oidc_modern.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	val.Clear()

	validator.ValidateIdentityProviders(&config.IdentityProviders, val)

	assert.Len(t, val.Errors(), 0)

	assert.Len(t, config.IdentityProviders.OIDC.IssuerJWKS.Keys, 2)
	assert.Equal(t, "keya", config.IdentityProviders.OIDC.IssuerJWKS.DefaultKeyID)

	assert.Equal(t, oidc.KeyUseSignature, config.IdentityProviders.OIDC.IssuerJWKS.Keys["keya"].Use)
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, config.IdentityProviders.OIDC.IssuerJWKS.Keys["keya"].Algorithm)

	assert.Equal(t, oidc.KeyUseSignature, config.IdentityProviders.OIDC.IssuerJWKS.Keys["ec521"].Use)
	assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, config.IdentityProviders.OIDC.IssuerJWKS.Keys["ec521"].Algorithm)

	assert.Contains(t, config.IdentityProviders.OIDC.Discovery.RegisteredJWKSigningAlgs, oidc.SigningAlgRSAUsingSHA256)
	assert.Contains(t, config.IdentityProviders.OIDC.Discovery.RegisteredJWKSigningAlgs, oidc.SigningAlgECDSAUsingP521AndSHA512)
}.

*/

func TestShouldConfigureConsent(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_oidc.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	require.Len(t, config.IdentityProviders.OIDC.Clients, 1)
	assert.Equal(t, config.IdentityProviders.OIDC.Clients[0].ConsentMode, "explicit")
}

func TestShouldValidateAndRaiseErrorsOnBadConfiguration(t *testing.T) {
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

func TestShouldValidateDeprecatedEnvNames(t *testing.T) {
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_URL", "ldap://from-env")

	val := schema.NewStructValidator()
	keys, c, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	require.Len(t, val.Warnings(), 1)

	assert.EqualError(t, val.Warnings()[0], "configuration key 'authentication_backend.ldap.url' is deprecated in 4.38.0 and has been replaced by 'authentication_backend.ldap.address': this has not been automatically mapped for you because the replacement key also exists and you will need to adjust your configuration to remove this message")

	assert.Equal(t, "ldap://127.0.0.1:389", c.AuthenticationBackend.LDAP.Address.String())
}

func TestShouldValidateDeprecatedEnvNamesWithDeprecatedKeys(t *testing.T) {
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_URL", "ldap://from-env")

	val := schema.NewStructValidator()
	keys, c, err := Load(val, NewDefaultSources([]string{"./test_resources/config.deprecated.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)

	warnings := val.Warnings()
	require.Len(t, warnings, 3)

	sort.Sort(utils.ErrSliceSortAlphabetical(warnings))

	assert.EqualError(t, warnings[0], "configuration key 'authentication_backend.ldap.url' is deprecated in 4.38.0 and has been replaced by 'authentication_backend.ldap.address': this has been automatically mapped for you but you will need to adjust your configuration to remove this message")
	assert.EqualError(t, warnings[1], "configuration key 'storage.mysql.host' is deprecated in 4.38.0 and has been replaced by 'storage.mysql.address' when combined with the 'storage.mysql.port' in the format of '[tcp://]<hostname>[:<port>]': this should be automatically mapped for you but you will need to adjust your configuration to remove this message")
	assert.EqualError(t, warnings[2], "configuration key 'storage.mysql.port' is deprecated in 4.38.0 and has been replaced by 'storage.mysql.address' when combined with the 'storage.mysql.host' in the format of '[tcp://]<hostname>[:<port>]': this should be automatically mapped for you but you will need to adjust your configuration to remove this message")

	assert.Equal(t, "ldap://from-env:389", c.AuthenticationBackend.LDAP.Address.String())
}

func TestShouldValidateDeprecatedEnvNamesWithDeprecatedKeysAlt(t *testing.T) {
	val := schema.NewStructValidator()
	keys, c, err := Load(val, NewDefaultSources([]string{"./test_resources/config.deprecated.alt.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)

	warnings := val.Warnings()
	require.Len(t, warnings, 3)

	sort.Sort(utils.ErrSliceSortAlphabetical(warnings))

	assert.EqualError(t, warnings[0], "configuration key 'authentication_backend.ldap.url' is deprecated in 4.38.0 and has been replaced by 'authentication_backend.ldap.address': this has been automatically mapped for you but you will need to adjust your configuration to remove this message")
	assert.EqualError(t, warnings[1], "configuration key 'storage.postgres.host' is deprecated in 4.38.0 and has been replaced by 'storage.postgres.address' when combined with the 'storage.postgres.port' in the format of '[tcp://]<hostname>[:<port>]': this should be automatically mapped for you but you will need to adjust your configuration to remove this message")
	assert.EqualError(t, warnings[2], "configuration key 'storage.postgres.port' is deprecated in 4.38.0 and has been replaced by 'storage.postgres.address' when combined with the 'storage.postgres.host' in the format of '[tcp://]<hostname>[:<port>]': this should be automatically mapped for you but you will need to adjust your configuration to remove this message")

	val.Clear()

	validator.ValidateConfiguration(c, val)

	require.NotNil(t, c.Storage.PostgreSQL.Address)
	assert.Equal(t, "tcp://127.0.0.1:5432", c.Storage.PostgreSQL.Address.String())
}

func TestShouldRaiseErrOnInvalidNotifierSMTPSender(t *testing.T) {
	val := schema.NewStructValidator()
	keys, _, err := Load(val, NewDefaultSources([]string{"./test_resources/config_smtp_sender_invalid.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	assert.EqualError(t, val.Errors()[0], "error occurred during unmarshalling configuration: 1 error(s) decoding:\n\n* error decoding 'notifier.smtp.sender': could not decode 'admin' to a mail.Address (RFC5322): mail: missing '@' or angle-addr")
}

func TestShouldHandleErrInvalidatorWhenSMTPSenderBlank(t *testing.T) {
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
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_alt.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "Admin", config.Notifier.SMTP.Sender.Name)
	assert.Equal(t, "admin@example.com", config.Notifier.SMTP.Sender.Address)
	assert.Equal(t, schema.RememberMeDisabled, config.Session.RememberMe)
}

func TestShouldParseRegex(t *testing.T) {
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

	dir := t.TempDir()

	cfg := filepath.Join(dir, "config.yml")
	assert.NoError(t, testCreateFile(filepath.Join(dir, "config.yml"), "port: 9091\n", 0000))

	val := schema.NewStructValidator()
	_, _, err := Load(val, NewFileSource(cfg))

	assert.NoError(t, err)
	require.Len(t, val.Errors(), 1)

	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("failed to load configuration from file path(%s) source: open %s: permission denied", cfg, cfg))
}

func TestShouldLoadDirectoryConfiguration(t *testing.T) {
	dir := t.TempDir()

	cfg := filepath.Join(dir, "myconf.yml")
	assert.NoError(t, testCreateFile(cfg, "server:\n  port: 9091\n", 0700))

	val := schema.NewStructValidator()
	_, _, err := Load(val, NewFileSource(dir))

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	require.Len(t, val.Warnings(), 1)

	assert.EqualError(t, val.Warnings()[0], "configuration key 'server.port' is deprecated in 4.38.0 and has been replaced by 'server.address' when combined with the 'server.host' in the format of '[tcp://]<hostname>[:<port>]': this should be automatically mapped for you but you will need to adjust your configuration to remove this message")
}

func testSetEnv(t *testing.T, key, value string) {
	t.Setenv(DefaultEnvPrefix+key, value)
}

func testCreateFile(path, value string, perm os.FileMode) (err error) {
	return os.WriteFile(path, []byte(value), perm)
}

func TestShouldErrorOnNoPath(t *testing.T) {
	val := schema.NewStructValidator()
	_, _, err := Load(val, NewFileSource(""))

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 1)
	assert.ErrorContains(t, val.Errors()[0], "invalid file path source configuration")
}

func TestShouldErrorOnInvalidPath(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "invalid-folder/config")

	val := schema.NewStructValidator()
	_, _, err := Load(val, NewFileSource(cfg))

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 1)
	assert.ErrorContains(t, val.Errors()[0], fmt.Sprintf("stat %s: no such file or directory", cfg))
}

func TestShouldErrorOnDirFSPermissionDenied(t *testing.T) {
	if runtime.GOOS == constWindows {
		t.Skip("skipping test due to being on windows")
	}

	dir := t.TempDir()
	err := os.Chmod(dir, 0200)
	assert.NoError(t, err)

	val := schema.NewStructValidator()
	_, _, err = Load(val, NewFileSource(dir))

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 1)
	assert.ErrorContains(t, val.Errors()[0], fmt.Sprintf("open %s: permission denied", dir))
}

func TestShouldSkipDirOnLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "some-dir")

	err := os.Mkdir(path, 0700)
	assert.NoError(t, err)

	val := schema.NewStructValidator()
	_, _, err = Load(val, NewFileSource(dir))

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
}

func TestShouldFailIfYmlIsInvalid(t *testing.T) {
	dir := t.TempDir()

	cfg := filepath.Join(dir, "myconf.yml")
	assert.NoError(t, testCreateFile(cfg, "an invalid contend\n", 0700))

	val := schema.NewStructValidator()
	_, _, err := Load(val, NewFileSource(dir))

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 1)
	assert.ErrorContains(t, val.Errors()[0], "unmarshal errors")
}
