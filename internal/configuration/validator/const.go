package validator

const (
	errFmtSessionSecretRedisProvider      = "The session secret must be set when using the %s session provider"
	errFmtSessionRedisPortRange           = "The port must be between 1 and 65535 for the %s session provider"
	errFmtSessionRedisHostRequired        = "The host must be provided when using the %s session provider"
	errFmtSessionRedisHostOrNodesRequired = "Either the host or a node must be provided when using the %s session provider"
	errFmtReplacedConfigurationKey        = "invalid configuration key `%s` was replaced by `%s`"

	errFileHashing  = "config key incorrect: authentication_backend.file.hashing should be authentication_backend.file.password"
	errFilePHashing = "config key incorrect: authentication_backend.file.password_hashing should be authentication_backend.file.password"
	errFilePOptions = "config key incorrect: authentication_backend.file.password_options should be authentication_backend.file.password"

	bypassPolicy    = "bypass"
	oneFactorPolicy = "one_factor"
	twoFactorPolicy = "two_factor"
	denyPolicy      = "deny"

	argon2id = "argon2id"
	sha512   = "sha512"

	schemeLDAP  = "ldap"
	schemeLDAPS = "ldaps"

	testBadTimer      = "-1"
	testInvalidPolicy = "invalid"
	testJWTSecret     = "a_secret"
	testLDAPBaseDN    = "base_dn"
	testLDAPPassword  = "password"
	testLDAPURL       = "ldap://ldap"
	testLDAPUser      = "user"
	testModeDisabled  = "disable"
	testTLSCert       = "/tmp/cert.pem"
	testTLSKey        = "/tmp/key.pem"

	errAccessControlInvalidPolicyWithSubjects = "Policy [bypass] for rule #%d domain %s with subjects %s is invalid. It is " +
		"not supported to configure both policy bypass and subjects. For more information see: " +
		"https://www.authelia.com/docs/configuration/access-control.html#combining-subjects-and-the-bypass-policy"
)

var validRequestMethods = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "TRACE", "CONNECT", "OPTIONS"}

// SecretNames contains a map of secret names.
var SecretNames = map[string]string{
	"JWTSecret":             "jwt_secret",
	"SessionSecret":         "session.secret",
	"DUOSecretKey":          "duo_api.secret_key",
	"RedisPassword":         "session.redis.password",
	"RedisSentinelPassword": "session.redis.high_availability.sentinel_password",
	"LDAPPassword":          "authentication_backend.ldap.password",
	"SMTPPassword":          "notifier.smtp.password",
	"MySQLPassword":         "storage.mysql.password",
	"PostgreSQLPassword":    "storage.postgres.password",
}

