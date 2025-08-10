package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateServerTLS checks a server TLS configuration is correct.
func ValidateServerTLS(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Server.TLS.Key != "" && config.Server.TLS.Certificate == "" {
		validator.Push(errors.New(errFmtServerTLSCert))
	} else if config.Server.TLS.Key == "" && config.Server.TLS.Certificate != "" {
		validator.Push(errors.New(errFmtServerTLSKey))
	}

	if config.Server.TLS.Key != "" {
		validateServerTLSFileExists("key", config.Server.TLS.Key, validator)
	}

	if config.Server.TLS.Certificate != "" {
		validateServerTLSFileExists("certificate", config.Server.TLS.Certificate, validator)
	}

	if config.Server.TLS.Key == "" && config.Server.TLS.Certificate == "" &&
		len(config.Server.TLS.ClientCertificates) > 0 {
		validator.Push(errors.New(errFmtServerTLSClientAuthNoAuth))
	}

	for _, clientCertPath := range config.Server.TLS.ClientCertificates {
		validateServerTLSFileExists("client_certificates", clientCertPath, validator)
	}
}

// validateServerTLSFileExists checks whether a file exist.
func validateServerTLSFileExists(name, path string, validator *schema.StructValidator) {
	var (
		info os.FileInfo
		err  error
	)

	switch info, err = os.Stat(path); {
	case os.IsNotExist(err):
		validator.Push(fmt.Errorf("server: tls: option '%s' with path '%s' refers to a file that doesn't exist", name, path))
	case err != nil:
		validator.Push(fmt.Errorf("server: tls: option '%s' with path '%s' could not be verified due to a file system error: %w", name, path, err))
	case info.IsDir():
		validator.Push(fmt.Errorf("server: tls: option '%s' with path '%s' refers to a directory but it should refer to a file", name, path))
	}
}

// ValidateServer checks the server configuration is correct.
func ValidateServer(config *schema.Configuration, validator *schema.StructValidator) {
	ValidateServerAddress(config, validator)
	ValidateServerTLS(config, validator)
	validateServerAssets(config, validator)

	if config.Server.Buffers.Read <= 0 {
		config.Server.Buffers.Read = schema.DefaultServerConfiguration.Buffers.Read
	}

	if config.Server.Buffers.Write <= 0 {
		config.Server.Buffers.Write = schema.DefaultServerConfiguration.Buffers.Write
	}

	if config.Server.Timeouts.Read <= 0 {
		config.Server.Timeouts.Read = schema.DefaultServerConfiguration.Timeouts.Read
	}

	if config.Server.Timeouts.Write <= 0 {
		config.Server.Timeouts.Write = schema.DefaultServerConfiguration.Timeouts.Write
	}

	if config.Server.Timeouts.Idle <= 0 {
		config.Server.Timeouts.Idle = schema.DefaultServerConfiguration.Timeouts.Idle
	}

	ValidateServerEndpoints(config, validator)
}

// ValidateServerAddress checks the configured server address is correct.
func ValidateServerAddress(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Server.Address == nil {
		config.Server.Address = schema.DefaultServerConfiguration.Address
	} else {
		var err error
		if err = config.Server.Address.ValidateHTTP(); err != nil {
			validator.Push(fmt.Errorf(errFmtServerAddress, config.Server.Address.String(), err))
		}
	}

	switch subpath := config.Server.Address.RouterPath(); {
	case subpath == "":
		config.Server.Address.SetPath("/")
	case subpath != "/":
		if p := strings.TrimPrefix(subpath, "/"); strings.Contains(p, "/") {
			validator.Push(fmt.Errorf(errFmtServerPathNotEndForwardSlash, subpath))
		} else if !utils.IsStringAlphaNumeric(p) {
			validator.Push(fmt.Errorf(errFmtServerPathAlphaNumeric, subpath))
		}
	}

	if config.Server.Address.IsUnixDomainSocket() || config.Server.Address.IsFileDescriptor() {
		config.Server.DisableHealthcheck = true
	}
}

