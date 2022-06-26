package suites

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/utils"
)

type CLISuite struct {
	*CommandSuite
}

func NewCLISuite() *CLISuite {
	return &CLISuite{CommandSuite: new(CommandSuite)}
}

func (s *CLISuite) SetupSuite() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/CLI/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
	})
	s.DockerEnvironment = dockerEnvironment
}

func (s *CLISuite) SetupTest() {
	testArg := ""
	coverageArg := ""

	if os.Getenv("CI") == t {
		testArg = "-test.coverprofile=/authelia/coverage-$(date +%s).txt"
		coverageArg = "COVERAGE"
	}

	s.testArg = testArg
	s.coverageArg = coverageArg
}

func (s *CLISuite) TestShouldPrintBuildInformation() {
	if os.Getenv("CI") == "false" {
		s.T().Skip("Skipping testing in dev environment")
	}

	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "build-info"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Last Tag: ")
	s.Assert().Contains(output, "State: ")
	s.Assert().Contains(output, "Branch: ")
	s.Assert().Contains(output, "Build Number: ")
	s.Assert().Contains(output, "Build OS: ")
	s.Assert().Contains(output, "Build Arch: ")
	s.Assert().Contains(output, "Build Date: ")

	r := regexp.MustCompile(`^Last Tag: v\d+\.\d+\.\d+\nState: (tagged|untagged) (clean|dirty)\nBranch: [^\s\n]+\nCommit: [0-9a-f]{40}\nBuild Number: \d+\nBuild OS: (linux|darwin|windows|freebsd)\nBuild Arch: (amd64|arm|arm64)\nBuild Date: (Sun|Mon|Tue|Wed|Thu|Fri|Sat), \d{2} (Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d{4} \d{2}:\d{2}:\d{2} [+-]\d{4}\nExtra: \n`)
	s.Assert().Regexp(r, output)
}

func (s *CLISuite) TestShouldPrintVersion() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "--version"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "authelia version")
}

func (s *CLISuite) TestShouldValidateConfig() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "validate-config", "--config=/config/configuration.yml"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Configuration parsed and loaded successfully without errors.")
}

func (s *CLISuite) TestShouldFailValidateConfig() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "validate-config", "--config=/config/invalid.yml"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "failed to load configuration from yaml file(/config/invalid.yml) source: open /config/invalid.yml: no such file or directory")
}

func (s *CLISuite) TestShouldHashPasswordArgon2id() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "hash-password", "test", "-m", "32", "-s", "test1234"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Password hash: $argon2id$v=19$m=32768,t=3,p=4$")
}

func (s *CLISuite) TestShouldHashPasswordSHA512() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "hash-password", "test", "-z"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Password hash: $6$rounds=50000")
}

func (s *CLISuite) TestShouldGenerateRSACertificateRequest() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "rsa", "generate", "--csr", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate Signing Request")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tSignature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate Signing Request: /tmp/request.csr")
}

func (s *CLISuite) TestShouldGenerateECDSACurveP224CertificateRequest() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "ecdsa", "generate", "--csr", "--curve=P224", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate Signing Request")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tSignature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-224")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate Signing Request: /tmp/request.csr")
}

func (s *CLISuite) TestShouldGenerateECDSACurveP256CertificateRequest() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "ecdsa", "generate", "--csr", "--curve=P256", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate Signing Request")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tSignature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-256")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate Signing Request: /tmp/request.csr")
}

func (s *CLISuite) TestShouldGenerateECDSACurveP384CertificateRequest() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "ecdsa", "generate", "--csr", "--curve=P384", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate Signing Request")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tSignature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-384")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate Signing Request: /tmp/request.csr")
}

func (s *CLISuite) TestShouldGenerateECDSACurveP521CertificateRequest() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "ecdsa", "generate", "--csr", "--curve=P521", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate Signing Request")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tSignature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-521")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate Signing Request: /tmp/request.csr")
}

func (s *CLISuite) TestShouldGenerateEd25519CertificateRequest() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "ed25519", "generate", "--csr", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate Signing Request")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tSignature Algorithm: Ed25519, Public Key Algorithm: Ed25519")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate Signing Request: /tmp/request.csr")
}

