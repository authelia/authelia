package configuration

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShouldErrorSecretNotExist(t *testing.T) {
	dir := t.TempDir()

	testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET_FILE", filepath.Join(dir, "jwt"))
	testSetEnv(t, "DUO_API_SECRET_KEY_FILE", filepath.Join(dir, "duo"))
	testSetEnv(t, "SESSION_SECRET_FILE", filepath.Join(dir, "session"))
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", dir)
	testSetEnv(t, "NOTIFIER_SMTP_PASSWORD_FILE", filepath.Join(dir, "notifier"))
	testSetEnv(t, "SESSION_REDIS_PASSWORD_FILE", filepath.Join(dir, "redis"))
	testSetEnv(t, "SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE", filepath.Join(dir, "redis-sentinel"))
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD_FILE", filepath.Join(dir, "mysql"))
	testSetEnv(t, "STORAGE_POSTGRES_PASSWORD_FILE", filepath.Join(dir, "postgres"))
	testSetEnv(t, "IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE", filepath.Join(dir, "oidc-key"))
	testSetEnv(t, "IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE", filepath.Join(dir, "oidc-hmac"))

	val := schema.NewStructValidator()
	_, _, err := Load(val, NewEnvironmentSource(DefaultEnvPrefix, DefaultEnvDelimiter), NewSecretsSource(DefaultEnvPrefix, DefaultEnvDelimiter))

	assert.NoError(t, err)
	assert.Len(t, val.Warnings(), 0)

	errs := val.Errors()
	require.Len(t, errs, 11)

	sort.Sort(utils.ErrSliceSortAlphabetical(errs))

	errFmt := utils.GetExpectedErrTxt("filenotfound")
	errFmtDir := utils.GetExpectedErrTxt("isdir")

	// ignore the errors before this as they are checked by the validator.
	assert.EqualError(t, errs[0], fmt.Sprintf("secrets: error loading secret path %s into key 'authentication_backend.ldap.password': %s", dir, fmt.Sprintf(errFmtDir, dir)))
	assert.EqualError(t, errs[1], fmt.Sprintf("secrets: error loading secret path %s into key 'duo_api.secret_key': file does not exist error occurred: %s", filepath.Join(dir, "duo"), fmt.Sprintf(errFmt, filepath.Join(dir, "duo"))))
	assert.EqualError(t, errs[2], fmt.Sprintf("secrets: error loading secret path %s into key 'identity_validation.reset_password.jwt_secret': file does not exist error occurred: %s", filepath.Join(dir, "jwt"), fmt.Sprintf(errFmt, filepath.Join(dir, "jwt"))))
	assert.EqualError(t, errs[3], fmt.Sprintf("secrets: error loading secret path %s into key 'storage.mysql.password': file does not exist error occurred: %s", filepath.Join(dir, "mysql"), fmt.Sprintf(errFmt, filepath.Join(dir, "mysql"))))
	assert.EqualError(t, errs[4], fmt.Sprintf("secrets: error loading secret path %s into key 'notifier.smtp.password': file does not exist error occurred: %s", filepath.Join(dir, "notifier"), fmt.Sprintf(errFmt, filepath.Join(dir, "notifier"))))
	assert.EqualError(t, errs[5], fmt.Sprintf("secrets: error loading secret path %s into key 'identity_providers.oidc.hmac_secret': file does not exist error occurred: %s", filepath.Join(dir, "oidc-hmac"), fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-hmac"))))
	assert.EqualError(t, errs[6], fmt.Sprintf("secrets: error loading secret path %s into key 'identity_providers.oidc.issuer_private_key': file does not exist error occurred: %s", filepath.Join(dir, "oidc-key"), fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-key"))))
	assert.EqualError(t, errs[7], fmt.Sprintf("secrets: error loading secret path %s into key 'storage.postgres.password': file does not exist error occurred: %s", filepath.Join(dir, "postgres"), fmt.Sprintf(errFmt, filepath.Join(dir, "postgres"))))
	assert.EqualError(t, errs[8], fmt.Sprintf("secrets: error loading secret path %s into key 'session.redis.password': file does not exist error occurred: %s", filepath.Join(dir, "redis"), fmt.Sprintf(errFmt, filepath.Join(dir, "redis"))))
	assert.EqualError(t, errs[9], fmt.Sprintf("secrets: error loading secret path %s into key 'session.redis.high_availability.sentinel_password': file does not exist error occurred: %s", filepath.Join(dir, "redis-sentinel"), fmt.Sprintf(errFmt, filepath.Join(dir, "redis-sentinel"))))
	assert.EqualError(t, errs[10], fmt.Sprintf("secrets: error loading secret path %s into key 'session.secret': file does not exist error occurred: %s", filepath.Join(dir, "session"), fmt.Sprintf(errFmt, filepath.Join(dir, "session"))))
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
	testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET", "abc")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
	assert.NotNil(t, config.Notifier)
}

func TestShouldHandleNoAutoMapEmptyNewKey(t *testing.T) {
	testSetEnv(t, "SESSION_SECRET", "abc")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "abc")
	testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET", "abc")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_no_automap.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	require.Len(t, val.Warnings(), 1)

	assert.EqualError(t, val.Warnings()[0], "configuration key 'authentication_backend.ldap.permit_feature_detection_failure' is deprecated in 4.39.16 and has been removed': you are not required to make any configuration changes right now but you may be required to in 5.0.0")
	assert.NotNil(t, config.Notifier)

	assert.NotContains(t, keys, "authentication_backend.ldap.permit_feature_detection_failure")
}

func TestShouldHaveEndpointSubPath(t *testing.T) {
	testSetEnv(t, "SESSION_SECRET", "abc")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "abc")
	testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET", "abc")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_authz_subpath.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)
	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
	assert.Contains(t, config.Server.Endpoints.Authz, "auth-request/basic")
}

func TestShouldConfigureRefreshIntervalDisable(t *testing.T) {
	testSetEnv(t, "SESSION_SECRET", "abc")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "abc")
	testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET", "abc")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	require.NotNil(t, config.AuthenticationBackend.RefreshInterval)
	assert.True(t, config.AuthenticationBackend.RefreshInterval.Never())
	assert.False(t, config.AuthenticationBackend.RefreshInterval.Always())
}

func TestShouldParseLargeIntegerDurations(t *testing.T) {
	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.durations.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, durationMax, config.Regulation.FindTime)
	assert.Equal(t, time.Second*1000, config.Regulation.BanTime)

	require.NotNil(t, config.AuthenticationBackend.RefreshInterval)
	assert.Equal(t, false, config.AuthenticationBackend.RefreshInterval.Always())
	assert.Equal(t, false, config.AuthenticationBackend.RefreshInterval.Never())
	assert.Equal(t, time.Minute*5, config.AuthenticationBackend.RefreshInterval.Value())
}

