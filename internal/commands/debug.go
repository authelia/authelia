package commands

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
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
		),
		DisableAutoGenTag: true,
	}

	cmd.Flags().String("hostname", "", "overrides the hostname to use for the TLS connection which is usually extracted from the address")

	return cmd
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
		MinVersion:         tls.VersionSSL30, //nolint:staticcheck // This is used solely to determine the TLS socket information.
		MaxVersion:         tls.VersionTLS13,
		CipherSuites:       suites,
	}

	if conn, err = tls.Dial(address.Network(), address.NetworkAddress(), config); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address.NetworkAddress(), err)
	}

	_, _ = fmt.Fprintf(os.Stdout, "General Information:\n")
	_, _ = fmt.Fprintf(os.Stdout, "\tServer Name: %s\n", conn.ConnectionState().ServerName)
	_, _ = fmt.Fprintf(os.Stdout, "\tNegotiated Protocol: %s\n", conn.ConnectionState().NegotiatedProtocol)
	_, _ = fmt.Fprintf(os.Stdout, "\tTLS Version: %s\n", tls.VersionName(conn.ConnectionState().Version))

	if utils.IsInsecureCipherSuite(conn.ConnectionState().CipherSuite) {
		_, _ = fmt.Fprintf(os.Stdout, "\tCipher Suite: %s (Unsupported and Insecure)\n", tls.CipherSuiteName(conn.ConnectionState().CipherSuite))
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "\tCipher Suite: %s (Supported)\n", tls.CipherSuiteName(conn.ConnectionState().CipherSuite))
	}

	_, _ = fmt.Fprintf(os.Stdout, "\nCertificate Information:\n")

	opts := x509.VerifyOptions{
		Roots:         ctx.trusted,
		Intermediates: x509.NewCertPool(),
	}

	certs := conn.ConnectionState().PeerCertificates

	for i, cert := range certs {
		if _, err = cert.Verify(opts); err == nil {
			if cert.IsCA {
				opts.Intermediates.AddCert(cert)
			}

			if i != 0 {
				if certs[i-1].IsCA {
					opts.Intermediates.AddCert(certs[i-1])
				}
			}
		}
	}

	valid := true
	validHostname := true

	for i, cert := range conn.ConnectionState().PeerCertificates {
		if _, err = cert.Verify(opts); err != nil {
			valid = false
		}

		_, _ = fmt.Fprintf(os.Stdout, "\n\tCertificate #%d:\n", i+1)
		_, _ = fmt.Fprintf(os.Stdout, "\t\tCertificate Authority: %t\n", cert.IsCA)
		_, _ = fmt.Fprintf(os.Stdout, "\t\tPublic Key Algorithm: %s\n", cert.PublicKeyAlgorithm)
		_, _ = fmt.Fprintf(os.Stdout, "\t\tSignature Algorithm: %s\n", cert.SignatureAlgorithm)
		_, _ = fmt.Fprintf(os.Stdout, "\t\tSubject: %s\n", cert.Subject)

		if len(cert.DNSNames) != 0 {
			_, _ = fmt.Fprintf(os.Stdout, "\t\tAlternative Names (DNS): %s\n", strings.Join(cert.DNSNames, ", "))
		}

		if len(cert.IPAddresses) != 0 {
			ips := make([]string, len(cert.IPAddresses))
			for j, ip := range cert.IPAddresses {
				ips[j] = ip.String()
			}

			_, _ = fmt.Fprintf(os.Stdout, "\t\tAlternative Names (IP): %s\n", strings.Join(ips, ", "))
		}

		_, _ = fmt.Fprintf(os.Stdout, "\t\tIssuer: %s\n", cert.Issuer)
		_, _ = fmt.Fprintf(os.Stdout, "\t\tNot Before: %s\n", cert.NotBefore)
		_, _ = fmt.Fprintf(os.Stdout, "\t\tNot After: %s\n", cert.NotAfter)
		_, _ = fmt.Fprintf(os.Stdout, "\t\tSerial Number: %s\n", cert.SerialNumber)
		_, _ = fmt.Fprintf(os.Stdout, "\t\tValid: %t\n", valid)

		if err != nil {
			var (
				errUA *x509.UnknownAuthorityError
				errH  *x509.HostnameError
				errCI *x509.CertificateInvalidError
			)

			switch {
			case errors.As(err, &errUA):
				_, _ = fmt.Fprintf(os.Stdout, "\t\tValidation Hint: Certificate signed by unknown authority\n")
			case errors.As(err, &errH):
				_, _ = fmt.Fprintf(os.Stdout, "\t\tValidation Hint: Certificate hostname mismatch\n")
			case errors.As(err, &errCI):
				_, _ = fmt.Fprintf(os.Stdout, "\t\tValidation Hint: Certificate is invalid\n")
			default:
				_, _ = fmt.Fprintf(os.Stdout, "\t\tValidation Hint: Unknown Error (%T)\n", err)
			}

			_, _ = fmt.Fprintf(os.Stdout, "\t\tValidation Error: %v\n", err)
		}

		if i == 0 {
			if err = cert.VerifyHostname(hostname); err != nil {
				validHostname = false

				_, _ = fmt.Fprintf(os.Stdout, "\t\tHostname Verification Error: %v\n", err)
			} else {
				_, _ = fmt.Fprintf(os.Stdout, "\t\tHostname Verification: pass\n")
			}
		}
	}

	_, _ = fmt.Fprintf(os.Stdout, "\n\tTrusted: %t\n", valid)

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

	data, err := yaml.Marshal(&c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "\nError marshaling suggested config: %v\n", err)
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "\nSuggested Configuration:\n\n")
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", string(data))
	}

	if !valid {
		_, _ = fmt.Fprintf(os.Stdout, "\nWARNING: The certificate is not valid for one reason or another. You may need to configure Authelia to trust certificate below.\n\n")

		block := &pem.Block{
			Type:  utils.BlockTypeCertificate,
			Bytes: conn.ConnectionState().PeerCertificates[0].Raw,
		}

		if err = pem.Encode(os.Stdout, block); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "Error writing certificate to stdout: %v\n", err)
		}
	}

	return conn.Close()
}