func (s *CLISuite) TestShouldGenerateCertificateRSA() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "rsa", "generate", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tSelf-Signed")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tCA: false, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateRSAWithIPAddress() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "rsa", "generate", "--common-name=example.com", "--sans", "*.example.com,127.0.0.1", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tSelf-Signed")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tCA: false, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com, IP.1:127.0.0.1")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateRSAWithNotBefore() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "rsa", "generate", "--common-name=example.com", "--sans='*.example.com'", "--not-before", "'Jan 1 15:04:05 2011'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tSelf-Signed")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tNot Before: 2011-01-01T15:04:05Z, Not After: 2012-01-01T15:04:05Z")
	s.Assert().Contains(output, "\tCA: false, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/public.crt")
}

func (s *CLISuite) TestShouldFailGenerateCertificateRSAWithInvalidNotBefore() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "rsa", "generate", "--common-name=example.com", "--sans='*.example.com'", "--not-before", "Jan", "--directory=/tmp/"})
	s.Assert().NotNil(err)
	s.Assert().Contains(output, "Error: failed to parse not before: parsing time \"Jan\" as \"Jan 2 15:04:05 2006\": cannot parse \"\" as \"2\"")
}

func (s *CLISuite) TestShouldGenerateCertificateRSAWith4096Bits() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "rsa", "generate", "--bits=4096", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tSelf-Signed")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tCA: false, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 4096")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateWithCustomizedSubject() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "rsa", "generate", "--common-name=example.com", "--sans='*.example.com'", "--country=Australia", "--organization='Acme Co.'", "--organizational-unit=Tech", "--province=QLD", "--street-address='123 Smith St'", "--postcode=4000", "--locality=Internet", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tSelf-Signed")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Acme Co.], Organizational Unit: [Tech]")
	s.Assert().Contains(output, "\tCountry: [Australia], Province: [QLD], Street Address: [123 Smith St], Postal Code: [4000], Locality: [Internet]")
	s.Assert().Contains(output, "\tCA: false, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateCA() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "rsa", "generate", "--common-name='Authelia Standalone Root Certificate Authority'", "--ca", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tSelf-Signed")

	s.Assert().Contains(output, "\tCommon Name: Authelia Standalone Root Certificate Authority, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tCA: true, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Assert().Contains(output, "\tSubject Alternative Names: ")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/ca.private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/ca.public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateCAAndSignCertificate() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "rsa", "generate", "--common-name='Authelia Standalone Root Certificate Authority'", "--ca", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tSelf-Signed")

	s.Assert().Contains(output, "\tCommon Name: Authelia Standalone Root Certificate Authority, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tCA: true, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Assert().Contains(output, "\tSubject Alternative Names: ")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/ca.private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/ca.public.crt")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "rsa", "generate", "--common-name=example.com", "--sans='*.example.com'", "--path.ca", "/tmp/", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tAuthelia Standalone Root Certificate Authority")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, ", Expires: ")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tCA: false, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/public.crt")

	// Check the certificates look fine.
	privateKeyData, err := os.ReadFile("/tmp/private.pem")
	s.Assert().NoError(err)

	certificateData, err := os.ReadFile("/tmp/public.crt")
	s.Assert().NoError(err)

	privateKeyCAData, err := os.ReadFile("/tmp/ca.private.pem")
	s.Assert().NoError(err)

	certificateCAData, err := os.ReadFile("/tmp/ca.public.crt")
	s.Assert().NoError(err)

	s.Assert().Equal("", string(privateKeyData))
	s.Assert().Equal("", string(privateKeyCAData))
	s.Assert().Equal("", string(certificateData))
	s.Assert().Equal("", string(certificateCAData))

	s.Assert().False(bytes.Equal(privateKeyData, privateKeyCAData))
	s.Assert().False(bytes.Equal(certificateData, certificateCAData))

	privateKey, err := utils.ParseX509FromPEM(privateKeyData)
	s.Assert().NoError(err)
	s.Assert().True(utils.IsX509PrivateKey(privateKey))

	privateCAKey, err := utils.ParseX509FromPEM(privateKeyCAData)
	s.Assert().NoError(err)
	s.Assert().True(utils.IsX509PrivateKey(privateCAKey))

	c, err := utils.ParseX509FromPEM(certificateData)
	s.Assert().NoError(err)
	s.Assert().True(utils.IsX509PrivateKey(privateKey))

	cCA, err := utils.ParseX509FromPEM(privateKeyCAData)
	s.Assert().NoError(err)
	s.Assert().True(utils.IsX509PrivateKey(privateCAKey))

	certificate, ok := utils.CastX509AsCertificate(c)
	s.Assert().True(ok)

	certificateCA, ok := utils.CastX509AsCertificate(cCA)
	s.Assert().True(ok)

	s.Require().NotNil(certificate)
	s.Require().NotNil(certificateCA)

	err = certificate.CheckSignatureFrom(certificateCA)

	s.Assert().NoError(err)
}

