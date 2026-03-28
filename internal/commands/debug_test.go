package commands

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewDebugCmds(t *testing.T) {
	var cmd *cobra.Command

	cmd = newDebugCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newDebugExpressionCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newDebugOIDCCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newDebugOIDCClaimsCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newDebugTLSCmd(&CmdCtx{})
	assert.NotNil(t, cmd)
}

// testUserDatabaseContent is a minimal user database for testing.
var testUserDatabaseContent = []byte(`
users:
  john:
    displayname: "John Doe"
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
`)

// newTestFileAuthConfig creates a schema.Configuration with a file-based auth backend
// pointing at a temp user database file.
func newTestFileAuthConfig(t *testing.T) *schema.Configuration {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "users.yml")

	require.NoError(t, os.WriteFile(dbPath, testUserDatabaseContent, 0600))

	return &schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			File: &schema.AuthenticationBackendFile{
				Path:     dbPath,
				Password: schema.DefaultCIPasswordConfig,
			},
		},
	}
}

// newTestTLSServer creates a TLS listener on a random local port and returns its address.
// The server is automatically cleaned up when the test finishes.
func newTestTLSServer(t *testing.T) (addr string, certPool *x509.CertPool) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	tlsCert := tls.Certificate{
		Certificate: [][]byte{certBytes},
		PrivateKey:  key,
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	})
	require.NoError(t, err)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			// Perform the TLS handshake before closing.
			if tlsConn, ok := conn.(*tls.Conn); ok {
				_ = tlsConn.Handshake()
			}

			conn.Close()
		}
	}()

	t.Cleanup(func() {
		ln.Close()
	})

	pool := x509.NewCertPool()
	pool.AddCert(cert)

	return ln.Addr().String(), pool
}