func TestShouldValidateConfigurationWithEnv(t *testing.T) {
	testSetEnv(t, "SESSION_SECRET", "abc")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "abc")
	testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET", "abc")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

	val := schema.NewStructValidator()
	_, _, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
}

func TestShouldValidateConfigurationWithOverridenDefaults(t *testing.T) {
	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSourcesWithDefaults([]string{"./test_resources/config.webauthn.yml"}, NewFileFiltersDefault(), DefaultEnvPrefix, DefaultEnvDelimiter, nil)...)

	require.NoError(t, err)

	validator.ValidateWebAuthn(config, val)

	assert.Equal(t, protocol.ResidentKeyRequirement(""), config.WebAuthn.SelectionCriteria.Discoverability)
	assert.Equal(t, protocol.AuthenticatorAttachment(""), config.WebAuthn.SelectionCriteria.Attachment)
	assert.Equal(t, protocol.UserVerificationRequirement(""), config.WebAuthn.SelectionCriteria.UserVerification)
}

func TestShouldValidateConfigurationWithoutOverridenDefaults(t *testing.T) {
	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSourcesWithDefaults([]string{"./test_resources/config.webauthn-defaults.yml"}, NewFileFiltersDefault(), DefaultEnvPrefix, DefaultEnvDelimiter, nil)...)

	require.NoError(t, err)

	validator.ValidateWebAuthn(config, val)

	assert.Equal(t, protocol.ResidentKeyRequirementPreferred, config.WebAuthn.SelectionCriteria.Discoverability)
	assert.Equal(t, protocol.AuthenticatorAttachment(""), config.WebAuthn.SelectionCriteria.Attachment)
	assert.Equal(t, protocol.VerificationPreferred, config.WebAuthn.SelectionCriteria.UserVerification)
}

func TestShouldValidateConfigurationWithFilters(t *testing.T) {
	testCases := []struct {
		name string
		path string
	}{
		{
			"ShouldHandleSingleFile",
			"./test_resources/config.filtered.yml",
		},
		{
			"ShouldHandleDirectory",
			"./test_resources/config-dir/filtered",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSetEnv(t, "SESSION_SECRET", "abc")
			testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "abc")
			testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET", "abc")
			testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

			t.Setenv("ABC_CLIENT_SECRET", "$plaintext$example-abc")
			t.Setenv("XYZ_CLIENT_SECRET", "$plaintext$example-xyz")
			t.Setenv("SERVICES_SERVER", "10.10.10.10")
			t.Setenv("ROOT_DOMAIN", "example.org")

			val := schema.NewStructValidator()
			_, config, err := Load(val, NewDefaultSourcesFiltered([]string{tc.path}, NewFileFiltersDefault(), DefaultEnvPrefix, DefaultEnvDelimiter)...)

			assert.NoError(t, err)
			require.Len(t, val.Errors(), 0)
			require.Len(t, val.Warnings(), 0)

			assert.Equal(t, "api-123456789.example.org", config.DuoAPI.Hostname)
			assert.Equal(t, "smtp://10.10.10.10:1025", config.Notifier.SMTP.Address.String())
			assert.Equal(t, "10.10.10.10", config.Session.Redis.Host)

			require.Len(t, config.IdentityProviders.OIDC.Clients, 4)
			assert.Equal(t, "$plaintext$example-abc", config.IdentityProviders.OIDC.Clients[0].Secret.String())
			assert.Equal(t, "$plaintext$example-xyz", config.IdentityProviders.OIDC.Clients[1].Secret.String())
			assert.Equal(t, "$plaintext$example_secret value", config.IdentityProviders.OIDC.Clients[2].Secret.String())
			assert.Equal(t, "$plaintext$abc", config.IdentityProviders.OIDC.Clients[3].Secret.String())

			require.Len(t, config.IdentityProviders.OIDC.JSONWebKeys, 1)

			key, ok := config.IdentityProviders.OIDC.JSONWebKeys[0].Key.(schema.CryptographicPrivateKey)
			assert.True(t, ok)
			require.NotNil(t, key)

			rsakey, ok := key.(*rsa.PrivateKey)
			assert.True(t, ok)
			require.NotNil(t, rsakey)

			assert.Equal(t, 65537, rsakey.E)
			assert.Equal(t, "27171434142509968675194232284375073019792572110439705540328918657232692168643195881620537202636198369160560799743144111431567452741046220953662805104932829188046044673961143220261310008810498023470535975681337666107808278037041152412426963982841905494490761888868583347468199094007084012384588888035364766072411615843478518353414183640511444802956354678240763665865557092671631235272029876735331399857244041249715616453815382050245467939750635216436773618819757152567487060661311335480594478902550197306956880336905504741940598285468339785455485086967213774716099196949673312743795439236046960995348506152278833238987", rsakey.N.String())
			assert.Equal(t, "5706925720915661669195242494994016816721008820974450261113990040996811079258641550734801632578349185215910392731806135371706455696484447433162465664729853270266472449716574399604756584391664331493231727196142834947800188400138417427667686333274620887920797982823077799989315356653608060034390741776504814150513570875362236882334931949786678793855564217596234691391113095918532726196507032878006343060796051755555405212832046478322407013172691936979796693050565243392092102513298609204623359016844719592078589959501078387650387089103850347191460557257744984924144972386173794776498508384237037750896668486369884278793", rsakey.D.String())
		})
	}
}

func TestShouldValidateConfigurationWithFiltersWalk(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSourcesFiltered([]string{"./test_resources/config_walk.yml"}, []BytesFilter{NewTemplateFileFilter()}, DefaultEnvPrefix, DefaultEnvDelimiter)...)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.NotNil(t, keys)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
}

func TestShouldValidateConfigurationWithFiltersGlob(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSourcesFiltered([]string{"./test_resources/config_glob.yml"}, []BytesFilter{NewTemplateFileFilter()}, DefaultEnvPrefix, DefaultEnvDelimiter)...)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.NotNil(t, keys)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
}