func (s *CLISuite) TestShouldGenerateCertificateEd25519() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "ed25519", "generate", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tSelf-Signed")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tCA: false, CSR: false, Signature Algorithm: Ed25519, Public Key Algorithm: Ed25519")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/public.crt")
}

func (s *CLISuite) TestShouldFailGenerateCertificateParseNotBefore() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "ecdsa", "generate", "--not-before=invalid", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NotNil(err)
	s.Assert().Contains(output, "Error: failed to parse not before: parsing time \"invalid\" as \"Jan 2 15:04:05 2006\": cannot parse \"invalid\" as \"Jan\"")
}

func (s *CLISuite) TestShouldFailGenerateCertificateECDSA() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "ecdsa", "generate", "--curve=invalid", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NotNil(err)
	s.Assert().Contains(output, "Error: invalid curve 'invalid' was specified: curve must be P224, P256, P384, or P521")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSACurveP224() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "ecdsa", "generate", "--curve=P224", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tSelf-Signed")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tCA: false, CSR: false, Signature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-224")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSACurveP256() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "ecdsa", "generate", "--curve=P256", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tSelf-Signed")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tCA: false, CSR: false, Signature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-256")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSACurveP384() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "ecdsa", "generate", "--curve=P384", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tSelf-Signed")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tCA: false, CSR: false, Signature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-384")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSACurveP521() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "certificate", "ecdsa", "generate", "--curve=P521", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating Certificate")
	s.Assert().Contains(output, "\tSerial: ")
	s.Assert().Contains(output, "Signed By:\n\tSelf-Signed")

	s.Assert().Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Assert().Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Assert().Contains(output, "\tCA: false, CSR: false, Signature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-521")
	s.Assert().Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tCertificate: /tmp/public.crt")
}

func (s *CLISuite) TestShouldGenerateRSAKeyPair() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "pair", "rsa", "generate", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating key pair")

	s.Assert().Contains(output, "Algorithm: RSA-256 2048 bits\n\n")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tPublic Key: /tmp/public.pem")
}

func (s *CLISuite) TestShouldGenerateRSAKeyPairWith4069Bits() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "pair", "rsa", "generate", "--bits=4096", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating key pair")

	s.Assert().Contains(output, "Algorithm: RSA-512 4096 bits\n\n")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tPublic Key: /tmp/public.pem")
}

func (s *CLISuite) TestShouldGenerateECDSAKeyPair() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "pair", "ecdsa", "generate", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating key pair")

	s.Assert().Contains(output, "Algorithm: ECDSA Curve P-256\n\n")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tPublic Key: /tmp/public.pem")
}

func (s *CLISuite) TestShouldGenerateECDSAKeyPairCurveP224() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "pair", "ecdsa", "generate", "--curve=P224", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating key pair")

	s.Assert().Contains(output, "Algorithm: ECDSA Curve P-224\n\n")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tPublic Key: /tmp/public.pem")
}

func (s *CLISuite) TestShouldGenerateECDSAKeyPairCurveP256() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "pair", "ecdsa", "generate", "--curve=P256", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating key pair")

	s.Assert().Contains(output, "Algorithm: ECDSA Curve P-256\n\n")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tPublic Key: /tmp/public.pem")
}

