package configuration

import (
	"github.com/authelia/authelia/v4/internal/model"
)

// Deprecation represents a deprecated configuration key.
type Deprecation struct {
	Version model.SemanticVersion
	Key     string
	NewKey  string
	AutoMap bool
	MapFunc func(value interface{}) interface{}
	ErrText string
}

var deprecations = map[string]Deprecation{
	"logs_level": {
		Version: model.SemanticVersion{Major: 4, Minor: 7},
		Key:     "logs_level",
		NewKey:  "log.level",
		AutoMap: true,
		MapFunc: nil,
	},
	"logs_file": {
		Version: model.SemanticVersion{Major: 4, Minor: 7},
		Key:     "logs_file",
		NewKey:  "log.file_path",
		AutoMap: true,
		MapFunc: nil,
	},
	"authentication_backend.ldap.skip_verify": {
		Version: model.SemanticVersion{Major: 4, Minor: 25},
		Key:     "authentication_backend.ldap.skip_verify",
		NewKey:  "authentication_backend.ldap.tls.skip_verify",
		AutoMap: true,
		MapFunc: nil,
	},
	"authentication_backend.ldap.minimum_tls_version": {
		Version: model.SemanticVersion{Major: 4, Minor: 25},
		Key:     "authentication_backend.ldap.minimum_tls_version",
		NewKey:  "authentication_backend.ldap.tls.minimum_version",
		AutoMap: true,
		MapFunc: nil,
	},
	"notifier.smtp.disable_verify_cert": {
		Version: model.SemanticVersion{Major: 4, Minor: 25},
		Key:     "notifier.smtp.disable_verify_cert",
		NewKey:  "notifier.smtp.tls.skip_verify",
		AutoMap: true,
		MapFunc: nil,
	},
	"notifier.smtp.trusted_cert": {
		Version: model.SemanticVersion{Major: 4, Minor: 25},
		Key:     "notifier.smtp.trusted_cert",
		NewKey:  "certificates_directory",
		AutoMap: false,
		MapFunc: nil,
	},
	"host": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "logs_file",
		NewKey:  "server.host",
		AutoMap: true,
		MapFunc: nil,
	},
	"port": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "port",
		NewKey:  "server.port",
		AutoMap: true,
		MapFunc: nil,
	},
	"tls_key": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "tls_key",
		NewKey:  "server.tls.key",
		AutoMap: true,
		MapFunc: nil,
	},
	"tls_cert": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "tls_cert",
		NewKey:  "server.tls.certificate",
		AutoMap: true,
		MapFunc: nil,
	},
	"log_level": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "log_level",
		NewKey:  "log.level",
		AutoMap: true,
		MapFunc: nil,
	},
	"log_file_path": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "log_file_path",
		NewKey:  "log.file_path",
		AutoMap: true,
		MapFunc: nil,
	},
	"log_format": {
		Version: model.SemanticVersion{Major: 4, Minor: 30},
		Key:     "log_format",
		NewKey:  "log.format",
		AutoMap: true,
		MapFunc: nil,
	},
	"authentication_backend.disable_reset_password": {
		Version: model.SemanticVersion{Major: 4, Minor: 36},
		Key:     "authentication_backend.disable_reset_password",
		NewKey:  "authentication_backend.password_reset.disable",
		AutoMap: true,
		MapFunc: nil,
	},
}
