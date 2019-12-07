package storage

// Keep table names in lower case because some DB does not support upper case.
var preferencesTableName = "user_preferences"
var identityVerificationTokensTableName = "identity_verification_tokens"
var totpSecretsTableName = "totp_secrets"
var u2fDeviceHandlesTableName = "u2f_devices"
var authenticationLogsTableName = "authentication_logs"
