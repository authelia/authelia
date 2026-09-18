// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"bufio"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

// TestRelyingPartySuitesTrustTheProviderCertificate asserts the Authelia instance of each suite which acts as an
// OpenID Connect 1.0 Relying Party trusts the certificate its external provider is served with, using nothing but the
// certificates mounted into its certificates_directory. Authelia's Relying Party only trusts the system roots and that
// directory, and the development PKI is in neither the system roots of the container nor anything else it is given,
// so a suite which gets this wrong fails at the first discovery request with an error which says nothing of the cause.
func TestRelyingPartySuitesTrustTheProviderCertificate(t *testing.T) {
	testCases := []struct {
		name  string
		suite string
		host  string
	}{
		{"ShouldTrustTheConformanceSuite", "OIDCConformance", "conformance.example.com"},
		{"ShouldTrustTheUpstreamProvider", "OpenIDConnectRelyingParty", "auth-upstream.example.com"},
	}

	// The portal serves both hosts with this chain; it is mounted at /pki and named by ssl_certificate in nginx.conf.
	chain := readTestCertificates(t, filepath.Join("common", "pki", "public.chain.pem"))
	require.NotEmpty(t, chain)

	intermediates := x509.NewCertPool()

	for _, certificate := range chain[1:] {
		intermediates.AddCert(certificate)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			directory := readTestCertificatesDirectory(t, filepath.Join(tc.suite, "configuration.yml"))
			require.NotEmpty(t, directory, "the suite must configure a certificates_directory")

			roots := x509.NewCertPool()
			authorities := 0

			for _, volume := range readTestComposeVolumes(t, filepath.Join(tc.suite, "compose.yml"), "authelia-backend") {
				parts := strings.Split(volume, ":")

				if len(parts) < 2 || path.Dir(parts[1]) != path.Clean(directory) || !isTestCertificateFile(parts[1]) {
					continue
				}

				for _, certificate := range readTestCertificates(t, parts[0]) {
					roots.AddCert(certificate)

					if certificate.IsCA {
						authorities++
					}
				}
			}

			assert.NotZero(t, authorities, "the certificate authority must be trusted rather than only the leaf, so the trust survives the leaf being reissued")

			// The validity period is not what is under test, so the chain is verified at a time it is known to be valid.
			_, err := chain[0].Verify(x509.VerifyOptions{
				DNSName:       tc.host,
				Roots:         roots,
				Intermediates: intermediates,
				CurrentTime:   chain[0].NotBefore.Add(time.Hour),
			})

			require.NoError(t, err)
		})
	}
}

func isTestCertificateFile(name string) bool {
	name = strings.ToLower(name)

	return strings.HasSuffix(name, ".crt") || strings.HasSuffix(name, ".cer") || strings.HasSuffix(name, ".pem")
}

func readTestCertificates(t *testing.T, name string) (certificates []*x509.Certificate) {
	t.Helper()

	data, err := os.ReadFile(name)
	require.NoError(t, err)

	for block, rest := pem.Decode(data); block != nil; block, rest = pem.Decode(rest) {
		if block.Type != "CERTIFICATE" {
			continue
		}

		certificate, err := x509.ParseCertificate(block.Bytes)
		require.NoError(t, err)

		certificates = append(certificates, certificate)
	}

	return certificates
}

// readTestCertificatesDirectory reads the certificates_directory option by scanning the file rather than decoding it,
// as the suite configurations carry template expressions which are not valid YAML until they are rendered.
func readTestCertificatesDirectory(t *testing.T, name string) string {
	t.Helper()

	f, err := os.Open(name)
	require.NoError(t, err)

	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if value, ok := strings.CutPrefix(scanner.Text(), "certificates_directory:"); ok {
			return strings.Trim(strings.TrimSpace(value), `'"`)
		}
	}

	require.NoError(t, scanner.Err())

	return ""
}

func readTestComposeVolumes(t *testing.T, name, service string) []string {
	t.Helper()

	data, err := os.ReadFile(name)
	require.NoError(t, err)

	compose := struct {
		Services map[string]struct {
			Volumes []string `yaml:"volumes"`
		} `yaml:"services"`
	}{}

	require.NoError(t, yaml.Unmarshal(data, &compose))

	return compose.Services[service].Volumes
}