func (s *CLISuite) TestShouldGenerateECDSAKeyPairCurveP384() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "pair", "ecdsa", "generate", "--curve=P384", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating key pair")

	s.Assert().Contains(output, "Algorithm: ECDSA Curve P-384\n\n")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tPublic Key: /tmp/public.pem")
}

func (s *CLISuite) TestShouldGenerateECDSAKeyPairCurveP521() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "pair", "ecdsa", "generate", "--curve=P521", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating key pair")

	s.Assert().Contains(output, "Algorithm: ECDSA Curve P-521\n\n")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tPublic Key: /tmp/public.pem")
}

func (s *CLISuite) TestShouldGenerateEd25519KeyPair() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "pair", "ed25519", "generate", "--directory=/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Generating key pair")

	s.Assert().Contains(output, "Algorithm: Ed25519\n\n")

	s.Assert().Contains(output, "Output Paths:")
	s.Assert().Contains(output, "\tPrivate Key: /tmp/private.pem")
	s.Assert().Contains(output, "\tPublic Key: /tmp/public.pem")
}

func (s *CLISuite) TestShouldNotGenerateECDSAKeyPairCurveInvalid() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "crypto", "pair", "ecdsa", "generate", "--curve=invalid", "--directory=/tmp/"})
	s.Assert().NotNil(err)
	s.Assert().Contains(output, "Error: invalid curve 'invalid' was specified: curve must be P224, P256, P384, or P521")
}

func (s *CLISuite) TestStorageShouldShowErrWithoutConfig() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "schema-info"})
	s.Assert().EqualError(err, "exit status 1")

	s.Assert().Contains(output, "Error: storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided, storage: option 'encryption_key' must is required\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "history"})
	s.Assert().EqualError(err, "exit status 1")

	s.Assert().Contains(output, "Error: storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided, storage: option 'encryption_key' must is required\n")
}

func (s *CLISuite) TestStorage00ShouldShowCorrectPreInitInformation() {
	_ = os.Remove("/tmp/db.sqlite3")

	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "schema-info", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	pattern := regexp.MustCompile(`^Schema Version: N/A\nSchema Upgrade Available: yes - version \d+\nSchema Tables: N/A\nSchema Encryption Key: unsupported \(schema version\)`)

	s.Assert().Regexp(pattern, output)

	patternOutdated := regexp.MustCompile(`Error: schema is version \d+ which is outdated please migrate to version \d+ in order to use this command or use an older binary`)

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "export", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Regexp(patternOutdated, output)

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "change-key", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Regexp(patternOutdated, output)

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Could not check encryption key for validity. The schema version doesn't support encryption.")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "down", "--target=0", "--destroy-data", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: schema migration target version 0 is the same current version 0")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "up", "--target=2147483640", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: schema up migration target version 2147483640 is greater then the latest version ")
	s.Assert().Contains(output, " which indicates it doesn't exist")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "history", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "No migration history is available for schemas that not version 1 or above.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "list-up", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Storage Schema Migration List (Up)\n\nVersion\t\tDescription\n1\t\tInitial Schema\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "list-down", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Storage Schema Migration List (Down)\n\nNo Migrations Available\n")
}

func (s *CLISuite) TestStorage01ShouldMigrateUp() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "up", "--config=/config/configuration.storage.yml"})
	s.Require().NoError(err)

	pattern0 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is being attempted"`)
	pattern1 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is complete"`)

	s.Regexp(pattern0, output)
	s.Regexp(pattern1, output)

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "up", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")

	s.Assert().Contains(output, "Error: schema already up to date\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "history", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Migration History:\n\nID\tDate\t\t\t\tBefore\tAfter\tAuthelia Version\n")
	s.Assert().Contains(output, "0\t1")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "list-up", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Storage Schema Migration List (Up)\n\nNo Migrations Available")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "list-down", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Storage Schema Migration List (Down)\n\nVersion\t\tDescription\n")
	s.Assert().Contains(output, "1\t\tInitial Schema")
}