func TestShouldReadFilesWithFiltersSingleFile(t *testing.T) {
	testSetEnv(t, "SESSION_SECRET", "abc")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "abc")
	testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET", "abc")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

	t.Setenv("ABC_CLIENT_SECRET", "$plaintext$example-abc")
	t.Setenv("XYZ_CLIENT_SECRET", "$plaintext$example-xyz")
	t.Setenv("SERVICES_SERVER", "10.10.10.10")
	t.Setenv("ROOT_DOMAIN", "example.org")

	sources := NewDefaultSourcesFiltered([]string{"./test_resources/config.filtered.yml"}, NewFileFiltersDefault(), DefaultEnvPrefix, DefaultEnvDelimiter)

	var (
		source *FileSource
		files  []*File
		err    error
		ok     bool
	)

	for _, s := range sources {
		if source, ok = s.(*FileSource); !ok {
			continue
		}

		var f []*File

		f, err = source.ReadFiles()

		require.NoError(t, err)

		files = append(files, f...)
	}

	assert.Len(t, files, 1)

	for _, file := range files {
		switch file.Path {
		case "./test_resources/config.filtered.yml":
			data := string(file.Data)

			assert.Contains(t, data, "- 'secure.example.org'")
			assert.Contains(t, data, "address: 'ldap://10.10.10.10'")
			assert.Contains(t, data, "address: 'tcp://10.10.10.10:9091'")
			assert.Contains(t, data, "hostname: 'api-123456789.example.org'")
			assert.Contains(t, data, "client_secret: 'example_secret value'")
			assert.Contains(t, data, "sender: 'admin@example.org'")
			assert.Contains(t, data, "domain: 'example.org'")
			assert.Contains(t, data, "address: 'tcp://10.10.10.10:3306'")
		default:
			assert.Fail(t, "Unexpected File", "The file with path %s is not expected.", file.Path)
		}
	}
}

func TestShouldReadFilesWithFilters(t *testing.T) {
	testSetEnv(t, "SESSION_SECRET", "abc")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "abc")
	testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET", "abc")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

	t.Setenv("ABC_CLIENT_SECRET", "$plaintext$example-abc")
	t.Setenv("XYZ_CLIENT_SECRET", "$plaintext$example-xyz")
	t.Setenv("SERVICES_SERVER", "10.10.10.10")
	t.Setenv("ROOT_DOMAIN", "example.org")

	sources := NewDefaultSourcesFiltered([]string{"./test_resources/config-dir/filtered"}, NewFileFiltersDefault(), DefaultEnvPrefix, DefaultEnvDelimiter)

	var (
		source *FileSource
		files  []*File
		err    error
		ok     bool
	)

	for _, s := range sources {
		if source, ok = s.(*FileSource); !ok {
			continue
		}

		var f []*File

		f, err = source.ReadFiles()

		require.NoError(t, err)

		files = append(files, f...)
	}

	assert.Len(t, files, 7)

	for _, file := range files {
		switch file.Path {
		case "test_resources/config-dir/filtered/access-control.yml":
			assert.Contains(t, string(file.Data), "- 'secure.example.org'")
		case "test_resources/config-dir/filtered/authentication-backend.yml":
			assert.Contains(t, string(file.Data), "address: 'ldap://10.10.10.10'")
		case "test_resources/config-dir/filtered/general.yml":
			data := string(file.Data)
			assert.Contains(t, data, "address: 'tcp://10.10.10.10:9091'")
			assert.Contains(t, data, "hostname: 'api-123456789.example.org'")
		case "test_resources/config-dir/filtered/identity-providers.yml":
			assert.Contains(t, string(file.Data), "client_secret: 'example_secret value'")
		case "test_resources/config-dir/filtered/notifier.yml":
			assert.Contains(t, string(file.Data), "sender: 'admin@example.org'")
		case "test_resources/config-dir/filtered/session.yml":
			assert.Contains(t, string(file.Data), "domain: 'example.org'")
		case "test_resources/config-dir/filtered/storage.yml":
			assert.Contains(t, string(file.Data), "address: 'tcp://10.10.10.10:3306'")
		default:
			assert.Fail(t, "Unexpected File", "The file with path %s is not expected.", file.Path)
		}
	}
}

func TestShouldHandleNoAddressMySQLWithHostEnv(t *testing.T) {
	testSetEnv(t, "STORAGE_MYSQL_HOST", "mysql")

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_no_address_mysql.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateConfiguration(config, val)

	require.Len(t, val.Warnings(), 1)
	require.Len(t, val.Errors(), 1)

	assert.EqualError(t, val.Warnings()[0], "configuration keys 'storage.mysql.host' and 'storage.mysql.port' are deprecated in 4.38.0 and has been replaced by 'storage.mysql.address' in the format of '[tcp://]<hostname>[:<port>]': you are not required to make any changes as this has been automatically mapped for you to the value 'tcp://mysql:3306', but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0")
	assert.Equal(t, "tcp://mysql:3306", config.Storage.MySQL.Address.String())
}

func TestShouldHandleNoAddressPostgreSQLWithHostEnv(t *testing.T) {
	testSetEnv(t, "STORAGE_POSTGRES_HOST", "postgres")

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_no_address_postgres.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateConfiguration(config, val)

	assert.Len(t, val.Warnings(), 1)
	assert.Len(t, val.Errors(), 1)

	assert.Equal(t, "tcp://postgres:5432", config.Storage.PostgreSQL.Address.String())
}

func TestShouldHandleNoAddressSMTPWithHostEnv(t *testing.T) {
	testSetEnv(t, "NOTIFIER_SMTP_HOST", "smtp")

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_no_address_smtp.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateConfiguration(config, val)

	assert.Len(t, val.Warnings(), 1)
	assert.Len(t, val.Errors(), 1)

	assert.Equal(t, "smtp://smtp:25", config.Notifier.SMTP.Address.String())
}

func TestShouldNotIgnoreInvalidEnvs(t *testing.T) {
	testSetEnv(t, "SESSION_SECRET", "an env session secret")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "an env storage mysql password")
	testSetEnv(t, "STORAGE_MYSQL", "a bad env")
	testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET", "an env jwt secret")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_ADDRESS", "an env authentication backend ldap password")

	val := schema.NewStructValidator()
	keys, _, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	require.Len(t, val.Warnings(), 1)
	assert.Len(t, val.Errors(), 1)

	assert.EqualError(t, val.Warnings()[0], fmt.Sprintf("configuration environment variable not expected: %sSTORAGE_MYSQL", DefaultEnvPrefix))
	assert.EqualError(t, val.Errors()[0], "error occurred during unmarshalling configuration: decoding failed due to the following error(s):\n\n'authentication_backend.ldap.address' could not decode 'an env authentication backend ldap password' to a *schema.AddressLDAP: could not parse string 'an env authentication backend ldap password' as address: expected format is [<scheme>://]<hostname>[:<port>]: parse \"ldaps://an env authentication backend ldap password\": invalid character \" \" in host name")
}

