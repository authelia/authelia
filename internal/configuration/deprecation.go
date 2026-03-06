package configuration

import (
	"fmt"
	"path"
	"strconv"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
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
	"identity_providers.oidc.clients[].userinfo_signing_algorithm": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.clients[].userinfo_signing_algorithm",
		NewKey:  "identity_providers.oidc.clients[].userinfo_signed_response_alg",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"identity_providers.oidc.clients[].id": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.clients[].id",
		NewKey:  "identity_providers.oidc.clients[].client_id",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"identity_providers.oidc.clients[].secret": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.clients[].secret",
		NewKey:  "identity_providers.oidc.clients[].client_secret",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"identity_providers.oidc.clients[].description": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.clients[].description",
		NewKey:  "identity_providers.oidc.clients[].client_name",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"identity_providers.oidc.clients[].sector_identifier": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.clients[].sector_identifier",
		NewKey:  "identity_providers.oidc.clients[].sector_identifier_uri",
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
		NewKey:  "identity_providers.oidc.jwks",
		AutoMap: false,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, keysFinal map[string]any, value any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf(errFmtAutoMapKey+" : see https://www.authelia.com/c/oidc for more information", d.Key, d.Version, d.NewKey, d.Version.NextMajor()))
		},
	},
	"identity_providers.oidc.issuer_certificate_chain": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "identity_providers.oidc.issuer_certificate_chain",
		NewKey:  "identity_providers.oidc.jwks",
		AutoMap: false,
		MapFunc: nil,
		ErrFunc: func(d Deprecation, keysFinal map[string]any, value any, val *schema.StructValidator) {
			val.PushWarning(fmt.Errorf("configuration key '%s' is deprecated in %s and should be configured using the new configuration key '%s': this has been automatically mapped for you but you will need to adjust your configuration (see https://www.authelia.com/c/oidc) to remove this message", d.Key, d.Version, d.NewKey))
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
	"jwt_secret": {
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Key:     "jwt_secret",
		NewKey:  "identity_validation.reset_password.jwt_secret",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"webauthn.user_verification": {
		Version: model.SemanticVersion{Major: 4, Minor: 39},
		Key:     "webauthn.user_verification",
		NewKey:  "webauthn.selection_criteria.user_verification",
		AutoMap: true,
		MapFunc: nil,
		ErrFunc: nil,
	},
	"authentication_backend.ldap.permit_feature_detection_failure": {
		Version: model.SemanticVersion{Major: 4, Minor: 39, Patch: 16},
		Key:     "authentication_backend.ldap.permit_feature_detection_failure",
		AutoMap: false,
	},
}

// MultiKeyMappedDeprecation represents a deprecated configuration key.
type MultiKeyMappedDeprecation struct {
	Version model.SemanticVersion
	Keys    []string
	NewKey  string
	MapFunc func(d MultiKeyMappedDeprecation, keys map[string]any, val *schema.StructValidator)
}

var deprecationsMKM = []MultiKeyMappedDeprecation{
	{
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Keys:    []string{"notifier.smtp.host", "notifier.smtp.port"},
		NewKey:  "notifier.smtp.address",
		MapFunc: func(d MultiKeyMappedDeprecation, keys map[string]any, val *schema.StructValidator) {
			host, port, err := getHostPort("notifier.smtp.host", "notifier.smtp.port", schema.DefaultSMTPNotifierConfiguration.Address.Host(), schema.DefaultSMTPNotifierConfiguration.Address.Port(), keys)
			if err != nil {
				val.Push(fmt.Errorf(errFmtMultiKeyMappingPortConvert, utils.StringJoinAnd(d.Keys), d.NewKey, err))

				return
			}

			address := schema.NewSMTPAddress("", host, port)

			val.PushWarning(fmt.Errorf(errFmtMultiRemappedKeys, utils.StringJoinAnd(d.Keys), d.Version, d.NewKey, "[tcp://]<hostname>[:<port>]", address.String(), d.Version.NextMajor()))

			keys[d.NewKey] = address.String()

			for _, key := range d.Keys {
				delete(keys, key)
			}
		},
	},
	{
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Keys:    []string{keyStoragePostgresHost, keyStoragePostgresPort},
		NewKey:  "storage.postgres.address",
		MapFunc: func(d MultiKeyMappedDeprecation, keys map[string]any, val *schema.StructValidator) {
			host, port, err := getHostPort(keyStoragePostgresHost, keyStoragePostgresPort, schema.DefaultPostgreSQLStorageConfiguration.Address.Host(), schema.DefaultPostgreSQLStorageConfiguration.Address.Port(), keys)
			if err != nil {
				val.Push(fmt.Errorf(errFmtMultiKeyMappingPortConvert, utils.StringJoinAnd(d.Keys), d.NewKey, err))

				return
			}

			if address, err := schema.NewAddressFromNetworkValuesDefault(host, port, schema.AddressSchemeTCP, schema.AddressSchemeUnix); err != nil {
				val.Push(fmt.Errorf("storage: %s: option 'address' failed to parse options 'host' and 'port' for mapping: %w", "postgres", err))
			} else {
				keys[d.NewKey] = address.String()

				val.PushWarning(fmt.Errorf(errFmtMultiRemappedKeys, utils.StringJoinAnd(d.Keys), d.Version, d.NewKey, "[tcp://]<hostname>[:<port>]", address.String(), d.Version.NextMajor()))

				for _, key := range d.Keys {
					delete(keys, key)
				}
			}
		},
	},
	{
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Keys:    []string{keyStorageMySQLHost, keyStorageMySQLPort},
		NewKey:  "storage.mysql.address",
		MapFunc: func(d MultiKeyMappedDeprecation, keys map[string]any, val *schema.StructValidator) {
			host, port, err := getHostPort(keyStorageMySQLHost, keyStorageMySQLPort, schema.DefaultMySQLStorageConfiguration.Address.Host(), schema.DefaultMySQLStorageConfiguration.Address.Port(), keys)
			if err != nil {
				val.Push(fmt.Errorf(errFmtMultiKeyMappingPortConvert, utils.StringJoinAnd(d.Keys), d.NewKey, err))

				return
			}

			if address, err := schema.NewAddressFromNetworkValuesDefault(host, port, schema.AddressSchemeTCP, schema.AddressSchemeUnix); err != nil {
				val.Push(fmt.Errorf("storage: %s: option 'address' failed to parse options 'host' and 'port' for mapping: %w", "mysql", err))
			} else {
				keys[d.NewKey] = address.String()

				val.PushWarning(fmt.Errorf(errFmtMultiRemappedKeys, utils.StringJoinAnd(d.Keys), d.Version, d.NewKey, "[tcp://]<hostname>[:<port>]", address.String(), d.Version.NextMajor()))

				for _, key := range d.Keys {
					delete(keys, key)
				}
			}
		},
	},
	{
		Version: model.SemanticVersion{Major: 4, Minor: 38},
		Keys:    []string{keyServerHost, keyServerPort, keyServerPath},
		NewKey:  "server.address",
		MapFunc: func(d MultiKeyMappedDeprecation, keys map[string]any, val *schema.StructValidator) {
			host, port, err := getHostPort(keyServerHost, keyServerPort, schema.DefaultServerConfiguration.Address.Hostname(), schema.DefaultServerConfiguration.Address.Port(), keys)
			if err != nil {
				val.Push(fmt.Errorf(errFmtMultiKeyMappingPortConvert, utils.StringJoinAnd(d.Keys), d.NewKey, err))

				return
			}

			var (
				v       any
				ok      bool
				subpath string
			)

			if v, ok = keys[keyServerPath]; ok {
				subpath, _ = v.(string)
			}

			switch subpath {
			case "":
				subpath = schema.DefaultServerConfiguration.Address.Path()
			default:
				subpath = path.Clean("/" + subpath)
			}

			address := &schema.AddressTCP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeTCP, host, port)}

			address.SetPath(subpath)

			val.PushWarning(fmt.Errorf(errFmtMultiRemappedKeys, utils.StringJoinAnd(d.Keys), d.Version, d.NewKey, "[tcp[(4|6)]://]<hostname>[:<port>][/<path>]' or 'tcp[(4|6)://][hostname]:<port>[/<path>]", address.String(), d.Version.NextMajor()))

			keys[d.NewKey] = address.String()

			for _, key := range d.Keys {
				delete(keys, key)
			}
		},
	},
}

func getHostPort(hostKey, portKey, hostFallback string, portFallback uint16, keys map[string]any) (host string, port uint16, err error) {
	var (
		ok bool
		v  any
	)

	if v, ok = keys[hostKey]; ok {
		host, _ = v.(string)
	}

	if v, ok = keys[portKey]; ok {
		switch value := v.(type) {
		case uint16:
			port = value
		case int:
			if value >= 0 && value <= 65535 {
				port = uint16(value)
			}
		case string:
			var p uint64

			if p, err = strconv.ParseUint(value, 10, 16); err != nil {
				return "", 0, fmt.Errorf("error occurred converting the port from a string: %w", err)
			}

			port = uint16(p)
		}
	}

	if host == "" {
		host = hostFallback
	}

	if port == 0 {
		port = portFallback
	}

	return host, port, nil
}

func GetMultiKeyMappedDeprecationKeys() (keys []string) {
	for _, mkm := range deprecationsMKM {
		keys = append(keys, mkm.Keys...)
	}

	return keys
}