func (s *CLISuite) TestStorage02ShouldShowSchemaInfo() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "schema-info", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Schema Version: ")
	s.Assert().Contains(output, "authentication_logs")
	s.Assert().Contains(output, "identity_verification")
	s.Assert().Contains(output, "duo_devices")
	s.Assert().Contains(output, "user_preferences")
	s.Assert().Contains(output, "migrations")
	s.Assert().Contains(output, "encryption")
	s.Assert().Contains(output, "encryption")
	s.Assert().Contains(output, "webauthn_devices")
	s.Assert().Contains(output, "totp_configurations")
	s.Assert().Contains(output, "Schema Encryption Key: valid")
}

func (s *CLISuite) TestStorage03ShouldExportTOTP() {
	storageProvider := storage.NewSQLiteProvider(&storageLocalTmpConfig)

	ctx := context.Background()

	var (
		err error
	)

	var (
		expectedLines    = make([]string, 0, 3)
		expectedLinesCSV = make([]string, 0, 4)
		output           string
	)

	expectedLinesCSV = append(expectedLinesCSV, "issuer,username,algorithm,digits,period,secret")

	testCases := []struct {
		config model.TOTPConfiguration
		png    bool
	}{
		{
			config: model.TOTPConfiguration{
				Username:  "john",
				Period:    30,
				Digits:    6,
				Algorithm: "SHA1",
			},
		},
		{
			config: model.TOTPConfiguration{
				Username:  "mary",
				Period:    45,
				Digits:    6,
				Algorithm: "SHA1",
			},
		},
		{
			config: model.TOTPConfiguration{
				Username:  "fred",
				Period:    30,
				Digits:    8,
				Algorithm: "SHA1",
			},
		},
		{
			config: model.TOTPConfiguration{
				Username:  "jone",
				Period:    30,
				Digits:    6,
				Algorithm: "SHA512",
			},
			png: true,
		},
	}

	var (
		config   *model.TOTPConfiguration
		fileInfo os.FileInfo
	)

	for _, testCase := range testCases {
		if testCase.png {
			output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "generate", testCase.config.Username, "--period", strconv.Itoa(int(testCase.config.Period)), "--algorithm", testCase.config.Algorithm, "--digits", strconv.Itoa(int(testCase.config.Digits)), "--path=/tmp/qr.png", "--config=/config/configuration.storage.yml"})
			s.Assert().NoError(err)
			s.Assert().Contains(output, " and saved it as a PNG image at the path '/tmp/qr.png'")

			fileInfo, err = os.Stat("/tmp/qr.png")
			s.Assert().NoError(err)
			s.Require().NotNil(fileInfo)
			s.Assert().False(fileInfo.IsDir())
			s.Assert().Greater(fileInfo.Size(), int64(1000))
		} else {
			output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "generate", testCase.config.Username, "--period", strconv.Itoa(int(testCase.config.Period)), "--algorithm", testCase.config.Algorithm, "--digits", strconv.Itoa(int(testCase.config.Digits)), "--config=/config/configuration.storage.yml"})
			s.Assert().NoError(err)
		}

		config, err = storageProvider.LoadTOTPConfiguration(ctx, testCase.config.Username)
		s.Assert().NoError(err)

		s.Assert().Contains(output, config.URI())

		expectedLinesCSV = append(expectedLinesCSV, fmt.Sprintf("%s,%s,%s,%d,%d,%s", "Authelia", config.Username, config.Algorithm, config.Digits, config.Period, string(config.Secret)))
		expectedLines = append(expectedLines, config.URI())
	}

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "export", "--format=uri", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	for _, expectedLine := range expectedLines {
		s.Assert().Contains(output, expectedLine)
	}

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "export", "--format=csv", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	for _, expectedLine := range expectedLinesCSV {
		s.Assert().Contains(output, expectedLine)
	}

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "export", "--format=wrong", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: format must be csv, uri, or png")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "export", "--format=png", "--dir=/tmp/qr", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Exported TOTP QR codes in PNG format in the '/tmp/qr' directory")

	for _, testCase := range testCases {
		fileInfo, err = os.Stat(fmt.Sprintf("/tmp/qr/%s.png", testCase.config.Username))

		s.Assert().NoError(err)
		s.Require().NotNil(fileInfo)

		s.Assert().False(fileInfo.IsDir())
		s.Assert().Greater(fileInfo.Size(), int64(1000))
	}

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "generate", "test", "--period=30", "--algorithm=SHA1", "--digits=6", "--path=/tmp/qr.png", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: image output filepath already exists")
}