func TestShouldValidateServerAddressValues(t *testing.T) {
	testCases := []struct {
		name string
		data []byte

		envHost, envPort, envAddress string

		envMetricsAddress      string
		expectedHTTP           string
		expectedNetAddrHTTP    string
		expectedMetrics        string
		expectedNetAddrMetrics string
		werrs                  []string
		errs                   []string
	}{
		{
			"ShouldSetDefaultValues",
			nil,
			"",
			"",
			"",
			"",
			"tcp://:9091/",
			":9091",
			"tcp://:9959/metrics",
			":9959",
			nil,
			nil,
		},
		{
			"ShouldMapEnvValuesWithConfigTemplate",
			func() []byte {
				data, err := os.ReadFile("config.template.yml")
				if err != nil {
					panic(err)
				}

				return data
			}(),
			"127.0.0.1",
			"8080",
			"",
			"",
			"tcp://127.0.0.1:8080/",
			"127.0.0.1:8080",
			"tcp://:9959/metrics",
			":9959",
			[]string{
				"configuration keys 'server.host', 'server.port', and 'server.path' are deprecated in 4.38.0 and has been replaced by 'server.address' in the format of '[tcp[(4|6)]://]<hostname>[:<port>][/<path>]' or 'tcp[(4|6)://][hostname]:<port>[/<path>]': you are not required to make any changes as this has been automatically mapped for you to the value 'tcp://127.0.0.1:8080/', but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0",
			},
			nil,
		},
		{
			"ShouldOverrideDefault",
			func() []byte {
				data, err := os.ReadFile("config.template.yml")
				if err != nil {
					panic(err)
				}

				return data
			}(),
			"",
			"",
			"tcp://127.0.0.2:7071",
			"tcp://127.0.0.3:8080",
			"tcp://127.0.0.2:7071/",
			"127.0.0.2:7071",
			"tcp://127.0.0.3:8080/metrics",
			"127.0.0.3:8080",
			nil,
			nil,
		},
		{
			"ShouldErrorOnDeprecatedEnvAndModernConfigFileListenerOptions",
			[]byte("server:\n  address: 'tcp://:1000'"),
			"127.0.0.1",
			"8080",
			"",
			"tcp://:",
			"tcp://:1000/",
			":1000",
			"tcp://:9959/metrics",
			":9959",
			nil,
			[]string{
				"error occurred performing deprecation mapping for keys 'server.host', 'server.port', and 'server.path' to new key server.address: the new key already exists with value 'tcp://:1000' but the deprecated keys and the new key can't both be configured",
			},
		},
		{
			"ShouldErrorOnDeprecatedEnvAndModernEnvListenerOptions",
			nil,
			"127.0.0.1",
			"8080",
			"tcp://:10000",
			"tcp://:",
			"tcp://:10000/",
			":10000",
			"tcp://:9959/metrics",
			":9959",
			nil,
			[]string{
				"error occurred performing deprecation mapping for keys 'server.host', 'server.port', and 'server.path' to new key server.address: the new key already exists with value 'tcp://:10000' but the deprecated keys and the new key can't both be configured",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSetEnv(t, "TELEMETRY_METRICS_ENABLED", "true")

			if tc.envHost != "" {
				testSetEnv(t, "SERVER_HOST", tc.envHost)
			}

			if tc.envPort != "" {
				testSetEnv(t, "SERVER_PORT", tc.envPort)
			}

			if tc.envAddress != "" {
				testSetEnv(t, "SERVER_ADDRESS", tc.envAddress)
			}

			if tc.envMetricsAddress != "" {
				testSetEnv(t, "TELEMETRY_METRICS_ADDRESS", tc.envMetricsAddress)
			}

			sources := []Source{
				NewBytesSource(tc.data),
				NewEnvironmentSource(DefaultEnvPrefix, DefaultEnvDelimiter),
				NewSecretsSource(DefaultEnvPrefix, DefaultEnvDelimiter),
			}

			val := schema.NewStructValidator()
			keys, config, err := Load(val, sources...)

			assert.NoError(t, err)

			validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

			assert.NotEmpty(t, config)

			validator.ValidateServer(config, val)
			validator.ValidateTelemetry(config, val)

			werrs := val.Warnings()
			if n := len(tc.werrs); n == 0 {
				assert.Len(t, werrs, 0)
			} else {
				require.Len(t, werrs, n)

				for i := 0; i < n; i++ {
					assert.EqualError(t, werrs[i], tc.werrs[i])
				}
			}

			errs := val.Errors()
			if n := len(tc.errs); n == 0 {
				assert.Len(t, errs, 0)
			} else {
				require.Len(t, errs, n)

				for i := 0; i < n; i++ {
					assert.EqualError(t, errs[i], tc.errs[i])
				}
			}

			assert.Equal(t, tc.expectedHTTP, config.Server.Address.String())
			assert.Equal(t, "tcp", config.Server.Address.Network())
			assert.Equal(t, tc.expectedNetAddrHTTP, config.Server.Address.NetworkAddress())

			assert.Equal(t, tc.expectedMetrics, config.Telemetry.Metrics.Address.String())
			assert.Equal(t, "tcp", config.Telemetry.Metrics.Address.Network())
			assert.Equal(t, tc.expectedNetAddrMetrics, config.Telemetry.Metrics.Address.NetworkAddress())
		})
	}
}

func TestShouldValidateAndRaiseErrorsOnNormalConfigurationAndSecret(t *testing.T) {
	testSetEnv(t, "SESSION_SECRET", "an env session secret")
	testSetEnv(t, "SESSION_SECRET_FILE", "./test_resources/example_secret")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "an env storage mysql password")
	testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET_FILE", "./test_resources/example_secret")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password")
	testSetEnv(t, "STORAGE_ENCRYPTION_KEY", "a_very_bad_encryption_key")

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	assert.EqualError(t, val.Errors()[0], "secrets: error loading secret into key 'session.secret': it's already defined in other configuration sources")

	assert.Equal(t, "example_secret value", config.IdentityValidation.ResetPassword.JWTSecret)
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
	testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET_FILE", "./test_resources/example_secret")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", "./test_resources/example_secret")
	testSetEnv(t, "STORAGE_ENCRYPTION_KEY_FILE", "./test_resources/example_secret")

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "example_secret value", config.IdentityValidation.ResetPassword.JWTSecret)
	assert.Equal(t, "example_secret value", config.Session.Secret)
	assert.Equal(t, "example_secret value", config.AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "example_secret value", config.Storage.MySQL.Password)
	assert.Equal(t, "example_secret value", config.Storage.EncryptionKey)
}