// ValidateServerEndpoints configures the default endpoints and checks the configuration of custom endpoints.
func ValidateServerEndpoints(config *schema.Configuration, validator *schema.StructValidator) {
	validateServerEndpointsRateLimits(config, validator)

	if config.Server.Endpoints.EnableExpvars {
		validator.PushWarning(fmt.Errorf("server: endpoints: option 'enable_expvars' should not be enabled in production"))
	}

	if config.Server.Endpoints.EnablePprof {
		validator.PushWarning(fmt.Errorf("server: endpoints: option 'enable_pprof' should not be enabled in production"))
	}

	if len(config.Server.Endpoints.Authz) == 0 {
		config.Server.Endpoints.Authz = schema.DefaultServerConfiguration.Endpoints.Authz

		return
	}

	authzs := make([]string, 0, len(config.Server.Endpoints.Authz))

	for name := range config.Server.Endpoints.Authz {
		authzs = append(authzs, name)
	}

	sort.Strings(authzs)

	for _, name := range authzs {
		endpoint := config.Server.Endpoints.Authz[name]

		validateServerEndpointsAuthzEndpoint(config, name, endpoint, validator)

		for _, oName := range authzs {
			oEndpoint := config.Server.Endpoints.Authz[oName]

			if oName == name || oName == legacy {
				continue
			}

			switch oEndpoint.Implementation {
			case schema.AuthzImplementationLegacy, schema.AuthzImplementationExtAuthz:
				if strings.HasPrefix(name, oName+"/") {
					validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzPrefixDuplicate, name, oName, oEndpoint.Implementation))
				}
			default:
				continue
			}
		}

		validateServerEndpointsAuthzStrategies(name, endpoint.Implementation, endpoint.AuthnStrategies, validator)
	}
}

func validateServerAssets(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Server.AssetPath == "" {
		return
	}

	if _, err := os.Stat(config.Server.AssetPath); err != nil {
		switch {
		case os.IsNotExist(err):
			validator.Push(fmt.Errorf("server: asset_path: error occurred reading the '%s' directory: the directory does not exist", config.Server.AssetPath))
		case os.IsPermission(err):
			validator.Push(fmt.Errorf("server: asset_path: error occurred reading the '%s' directory: a permission error occurred trying to read the directory", config.Server.AssetPath))
		default:
			validator.Push(fmt.Errorf("server: asset_path: error occurred reading the '%s' directory: %w", config.Server.AssetPath, err))
		}

		return
	}

	var (
		entries []fs.DirEntry
		err     error
	)
	if entries, err = os.ReadDir(filepath.Join(config.Server.AssetPath, "locales")); err != nil {
		if !os.IsNotExist(err) {
			validator.Push(fmt.Errorf("server: asset_path: error occurred reading the '%s' directory: %w", filepath.Join(config.Server.AssetPath, "locales"), err))
		}

		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		locale := entry.Name()

		var namespaceEntries []fs.DirEntry

		if namespaceEntries, err = os.ReadDir(filepath.Join(config.Server.AssetPath, "locales", locale)); err != nil {
			validator.Push(fmt.Errorf("server: asset_path: error occurred reading the '%s' directory: %w", filepath.Join(config.Server.AssetPath, "locales", locale), err))
		}

		for _, namespaceEntry := range namespaceEntries {
			if namespaceEntry.IsDir() || !strings.HasSuffix(namespaceEntry.Name(), ".json") {
				continue
			}

			path := filepath.Join(config.Server.AssetPath, "locales", locale, namespaceEntry.Name())

			var (
				data         []byte
				translations map[string]any
			)

			if data, err = os.ReadFile(path); err != nil {
				validator.Push(fmt.Errorf("server: asset_path: error occurred reading the '%s' file: %w", path, err))

				continue
			}

			if err = json.Unmarshal(data, &translations); err != nil {
				validator.Push(fmt.Errorf("server: asset_path: error occurred decoding the '%s' file: %w", path, err))

				continue
			}

			validateServerAssetsIterate("", path, translations, validator)
		}
	}
}

