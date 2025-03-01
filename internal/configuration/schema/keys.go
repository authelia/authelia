// Code generated by go generate. DO NOT EDIT.
//
// Run the following command to generate this file:
// 		go run ./cmd/authelia-gen code keys
//

package schema

// Keys is a list of valid schema keys detected by reflecting over a schema.Configuration struct.
var Keys = []string{
	"access_control.default_policy",
	"access_control.networks",
	"access_control.networks[].name",
	"access_control.networks[].networks",
	"access_control.rules",
	"access_control.rules[].domain",
	"access_control.rules[].domain_regex",
	"access_control.rules[].methods",
	"access_control.rules[].networks",
	"access_control.rules[].policy",
	"access_control.rules[].query",
	"access_control.rules[].query[][].key",
	"access_control.rules[].query[][].operator",
	"access_control.rules[].query[][].value",
	"access_control.rules[].resources",
	"access_control.rules[].subject",
	"authentication_backend.file.extra_attributes",
	"authentication_backend.file.extra_attributes.*",
	"authentication_backend.file.extra_attributes.*.multi_valued",
	"authentication_backend.file.extra_attributes.*.value_type",
	"authentication_backend.file.password.algorithm",
	"authentication_backend.file.password.argon2.iterations",
	"authentication_backend.file.password.argon2.key_length",
	"authentication_backend.file.password.argon2.memory",
	"authentication_backend.file.password.argon2.parallelism",
	"authentication_backend.file.password.argon2.salt_length",
	"authentication_backend.file.password.argon2.variant",
	"authentication_backend.file.password.bcrypt.cost",
	"authentication_backend.file.password.bcrypt.variant",
	"authentication_backend.file.password.iterations",
	"authentication_backend.file.password.key_length",
	"authentication_backend.file.password.memory",
	"authentication_backend.file.password.parallelism",
	"authentication_backend.file.password.pbkdf2.iterations",
	"authentication_backend.file.password.pbkdf2.salt_length",
	"authentication_backend.file.password.pbkdf2.variant",
	"authentication_backend.file.password.salt_length",
	"authentication_backend.file.password.scrypt.block_size",
	"authentication_backend.file.password.scrypt.iterations",
	"authentication_backend.file.password.scrypt.key_length",
	"authentication_backend.file.password.scrypt.parallelism",
	"authentication_backend.file.password.scrypt.salt_length",
	"authentication_backend.file.password.sha2crypt.iterations",
	"authentication_backend.file.password.sha2crypt.salt_length",
	"authentication_backend.file.password.sha2crypt.variant",
	"authentication_backend.file.path",
	"authentication_backend.file.search.case_insensitive",
	"authentication_backend.file.search.email",
	"authentication_backend.file.watch",
	"authentication_backend.ldap.additional_groups_dn",
	"authentication_backend.ldap.additional_users_dn",
	"authentication_backend.ldap.address",
	"authentication_backend.ldap.attributes.birthdate",
	"authentication_backend.ldap.attributes.country",
	"authentication_backend.ldap.attributes.display_name",
	"authentication_backend.ldap.attributes.distinguished_name",
	"authentication_backend.ldap.attributes.extra",
	"authentication_backend.ldap.attributes.extra.*",
	"authentication_backend.ldap.attributes.extra.*.multi_valued",
	"authentication_backend.ldap.attributes.extra.*.name",
	"authentication_backend.ldap.attributes.extra.*.value_type",
	"authentication_backend.ldap.attributes.family_name",
	"authentication_backend.ldap.attributes.gender",
	"authentication_backend.ldap.attributes.given_name",
	"authentication_backend.ldap.attributes.group_name",
	"authentication_backend.ldap.attributes.locale",
	"authentication_backend.ldap.attributes.locality",
	"authentication_backend.ldap.attributes.mail",
	"authentication_backend.ldap.attributes.member_of",
	"authentication_backend.ldap.attributes.middle_name",
	"authentication_backend.ldap.attributes.nickname",
	"authentication_backend.ldap.attributes.phone_extension",
	"authentication_backend.ldap.attributes.phone_number",
	"authentication_backend.ldap.attributes.picture",
	"authentication_backend.ldap.attributes.postal_code",
	"authentication_backend.ldap.attributes.profile",
	"authentication_backend.ldap.attributes.region",
	"authentication_backend.ldap.attributes.street_address",
	"authentication_backend.ldap.attributes.username",
	"authentication_backend.ldap.attributes.website",
	"authentication_backend.ldap.attributes.zoneinfo",
	"authentication_backend.ldap.base_dn",
	"authentication_backend.ldap.group_search_mode",
	"authentication_backend.ldap.groups_filter",
	"authentication_backend.ldap.implementation",
	"authentication_backend.ldap.password",
	"authentication_backend.ldap.permit_feature_detection_failure",
	"authentication_backend.ldap.permit_referrals",
	"authentication_backend.ldap.permit_unauthenticated_bind",
	"authentication_backend.ldap.pooling.count",
	"authentication_backend.ldap.pooling.enable",
	"authentication_backend.ldap.pooling.retries",
	"authentication_backend.ldap.pooling.timeout",
	"authentication_backend.ldap.start_tls",
	"authentication_backend.ldap.timeout",
	"authentication_backend.ldap.tls.certificate_chain",
	"authentication_backend.ldap.tls.maximum_version",
	"authentication_backend.ldap.tls.minimum_version",
	"authentication_backend.ldap.tls.private_key",
	"authentication_backend.ldap.tls.server_name",
	"authentication_backend.ldap.tls.skip_verify",
	"authentication_backend.ldap.user",
	"authentication_backend.ldap.users_filter",
	"authentication_backend.password_reset.custom_url",
	"authentication_backend.password_reset.disable",
	"authentication_backend.refresh_interval",
	"certificates_directory",
	"default_2fa_method",
	"default_redirection_url",
	"definitions.network",
	"definitions.network.*",
	"definitions.user_attributes",
	"definitions.user_attributes.*",
	"definitions.user_attributes.*.expression",
	"duo_api.disable",
	"duo_api.enable_self_enrollment",
	"duo_api.hostname",
	"duo_api.integration_key",
	"duo_api.secret_key",
	"identity_providers.oidc",
	"identity_providers.oidc.authorization_policies",
	"identity_providers.oidc.authorization_policies.*",
	"identity_providers.oidc.authorization_policies.*.default_policy",
	"identity_providers.oidc.authorization_policies.*.rules",
	"identity_providers.oidc.authorization_policies.*.rules[].networks",
	"identity_providers.oidc.authorization_policies.*.rules[].policy",
	"identity_providers.oidc.authorization_policies.*.rules[].subject",
	"identity_providers.oidc.claims_policies",
	"identity_providers.oidc.claims_policies.*",
	"identity_providers.oidc.claims_policies.*.access_token",
	"identity_providers.oidc.claims_policies.*.custom_claims",
	"identity_providers.oidc.claims_policies.*.custom_claims.*",
	"identity_providers.oidc.claims_policies.*.custom_claims.*.attribute",
	"identity_providers.oidc.claims_policies.*.id_token",
	"identity_providers.oidc.claims_policies.*.id_token_audience_mode",
	"identity_providers.oidc.clients",
	"identity_providers.oidc.clients[]",
	"identity_providers.oidc.clients[].access_token_encrypted_response_alg",
	"identity_providers.oidc.clients[].access_token_encrypted_response_enc",
	"identity_providers.oidc.clients[].access_token_encrypted_response_key_id",
	"identity_providers.oidc.clients[].access_token_signed_response_alg",
	"identity_providers.oidc.clients[].access_token_signed_response_key_id",
	"identity_providers.oidc.clients[].allow_multiple_auth_methods",
	"identity_providers.oidc.clients[].audience",
	"identity_providers.oidc.clients[].authorization_encrypted_response_alg",
	"identity_providers.oidc.clients[].authorization_encrypted_response_enc",
	"identity_providers.oidc.clients[].authorization_encrypted_response_key_id",
	"identity_providers.oidc.clients[].authorization_policy",
	"identity_providers.oidc.clients[].authorization_signed_response_alg",
	"identity_providers.oidc.clients[].authorization_signed_response_key_id",
	"identity_providers.oidc.clients[].claims_policy",
	"identity_providers.oidc.clients[].client_id",
	"identity_providers.oidc.clients[].client_name",
	"identity_providers.oidc.clients[].client_secret",
	"identity_providers.oidc.clients[].consent_mode",
	"identity_providers.oidc.clients[].grant_types",
	"identity_providers.oidc.clients[].id_token_encrypted_response_alg",
	"identity_providers.oidc.clients[].id_token_encrypted_response_enc",
	"identity_providers.oidc.clients[].id_token_encrypted_response_key_id",
	"identity_providers.oidc.clients[].id_token_signed_response_alg",
	"identity_providers.oidc.clients[].id_token_signed_response_key_id",
	"identity_providers.oidc.clients[].introspection_encrypted_response_alg",
	"identity_providers.oidc.clients[].introspection_encrypted_response_enc",
	"identity_providers.oidc.clients[].introspection_encrypted_response_key_id",
	"identity_providers.oidc.clients[].introspection_endpoint_auth_method",
	"identity_providers.oidc.clients[].introspection_endpoint_auth_signing_alg",
	"identity_providers.oidc.clients[].introspection_signed_response_alg",
	"identity_providers.oidc.clients[].introspection_signed_response_key_id",
	"identity_providers.oidc.clients[].jwks",
	"identity_providers.oidc.clients[].jwks[].algorithm",
	"identity_providers.oidc.clients[].jwks[].certificate_chain",
	"identity_providers.oidc.clients[].jwks[].key",
	"identity_providers.oidc.clients[].jwks[].key_id",
	"identity_providers.oidc.clients[].jwks[].use",
	"identity_providers.oidc.clients[].jwks_uri",
	"identity_providers.oidc.clients[].lifespan",
	"identity_providers.oidc.clients[].pkce_challenge_method",
	"identity_providers.oidc.clients[].pre_configured_consent_duration",
	"identity_providers.oidc.clients[].public",
	"identity_providers.oidc.clients[].pushed_authorization_request_endpoint_auth_method",
	"identity_providers.oidc.clients[].pushed_authorization_request_endpoint_auth_signing_alg",
	"identity_providers.oidc.clients[].redirect_uris",
	"identity_providers.oidc.clients[].request_object_encryption_alg",
	"identity_providers.oidc.clients[].request_object_encryption_enc",
	"identity_providers.oidc.clients[].request_object_signing_alg",
	"identity_providers.oidc.clients[].request_uris",
	"identity_providers.oidc.clients[].requested_audience_mode",
	"identity_providers.oidc.clients[].require_pkce",
	"identity_providers.oidc.clients[].require_pushed_authorization_requests",
	"identity_providers.oidc.clients[].response_modes",
	"identity_providers.oidc.clients[].response_types",
	"identity_providers.oidc.clients[].revocation_endpoint_auth_method",
	"identity_providers.oidc.clients[].revocation_endpoint_auth_signing_alg",
	"identity_providers.oidc.clients[].scopes",
	"identity_providers.oidc.clients[].sector_identifier_uri",
	"identity_providers.oidc.clients[].token_endpoint_auth_method",
	"identity_providers.oidc.clients[].token_endpoint_auth_signing_alg",
	"identity_providers.oidc.clients[].userinfo_encrypted_response_alg",
	"identity_providers.oidc.clients[].userinfo_encrypted_response_enc",
	"identity_providers.oidc.clients[].userinfo_encrypted_response_key_id",
	"identity_providers.oidc.clients[].userinfo_signed_response_alg",
	"identity_providers.oidc.clients[].userinfo_signed_response_key_id",
	"identity_providers.oidc.cors.allowed_origins",
	"identity_providers.oidc.cors.allowed_origins_from_client_redirect_uris",
	"identity_providers.oidc.cors.endpoints",
	"identity_providers.oidc.discovery_signed_response_alg",
	"identity_providers.oidc.discovery_signed_response_key_id",
	"identity_providers.oidc.enable_client_debug_messages",
	"identity_providers.oidc.enable_jwt_access_token_stateless_introspection",
	"identity_providers.oidc.enable_pkce_plain_challenge",
	"identity_providers.oidc.enforce_pkce",
	"identity_providers.oidc.hmac_secret",
	"identity_providers.oidc.issuer_certificate_chain",
	"identity_providers.oidc.issuer_private_key",
	"identity_providers.oidc.jwks",
	"identity_providers.oidc.jwks[].algorithm",
	"identity_providers.oidc.jwks[].certificate_chain",
	"identity_providers.oidc.jwks[].key",
	"identity_providers.oidc.jwks[].key_id",
	"identity_providers.oidc.jwks[].use",
	"identity_providers.oidc.lifespans.access_token",
	"identity_providers.oidc.lifespans.authorize_code",
	"identity_providers.oidc.lifespans.custom",
	"identity_providers.oidc.lifespans.custom.*",
	"identity_providers.oidc.lifespans.custom.*.access_token",
	"identity_providers.oidc.lifespans.custom.*.authorize_code",
	"identity_providers.oidc.lifespans.custom.*.grants.authorize_code.access_token",
	"identity_providers.oidc.lifespans.custom.*.grants.authorize_code.authorize_code",
	"identity_providers.oidc.lifespans.custom.*.grants.authorize_code.id_token",
	"identity_providers.oidc.lifespans.custom.*.grants.authorize_code.refresh_token",
	"identity_providers.oidc.lifespans.custom.*.grants.client_credentials.access_token",
	"identity_providers.oidc.lifespans.custom.*.grants.client_credentials.authorize_code",
	"identity_providers.oidc.lifespans.custom.*.grants.client_credentials.id_token",
	"identity_providers.oidc.lifespans.custom.*.grants.client_credentials.refresh_token",
	"identity_providers.oidc.lifespans.custom.*.grants.implicit.access_token",
	"identity_providers.oidc.lifespans.custom.*.grants.implicit.authorize_code",
	"identity_providers.oidc.lifespans.custom.*.grants.implicit.id_token",
	"identity_providers.oidc.lifespans.custom.*.grants.implicit.refresh_token",
	"identity_providers.oidc.lifespans.custom.*.grants.jwt_bearer.access_token",
	"identity_providers.oidc.lifespans.custom.*.grants.jwt_bearer.authorize_code",
	"identity_providers.oidc.lifespans.custom.*.grants.jwt_bearer.id_token",
	"identity_providers.oidc.lifespans.custom.*.grants.jwt_bearer.refresh_token",
	"identity_providers.oidc.lifespans.custom.*.grants.refresh_token.access_token",
	"identity_providers.oidc.lifespans.custom.*.grants.refresh_token.authorize_code",
	"identity_providers.oidc.lifespans.custom.*.grants.refresh_token.id_token",
	"identity_providers.oidc.lifespans.custom.*.grants.refresh_token.refresh_token",
	"identity_providers.oidc.lifespans.custom.*.id_token",
	"identity_providers.oidc.lifespans.custom.*.refresh_token",
	"identity_providers.oidc.lifespans.id_token",
	"identity_providers.oidc.lifespans.jwt_secured_authorization",
	"identity_providers.oidc.lifespans.refresh_token",
	"identity_providers.oidc.minimum_parameter_entropy",
	"identity_providers.oidc.require_pushed_authorization_requests",
	"identity_providers.oidc.scopes",
	"identity_providers.oidc.scopes.*",
	"identity_providers.oidc.scopes.*.claims",
	"identity_validation.elevated_session.characters",
	"identity_validation.elevated_session.code_lifespan",
	"identity_validation.elevated_session.elevation_lifespan",
	"identity_validation.elevated_session.require_second_factor",
	"identity_validation.elevated_session.skip_second_factor",
	"identity_validation.reset_password.jwt_algorithm",
	"identity_validation.reset_password.jwt_lifespan",
	"identity_validation.reset_password.jwt_secret",
	"log.file_path",
	"log.format",
	"log.keep_stdout",
	"log.level",
	"notifier.disable_startup_check",
	"notifier.filesystem.filename",
	"notifier.smtp.address",
	"notifier.smtp.disable_html_emails",
	"notifier.smtp.disable_require_tls",
	"notifier.smtp.disable_starttls",
	"notifier.smtp.host",
	"notifier.smtp.identifier",
	"notifier.smtp.password",
	"notifier.smtp.port",
	"notifier.smtp.sender",
	"notifier.smtp.startup_check_address",
	"notifier.smtp.subject",
	"notifier.smtp.timeout",
	"notifier.smtp.tls.certificate_chain",
	"notifier.smtp.tls.maximum_version",
	"notifier.smtp.tls.minimum_version",
	"notifier.smtp.tls.private_key",
	"notifier.smtp.tls.server_name",
	"notifier.smtp.tls.skip_verify",
	"notifier.smtp.username",
	"notifier.template_path",
	"ntp.address",
	"ntp.disable_failure",
	"ntp.disable_startup_check",
	"ntp.max_desync",
	"ntp.version",
	"password_policy.standard.enabled",
	"password_policy.standard.max_length",
	"password_policy.standard.min_length",
	"password_policy.standard.require_lowercase",
	"password_policy.standard.require_number",
	"password_policy.standard.require_special",
	"password_policy.standard.require_uppercase",
	"password_policy.zxcvbn.enabled",
	"password_policy.zxcvbn.min_score",
	"privacy_policy.enabled",
	"privacy_policy.policy_url",
	"privacy_policy.require_user_acceptance",
	"regulation.ban_time",
	"regulation.find_time",
	"regulation.max_retries",
	"regulation.modes",
	"server.address",
	"server.asset_path",
	"server.buffers.read",
	"server.buffers.write",
	"server.disable_healthcheck",
	"server.endpoints.authz",
	"server.endpoints.authz.*",
	"server.endpoints.authz.*.authn_strategies",
	"server.endpoints.authz.*.authn_strategies[].name",
	"server.endpoints.authz.*.authn_strategies[].scheme_basic_cache_lifespan",
	"server.endpoints.authz.*.authn_strategies[].schemes",
	"server.endpoints.authz.*.implementation",
	"server.endpoints.enable_expvars",
	"server.endpoints.enable_pprof",
	"server.endpoints.rate_limits.reset_password_finish.buckets",
	"server.endpoints.rate_limits.reset_password_finish.buckets[].period",
	"server.endpoints.rate_limits.reset_password_finish.buckets[].requests",
	"server.endpoints.rate_limits.reset_password_finish.enable",
	"server.endpoints.rate_limits.reset_password_start.buckets",
	"server.endpoints.rate_limits.reset_password_start.buckets[].period",
	"server.endpoints.rate_limits.reset_password_start.buckets[].requests",
	"server.endpoints.rate_limits.reset_password_start.enable",
	"server.endpoints.rate_limits.second_factor_duo.buckets",
	"server.endpoints.rate_limits.second_factor_duo.buckets[].period",
	"server.endpoints.rate_limits.second_factor_duo.buckets[].requests",
	"server.endpoints.rate_limits.second_factor_duo.enable",
	"server.endpoints.rate_limits.second_factor_totp.buckets",
	"server.endpoints.rate_limits.second_factor_totp.buckets[].period",
	"server.endpoints.rate_limits.second_factor_totp.buckets[].requests",
	"server.endpoints.rate_limits.second_factor_totp.enable",
	"server.endpoints.rate_limits.session_elevation_finish.buckets",
	"server.endpoints.rate_limits.session_elevation_finish.buckets[].period",
	"server.endpoints.rate_limits.session_elevation_finish.buckets[].requests",
	"server.endpoints.rate_limits.session_elevation_finish.enable",
	"server.endpoints.rate_limits.session_elevation_start.buckets",
	"server.endpoints.rate_limits.session_elevation_start.buckets[].period",
	"server.endpoints.rate_limits.session_elevation_start.buckets[].requests",
	"server.endpoints.rate_limits.session_elevation_start.enable",
	"server.headers.csp_template",
	"server.timeouts.idle",
	"server.timeouts.read",
	"server.timeouts.write",
	"server.tls.certificate",
	"server.tls.client_certificates",
	"server.tls.key",
	"session",
	"session.cookies",
	"session.cookies[]",
	"session.cookies[].authelia_url",
	"session.cookies[].default_redirection_url",
	"session.cookies[].domain",
	"session.cookies[].expiration",
	"session.cookies[].inactivity",
	"session.cookies[].name",
	"session.cookies[].remember_me",
	"session.cookies[].same_site",
	"session.domain",
	"session.expiration",
	"session.inactivity",
	"session.name",
	"session.redis.database_index",
	"session.redis.high_availability.nodes",
	"session.redis.high_availability.nodes[].host",
	"session.redis.high_availability.nodes[].port",
	"session.redis.high_availability.route_by_latency",
	"session.redis.high_availability.route_randomly",
	"session.redis.high_availability.sentinel_name",
	"session.redis.high_availability.sentinel_password",
	"session.redis.high_availability.sentinel_username",
	"session.redis.host",
	"session.redis.max_retries",
	"session.redis.maximum_active_connections",
	"session.redis.minimum_idle_connections",
	"session.redis.password",
	"session.redis.port",
	"session.redis.timeout",
	"session.redis.tls.certificate_chain",
	"session.redis.tls.maximum_version",
	"session.redis.tls.minimum_version",
	"session.redis.tls.private_key",
	"session.redis.tls.server_name",
	"session.redis.tls.skip_verify",
	"session.redis.username",
	"session.remember_me",
	"session.same_site",
	"session.secret",
	"storage.encryption_key",
	"storage.local.path",
	"storage.mysql.address",
	"storage.mysql.database",
	"storage.mysql.password",
	"storage.mysql.timeout",
	"storage.mysql.tls.certificate_chain",
	"storage.mysql.tls.maximum_version",
	"storage.mysql.tls.minimum_version",
	"storage.mysql.tls.private_key",
	"storage.mysql.tls.server_name",
	"storage.mysql.tls.skip_verify",
	"storage.mysql.username",
	"storage.postgres.address",
	"storage.postgres.database",
	"storage.postgres.password",
	"storage.postgres.schema",
	"storage.postgres.servers",
	"storage.postgres.servers[].address",
	"storage.postgres.servers[].tls",
	"storage.postgres.servers[].tls.certificate_chain",
	"storage.postgres.servers[].tls.maximum_version",
	"storage.postgres.servers[].tls.minimum_version",
	"storage.postgres.servers[].tls.private_key",
	"storage.postgres.servers[].tls.server_name",
	"storage.postgres.servers[].tls.skip_verify",
	"storage.postgres.ssl.certificate",
	"storage.postgres.ssl.key",
	"storage.postgres.ssl.mode",
	"storage.postgres.ssl.root_certificate",
	"storage.postgres.timeout",
	"storage.postgres.tls.certificate_chain",
	"storage.postgres.tls.maximum_version",
	"storage.postgres.tls.minimum_version",
	"storage.postgres.tls.private_key",
	"storage.postgres.tls.server_name",
	"storage.postgres.tls.skip_verify",
	"storage.postgres.username",
	"telemetry.metrics.address",
	"telemetry.metrics.buffers.read",
	"telemetry.metrics.buffers.write",
	"telemetry.metrics.enabled",
	"telemetry.metrics.timeouts.idle",
	"telemetry.metrics.timeouts.read",
	"telemetry.metrics.timeouts.write",
	"theme",
	"totp.algorithm",
	"totp.allowed_algorithms",
	"totp.allowed_digits",
	"totp.allowed_periods",
	"totp.digits",
	"totp.disable",
	"totp.disable_reuse_security_policy",
	"totp.issuer",
	"totp.period",
	"totp.secret_size",
	"totp.skew",
	"webauthn.attestation_conveyance_preference",
	"webauthn.disable",
	"webauthn.display_name",
	"webauthn.enable_passkey_login",
	"webauthn.experimental_enable_passkey_uv_two_factors",
	"webauthn.filtering.permitted_aaguids",
	"webauthn.filtering.prohibit_backup_eligibility",
	"webauthn.filtering.prohibited_aaguids",
	"webauthn.metadata.enabled",
	"webauthn.metadata.validate_entry",
	"webauthn.metadata.validate_entry_permit_zero_aaguid",
	"webauthn.metadata.validate_status",
	"webauthn.metadata.validate_status_permitted",
	"webauthn.metadata.validate_status_prohibited",
	"webauthn.metadata.validate_trust_anchor",
	"webauthn.selection_criteria.attachment",
	"webauthn.selection_criteria.discoverability",
	"webauthn.selection_criteria.user_verification",
	"webauthn.timeout",
}