func TestShouldNotErrorOnLogLevel(t *testing.T) {
	testSetEnv(t, "LOG_LEVEL", "warn")

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_nolog.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "warn", config.Log.Level)
}

func TestShouldLoadURLList(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_oidc.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	require.Len(t, config.IdentityProviders.OIDC.CORS.AllowedOrigins, 2)
	assert.Equal(t, "https://google.com", config.IdentityProviders.OIDC.CORS.AllowedOrigins[0].String())
	assert.Equal(t, "https://example.com", config.IdentityProviders.OIDC.CORS.AllowedOrigins[1].String())
}

func TestShouldNotPanicJWKNilKey(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_oidc_empty_jwk_key.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	require.NotPanics(t, func() {
		validator.ValidateIdentityProviders(validator.NewValidateCtx(), config, val)
	})

	assert.Len(t, val.Warnings(), 1)
	require.Len(t, val.Errors(), 5)

	assert.EqualError(t, val.Errors()[0], "identity_providers: oidc: jwks: key #1 with key id 'abc': option 'key' must be provided")
	assert.EqualError(t, val.Errors()[1], "identity_providers: oidc: jwks: key #2: option 'key' must be provided")
	assert.EqualError(t, val.Errors()[2], "identity_providers: oidc: clients: client 'abc': jwks: key #1 with key id 'client_abc': option 'key' must be provided")
	assert.EqualError(t, val.Errors()[3], "identity_providers: oidc: clients: client 'abc': jwks: key #2: option 'key_id' must be provided")
	assert.EqualError(t, val.Errors()[4], "identity_providers: oidc: clients: client 'abc': jwks: key #2: option 'key' must be provided")
}

func TestShouldDisableOIDCEntropy(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_oidc_disable_entropy.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, -1, config.IdentityProviders.OIDC.MinimumParameterEntropy)

	validator.ValidateIdentityProviders(validator.NewValidateCtx(), config, val)

	assert.Len(t, val.Errors(), 1)
	require.Len(t, val.Warnings(), 2)

	assert.EqualError(t, val.Warnings()[0], "identity_providers: oidc: option 'minimum_parameter_entropy' is disabled which is considered unsafe and insecure")
	assert.EqualError(t, val.Warnings()[1], "identity_providers: oidc: clients: client 'abc': option 'client_secret' is plaintext but for clients not using any endpoint authentication method 'client_secret_jwt' it should be a hashed value as plaintext values are deprecated with the exception of 'client_secret_jwt' and will be removed in the near future")
	assert.Equal(t, -1, config.IdentityProviders.OIDC.MinimumParameterEntropy)
}

func TestShouldHandleOIDCClaims(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_oidc_claims.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)

	val.Clear()

	validator.ValidateIdentityProviders(validator.NewValidateCtx(), config, val)

	require.Len(t, val.Warnings(), 1)
	assert.EqualError(t, val.Warnings()[0], "identity_providers: oidc: clients: client 'abc': option 'client_secret' is plaintext but for clients not using any endpoint authentication method 'client_secret_jwt' it should be a hashed value as plaintext values are deprecated with the exception of 'client_secret_jwt' and will be removed in the near future")

	require.Len(t, val.Errors(), 2)
	assert.Regexp(t, regexp.MustCompile(`^identity_providers: oidc: jwks: key #1 with key id 'keya': option 'certificate_chain' produced an error during validation of the chain: certificate #1 in chain is invalid after 1713180174 but the time is \d+$`), val.Errors()[0].Error())
	assert.Regexp(t, regexp.MustCompile(`^identity_providers: oidc: jwks: key #2 with key id 'ec521': option 'certificate_chain' produced an error during validation of the chain: certificate #1 in chain is invalid after 1713180101 but the time is \d+$`), val.Errors()[1].Error())

	require.Len(t, config.IdentityProviders.OIDC.JSONWebKeys, 3)
	require.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[0].Key)
	require.IsType(t, &rsa.PrivateKey{}, config.IdentityProviders.OIDC.JSONWebKeys[0].Key)
	assert.Equal(t, "sig", config.IdentityProviders.OIDC.JSONWebKeys[0].Use)
	assert.Equal(t, "RS256", config.IdentityProviders.OIDC.JSONWebKeys[0].Algorithm)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[0].Key.(*rsa.PrivateKey).D)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[0].Key.(*rsa.PrivateKey).N)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[0].Key.(*rsa.PrivateKey).E)
	assert.Equal(t, 256, config.IdentityProviders.OIDC.JSONWebKeys[0].Key.(*rsa.PrivateKey).Size())
	require.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[0].CertificateChain)
	assert.True(t, config.IdentityProviders.OIDC.JSONWebKeys[0].CertificateChain.HasCertificates())

	require.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[1].Key)
	require.IsType(t, &ecdsa.PrivateKey{}, config.IdentityProviders.OIDC.JSONWebKeys[1].Key)
	assert.Equal(t, "sig", config.IdentityProviders.OIDC.JSONWebKeys[1].Use)
	assert.Equal(t, "ES512", config.IdentityProviders.OIDC.JSONWebKeys[1].Algorithm)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[1].Key.(*ecdsa.PrivateKey).D)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[1].Key.(*ecdsa.PrivateKey).Y)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[1].Key.(*ecdsa.PrivateKey).X)
	assert.Equal(t, elliptic.P521(), config.IdentityProviders.OIDC.JSONWebKeys[1].Key.(*ecdsa.PrivateKey).Curve)
	require.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[1].CertificateChain)
	assert.True(t, config.IdentityProviders.OIDC.JSONWebKeys[1].CertificateChain.HasCertificates())

	require.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[2].Key)
	assert.Equal(t, "sig", config.IdentityProviders.OIDC.JSONWebKeys[2].Use)
	assert.Equal(t, "RS256", config.IdentityProviders.OIDC.JSONWebKeys[2].Algorithm)
	require.IsType(t, &rsa.PrivateKey{}, config.IdentityProviders.OIDC.JSONWebKeys[2].Key)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[2].Key.(*rsa.PrivateKey).D)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[2].Key.(*rsa.PrivateKey).N)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[2].Key.(*rsa.PrivateKey).E)
	assert.Equal(t, 512, config.IdentityProviders.OIDC.JSONWebKeys[2].Key.(*rsa.PrivateKey).Size())
}