func validateServerAssetsIterate(keyRoot, path string, translations map[string]any, validator *schema.StructValidator) {
	for key, raw := range translations {
		var (
			value   string
			fullkey string
			sub     map[string]any
			ok      bool
		)

		if keyRoot == "" {
			fullkey = key
		} else {
			fullkey = strings.Join([]string{keyRoot, key}, ".")
		}

		if sub, ok = raw.(map[string]any); ok {
			validateServerAssetsIterate(fullkey, path, sub, validator)

			continue
		}

		if !strings.Contains(key, i18nAuthelia) {
			continue
		}

		if value, ok = raw.(string); !ok {
			validator.Push(fmt.Errorf("server: asset_path: error occurred decoding the '%s' file: translation key '%s' has a value which is not the required type", path, fullkey))

			continue
		}

		if !strings.Contains(value, i18nAuthelia) {
			validator.Push(fmt.Errorf("server: asset_path: error occurred decoding the '%s' file: translation key '%s' has a value which is missing a required placeholder", path, fullkey))

			continue
		}
	}
}

func validateServerEndpointsRateLimits(config *schema.Configuration, validator *schema.StructValidator) {
	validateServerEndpointsRateLimitDefault("reset_password_start", &config.Server.Endpoints.RateLimits.ResetPasswordStart, schema.DefaultServerConfiguration.Endpoints.RateLimits.ResetPasswordStart, validator)
	validateServerEndpointsRateLimitDefault("reset_password_finish", &config.Server.Endpoints.RateLimits.ResetPasswordFinish, schema.DefaultServerConfiguration.Endpoints.RateLimits.ResetPasswordFinish, validator)
	validateServerEndpointsRateLimitDefault("second_factor_totp", &config.Server.Endpoints.RateLimits.SecondFactorTOTP, schema.DefaultServerConfiguration.Endpoints.RateLimits.SecondFactorTOTP, validator)
	validateServerEndpointsRateLimitDefault("second_factor_duo", &config.Server.Endpoints.RateLimits.SecondFactorDuo, schema.DefaultServerConfiguration.Endpoints.RateLimits.SecondFactorDuo, validator)

	validateServerEndpointsRateLimitDefaultWeighted("session_elevation_start", &config.Server.Endpoints.RateLimits.SessionElevationStart, schema.DefaultServerConfiguration.Endpoints.RateLimits.SessionElevationStart, config.IdentityValidation.ElevatedSession.CodeLifespan, validator)
	validateServerEndpointsRateLimitDefaultWeighted("session_elevation_finish", &config.Server.Endpoints.RateLimits.SessionElevationFinish, schema.DefaultServerConfiguration.Endpoints.RateLimits.SessionElevationFinish, config.IdentityValidation.ElevatedSession.ElevationLifespan, validator)
}

func validateServerEndpointsRateLimitDefault(name string, config *schema.ServerEndpointRateLimit, defaults schema.ServerEndpointRateLimit, validator *schema.StructValidator) {
	if len(config.Buckets) == 0 {
		config.Buckets = make([]schema.ServerEndpointRateLimitBucket, len(defaults.Buckets))

		copy(config.Buckets, defaults.Buckets)

		return
	}

	validateServerEndpointsRateLimitBuckets(name, config, validator)
}

func validateServerEndpointsRateLimitDefaultWeighted(name string, config *schema.ServerEndpointRateLimit, defaults schema.ServerEndpointRateLimit, weight time.Duration, validator *schema.StructValidator) {
	if len(config.Buckets) == 0 {
		config.Buckets = make([]schema.ServerEndpointRateLimitBucket, len(defaults.Buckets))

		for i := range defaults.Buckets {
			config.Buckets[i] = schema.ServerEndpointRateLimitBucket{
				Period:   weight * defaults.Buckets[i].Period,
				Requests: defaults.Buckets[i].Requests,
			}
		}

		return
	}

	validateServerEndpointsRateLimitBuckets(name, config, validator)
}

func validateServerEndpointsRateLimitBuckets(name string, config *schema.ServerEndpointRateLimit, validator *schema.StructValidator) {
	for i, bucket := range config.Buckets {
		if bucket.Period == 0 {
			validator.Push(fmt.Errorf(errFmtServerEndpointsRateLimitsBucketPeriodZero, name, i+1))
		} else if bucket.Period < (time.Second * 10) {
			validator.Push(fmt.Errorf(errFmtServerEndpointsRateLimitsBucketPeriodTooLow, name, i+1, bucket.Period))
		}

		if bucket.Requests <= 0 {
			validator.Push(fmt.Errorf(errFmtServerEndpointsRateLimitsBucketRequestsZero, name, i+1, bucket.Requests))
		}
	}
}