// validKeys is a list of valid keys that are not secret names. For the sake of consistency please place any secret in
// the secret names map and reuse it in relevant sections.
var validKeys = []string{
	// Root Keys.
	"host",
	"port",
	"log_level",
	"log_format",
	"log_file_path",
	"default_redirection_url",
	"theme",
	"tls_key",
	"tls_cert",
	"certificates_directory",

	// Server Keys.
	"server.read_buffer_size",
	"server.write_buffer_size",
	"server.path",

	// TOTP Keys.
	"totp.issuer",
	"totp.period",
	"totp.skew",

	// Access Control Keys.
	"access_control.rules",
	"access_control.default_policy",
	"access_control.networks",

	// Session Keys.
	"session.name",
	"session.expiration",
	"session.inactivity",
	"session.remember_me_duration",
	"session.domain",

	// Redis Session Keys.
	"session.redis.host",
	"session.redis.port",
	"session.redis.username",
	"session.redis.database_index",
	"session.redis.maximum_active_connections",
	"session.redis.minimum_idle_connections",
	"session.redis.tls.minimum_version",
	"session.redis.tls.skip_verify",
	"session.redis.tls.server_name",
	"session.redis.high_availability.sentinel_name",
	"session.redis.high_availability.nodes",
	"session.redis.high_availability.route_by_latency",
	"session.redis.high_availability.route_randomly",
	"session.redis.timeouts.dial",
	"session.redis.timeouts.idle",
	"session.redis.timeouts.pool",
	"session.redis.timeouts.read",
	"session.redis.timeouts.write",

	// Local Storage Keys.
	"storage.local.path",

	// MySQL Storage Keys.
	"storage.mysql.host",
	"storage.mysql.port",
	"storage.mysql.database",
	"storage.mysql.username",

	// PostgreSQL Storage Keys.
	"storage.postgres.host",
	"storage.postgres.port",
	"storage.postgres.database",
	"storage.postgres.username",
	"storage.postgres.sslmode",

	// FileSystem Notifier Keys.
	"notifier.filesystem.filename",
	"notifier.disable_startup_check",

	// SMTP Notifier Keys.
	"notifier.smtp.username",
	"notifier.smtp.host",
	"notifier.smtp.port",
	"notifier.smtp.identifier",
	"notifier.smtp.sender",
	"notifier.smtp.subject",
	"notifier.smtp.startup_check_address",
	"notifier.smtp.disable_require_tls",
	"notifier.smtp.disable_html_emails",
	"notifier.smtp.tls.minimum_version",
	"notifier.smtp.tls.skip_verify",
	"notifier.smtp.tls.server_name",

	// Regulation Keys.
	"regulation.max_retries",
	"regulation.find_time",
	"regulation.ban_time",

	// DUO API Keys.
	"duo_api.hostname",
	"duo_api.integration_key",

	// Authentication Backend Keys.
	"authentication_backend.disable_reset_password",
	"authentication_backend.refresh_interval",

	// LDAP Authentication Backend Keys.
	"authentication_backend.ldap.implementation",
	"authentication_backend.ldap.url",
	"authentication_backend.ldap.base_dn",
	"authentication_backend.ldap.username_attribute",
	"authentication_backend.ldap.additional_users_dn",
	"authentication_backend.ldap.users_filter",
	"authentication_backend.ldap.additional_groups_dn",
	"authentication_backend.ldap.groups_filter",
	"authentication_backend.ldap.group_name_attribute",
	"authentication_backend.ldap.mail_attribute",
	"authentication_backend.ldap.display_name_attribute",
	"authentication_backend.ldap.user",
	"authentication_backend.ldap.start_tls",
	"authentication_backend.ldap.tls.minimum_version",
	"authentication_backend.ldap.tls.skip_verify",
	"authentication_backend.ldap.tls.server_name",

	// File Authentication Backend Keys.
	"authentication_backend.file.path",
	"authentication_backend.file.password.algorithm",
	"authentication_backend.file.password.iterations",
	"authentication_backend.file.password.key_length",
	"authentication_backend.file.password.salt_length",
	"authentication_backend.file.password.memory",
	"authentication_backend.file.password.parallelism",
}

var replacedKeys = map[string]string{
	"authentication_backend.ldap.skip_verify":         "authentication_backend.ldap.tls.skip_verify",
	"authentication_backend.ldap.minimum_tls_version": "authentication_backend.ldap.tls.minimum_version",
	"notifier.smtp.disable_verify_cert":               "notifier.smtp.tls.skip_verify",
	"logs_file_path":                                  "log_file",
	"logs_level":                                      "log_level",
}

var specificErrorKeys = map[string]string{
	"google_analytics": "config key removed: google_analytics - this functionality has been deprecated",
	"notifier.smtp.trusted_cert": "invalid configuration key `notifier.smtp.trusted_cert` it has been removed, " +
		"option has been replaced by the global option `certificates_directory`",

	"authentication_backend.file.password_options.algorithm":   errFilePOptions,
	"authentication_backend.file.password_options.iterations":  errFilePOptions,
	"authentication_backend.file.password_options.key_length":  errFilePOptions,
	"authentication_backend.file.password_options.salt_length": errFilePOptions,
	"authentication_backend.file.password_options.memory":      errFilePOptions,
	"authentication_backend.file.password_options.parallelism": errFilePOptions,
	"authentication_backend.file.password_hashing.algorithm":   errFilePHashing,
	"authentication_backend.file.password_hashing.iterations":  errFilePHashing,
	"authentication_backend.file.password_hashing.key_length":  errFilePHashing,
	"authentication_backend.file.password_hashing.salt_length": errFilePHashing,
	"authentication_backend.file.password_hashing.memory":      errFilePHashing,
	"authentication_backend.file.password_hashing.parallelism": errFilePHashing,
	"authentication_backend.file.hashing.algorithm":            errFileHashing,
	"authentication_backend.file.hashing.iterations":           errFileHashing,
	"authentication_backend.file.hashing.key_length":           errFileHashing,
	"authentication_backend.file.hashing.salt_length":          errFileHashing,
	"authentication_backend.file.hashing.memory":               errFileHashing,
	"authentication_backend.file.hashing.parallelism":          errFileHashing,
}