func TestShouldDisableOIDCModern(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_oidc_modern.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)

	val.Clear()

	validator.ValidateConfiguration(config, val)

	assert.Len(t, val.Warnings(), 0)

	require.Len(t, val.Errors(), 3)
	assert.EqualError(t, val.Errors()[0], "access_control: rule #14 (domain 'dev.example.com'): option 'subject' with value 'oauth2:client:not_a_client' is invalid: the client id 'not_a_client' does not belong to a registered client")
	assert.Regexp(t, regexp.MustCompile(`^identity_providers: oidc: jwks: key #1 with key id 'keya': option 'certificate_chain' produced an error during validation of the chain: certificate #1 in chain is invalid after 1713180174 but the time is \d+$`), val.Errors()[1].Error())
	assert.Regexp(t, regexp.MustCompile(`^identity_providers: oidc: jwks: key #2 with key id 'ec521': option 'certificate_chain' produced an error during validation of the chain: certificate #1 in chain is invalid after 1713180101 but the time is \d+$`), val.Errors()[2].Error())

	require.Len(t, config.IdentityProviders.OIDC.JSONWebKeys, 3)
	require.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[0].Key)
	require.IsType(t, &rsa.PrivateKey{}, config.IdentityProviders.OIDC.JSONWebKeys[0].Key)
	assert.Equal(t, "sig", config.IdentityProviders.OIDC.JSONWebKeys[0].Use)
	assert.Equal(t, "RS256", config.IdentityProviders.OIDC.JSONWebKeys[0].Algorithm)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[0].Key.(*rsa.PrivateKey).D)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[0].Key.(*rsa.PrivateKey).N)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[0].Key.(*rsa.PrivateKey).E)
	assert.Equal(t, 256, config.IdentityProviders.OIDC.JSONWebKeys[0].Key.(*rsa.PrivateKey).Size())
	require.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[0].CertificateChain)
	assert.True(t, config.IdentityProviders.OIDC.JSONWebKeys[0].CertificateChain.HasCertificates())

	require.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[1].Key)
	require.IsType(t, &ecdsa.PrivateKey{}, config.IdentityProviders.OIDC.JSONWebKeys[1].Key)
	assert.Equal(t, "sig", config.IdentityProviders.OIDC.JSONWebKeys[1].Use)
	assert.Equal(t, "ES512", config.IdentityProviders.OIDC.JSONWebKeys[1].Algorithm)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[1].Key.(*ecdsa.PrivateKey).D)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[1].Key.(*ecdsa.PrivateKey).Y)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[1].Key.(*ecdsa.PrivateKey).X)
	assert.Equal(t, elliptic.P521(), config.IdentityProviders.OIDC.JSONWebKeys[1].Key.(*ecdsa.PrivateKey).Curve)
	require.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[1].CertificateChain)
	assert.True(t, config.IdentityProviders.OIDC.JSONWebKeys[1].CertificateChain.HasCertificates())

	require.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[2].Key)
	assert.Equal(t, "sig", config.IdentityProviders.OIDC.JSONWebKeys[2].Use)
	assert.Equal(t, "RS256", config.IdentityProviders.OIDC.JSONWebKeys[2].Algorithm)
	require.IsType(t, &rsa.PrivateKey{}, config.IdentityProviders.OIDC.JSONWebKeys[2].Key)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[2].Key.(*rsa.PrivateKey).D)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[2].Key.(*rsa.PrivateKey).N)
	assert.NotNil(t, config.IdentityProviders.OIDC.JSONWebKeys[2].Key.(*rsa.PrivateKey).E)
	assert.Equal(t, 512, config.IdentityProviders.OIDC.JSONWebKeys[2].Key.(*rsa.PrivateKey).Size())
}

func TestShouldConfigureConsent(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_oidc.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	require.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	require.Len(t, config.IdentityProviders.OIDC.Clients, 1)
	assert.Equal(t, config.IdentityProviders.OIDC.Clients[0].ConsentMode, "explicit")
	assert.Equal(t, "none", config.IdentityProviders.OIDC.Clients[0].UserinfoSignedResponseAlg)
}