func validateServerEndpointsAuthzEndpoint(config *schema.Configuration, name string, endpoint schema.ServerEndpointsAuthz, validator *schema.StructValidator) {
	if name == legacy {
		switch endpoint.Implementation {
		case schema.AuthzImplementationLegacy:
			break
		case "":
			endpoint.Implementation = schema.AuthzImplementationLegacy

			config.Server.Endpoints.Authz[name] = endpoint
		default:
			if !utils.IsStringInSlice(endpoint.Implementation, validAuthzImplementations) {
				validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzImplementation, name, utils.StringJoinOr(validAuthzImplementations), endpoint.Implementation))
			} else {
				validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzLegacyInvalidImplementation, name))
			}
		}
	} else if !utils.IsStringInSlice(endpoint.Implementation, validAuthzImplementations) {
		validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzImplementation, name, utils.StringJoinOr(validAuthzImplementations), endpoint.Implementation))
	}

	if !reAuthzEndpointName.MatchString(name) {
		validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzInvalidName, name))
	}
}

//nolint:gocyclo
func validateServerEndpointsAuthzStrategies(name, implementation string, strategies []schema.ServerEndpointsAuthzAuthnStrategy, validator *schema.StructValidator) {
	var defaults []schema.ServerEndpointsAuthzAuthnStrategy

	switch implementation {
	case schema.AuthzImplementationLegacy:
		defaults = schema.DefaultServerConfiguration.Endpoints.Authz[schema.AuthzEndpointNameLegacy].AuthnStrategies
	case schema.AuthzImplementationAuthRequest:
		defaults = schema.DefaultServerConfiguration.Endpoints.Authz[schema.AuthzEndpointNameAuthRequest].AuthnStrategies
	case schema.AuthzImplementationExtAuthz:
		defaults = schema.DefaultServerConfiguration.Endpoints.Authz[schema.AuthzEndpointNameExtAuthz].AuthnStrategies
	case schema.AuthzImplementationForwardAuth:
		defaults = schema.DefaultServerConfiguration.Endpoints.Authz[schema.AuthzEndpointNameForwardAuth].AuthnStrategies
	}

	if len(strategies) == 0 {
		copy(strategies, defaults)

		return
	}

	names := make([]string, 0, len(strategies))

	for i, strategy := range strategies {
		if strategy.Name != "" && utils.IsStringInSlice(strategy.Name, names) {
			validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzStrategyDuplicate, name, strategy.Name))
		}

		names = append(names, strategy.Name)

		if strategy.SchemeBasicCacheLifespan > 0 && !utils.IsStringInSliceFold(schema.SchemeBasic, strategy.Schemes) {
			validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzStrategySchemeOnlyOption, name, i+1, "scheme_basic_cache_lifespan", schema.SchemeBasic, utils.StringJoinAnd(strategy.Schemes)))
		}

		switch {
		case strategy.Name == "":
			validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzStrategyNoName, name, i+1))
		case !utils.IsStringInSlice(strategy.Name, validAuthzAuthnStrategies):
			validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzStrategy, name, utils.StringJoinOr(validAuthzAuthnStrategies), strategy.Name))
		default:
			if utils.IsStringInSlice(strategy.Name, validAuthzAuthnHeaderStrategies) {
				if len(strategy.Schemes) == 0 {
					strategies[i].Schemes = defaults[0].Schemes
				} else {
					for _, scheme := range strategy.Schemes {
						if !utils.IsStringInSliceFold(scheme, validAuthzAuthnStrategySchemes) {
							validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzSchemes, name, i+1, strategy.Name, utils.StringJoinOr(validAuthzAuthnStrategySchemes), scheme))
						}
					}
				}
			} else if len(strategy.Schemes) != 0 {
				validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzSchemesInvalidForStrategy, name, i+1, strategy.Name))
			}
		}
	}
}
