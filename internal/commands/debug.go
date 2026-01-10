package commands

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/spf13/cobra"
	goyaml "go.yaml.in/yaml/v4"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

func newDebugCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "debug",
		Short:   cmdAutheliaDebugShort,
		Long:    cmdAutheliaDebugLong,
		Example: cmdAutheliaDebugExample,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newDebugTLSCmd(ctx),
		newDebugExpressionCmd(ctx),
		newDebugOIDCCmd(ctx),
	)

	return cmd
}

func newDebugTLSCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "tls [address]",
		Short:   cmdAutheliaDebugTLSShort,
		Long:    cmdAutheliaDebugTLSLong,
		Example: cmdAutheliaDebugTLSExample,
		Args:    cobra.ExactArgs(1),
		RunE:    ctx.DebugTLSRunE,
		PreRunE: ctx.ChainRunE(
			ctx.HelperConfigLoadRunE,
			ctx.HelperConfigValidateKeysRunE,
			ctx.HelperConfigValidateRunE,
			ctx.LoadTrustedCertificatesRunE,
		),
		DisableAutoGenTag: true,
	}

	cmd.Flags().String("hostname", "", "overrides the hostname to use for the TLS connection which is usually extracted from the address")

	return cmd
}

func newDebugExpressionCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "expression <username> <expression>",
		Short:   cmdAutheliaDebugExpressionShort,
		Long:    cmdAutheliaDebugExpressionLong,
		Example: cmdAutheliaDebugExpressionExample,
		Args:    cobra.MinimumNArgs(2),
		RunE:    ctx.DebugExpressionRunE,
		PreRunE: ctx.ChainRunE(
			ctx.HelperConfigLoadRunE,
			ctx.HelperConfigValidateKeysRunE,
			ctx.HelperConfigValidateRunE,
			ctx.LoadTrustedCertificatesRunE,
		),
		DisableAutoGenTag: true,
	}

	return cmd
}

func newDebugOIDCCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "oidc",
		Short:   cmdAutheliaDebugOIDCShort,
		Long:    cmdAutheliaDebugOIDCLong,
		Example: cmdAutheliaDebugOIDCExample,
		PersistentPreRunE: ctx.ChainRunE(
			ctx.HelperConfigLoadRunE,
			ctx.HelperConfigValidateKeysRunE,
			ctx.HelperConfigValidateRunE,
		),
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newDebugOIDCClaimsCmd(ctx),
	)

	return cmd
}

func newDebugOIDCClaimsCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "claims <username>",
		Args:    cobra.ExactArgs(1),
		Short:   cmdAutheliaDebugOIDCClaimsShort,
		Long:    cmdAutheliaDebugOIDCClaimsLong,
		Example: cmdAutheliaDebugOIDCClaimsExample,
		RunE:    ctx.DebugOIDCClaimsRunE,

		DisableAutoGenTag: true,
	}

	cmd.Flags().String("policy", "", "claims policy name to use")
	cmd.Flags().String("client-id", "example", "arbitrary client id for the client")
	cmd.Flags().StringSlice("scopes", []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail, oidc.ScopePhone, oidc.ScopeAddress, oidc.ScopeGroups}, "granted scopes to use for this request")
	cmd.Flags().StringSlice("claims", nil, "granted claims to use for this request")
	cmd.Flags().String("response-type", oidc.ResponseTypeAuthorizationCodeFlow, "response type to use for this request")
	cmd.Flags().String("grant-type", oidc.GrantTypeAuthorizationCode, "grant type to use for this request")

	return cmd
}