func (s *CLISuite) TestStorage04ShouldManageUniqueID() {
	_ = os.Mkdir("/tmp/out", 0777)

	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "export", "--file=out.yml", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: no data to export")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=webauthn", "--sector=''", "--identifier=1097c8f8-83f2-4506-8138-5f40e83a1285", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: the service name 'webauthn' is invalid, the valid values are: 'openid'")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector=''", "--identifier=1097c8f8-83f2-4506-8138-5f40e83a1285", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Added User Opaque Identifier:\n\tService: openid\n\tSector: \n\tUsername: john\n\tIdentifier: 1097c8f8-83f2-4506-8138-5f40e83a1285\n\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "export", "--file=/a/no/path/fileout.yml", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: error occurred writing to file '/a/no/path/fileout.yml': open /a/no/path/fileout.yml: no such file or directory")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "export", "--file=out.yml", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: error occurred writing to file 'out.yml': open out.yml: permission denied")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "export", "--file=/tmp/out/1.yml", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Exported 1 User Opaque Identifiers to /tmp/out/1.yml")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "export", "--file=/tmp/out/1.yml", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: must specify a file that doesn't exist but '/tmp/out/1.yml' exists")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector=''", "--identifier=1097c8f8-83f2-4506-8138-5f40e83a1285", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: error inserting user opaque id for user 'john' with opaque id '1097c8f8-83f2-4506-8138-5f40e83a1285':")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector=''", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: error inserting user opaque id for user 'john' with opaque id")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector='openidconnect.com'", "--identifier=1097c8f8-83f2-4506-8138-5f40e83a1285", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: error inserting user opaque id for user 'john' with opaque id '1097c8f8-83f2-4506-8138-5f40e83a1285':")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector='openidconnect.net'", "--identifier=b0e17f48-933c-4cba-8509-ee9bfadf8ce5", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Added User Opaque Identifier:\n\tService: openid\n\tSector: openidconnect.net\n\tUsername: john\n\tIdentifier: b0e17f48-933c-4cba-8509-ee9bfadf8ce5\n\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector='bad-uuid.com'", "--identifier=d49564dc-b7a1-11ec-8429-fcaa147128ea", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: the identifier providerd 'd49564dc-b7a1-11ec-8429-fcaa147128ea' is a version 1 UUID but only version 4 UUID's accepted as identifiers")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector='bad-uuid.com'", "--identifier=asdmklasdm", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: the identifier provided 'asdmklasdm' is invalid as it must be a version 4 UUID but parsing it had an error: invalid UUID length: 10")

	data, err := os.ReadFile("/tmp/out/1.yml")
	s.Assert().NoError(err)

	var export model.UserOpaqueIdentifiersExport

	s.Assert().NoError(yaml.Unmarshal(data, &export))

	s.Require().Len(export.Identifiers, 1)

	s.Assert().Equal(1, export.Identifiers[0].ID)
	s.Assert().Equal("1097c8f8-83f2-4506-8138-5f40e83a1285", export.Identifiers[0].Identifier.String())
	s.Assert().Equal("john", export.Identifiers[0].Username)
	s.Assert().Equal("", export.Identifiers[0].SectorID)
	s.Assert().Equal("openid", export.Identifiers[0].Service)

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "export", "--file=/tmp/out/2.yml", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Exported 2 User Opaque Identifiers to /tmp/out/2.yml")

	export = model.UserOpaqueIdentifiersExport{}

	data, err = os.ReadFile("/tmp/out/2.yml")
	s.Assert().NoError(err)

	s.Assert().NoError(yaml.Unmarshal(data, &export))

	s.Require().Len(export.Identifiers, 2)

	s.Assert().Equal(1, export.Identifiers[0].ID)
	s.Assert().Equal("1097c8f8-83f2-4506-8138-5f40e83a1285", export.Identifiers[0].Identifier.String())
	s.Assert().Equal("john", export.Identifiers[0].Username)
	s.Assert().Equal("", export.Identifiers[0].SectorID)
	s.Assert().Equal("openid", export.Identifiers[0].Service)

	s.Assert().Equal(2, export.Identifiers[1].ID)
	s.Assert().Equal("b0e17f48-933c-4cba-8509-ee9bfadf8ce5", export.Identifiers[1].Identifier.String())
	s.Assert().Equal("john", export.Identifiers[1].Username)
	s.Assert().Equal("openidconnect.net", export.Identifiers[1].SectorID)
	s.Assert().Equal("openid", export.Identifiers[1].Service)
}

