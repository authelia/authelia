// Code generated by go generate. DO NOT EDIT.
//
// Run the following command to generate this file:
// 		go run ./cmd/authelia-gen code keys
//

package schema

// Keys is a list of valid schema keys detected by reflecting over a schema.Configuration struct.
var Keys = []string{
	"theme",
	"certificates_directory",
	"default_2fa_method",
	"log.level",
	"log.format",
	"log.file_path",
	"log.keep_stdout",
	"identity_providers.oidc.hmac_secret",
	"identity_providers.oidc.issuer_private_keys",
	"identity_providers.oidc.issuer_private_keys[].key_id",
	"identity_providers.oidc.issuer_private_keys[].use",
	"identity_providers.oidc.issuer_private_keys[].algorithm",
	"identity_providers.oidc.issuer_private_keys[].key",
	"identity_providers.oidc.issuer_private_keys[].certificate_chain",
	"identity_providers.oidc.issuer_certificate_chain",
	"identity_providers.oidc.issuer_private_key",
	"identity_providers.oidc.enable_client_debug_messages",
	"identity_providers.oidc.minimum_parameter_entropy",
	"identity_providers.oidc.enforce_pkce",
	"identity_providers.oidc.enable_pkce_plain_challenge",
	"identity_providers.oidc.enable_jwt_access_token_stateless_introspection",
	"identity_providers.oidc.pushed_authorizations.enforce",
	"identity_providers.oidc.pushed_authorizations.context_lifespan",
	"identity_providers.oidc.cors.endpoints",
	"identity_providers.oidc.cors.allowed_origins",
	"identity_providers.oidc.cors.allowed_origins_from_client_redirect_uris",
	"identity_providers.oidc.clients",
	"identity_providers.oidc.clients[].id",
	"identity_providers.oidc.clients[].description",
	"identity_providers.oidc.clients[].secret",
	"identity_providers.oidc.clients[].sector_identifier",
	"identity_providers.oidc.clients[].public",
	"identity_providers.oidc.clients[].redirect_uris",
	"identity_providers.oidc.clients[].audience",
	"identity_providers.oidc.clients[].scopes",
	"identity_providers.oidc.clients[].grant_types",
	"identity_providers.oidc.clients[].response_types",
	"identity_providers.oidc.clients[].response_modes",
	"identity_providers.oidc.clients[].authorization_policy",
	"identity_providers.oidc.clients[].lifespan",
	"identity_providers.oidc.clients[].requested_audience_mode",
	"identity_providers.oidc.clients[].consent_mode",
	"identity_providers.oidc.clients[].pre_configured_consent_duration",
	"identity_providers.oidc.clients[].enforce_par",
	"identity_providers.oidc.clients[].enforce_pkce",
	"identity_providers.oidc.clients[].pkce_challenge_method",
	"identity_providers.oidc.clients[].authorization_signed_response_alg",
	"identity_providers.oidc.clients[].authorization_signed_response_key_id",
	"identity_providers.oidc.clients[].id_token_signed_response_alg",
	"identity_providers.oidc.clients[].id_token_signed_response_key_id",
	"identity_providers.oidc.clients[].access_token_signed_response_alg",
	"identity_providers.oidc.clients[].access_token_signed_response_key_id",
	"identity_providers.oidc.clients[].userinfo_signed_response_alg",
	"identity_providers.oidc.clients[].userinfo_signed_response_key_id",
	"identity_providers.oidc.clients[].introspection_signed_response_alg",
	"identity_providers.oidc.clients[].introspection_signed_response_key_id",
	"identity_providers.oidc.clients[].request_object_signing_alg",
	"identity_providers.oidc.clients[].token_endpoint_auth_signing_alg",
	"identity_providers.oidc.clients[].token_endpoint_auth_method",
	"identity_providers.oidc.clients[].public_keys.uri",
	"identity_providers.oidc.clients[].public_keys.values",
	"identity_providers.oidc.clients[].public_keys.values[].key_id",
	"identity_providers.oidc.clients[].public_keys.values[].use",
	"identity_providers.oidc.clients[].public_keys.values[].algorithm",
	"identity_providers.oidc.clients[].public_keys.values[].key",
	"identity_providers.oidc.clients[].public_keys.values[].certificate_chain",
	"identity_providers.oidc.clients[]",
	"identity_providers.oidc.authorization_policies",
	"identity_providers.oidc.authorization_policies.*.default_policy",
	"identity_providers.oidc.authorization_policies.*.rules",
	"identity_providers.oidc.authorization_policies.*.rules[].policy",
	"identity_providers.oidc.authorization_policies.*.rules[].subject",
	"identity_providers.oidc.lifespans.access_token",
	"identity_providers.oidc.lifespans.authorize_code",
	"identity_providers.oidc.lifespans.id_token",
	"identity_providers.oidc.lifespans.refresh_token",
	"identity_providers.oidc.lifespans.jwt_secured_authorization",
	"identity_providers.oidc.lifespans.custom",
	"identity_providers.oidc.lifespans.custom.*.access_token",
	"identity_providers.oidc.lifespans.custom.*.authorize_code",
	"identity_providers.oidc.lifespans.custom.*.id_token",
	"identity_providers.oidc.lifespans.custom.*.refresh_token",
	"identity_providers.oidc.lifespans.custom.*.grants.authorize_code.access_token",
	"identity_providers.oidc.lifespans.custom.*.grants.authorize_code.authorize_code",
	"identity_providers.oidc.lifespans.custom.*.grants.authorize_code.id_token",
	"identity_providers.oidc.lifespans.custom.*.grants.authorize_code.refresh_token",
	"identity_providers.oidc.lifespans.custom.*.grants.implicit.access_token",
	"identity_providers.oidc.lifespans.custom.*.grants.implicit.authorize_code",
	"identity_providers.oidc.lifespans.custom.*.grants.implicit.id_token",
	"identity_providers.oidc.lifespans.custom.*.grants.implicit.refresh_token",
	"identity_providers.oidc.lifespans.custom.*.grants.client_credentials.access_token",
	"identity_providers.oidc.lifespans.custom.*.grants.client_credentials.authorize_code",
	"identity_providers.oidc.lifespans.custom.*.grants.client_credentials.id_token",
	"identity_providers.oidc.lifespans.custom.*.grants.client_credentials.refresh_token",
	"identity_providers.oidc.lifespans.custom.*.grants.refresh_token.access_token",
	"identity_providers.oidc.lifespans.custom.*.grants.refresh_token.authorize_code",
	"identity_providers.oidc.lifespans.custom.*.grants.refresh_token.id_token",
	"identity_providers.oidc.lifespans.custom.*.grants.refresh_token.refresh_token",
	"identity_providers.oidc.lifespans.custom.*.grants.jwt_bearer.access_token",
	"identity_providers.oidc.lifespans.custom.*.grants.jwt_bearer.authorize_code",
	"identity_providers.oidc.lifespans.custom.*.grants.jwt_bearer.id_token",
	"identity_providers.oidc.lifespans.custom.*.grants.jwt_bearer.refresh_token",
	"identity_providers.oidc",
	"authentication_backend.password_reset.disable",
	"authentication_backend.password_reset.custom_url",
	"authentication_backend.refresh_interval",
	"authentication_backend.file.path",
	"authentication_backend.file.watch",
	"authentication_backend.file.password.algorithm",
	"authentication_backend.file.password.argon2.variant",
	"authentication_backend.file.password.argon2.iterations",
	"authentication_backend.file.password.argon2.memory",
	"authentication_backend.file.password.argon2.parallelism",
	"authentication_backend.file.password.argon2.key_length",
	"authentication_backend.file.password.argon2.salt_length",
	"authentication_backend.file.password.sha2crypt.variant",
	"authentication_backend.file.password.sha2crypt.iterations",
	"authentication_backend.file.password.sha2crypt.salt_length",
	"authentication_backend.file.password.pbkdf2.variant",
	"authentication_backend.file.password.pbkdf2.iterations",
	"authentication_backend.file.password.pbkdf2.salt_length",
	"authentication_backend.file.password.bcrypt.variant",
	"authentication_backend.file.password.bcrypt.cost",
	"authentication_backend.file.password.scrypt.iterations",
	"authentication_backend.file.password.scrypt.block_size",
	"authentication_backend.file.password.scrypt.parallelism",
	"authentication_backend.file.password.scrypt.key_length",
	"authentication_backend.file.password.scrypt.salt_length",
	"authentication_backend.file.password.iterations",
	"authentication_backend.file.password.memory",
	"authentication_backend.file.password.parallelism",
	"authentication_backend.file.password.key_length",
	"authentication_backend.file.password.salt_length",
	"authentication_backend.file.search.email",
	"authentication_backend.file.search.case_insensitive",
	"authentication_backend.ldap.address",
	"authentication_backend.ldap.implementation",
	"authentication_backend.ldap.timeout",
	"authentication_backend.ldap.start_tls",
	"authentication_backend.ldap.tls.minimum_version",
	"authentication_backend.ldap.tls.maximum_version",
	"authentication_backend.ldap.tls.skip_verify",
	"authentication_backend.ldap.tls.server_name",
	"authentication_backend.ldap.tls.private_key",
	"authentication_backend.ldap.tls.certificate_chain",
	"authentication_backend.ldap.base_dn",
	"authentication_backend.ldap.additional_users_dn",
	"authentication_backend.ldap.users_filter",
	"authentication_backend.ldap.additional_groups_dn",
	"authentication_backend.ldap.groups_filter",
	"authentication_backend.ldap.group_search_mode",
	"authentication_backend.ldap.attributes.distinguished_name",
	"authentication_backend.ldap.attributes.username",
	"authentication_backend.ldap.attributes.display_name",
	"authentication_backend.ldap.attributes.mail",
	"authentication_backend.ldap.attributes.member_of",
	"authentication_backend.ldap.attributes.group_name",
	"authentication_backend.ldap.permit_referrals",
	"authentication_backend.ldap.permit_unauthenticated_bind",
	"authentication_backend.ldap.permit_feature_detection_failure",
	"authentication_backend.ldap.user",
	"authentication_backend.ldap.password",
	"session.name",
	"session.same_site",
	"session.expiration",
	"session.inactivity",
	"session.remember_me",
	"session",
	"session.secret",
	"session.cookies",
	"session.cookies[].name",
	"session.cookies[].same_site",
	"session.cookies[].expiration",
	"session.cookies[].inactivity",
	"session.cookies[].remember_me",
	"session.cookies[]",
	"session.cookies[].domain",
	"session.cookies[].authelia_url",
	"session.cookies[].default_redirection_url",
	"session.cookies[]",
	"session.redis.host",
	"session.redis.port",
	"session.redis.username",
	"session.redis.password",
	"session.redis.database_index",
	"session.redis.maximum_active_connections",
	"session.redis.minimum_idle_connections",
	"session.redis.tls.minimum_version",
	"session.redis.tls.maximum_version",
	"session.redis.tls.skip_verify",
	"session.redis.tls.server_name",
	"session.redis.tls.private_key",
	"session.redis.tls.certificate_chain",
	"session.redis.high_availability.sentinel_name",
	"session.redis.high_availability.sentinel_username",
	"session.redis.high_availability.sentinel_password",
	"session.redis.high_availability.route_by_latency",
	"session.redis.high_availability.route_randomly",
	"session.redis.high_availability.nodes",
	"session.redis.high_availability.nodes[].host",
	"session.redis.high_availability.nodes[].port",
	"session.domain",
	"totp.disable",
	"totp.issuer",
	"totp.algorithm",
	"totp.digits",
	"totp.period",
	"totp.skew",
	"totp.secret_size",
	"totp.allowed_algorithms",
	"totp.allowed_digits",
	"totp.allowed_periods",
	"totp.disable_reuse_security_policy",
	"duo_api.disable",
	"duo_api.hostname",
	"duo_api.integration_key",
	"duo_api.secret_key",
	"duo_api.enable_self_enrollment",
	"access_control.default_policy",
	"access_control.networks",
	"access_control.networks[].name",
	"access_control.networks[].networks",
	"access_control.rules",
	"access_control.rules[].domain",
	"access_control.rules[].domain_regex",
	"access_control.rules[].policy",
	"access_control.rules[].subject",
	"access_control.rules[].networks",
	"access_control.rules[].resources",
	"access_control.rules[].methods",
	"access_control.rules[].query[][].operator",
	"access_control.rules[].query[][].key",
	"access_control.rules[].query[][].value",
	"access_control.rules[].query",
	"ntp.address",
	"ntp.version",
	"ntp.max_desync",
	"ntp.disable_startup_check",
	"ntp.disable_failure",
	"regulation.max_retries",
	"regulation.find_time",
	"regulation.ban_time",
	"storage.local.path",
	"storage.mysql.address",
	"storage.mysql.database",
	"storage.mysql.username",
	"storage.mysql.password",
	"storage.mysql.timeout",
	"storage.mysql.host",
	"storage.mysql.port",
	"storage.mysql.tls.minimum_version",
	"storage.mysql.tls.maximum_version",
	"storage.mysql.tls.skip_verify",
	"storage.mysql.tls.server_name",
	"storage.mysql.tls.private_key",
	"storage.mysql.tls.certificate_chain",
	"storage.postgres.address",
	"storage.postgres.database",
	"storage.postgres.username",
	"storage.postgres.password",
	"storage.postgres.timeout",
	"storage.postgres.host",
	"storage.postgres.port",
	"storage.postgres.schema",
	"storage.postgres.tls.minimum_version",
	"storage.postgres.tls.maximum_version",
	"storage.postgres.tls.skip_verify",
	"storage.postgres.tls.server_name",
	"storage.postgres.tls.private_key",
	"storage.postgres.tls.certificate_chain",
	"storage.postgres.ssl.mode",
	"storage.postgres.ssl.root_certificate",
	"storage.postgres.ssl.certificate",
	"storage.postgres.ssl.key",
	"storage.encryption_key",
	"notifier.disable_startup_check",
	"notifier.filesystem.filename",
	"notifier.smtp.address",
	"notifier.smtp.timeout",
	"notifier.smtp.username",
	"notifier.smtp.password",
	"notifier.smtp.identifier",
	"notifier.smtp.sender",
	"notifier.smtp.subject",
	"notifier.smtp.startup_check_address",
	"notifier.smtp.disable_require_tls",
	"notifier.smtp.disable_html_emails",
	"notifier.smtp.disable_starttls",
	"notifier.smtp.tls.minimum_version",
	"notifier.smtp.tls.maximum_version",
	"notifier.smtp.tls.skip_verify",
	"notifier.smtp.tls.server_name",
	"notifier.smtp.tls.private_key",
	"notifier.smtp.tls.certificate_chain",
	"notifier.smtp.host",
	"notifier.smtp.port",
	"notifier.template_path",
	"server.address",
	"server.asset_path",
	"server.disable_healthcheck",
	"server.tls.certificate",
	"server.tls.key",
	"server.tls.client_certificates",
	"server.headers.csp_template",
	"server.endpoints.enable_pprof",
	"server.endpoints.enable_expvars",
	"server.endpoints.authz",
	"server.endpoints.authz.*.implementation",
	"server.endpoints.authz.*.authn_strategies",
	"server.endpoints.authz.*.authn_strategies[].name",
	"server.endpoints.authz.*.authn_strategies[].schemes",
	"server.buffers.read",
	"server.buffers.write",
	"server.timeouts.read",
	"server.timeouts.write",
	"server.timeouts.idle",
	"server.host",
	"server.port",
	"server.path",
	"telemetry.metrics.enabled",
	"telemetry.metrics.address",
	"telemetry.metrics.buffers.read",
	"telemetry.metrics.buffers.write",
	"telemetry.metrics.timeouts.read",
	"telemetry.metrics.timeouts.write",
	"telemetry.metrics.timeouts.idle",
	"webauthn.disable",
	"webauthn.display_name",
	"webauthn.attestation_conveyance_preference",
	"webauthn.user_verification",
	"webauthn.timeout",
	"password_policy.standard.enabled",
	"password_policy.standard.min_length",
	"password_policy.standard.max_length",
	"password_policy.standard.require_uppercase",
	"password_policy.standard.require_lowercase",
	"password_policy.standard.require_number",
	"password_policy.standard.require_special",
	"password_policy.zxcvbn.enabled",
	"password_policy.zxcvbn.min_score",
	"privacy_policy.enabled",
	"privacy_policy.require_user_acceptance",
	"privacy_policy.policy_url",
	"identity_validation.reset_password.expiration",
	"identity_validation.reset_password.jwt_algorithm",
	"identity_validation.reset_password.jwt_secret",
	"identity_validation.elevated_session.expiration",
	"identity_validation.elevated_session.elevation_expiration",
	"identity_validation.elevated_session.characters",
	"identity_validation.elevated_session.require_second_factor",
	"identity_validation.elevated_session.skip_second_factor",
	"default_redirection_url",
}
