package validator

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateConfiguration and adapt the configuration read from file.
func ValidateConfiguration(config *schema.Configuration, validator *schema.StructValidator, opts ...func(ctx *ValidateCtx)) {
	var err error

	ctx := NewValidateCtx()

	for _, opt := range opts {
		opt(ctx)
	}

	if config.CertificatesDirectory != "" {
		var info os.FileInfo

		if info, err = os.Stat(config.CertificatesDirectory); err != nil {
			validator.Push(fmt.Errorf("the location 'certificates_directory' could not be inspected: %w", err))
		} else if !info.IsDir() {
			validator.Push(fmt.Errorf("the location 'certificates_directory' refers to '%s' is not a directory", config.CertificatesDirectory))
		}
	}

	validateDefault2FAMethod(config, validator)

	ValidateTheme(config, validator)

	ValidateLog(config, validator)

	ValidateDuo(config, validator)

	ValidateTOTP(config, validator)

	ValidateWebAuthn(config, validator)

	ValidateIdentityValidation(config, validator)

	ValidateAuthenticationBackend(&config.AuthenticationBackend, validator)

	ValidateDefinitions(config, validator)

	ValidateAccessControl(config, validator)

	ValidateRules(config, validator)

	ValidateSession(config, validator)

	ValidateRegulation(config, validator)

	ValidateServer(config, validator)

	ValidateTelemetry(config, validator)

	ValidateStorage(config.Storage, validator)

	ValidateNotifier(&config.Notifier, validator, config.Definitions.Webhooks)

	ValidateIdentityProviders(ctx, config, validator)

	ValidateNTP(config, validator)

	ValidatePasswordPolicy(&config.PasswordPolicy, validator)

	ValidatePrivacyPolicy(&config.PrivacyPolicy, validator)
}

func validateDefault2FAMethod(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Default2FAMethod == "" {
		return
	}

	if !utils.IsStringInSlice(config.Default2FAMethod, validDefault2FAMethods) {
		validator.Push(fmt.Errorf(errFmtInvalidDefault2FAMethod, utils.StringJoinOr(validDefault2FAMethods), config.Default2FAMethod))

		return
	}

	var enabledMethods []string

	if !config.TOTP.Disable {
		enabledMethods = append(enabledMethods, "totp")
	}

	if !config.WebAuthn.Disable {
		enabledMethods = append(enabledMethods, "webauthn")
	}

	if !config.DuoAPI.Disable {
		enabledMethods = append(enabledMethods, "mobile_push")
	}

	if !utils.IsStringInSlice(config.Default2FAMethod, enabledMethods) {
		validator.Push(fmt.Errorf(errFmtInvalidDefault2FAMethodDisabled, utils.StringJoinOr(enabledMethods), config.Default2FAMethod))
	}
}

func NewValidateCtx() *ValidateCtx {
	return &ValidateCtx{
		Context: context.Background(),
	}
}

type ValidateCtx struct {
	client *http.Client

	tlsconfig *tls.Config

	cacheSectorIdentifierURIs map[string][]string

	context.Context
}

func (ctx *ValidateCtx) GetHTTPClient() (client *http.Client) {
	if ctx.client == nil {
		dialer := &net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}

		transport := &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       ctx.tlsconfig,
		}

		ctx.client = &http.Client{Transport: transport}
	}

	return ctx.client
}

func WithTLSConfig(config *tls.Config) func(ctx *ValidateCtx) {
	return func(ctx *ValidateCtx) {
		ctx.tlsconfig, ctx.client = config, nil
	}
}