func (s *CLISuite) TestStorage05ShouldChangeEncryptionKey() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "change-key", "--new-encryption-key=apple-apple-apple-apple", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Completed the encryption key change. Please adjust your configuration to use the new key.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "schema-info", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Schema Version: ")
	s.Assert().Contains(output, "authentication_logs")
	s.Assert().Contains(output, "identity_verification")
	s.Assert().Contains(output, "duo_devices")
	s.Assert().Contains(output, "user_preferences")
	s.Assert().Contains(output, "migrations")
	s.Assert().Contains(output, "encryption")
	s.Assert().Contains(output, "encryption")
	s.Assert().Contains(output, "webauthn_devices")
	s.Assert().Contains(output, "totp_configurations")
	s.Assert().Contains(output, "Schema Encryption Key: invalid")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Encryption key validation: failed.\n\nError: the encryption key is not valid against the schema check value.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--verbose", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Encryption key validation: failed.\n\nError: the encryption key is not valid against the schema check value, 4 of 4 total TOTP secrets were invalid.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--encryption-key=apple-apple-apple-apple", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Encryption key validation: success.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--verbose", "--encryption-key=apple-apple-apple-apple", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Encryption key validation: success.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "change-key", "--encryption-key=apple-apple-apple-apple", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")

	s.Assert().Contains(output, "Error: you must set the --new-encryption-key flag\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "change-key", "--encryption-key=apple-apple-apple-apple", "--new-encryption-key=abc", "--config=/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")

	s.Assert().Contains(output, "Error: the new encryption key must be at least 20 characters\n")
}

func (s *CLISuite) TestStorage06ShouldMigrateDown() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "down", "--target=0", "--destroy-data", "--config=/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	pattern0 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is being attempted"`)
	pattern1 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is complete"`)

	s.Regexp(pattern0, output)
	s.Regexp(pattern1, output)
}

