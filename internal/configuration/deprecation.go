package configuration

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

// Deprecation represents a deprecated configuration key.
type Deprecation struct {
	Version model.SemanticVersion
	Key     string
	NewKey  string
	AutoMap bool
	Keep    bool
	MapFunc func(value any) any
	ErrFunc func(d Deprecation, keysFinal map[string]any, value any, val *schema.StructValidator)
}

var deprecations = map[string]Deprecation{
	"logs_level": {
		Version: model.SemanticVersion{Major: 4, Minor: 7},
		Key:     "logs_level",
		NewKey:  "log.level",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"logs_file": {
		Version: model.SemanticVersion{Major: 4, Minor: 7},
		Key:     "logs_file",
		NewKey:  "log.file_path",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"authentication_backend.ldap.skip_verify": {
		Version: model.SemanticVersion{Major: 4, Minor: 25},
		Key:     "authentication_backend.ldap.skip_verify",
		NewKey:  "authentication_backend.ldap.tls.skip_verify",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"authentication_backend.ldap.minimum_tls_version": {
		Version: model.SemanticVersion{Major: 4, Minor: 25},
		Key:     "authentication_backend.ldap.minimum_tls_version",
		NewKey:  "authentication_backend.ldap.tls.minimum_version",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"notifier.smtp.disable_verify_cert": {
		Version: model.SemanticVersion{Major: 4, Minor: 25},
		Key:     "notifier.smtp.disable_verify_cert",
		NewKey:  "notifier.smtp.tls.skip_verify",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"notifier.smtp.trusted_cert": {
		Version: model.SemanticVersion{Major: 4, Minor: 25},
		Key:     "notifier.smtp.trusted_cert",
		NewKey:  "certificates_directory",
		AutoMap: false,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"host": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "host",
		NewKey:  "server.host",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"port": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "port",
		NewKey:  "server.port",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"tls_key": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "tls_key",
		NewKey:  "server.tls.key",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"tls_cert": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "tls_cert",
		NewKey:  "server.tls.certificate",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"log_level": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "log_level",
		NewKey:  "log.level",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"log_file_path": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "log_file_path",
		NewKey:  "log.file_path",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"log_format": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "log_format",
		NewKey:  "log.format",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"storage.postgres.sslmode": {
		Version: model.SemanticVersion{Major: 4, Minor: 36},
		Key:     "storage.postgres.sslmode",
		NewKey:  "storage.postgres.ssl.mode",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"authentication_backend.disable_reset_password": {
		Version: model.SemanticVersion{Major: 4, Minor: 36},
		Key:     "authentication_backend.disable_reset_password",
		NewKey:  "authentication_backend.password_reset.disable",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"server.read_buffer_size": {
		Version: model.SemanticVersion{Major: 4, Minor: 36},
		Key:     "server.read_buffer_size",
		NewKey:  "server.buffers.read",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"server.write_buffer_size": {
		Version: model.SemanticVersion{Major: 4, Minor: 36},
		Key:     "server.write_buffer_size",
		NewKey:  "server.buffers.write",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"session.remember_me_duration": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "session.remember_me_duration",
		NewKey:  "session.remember_me",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"server.enable_pprof": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "server.enable_pprof",
		NewKey:  "server.endpoints.enable_pprof",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"server.enable_expvars": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "server.enable_expvars",
		NewKey:  "server.endpoints.enable_expvars",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"server.host": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "server.host",
		NewKey:  "server.address",
		AutoMap: false,
		Keep:    true,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, _ map[string]any, _ any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf("configuration key '%s' is deprecated in %s and has been replaced by '%s' when combined with the 'server.port' and 'server.path' in the format of %s: this should be automatically mapped for you but you will need to adjust your configuration to remove this message", d.Key, d.Version.String(), d.NewKey, "'[tcp[(4|6)]://]<hostname>[:<port>][/<path>]' or 'tcp[(4|6)://][hostname]:<port>[/<path>]'"))
		},
	},
	"server.port": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "server.port",
		NewKey:  "server.address",
		AutoMap: false,
		Keep:    true,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, _ map[string]any, _ any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf("configuration key '%s' is deprecated in %s and has been replaced by '%s' when combined with the 'server.host' and 'server.path' in the format of %s: this should be automatically mapped for you but you will need to adjust your configuration to remove this message", d.Key, d.Version.String(), d.NewKey, "'[tcp[(4|6)]://]<hostname>[:<port>][/<path>]' or 'tcp[(4|6)://][hostname]:<port>[/<path>]'"))
		},
	},
	"server.path": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "server.path",
		NewKey:  "server.address",
		AutoMap: false,
		Keep:    true,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, _ map[string]any, _ any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf("configuration key '%s' is deprecated in %s and has been replaced by '%s' when combined with the 'server.host' and 'server.port' in the format of %s: this should be automatically mapped for you but you will need to adjust your configuration to remove this message", d.Key, d.Version.String(), d.NewKey, "'[tcp[(4|6)]://]<hostname>[:<port>][/<path>]' or 'tcp[(4|6)://][hostname]:<port>[/<path>]'"))
		},
	},
	"storage.mysql.host": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "storage.mysql.host",
		NewKey:  "storage.mysql.address",
		AutoMap: false,
		Keep:    true,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, _ map[string]any, _ any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf(errFmtSpecialRemappedKey, d.Key, d.Version.String(), d.NewKey, "storage.mysql.port", "[tcp://]<hostname>[:<port>]"))
		},
	},
	"storage.mysql.port": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "storage.mysql.port",
		NewKey:  "storage.mysql.address",
		AutoMap: false,
		Keep:    true,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, _ map[string]any, _ any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf(errFmtSpecialRemappedKey, d.Key, d.Version.String(), d.NewKey, "storage.mysql.host", "[tcp://]<hostname>[:<port>]"))
		},
	},
	"storage.postgres.host": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "storage.postgres.host",
		NewKey:  "storage.postgres.address",
		AutoMap: false,
		Keep:    true,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, _ map[string]any, _ any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf(errFmtSpecialRemappedKey, d.Key, d.Version.String(), d.NewKey, "storage.postgres.port", "[tcp://]<hostname>[:<port>]"))
		},
	},
	"storage.postgres.port": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "storage.postgres.port",
		NewKey:  "storage.postgres.address",
		AutoMap: false,
		Keep:    true,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, _ map[string]any, _ any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf(errFmtSpecialRemappedKey, d.Key, d.Version.String(), d.NewKey, "storage.postgres.host", "[tcp://]<hostname>[:<port>]"))
		},
	},
	"notifier.smtp.host": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "notifier.smtp.host",
		NewKey:  "notifier.smtp.address",
		AutoMap: false,
		Keep:    true,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, _ map[string]any, _ any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf(errFmtSpecialRemappedKey, d.Key, d.Version.String(), d.NewKey, "notifier.smtp.port", "[tcp://]<hostname>[:<port>]"))
		},
	},
	"notifier.smtp.port": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "notifier.smtp.port",
		NewKey:  "notifier.smtp.address",
		AutoMap: false,
		Keep:    true,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, _ map[string]any, _ any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf(errFmtSpecialRemappedKey, d.Key, d.Version.String(), d.NewKey, "notifier.smtp.host", "[tcp://]<hostname>[:<port>]"))
		},
	},
	"authentication_backend.ldap.url": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "authentication_backend.ldap.url",
		NewKey:  "authentication_backend.ldap.address",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"identity_providers.oidc.clients[].userinfo_signing_algorithm": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.clients[].userinfo_signing_algorithm",
		NewKey:  "identity_providers.oidc.clients[].userinfo_signed_response_alg",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"authentication_backend.ldap.username_attribute": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "authentication_backend.ldap.username_attribute",
		NewKey:  "authentication_backend.ldap.attributes.username",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"authentication_backend.ldap.mail_attribute": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "authentication_backend.ldap.mail_attribute",
		NewKey:  "authentication_backend.ldap.attributes.mail",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"authentication_backend.ldap.display_name_attribute": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "authentication_backend.ldap.display_name_attribute",
		NewKey:  "authentication_backend.ldap.attributes.display_name",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"authentication_backend.ldap.group_name_attribute": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "authentication_backend.ldap.group_name_attribute",
		NewKey:  "authentication_backend.ldap.attributes.group_name",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"identity_providers.oidc.access_token_lifespan": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.access_token_lifespan",
		NewKey:  "identity_providers.oidc.lifespans.access_token",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"identity_providers.oidc.authorize_code_lifespan": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.authorize_code_lifespan",
		NewKey:  "identity_providers.oidc.lifespans.authorize_code",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"identity_providers.oidc.id_token_lifespan": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.id_token_lifespan",
		NewKey:  "identity_providers.oidc.lifespans.id_token",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"identity_providers.oidc.refresh_token_lifespan": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.refresh_token_lifespan",
		NewKey:  "identity_providers.oidc.lifespans.refresh_token",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"identity_providers.oidc.issuer_private_key": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.issuer_private_key",
		NewKey:  "identity_providers.oidc.issuer_private_keys",
		AutoMap: false,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, keysFinal map[string]any, value any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf("configuration key '%s' is deprecated in %s and should be configured using the new configuration key '%s': this has been automatically mapped for you but you will need to adjust your configuration (see https://www.authelia.com/c/oidc) to remove this message", d.Key, d.Version, d.NewKey))
		},
	},
	"identity_providers.oidc.issuer_certificate_chain": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.issuer_certificate_chain",
		NewKey:  "identity_providers.oidc.issuer_private_keys",
		AutoMap: false,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, keysFinal map[string]any, value any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf("configuration key '%s' is deprecated in %s and should be configured using the new configuration key '%s': this has been automatically mapped for you but you will need to adjust your configuration (see https://www.authelia.com/c/oidc) to remove this message", d.Key, d.Version, d.NewKey))
		},
	},
}