func TestDebugTLSRunE(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(t *testing.T) (string, *pflag.FlagSet, *x509.CertPool)
		err      string
		expected []string
	}{
		{
			"ShouldSucceedTLSConnection",
			func(t *testing.T) (string, *pflag.FlagSet, *x509.CertPool) {
				addr, pool := newTestTLSServer(t)
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
				flags.String("hostname", "", "")

				return addr, flags, pool
			},
			"",
			[]string{"General Information:", "Server Name:", "TLS Version:", "Certificate Information:", "Certificate #1:"},
		},
		{
			"ShouldSucceedWithHostnameOverride",
			func(t *testing.T) (string, *pflag.FlagSet, *x509.CertPool) {
				addr, pool := newTestTLSServer(t)
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
				flags.String("hostname", "", "")

				require.NoError(t, flags.Set("hostname", "localhost"))

				return addr, flags, pool
			},
			"",
			[]string{"General Information:", "Server Name:", "Hostname Verification: pass"},
		},
		{
			"ShouldSucceedWithHostnameMismatch",
			func(t *testing.T) (string, *pflag.FlagSet, *x509.CertPool) {
				addr, pool := newTestTLSServer(t)
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
				flags.String("hostname", "", "")

				require.NoError(t, flags.Set("hostname", "wronghost.example.com"))

				return addr, flags, pool
			},
			"",
			[]string{"General Information:", "Hostname Verification: fail"},
		},
		{
			"ShouldErrInvalidAddress",
			func(t *testing.T) (string, *pflag.FlagSet, *x509.CertPool) {
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
				flags.String("hostname", "", "")

				return "://invalid", flags, nil
			},
			"",
			nil,
		},
		{
			"ShouldErrConnectionRefused",
			func(t *testing.T) (string, *pflag.FlagSet, *x509.CertPool) {
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
				flags.String("hostname", "", "")

				return "127.0.0.1:1", flags, nil
			},
			"failed to connect to",
			nil,
		},
		{
			"ShouldShowUntrustedCertificateWarning",
			func(t *testing.T) (string, *pflag.FlagSet, *x509.CertPool) {
				addr, _ := newTestTLSServer(t)
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
				flags.String("hostname", "", "")

				// Use an empty cert pool so the cert is untrusted.
				return addr, flags, x509.NewCertPool()
			},
			"",
			[]string{"General Information:", "WARNING: The certificate is not valid", "BEGIN CERTIFICATE"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			addr, flags, pool := tc.setup(t)

			buf := new(bytes.Buffer)

			err := runDebugTLS(buf, flags, pool, addr)

			if tc.err == "" {
				if tc.expected != nil {
					assert.NoError(t, err)

					for _, s := range tc.expected {
						assert.Contains(t, buf.String(), s)
					}
				}
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestDebugTLSRunENotTLS(t *testing.T) {
	// Start a plain TCP server (no TLS).
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

			conn.Close()
		}
	}()

	t.Cleanup(func() {
		ln.Close()
	})

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("hostname", "", "")

	buf := new(bytes.Buffer)

	err = runDebugTLS(buf, flags, nil, ln.Addr().String())

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Did not receive a TLS handshake")
}

func TestDebugExpressionRunE(t *testing.T) {
	testCases := []struct {
		name     string
		config   func(t *testing.T) *schema.Configuration
		username string
		expr     string
		err      string
		expected []string
	}{
		{
			"ShouldSucceedResolveTrue",
			newTestFileAuthConfig,
			"john",
			"'admins' in groups",
			"",
			[]string{"Resolved: true", "Resolved Value:"},
		},
		{
			"ShouldSucceedResolveFalse",
			newTestFileAuthConfig,
			"john",
			"'nonexistent' in groups",
			"",
			[]string{"Resolved: true"},
		},
		{
			"ShouldErrNoProvider",
			func(t *testing.T) *schema.Configuration {
				return &schema.Configuration{}
			},
			"john",
			"'admins' in groups",
			"error occurred initializing user authentication provider: a provider is not configured",
			nil,
		},
		{
			"ShouldErrUserNotFound",
			newTestFileAuthConfig,
			"nonexistent",
			"'admins' in groups",
			"error occurred getting extended user details from the user authentication provider:",
			nil,
		},
		{
			"ShouldErrInvalidExpression",
			newTestFileAuthConfig,
			"john",
			"invalid %%% expression",
			"error occurred initializing user attributes expression provider:",
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := tc.config(t)

			buf := new(bytes.Buffer)

			err := runDebugExpression(buf, config, nil, tc.username, tc.expr)

			if tc.err == "" {
				assert.NoError(t, err)

				for _, s := range tc.expected {
					assert.Contains(t, buf.String(), s)
				}
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestDebugOIDCClaimsRunE(t *testing.T) {
	newOIDCConfig := func(t *testing.T) *schema.Configuration {
		config := newTestFileAuthConfig(t)

		config.IdentityProviders.OIDC = &schema.IdentityProvidersOpenIDConnect{
			Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
				oidc.ScopeOpenID: {},
			},
		}

		return config
	}

	testCases := []struct {
		name     string
		config   func(t *testing.T) *schema.Configuration
		username string
		flags    map[string]string
		err      string
		expected []string
	}{
		{
			"ShouldSucceedDefaultFlags",
			newOIDCConfig,
			"john",
			nil,
			"",
			[]string{"Results:", "ID Token:", "User Information:"},
		},
		{
			"ShouldSucceedImplicitFlow",
			newOIDCConfig,
			"john",
			map[string]string{"response-type": oidc.ResponseTypeImplicitFlowIDToken},
			"",
			[]string{"Results:", "ID Token:", "User Information:"},
		},
		{
			"ShouldSucceedClientCredentials",
			newOIDCConfig,
			"john",
			map[string]string{"grant-type": oidc.GrantTypeClientCredentials},
			"",
			[]string{"Results:", "ID Token:", "User Information:"},
		},
		{
			"ShouldErrNoProvider",
			func(t *testing.T) *schema.Configuration {
				return &schema.Configuration{}
			},
			"john",
			nil,
			"error occurred initializing user authentication provider: a provider is not configured",
			nil,
		},
		{
			"ShouldErrNoOIDCProvider",
			newTestFileAuthConfig,
			"john",
			nil,
			"error occurred initializing oidc provider: a provider is not configured",
			nil,
		},
		{
			"ShouldErrUserNotFound",
			newOIDCConfig,
			"nonexistent",
			nil,
			"error occurred getting extended user details from the user authentication provider:",
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := tc.config(t)

			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.String("client-id", "example", "")
			flags.String("policy", "", "")
			flags.String("response-type", oidc.ResponseTypeAuthorizationCodeFlow, "")
			flags.String("grant-type", oidc.GrantTypeAuthorizationCode, "")
			flags.StringSlice("scopes", []string{oidc.ScopeOpenID}, "")
			flags.StringSlice("claims", nil, "")

			for k, v := range tc.flags {
				require.NoError(t, flags.Set(k, v))
			}

			buf := new(bytes.Buffer)

			err := runDebugOIDCClaims(context.Background(), buf, flags, config, nil, tc.username)

			if tc.err == "" {
				assert.NoError(t, err)

				for _, s := range tc.expected {
					assert.Contains(t, buf.String(), s)
				}
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestDebugOIDCClaimsRunECmd(t *testing.T) {
	config := newTestFileAuthConfig(t)

	config.IdentityProviders.OIDC = &schema.IdentityProvidersOpenIDConnect{
		Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
			oidc.ScopeOpenID: {},
		},
	}

	cmdCtx := NewCmdCtx()
	cmdCtx.config = config

	cmd, buf := newTestCmdWithBuf()
	cmd.Flags().String("client-id", "example", "")
	cmd.Flags().String("policy", "", "")
	cmd.Flags().String("response-type", oidc.ResponseTypeAuthorizationCodeFlow, "")
	cmd.Flags().String("grant-type", oidc.GrantTypeAuthorizationCode, "")
	cmd.Flags().StringSlice("scopes", []string{oidc.ScopeOpenID}, "")
	cmd.Flags().StringSlice("claims", nil, "")

	err := cmdCtx.DebugOIDCClaimsRunE(cmd, []string{"john"})

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Results:")
}

func TestDebugExpressionRunECmd(t *testing.T) {
	config := newTestFileAuthConfig(t)

	cmdCtx := NewCmdCtx()
	cmdCtx.config = config

	cmd, buf := newTestCmdWithBuf()

	err := cmdCtx.DebugExpressionRunE(cmd, []string{"john", "'admins'", "in", "groups"})

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Resolved: true")
}

func TestDebugTLSRunECmd(t *testing.T) {
	addr, pool := newTestTLSServer(t)

	cmdCtx := NewCmdCtx()
	cmdCtx.trusted = pool

	cmd, buf := newTestCmdWithBuf()
	cmd.Flags().String("hostname", "", "")

	err := cmdCtx.DebugTLSRunE(cmd, []string{addr})

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "General Information:")
	assert.Contains(t, buf.String(), "Certificate Information:")
}

func TestDebugClaimsStrategyContext(t *testing.T) {
	ctx := &debugClaimsStrategyContext{
		Context:  context.Background(),
		resolver: nil,
	}

	assert.Nil(t, ctx.GetProviderUserAttributeResolver())

	assert.NotNil(t, ctx.Context)
}

func TestDebugTLSRunESuggestedConfig(t *testing.T) {
	addr, pool := newTestTLSServer(t)

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("hostname", "", "")

	buf := new(bytes.Buffer)

	err := runDebugTLS(buf, flags, pool, addr)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Suggested Configuration:")
	assert.Contains(t, buf.String(), "tls:")
	assert.Contains(t, buf.String(), "minimum_version:")
}