func (s *CLISuite) TestACLPolicyCheckVerbose() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "access-control", "check-policy", "--url=https://public.example.com", "--verbose", "--config=/config/configuration.yml"})
	s.Assert().NoError(err)

	// This is an example of `authelia access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://public.example.com --verbose`.
	s.Contains(output, "Performing policy check for request to 'https://public.example.com' method 'GET'.\n\n")
	s.Contains(output, "  #\tDomain\tResource\tMethod\tNetwork\tSubject\n")
	s.Contains(output, "* 1\thit\thit\t\thit\thit\thit\n")
	s.Contains(output, "  2\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  3\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  4\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  5\tmiss\tmiss\t\thit\thit\thit\n")
	s.Contains(output, "  6\tmiss\thit\t\tmiss\thit\thit\n")
	s.Contains(output, "  7\tmiss\thit\t\thit\tmiss\thit\n")
	s.Contains(output, "  8\tmiss\thit\t\thit\thit\tmay\n")
	s.Contains(output, "  9\tmiss\thit\t\thit\thit\tmay\n")
	s.Contains(output, "The policy 'bypass' from rule #1 will be applied to this request.")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "access-control", "check-policy", "--url=https://admin.example.com", "--method=HEAD", "--username=tom", "--groups=basic,test", "--ip=192.168.2.3", "--verbose", "--config=/config/configuration.yml"})
	s.Assert().NoError(err)

	// This is an example of `authelia access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://admin.example.com --method=HEAD --username=tom --groups=basic,test --ip=192.168.2.3 --verbose`.
	s.Contains(output, "Performing policy check for request to 'https://admin.example.com' method 'HEAD' username 'tom' groups 'basic,test' from IP '192.168.2.3'.\n\n")
	s.Contains(output, "  #\tDomain\tResource\tMethod\tNetwork\tSubject\n")
	s.Contains(output, "  #\tDomain\tResource\tMethod\tNetwork\tSubject\n")
	s.Contains(output, "  1\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "* 2\thit\thit\t\thit\thit\thit\n")
	s.Contains(output, "  3\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  4\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  5\tmiss\tmiss\t\thit\thit\thit\n")
	s.Contains(output, "  6\tmiss\thit\t\tmiss\thit\thit\n")
	s.Contains(output, "  7\tmiss\thit\t\thit\tmiss\thit\n")
	s.Contains(output, "  8\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  9\tmiss\thit\t\thit\thit\tmiss\n")
	s.Contains(output, "The policy 'two_factor' from rule #2 will be applied to this request.")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "access-control", "check-policy", "--url=https://resources.example.com/resources/test", "--method=POST", "--username=john", "--groups=admin,test", "--ip=192.168.1.3", "--verbose", "--config=/config/configuration.yml"})
	s.Assert().NoError(err)

	// This is an example of `authelia access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://resources.example.com/resources/test --method=POST --username=john --groups=admin,test --ip=192.168.1.3 --verbose`.
	s.Contains(output, "Performing policy check for request to 'https://resources.example.com/resources/test' method 'POST' username 'john' groups 'admin,test' from IP '192.168.1.3'.\n\n")
	s.Contains(output, "  #\tDomain\tResource\tMethod\tNetwork\tSubject\n")
	s.Contains(output, "  1\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  2\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  3\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  4\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "* 5\thit\thit\t\thit\thit\thit\n")
	s.Contains(output, "  6\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  7\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  8\tmiss\thit\t\thit\thit\tmiss\n")
	s.Contains(output, "  9\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "The policy 'one_factor' from rule #5 will be applied to this request.")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "access-control", "check-policy", "--url=https://user.example.com/resources/test", "--method=HEAD", "--username=john", "--groups=admin,test", "--ip=192.168.1.3", "--verbose", "--config=/config/configuration.yml"})
	s.Assert().NoError(err)

	// This is an example of `access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://user.example.com --method=HEAD --username=john --groups=admin,test --ip=192.168.1.3 --verbose`.
	s.Contains(output, "Performing policy check for request to 'https://user.example.com/resources/test' method 'HEAD' username 'john' groups 'admin,test' from IP '192.168.1.3'.\n\n")
	s.Contains(output, "  #\tDomain\tResource\tMethod\tNetwork\tSubject\n")
	s.Contains(output, "  1\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  2\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  3\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  4\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  5\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  6\tmiss\thit\t\tmiss\thit\thit\n")
	s.Contains(output, "  7\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  8\tmiss\thit\t\thit\thit\tmiss\n")
	s.Contains(output, "* 9\thit\thit\t\thit\thit\thit\n")
	s.Contains(output, "The policy 'one_factor' from rule #9 will be applied to this request.")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "access-control", "check-policy", "--url=https://user.example.com", "--method=HEAD", "--ip=192.168.1.3", "--verbose", "--config=/config/configuration.yml"})
	s.Assert().NoError(err)

	// This is an example of `authelia access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://user.example.com --method=HEAD --ip=192.168.1.3 --verbose`.
	s.Contains(output, "Performing policy check for request to 'https://user.example.com' method 'HEAD' from IP '192.168.1.3'.\n\n")
	s.Contains(output, "  #\tDomain\tResource\tMethod\tNetwork\tSubject\n")
	s.Contains(output, "  1\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  2\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  3\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  4\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  5\tmiss\tmiss\t\thit\thit\thit\n")
	s.Contains(output, "  6\tmiss\thit\t\tmiss\thit\thit\n")
	s.Contains(output, "  7\tmiss\thit\t\thit\thit\thit\n")
	s.Contains(output, "  8\tmiss\thit\t\thit\thit\tmay\n")
	s.Contains(output, "~ 9\thit\thit\t\thit\thit\tmay\n")
	s.Contains(output, "The policy 'one_factor' from rule #9 will potentially be applied to this request. Otherwise the policy 'bypass' from the default policy will be.")
}

func TestCLISuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewCLISuite())
}
