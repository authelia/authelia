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
			val.PushWarning(fmt.Errorf("configuration key 'server.host' is deprecated in %s and has been replaced by 'server.address' when combined with the 'server.port' in the format of '[tcp://]<hostname>[:<port>]': this should be automatically mapped for you but you will need to adjust your configuration to remove this message", d.Version.String()))
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
			val.PushWarning(fmt.Errorf("configuration key 'server.port' is deprecated in %s and has been replaced by 'server.address' when combined with the 'server.host' in the format of '[tcp://]<hostname>[:<port>]': this should be automatically mapped for you but you will need to adjust your configuration to remove this message", d.Version.String()))
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
			val.PushWarning(fmt.Errorf("configuration key 'storage.mysql.host' is deprecated in %s and has been replaced by 'storage.mysql.address' when combined with the 'storage.mysql.port' in the format of '[tcp://]<hostname>[:<port>]': this should be automatically mapped for you but you will need to adjust your configuration to remove this message", d.Version.String()))
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
			val.PushWarning(fmt.Errorf("configuration key 'storage.mysql.port' is deprecated in %s and has been replaced by 'storage.mysql.address' when combined with the 'storage.mysql.host' in the format of '[tcp://]<hostname>[:<port>]': this should be automatically mapped for you but you will need to adjust your configuration to remove this message", d.Version.String()))
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
			val.PushWarning(fmt.Errorf("configuration key 'storage.postgres.host' is deprecated in %s and has been replaced by 'storage.postgres.address' when combined with the 'storage.postgres.port' in the format of '[tcp://]<hostname>[:<port>]': this should be automatically mapped for you but you will need to adjust your configuration to remove this message", d.Version.String()))
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
			val.PushWarning(fmt.Errorf("configuration key 'storage.postgres.port' is deprecated in %s and has been replaced by 'storage.postgres.address' when combined with the 'storage.postgres.host' in the format of '[tcp://]<hostname>[:<port>]': this should be automatically mapped for you but you will need to adjust your configuration to remove this message", d.Version.String()))
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
}