func TestShouldValidateAndRaiseErrorsOnBadConfiguration(t *testing.T) {
	testSetEnv(t, "SESSION_SECRET", "abc")
	testSetEnv(t, "STORAGE_MYSQL_PASSWORD", "abc")
	testSetEnv(t, "IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET", "abc")
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc")

	val := schema.NewStructValidator()
	keys, c, err := Load(val, NewDefaultSources([]string{"./test_resources/config_bad_keys.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	require.Len(t, val.Warnings(), 1)
	assert.EqualError(t, val.Warnings()[0], "configuration key 'logs_level' is deprecated in 4.7.0 and has been replaced by 'log.level': you are not required to make any changes as this has been automatically mapped for you, but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0")

	require.Len(t, val.Errors(), 1)
	assert.EqualError(t, val.Errors()[0], "configuration key not expected: loggy_file")

	assert.Equal(t, "debug", c.Log.Level)
}

func TestShouldValidateDeprecatedEnvNames(t *testing.T) {
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_URL", "ldap://from-env")

	val := schema.NewStructValidator()
	keys, c, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	require.Len(t, val.Warnings(), 1)

	assert.EqualError(t, val.Warnings()[0], "configuration key 'authentication_backend.ldap.url' is deprecated in 4.38.0 and has been replaced by 'authentication_backend.ldap.address': this has not been automatically mapped for you because the replacement key also exists and you will need to adjust your configuration to remove this message")

	assert.Equal(t, "ldap://127.0.0.1:389", c.AuthenticationBackend.LDAP.Address.String())
}

func TestShouldValidateDeprecatedEnvNamesWithDeprecatedKeys(t *testing.T) {
	testSetEnv(t, "AUTHENTICATION_BACKEND_LDAP_URL", "ldap://from-env")
	testSetEnv(t, "JWT_SECRET_FILE", "./test_resources/example_secret")

	val := schema.NewStructValidator()
	keys, c, err := Load(val, NewDefaultSources([]string{"./test_resources/config.deprecated.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)

	warnings := val.Warnings()
	require.Len(t, warnings, 6)

	sort.Sort(utils.ErrSliceSortAlphabetical(warnings))

	assert.EqualError(t, warnings[0], "configuration key 'authentication_backend.ldap.url' is deprecated in 4.38.0 and has been replaced by 'authentication_backend.ldap.address': you are not required to make any changes as this has been automatically mapped for you, but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0")
	assert.EqualError(t, warnings[1], "configuration key 'jwt_secret' is deprecated in 4.38.0 and has been replaced by 'identity_validation.reset_password.jwt_secret': you are not required to make any changes as this has been automatically mapped for you, but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0")
	assert.EqualError(t, warnings[2], "configuration keys 'notifier.smtp.host' and 'notifier.smtp.port' are deprecated in 4.38.0 and has been replaced by 'notifier.smtp.address' in the format of '[tcp://]<hostname>[:<port>]': you are not required to make any changes as this has been automatically mapped for you to the value 'smtp://127.0.0.47:2025', but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0")
	assert.EqualError(t, warnings[3], "configuration keys 'server.host', 'server.port', and 'server.path' are deprecated in 4.38.0 and has been replaced by 'server.address' in the format of '[tcp[(4|6)]://]<hostname>[:<port>][/<path>]' or 'tcp[(4|6)://][hostname]:<port>[/<path>]': you are not required to make any changes as this has been automatically mapped for you to the value 'tcp://127.0.0.44:90/abc', but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0")
	assert.EqualError(t, warnings[4], "configuration keys 'storage.mysql.host' and 'storage.mysql.port' are deprecated in 4.38.0 and has been replaced by 'storage.mysql.address' in the format of '[tcp://]<hostname>[:<port>]': you are not required to make any changes as this has been automatically mapped for you to the value 'tcp://127.0.0.45:13306', but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0")
	assert.EqualError(t, warnings[5], "configuration keys 'storage.postgres.host' and 'storage.postgres.port' are deprecated in 4.38.0 and has been replaced by 'storage.postgres.address' in the format of '[tcp://]<hostname>[:<port>]': you are not required to make any changes as this has been automatically mapped for you to the value 'tcp://127.0.0.46:15432', but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0")

	assert.Equal(t, "tcp://127.0.0.44:90/abc", c.Server.Address.String())
	assert.Equal(t, "tcp://127.0.0.45:13306", c.Storage.MySQL.Address.String())
	assert.Equal(t, "tcp://127.0.0.46:15432", c.Storage.PostgreSQL.Address.String())
	assert.Equal(t, "smtp://127.0.0.47:2025", c.Notifier.SMTP.Address.String())
	assert.Equal(t, "ldap://from-env:389", c.AuthenticationBackend.LDAP.Address.String())
	assert.Equal(t, "example_secret value", c.IdentityValidation.ResetPassword.JWTSecret)
}

func TestShouldRaiseErrOnInvalidNotifierSMTPSender(t *testing.T) {
	val := schema.NewStructValidator()
	keys, _, err := Load(val, NewDefaultSources([]string{"./test_resources/config_smtp_sender_invalid.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	assert.EqualError(t, val.Errors()[0], "error occurred during unmarshalling configuration: decoding failed due to the following error(s):\n\n'notifier.smtp.sender' could not decode 'admin' to a mail.Address (RFC5322): mail: missing '@' or angle-addr")
}

func TestShouldHandleErrInvalidatorWhenSMTPSenderBlank(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_smtp_sender_blank.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

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

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "", config.Notifier.SMTP.Sender.Name)
	assert.Equal(t, "admin@example.com", config.Notifier.SMTP.Sender.Address)
}

func TestShouldDecodeServerTLS(t *testing.T) {
	testSetEnv(t, "SERVER_TLS_KEY", "abc")
	testSetEnv(t, "SERVER_TLS_CERTIFICATE", "123")
	testSetEnv(t, "SERVER_TLS_CLIENT_CERTIFICATES", "abc,123")

	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "abc", config.Server.TLS.Key)
	assert.Equal(t, "123", config.Server.TLS.Certificate)
	assert.Equal(t, []string{"abc", "123"}, config.Server.TLS.ClientCertificates)
}

func TestShouldDecodeSMTPSenderWithName(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_alt.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "Admin", config.Notifier.SMTP.Sender.Name)
	assert.Equal(t, "admin@example.com", config.Notifier.SMTP.Sender.Address)
	assert.Equal(t, schema.RememberMeDisabled, config.Session.RememberMe)
}

func TestShouldConfigureRefreshIntervalAlways(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_alt.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	require.NotNil(t, config.AuthenticationBackend.RefreshInterval)
	assert.False(t, config.AuthenticationBackend.RefreshInterval.Never())
	assert.True(t, config.AuthenticationBackend.RefreshInterval.Always())
}

func TestShouldConfigureRefreshIntervalDefault(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.no-refresh.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	validator.ValidateAuthenticationBackend(&config.AuthenticationBackend, val)

	require.NotNil(t, config.AuthenticationBackend.RefreshInterval)
	assert.False(t, config.AuthenticationBackend.RefreshInterval.Always())
	assert.False(t, config.AuthenticationBackend.RefreshInterval.Never())
	assert.Equal(t, time.Minute*5, config.AuthenticationBackend.RefreshInterval.Value())
}

func TestShouldParseRegex(t *testing.T) {
	val := schema.NewStructValidator()
	keys, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config_domain_regex.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

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

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	assert.EqualError(t, val.Errors()[0], "error occurred during unmarshalling configuration: decoding failed due to the following error(s):\n\n'access_control.rules[0].domain_regex[0]' could not decode '^\\K(public|public2).example.com$' to a regexp.Regexp: error parsing regexp: invalid escape sequence: `\\K`")
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

	assert.EqualError(t, val.Warnings()[0], "configuration keys 'server.host', 'server.port', and 'server.path' are deprecated in 4.38.0 and has been replaced by 'server.address' in the format of '[tcp[(4|6)]://]<hostname>[:<port>][/<path>]' or 'tcp[(4|6)://][hostname]:<port>[/<path>]': you are not required to make any changes as this has been automatically mapped for you to the value 'tcp://:9091/', but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0")
}

func testSetEnv(t *testing.T, key, value string) {
	t.Helper()

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

func TestConfigurationDefinitions(t *testing.T) {
	var (
		definitions *schema.Definitions
		err         error
	)

	val := schema.NewStructValidator()

	config := &schema.Configuration{}

	sources := NewDefaultSourcesWithDefaults([]string{"./test_resources/config_with_definitions.yml"}, nil, DefaultEnvPrefix, DefaultEnvDelimiter, nil)

	definitions, err = LoadDefinitions(val, sources...)

	require.NoError(t, err)

	var keys []string

	keys, err = LoadAdvanced(val, "", config, definitions, sources...)

	assert.Len(t, val.Warnings(), 0)
	assert.Len(t, val.Errors(), 0)

	val.Clear()

	require.NoError(t, err)

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Warnings(), 0)
	assert.Len(t, val.Errors(), 0)

	require.Len(t, config.Definitions.Network, 2)

	require.Contains(t, config.Definitions.Network, "lan")
	require.Len(t, config.Definitions.Network["lan"], 2)
	assert.Equal(t, "192.168.1.0/24", config.Definitions.Network["lan"][0].String())
	assert.Equal(t, "192.168.2.0/24", config.Definitions.Network["lan"][1].String())

	require.Contains(t, config.Definitions.Network, "abc")
	require.Len(t, config.Definitions.Network["abc"], 2)
	assert.Equal(t, "192.168.3.0/24", config.Definitions.Network["abc"][0].String())
	assert.Equal(t, "192.168.4.0/24", config.Definitions.Network["abc"][1].String())

	require.Len(t, config.AccessControl.Rules, 12)
	require.Len(t, config.AccessControl.Rules[1].Networks, 5)

	assert.Equal(t, "192.168.0.0/24", config.AccessControl.Rules[1].Networks[0].String())
	assert.Equal(t, "192.168.1.0/24", config.AccessControl.Rules[1].Networks[1].String())
	assert.Equal(t, "192.168.2.0/24", config.AccessControl.Rules[1].Networks[2].String())
	assert.Equal(t, "192.168.3.0/24", config.AccessControl.Rules[1].Networks[3].String())
	assert.Equal(t, "192.168.4.0/24", config.AccessControl.Rules[1].Networks[4].String())

	assert.Contains(t, config.Definitions.UserAttributes, "example")
}

func TestConfigurationTemplate(t *testing.T) {
	buf := &bytes.Buffer{}

	setup := func() {
		f, err := os.Open("config.template.yml")
		require.NoError(t, err)

		defer f.Close()

		lints := regexp.MustCompile(`^(\s+)?# yamllint`)
		doc := regexp.MustCompile(`^\s+?## `)
		commented := regexp.MustCompile(`^(\s+)?# (.*)$`)
		uncommented := regexp.MustCompile(`^(\s+)?\w+`)
		ignore := regexp.MustCompile(`^(\s+)?# host: '/var/run/redis/redis.sock'`)
		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			line := scanner.Bytes()
			if doc.Match(line) || lints.Match(line) || ignore.Match(line) {
				continue
			}

			if commented.Match(line) {
				buf.Write(commented.ReplaceAll(line, []byte("$1$2")))
				buf.WriteString("\n")
			} else if uncommented.Match(line) {
				buf.Write(line)
				buf.WriteString("\n")
			}
		}
	}

	setup()

	config := buf.Bytes()

	ca, _, cert, key := MustLoadCryptoSet("RSA", false, "2048")

	publickey, err := os.ReadFile("./test_resources/crypto/rsa.pair.2048.public.pem")
	require.NoError(t, err)

	certchain := regexp.MustCompile(`[\t ]+-----BEGIN CERTIFICATE-----\n(?P<Padding>[\t ]+)\.\.\.\n[\t ]+-----END CERTIFICATE-----\n[\t ]+-----BEGIN CERTIFICATE-----\n[\t ]+\.\.\.\n[\t ]+-----END CERTIFICATE-----\n`)

	for _, match := range certchain.FindAllSubmatch(config, -1) {
		padding := match[certchain.SubexpIndex("Padding")]

		before := string(padding) + "-----BEGIN CERTIFICATE-----\n" + string(padding) + pemMaterialPlaceholder + string(padding) + "-----END CERTIFICATE-----\n"
		before += before

		after := strings.ReplaceAll(strings.TrimSuffix(string(padding)+cert, "\n"), "\n", "\n"+string(padding)) + "\n" + string(padding)
		after += strings.ReplaceAll(strings.TrimSuffix(ca, "\n"), "\n", "\n"+string(padding)) + "\n"

		config = bytes.ReplaceAll(config, []byte(before), []byte(after))
	}

	pem := regexp.MustCompile(`[\t ]+-----BEGIN (?P<BlockType>(RSA )?(?P<KeyType>PRIVATE|PUBLIC) KEY)-----\n(?P<Padding>[\t ]+)\.\.\.\n[\t ]+-----(END (RSA )?(PUBLIC|PRIVATE) KEY)-----\n`)

	for _, match := range pem.FindAllSubmatch(config, -1) {
		padding := match[pem.SubexpIndex("Padding")]
		blocktype := match[pem.SubexpIndex("BlockType")]
		keytype := match[pem.SubexpIndex("KeyType")]

		material := key

		if string(keytype) == "PUBLIC" {
			material = string(publickey)
		}

		before := string(padding) + "-----BEGIN " + string(blocktype) + pemEnd + string(padding) + pemMaterialPlaceholder + string(padding) + "-----END " + string(blocktype) + pemEnd

		after := strings.ReplaceAll(strings.TrimSuffix(string(padding)+material, "\n"), "\n", "\n"+string(padding)) + "\n"

		config = bytes.ReplaceAll(config, []byte(before), []byte(after))
	}

	val := schema.NewStructValidator()

	var (
		keys        []string
		definitions *schema.Definitions
	)

	c := &schema.Configuration{}

	src := NewBytesSource(config)

	definitions, err = LoadDefinitions(val, src)

	require.NoError(t, err)

	keys, err = LoadAdvanced(val, "", c, definitions, src)

	require.NoError(t, err)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	val.Clear()

	validator.ValidateKeys(keys, GetMultiKeyMappedDeprecationKeys(), DefaultEnvPrefix, val)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
}

func MustLoadCryptoSet(alg string, legacy bool, extra ...string) (certCA, keyCA, cert, key string) {
	extraAlt := make([]string, len(extra))

	copy(extraAlt, extra)

	if legacy {
		extraAlt = append(extraAlt, "legacy")
	}

	return MustLoadCryptoRaw(true, alg, "crt", extra...), MustLoadCryptoRaw(true, alg, "pem", extra...), MustLoadCryptoRaw(false, alg, "crt", extraAlt...), MustLoadCryptoRaw(false, alg, "pem", extraAlt...)
}

func MustLoadCryptoRaw(ca bool, alg, ext string, extra ...string) string {
	var fparts []string

	if ca {
		fparts = append(fparts, "ca")
	}

	fparts = append(fparts, strings.ToLower(alg))

	if len(extra) != 0 {
		fparts = append(fparts, extra...)
	}

	var (
		data []byte
		err  error
	)
	if data, err = os.ReadFile(fmt.Sprintf(pathCrypto, strings.Join(fparts, "."), ext)); err != nil {
		panic(err)
	}

	return string(data)
}

const (
	pathCrypto = "./test_resources/crypto/%s.%s"
)