//nolint:gocyclo
func (ctx *CmdCtx) DebugOIDCClaimsRunE(cmd *cobra.Command, args []string) (err error) {
	filters, err := configuration.NewFileFilters(ctx.cconfig.filters)

	if err != nil {
		return fmt.Errorf("error occurred initializing user authentication provider: filters could not be initialized")
	}

	provider := middlewares.NewAuthenticationProvider(ctx.config, ctx.trusted, filters)

	if provider == nil {
		return fmt.Errorf("error occurred initializing user authentication provider: a provider is not configured")
	}

	if err = provider.StartupCheck(); err != nil {
		return fmt.Errorf("error occurred initializing user authentication provider: %w", err)
	}

	resolver := expression.NewUserAttributes(ctx.config)

	if err = resolver.StartupCheck(); err != nil {
		return fmt.Errorf("error occurred initializing user attributes expression provider: %w", err)
	}

	if ctx.config.IdentityProviders.OIDC == nil {
		return fmt.Errorf("error occurred initializing oidc provider: a provider is not configured")
	}

	var (
		id, policy, responseType, grantType string

		scopes, claims []string

		detailer *authentication.UserDetailsExtended
	)

	username := args[0]

	if id, err = cmd.Flags().GetString("client-id"); err != nil {
		return err
	}

	if policy, err = cmd.Flags().GetString("policy"); err != nil {
		return err
	}

	if responseType, err = cmd.Flags().GetString("response-type"); err != nil {
		return err
	}

	if grantType, err = cmd.Flags().GetString("grant-type"); err != nil {
		return err
	}

	if scopes, err = cmd.Flags().GetStringSlice("scopes"); err != nil {
		return err
	}

	if claims, err = cmd.Flags().GetStringSlice("claims"); err != nil {
		return err
	}

	if detailer, err = provider.GetDetailsExtended(username); err != nil {
		return fmt.Errorf("error occurred getting extended user details from the user authentication provider: %w", err)
	}

	strategy := oidc.NewCustomClaimsStrategy(policy, scopes, ctx.config.IdentityProviders.OIDC.Scopes, ctx.config.IdentityProviders.OIDC.ClaimsPolicies)

	resolverctx := &debugClaimsStrategyContext{Context: ctx.Context, resolver: resolver}

	idtoken := map[string]any{}
	userinfo := map[string]any{}

	client := &oidc.RegisteredClient{
		ID: id,
	}

	implicit := responseType == oidc.ResponseTypeImplicitFlowIDToken

	if err = strategy.HydrateIDTokenClaims(resolverctx, oauthelia2.ExactScopeStrategy, client, scopes, claims, nil, detailer, time.Now(), time.Now().Add(time.Second*-10), nil, idtoken, implicit); err != nil {
		return fmt.Errorf("error occurred populating user ID token claims: %w", err)
	}

	if grantType == oidc.GrantTypeClientCredentials {
		if err = strategy.HydrateClientCredentialsUserInfoClaims(resolverctx, client, nil, userinfo); err != nil {
			return fmt.Errorf("error occurred populating user info claims: %w", err)
		}
	} else if err = strategy.HydrateUserInfoClaims(resolverctx, oauthelia2.ExactScopeStrategy, client, scopes, claims, nil, detailer, time.Now(), time.Now().Add(time.Second*-10), nil, userinfo); err != nil {
		return fmt.Errorf("error occurred populating user info claims: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Results:\n\n")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\tID Token:\n\t\t")

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("\t\t", "  ")
	encoder.SetEscapeHTML(false)

	if err = encoder.Encode(idtoken); err != nil {
		return fmt.Errorf("error occurred encoding ID Token claims: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n\tUser Information:\n\t\t")

	if err = encoder.Encode(userinfo); err != nil {
		return fmt.Errorf("error occurred encoding User Information claims: %w", err)
	}

	return nil
}

func (ctx *CmdCtx) DebugExpressionRunE(cmd *cobra.Command, args []string) (err error) {
	filters, _ := configuration.NewFileFilters(ctx.cconfig.filters)
	provider := middlewares.NewAuthenticationProvider(ctx.config, ctx.trusted, filters)

	if provider == nil {
		return fmt.Errorf("error occurred initializing user authentication provider: a provider is not configured")
	}

	if err = provider.StartupCheck(); err != nil {
		return fmt.Errorf("error occurred initializing user authentication provider: %w", err)
	}

	exp := strings.Join(args[1:], " ")

	e := expression.NewUserAttributes(&schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{File: &schema.AuthenticationBackendFile{}},
		Definitions: schema.Definitions{
			UserAttributes: map[string]schema.UserAttribute{
				"example": {
					Expression: exp,
				},
			},
		},
	})

	if err = e.StartupCheck(); err != nil {
		return fmt.Errorf("error occurred initializing user attributes expression provider: %w", err)
	}

	username := args[0]

	var details *authentication.UserDetailsExtended

	if details, err = provider.GetDetailsExtended(username); err != nil {
		return fmt.Errorf("error occurred getting extended user details from the user authentication provider: %w", err)
	}

	resolved, found := e.Resolve("example", details, time.Now())

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Resolved: %t\n", found)

	if found {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Resolved Value: %v\n", resolved)
	}

	return nil
}

//nolint:gocyclo
func (ctx *CmdCtx) DebugTLSRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		address *schema.Address
		conn    *tls.Conn
	)

	if address, err = schema.NewAddress(args[0]); err != nil {
		return err
	}

	var hostnameOverride string

	hostname := address.Hostname()

	if hostnameOverride, err = cmd.Flags().GetString("hostname"); err == nil && hostnameOverride != "" {
		hostname = hostnameOverride
	} else if err != nil {
		return err
	}

	n := len(tls.CipherSuites())

	suites := make([]uint16, n+len(tls.InsecureCipherSuites()))

	for i, suite := range tls.CipherSuites() {
		suites[i] = suite.ID
	}

	for i, suite := range tls.InsecureCipherSuites() {
		suites[i+n] = suite.ID
	}

	config := &tls.Config{
		ServerName:         hostname,
		InsecureSkipVerify: true,             //nolint:gosec // This is used solely to determine the TLS socket information.
		MinVersion:         tls.VersionSSL30, //nolint:staticcheck
		MaxVersion:         tls.VersionTLS13,
		CipherSuites:       suites,
	}

	if conn, err = tls.Dial(address.Network(), address.NetworkAddress(), config); err != nil {
		switch errStr := err.Error(); {
		case strings.Contains(errStr, "first record does not look like a TLS handshake"):
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "General Information:\n")
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\tFailure: Did not receive a TLS handshake from %s\n", address.NetworkAddress())

			return nil
		default:
			return fmt.Errorf("failed to connect to '%s' with unknown error: %w", address.NetworkAddress(), err)
		}
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "General Information:\n")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\tServer Name: %s\n", conn.ConnectionState().ServerName)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\tRemote Address: %s\n", conn.RemoteAddr().String())
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\tNegotiated Protocol: %s\n", conn.ConnectionState().NegotiatedProtocol)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\tTLS Version: %s\n", tls.VersionName(conn.ConnectionState().Version))

	if utils.IsInsecureCipherSuite(conn.ConnectionState().CipherSuite) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\tCipher Suite: %s (Unsupported and Insecure)\n", tls.CipherSuiteName(conn.ConnectionState().CipherSuite))
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\tCipher Suite: %s (Supported)\n", tls.CipherSuiteName(conn.ConnectionState().CipherSuite))
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nCertificate Information:\n")

	system, err := x509.SystemCertPool()
	if err != nil {
		system = x509.NewCertPool()
	}

	optsSystem := x509.VerifyOptions{
		Roots:         system,
		Intermediates: x509.NewCertPool(),
	}

	opts := x509.VerifyOptions{
		Roots:         ctx.trusted,
		Intermediates: x509.NewCertPool(),
	}

	certs := conn.ConnectionState().PeerCertificates

	opts.Intermediates = utils.UnsafeGetIntermediatesFromPeerCertificates(conn.ConnectionState().PeerCertificates, opts.Roots, opts.Intermediates)
	optsSystem.Intermediates = utils.UnsafeGetIntermediatesFromPeerCertificates(conn.ConnectionState().PeerCertificates, optsSystem.Roots, optsSystem.Intermediates)

	valid, validSystem, validHostname := true, true, true

	for i, cert := range conn.ConnectionState().PeerCertificates {
		if _, err = cert.Verify(optsSystem); err != nil {
			validSystem = false
		}

		if _, err = cert.Verify(opts); err != nil {
			valid = false
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n\tCertificate #%d:\n", i+1)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tCertificate Authority: %t\n", cert.IsCA)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tPublic Key Algorithm: %s\n", cert.PublicKeyAlgorithm)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tSignature Algorithm: %s\n", cert.SignatureAlgorithm)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tSubject: %s\n", cert.Subject)

		if len(cert.DNSNames) != 0 {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tAlternative Names (DNS): %s\n", strings.Join(cert.DNSNames, ", "))
		}

		if len(cert.IPAddresses) != 0 {
			ips := make([]string, len(cert.IPAddresses))
			for j, ip := range cert.IPAddresses {
				ips[j] = ip.String()
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tAlternative Names (IP): %s\n", strings.Join(ips, ", "))
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tIssuer: %s\n", cert.Issuer)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tNot Before: %s\n", cert.NotBefore)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tNot After: %s\n", cert.NotAfter)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tSerial Number: %s\n", cert.SerialNumber)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tValid: %t\n", valid)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tValid (System): %t\n", validSystem)

		if err != nil {
			var (
				errUA x509.UnknownAuthorityError
				errH  x509.HostnameError
				errCI x509.CertificateInvalidError
			)

			switch {
			case errors.As(err, &errUA):
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tValidation Hint: Certificate signed by unknown authority\n")
			case errors.As(err, &errH):
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tValidation Hint: Certificate hostname mismatch\n")
			case errors.As(err, &errCI):
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tValidation Hint: Certificate is invalid\n")
			default:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tValidation Hint: Unknown Error (%T)\n", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tValidation Error: %v\n", err)
		}

		if i == 0 {
			if err = cert.VerifyHostname(hostname); err != nil {
				validHostname = false

				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tHostname Verification: fail\n\t\tHostname Verification Error: %v\n", err)
			} else {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\t\tHostname Verification: pass\n")
			}
		}
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n\tCertificate Trusted: %t\n\tCertificate Matches Hostname: %t\n", valid, validHostname)

	c := struct {
		TLS schema.TLS `yaml:"tls"`
	}{
		TLS: schema.TLS{
			ServerName:     conn.ConnectionState().ServerName,
			SkipVerify:     false,
			MinimumVersion: schema.TLSVersion{Value: conn.ConnectionState().Version},
			MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
		},
	}

	if !validHostname && len(certs[0].DNSNames) != 0 {
		c.TLS.ServerName = certs[0].DNSNames[0]
	} else if validHostname && hostnameOverride != "" {
		c.TLS.ServerName = hostnameOverride
	}

	data, err := goyaml.Marshal(&c)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nError marshaling suggested config: %v\n", err)
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nSuggested Configuration:\n\n")
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(data))
	}

	if !valid {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nWARNING: The certificate is not valid for one reason or another. You may need to configure Authelia to trust certificate below.\n\n")

		block := &pem.Block{
			Type:  utils.BlockTypeCertificate,
			Bytes: conn.ConnectionState().PeerCertificates[0].Raw,
		}

		if err = pem.Encode(cmd.OutOrStdout(), block); err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error writing certificate to stdout: %v\n", err)
		}
	}

	return conn.Close()
}

type debugClaimsStrategyContext struct {
	resolver expression.UserAttributeResolver

	context.Context
}

func (ctx *debugClaimsStrategyContext) GetProviderUserAttributeResolver() expression.UserAttributeResolver {
	return ctx.resolver
}
