package suites

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.yaml.in/yaml/v4"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/utils"
)

type CLISuite struct {
	*CommandSuite
}

func NewCLISuite() *CLISuite {
	return &CLISuite{
		CommandSuite: &CommandSuite{
			BaseSuite: &BaseSuite{
				Name: cliSuiteName,
			},
		},
	}
}

func (s *CLISuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/compose.yml",
		"internal/suites/CLI/compose.yml",
		"internal/suites/example/compose/authelia/compose.backend.{}.yml",
	})
	s.DockerEnvironment = dockerEnvironment
}

func (s *CLISuite) TestShouldPrintBuildInformation() {
	if os.Getenv("CI") == "false" {
		s.T().Skip("Skipping testing in dev environment")
	}

	output, err := s.Exec("authelia-backend", []string{"authelia", "build-info"})
	s.NoError(err)
	s.Contains(output, "Last Tag: ")
	s.Contains(output, "State: ")
	s.Contains(output, "Branch: ")
	s.Contains(output, "Build Number: ")
	s.Contains(output, "Build OS: ")
	s.Contains(output, "Build Arch: ")
	s.Contains(output, "Build Date: ")
	s.Contains(output, "Development: ")

	r := regexp.MustCompile(`^Last Tag: v\d+\.\d+\.\d+\nState: (tagged|untagged) (clean|dirty)\nBranch: [^\s\n]+\nCommit: [0-9a-f]{40}\nBuild Number: \d+\nBuild OS: (linux|darwin|windows|freebsd)\nBuild Arch: (amd64|arm|arm64)\nBuild Compiler: gc\nBuild Date: (Sun|Mon|Tue|Wed|Thu|Fri|Sat), \d{2} (Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d{4} \d{2}:\d{2}:\d{2} [+-]\d{4}\nDevelopment: (true|false)\nExtra: \n\nGo:\n\s+Version: go\d+\.\d+\.\d+ X:nosynchashtriemap\n\s+Module Path: github.com/authelia/authelia/v4\n\s+Executable Path: github.com/authelia/authelia/v4/cmd/authelia`)
	s.Regexp(r, output)

	output, err = s.Exec("authelia-backend", []string{"authelia", "build-info", "-v"})
	s.NoError(err)
	s.Contains(output, "Last Tag: ")
	s.Contains(output, "State: ")
	s.Contains(output, "Branch: ")
	s.Contains(output, "Build Number: ")
	s.Contains(output, "Build OS: ")
	s.Contains(output, "Build Arch: ")
	s.Contains(output, "Build Date: ")
	s.Contains(output, "Development: ")

	s.Regexp(r, output)
}

func (s *CLISuite) TestShouldPrintVersion() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "--version"})
	s.NoError(err)
	s.Contains(output, "authelia version")
}

func (s *CLISuite) TestShouldValidateConfig() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "validate-config"})
	s.NoError(err)
	s.Contains(output, "Configuration parsed and loaded successfully without errors.")
}

func (s *CLISuite) TestShouldFailValidateConfig() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "validate-config", "--config=/config/invalid.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "failed to load configuration from file path(/config/invalid.yml) source: stat /config/invalid.yml: no such file or directory\n")
}

func (s *CLISuite) TestShouldHashPasswordArgon2() {
	var (
		output string
		err    error
	)

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "argon2", "--password=apple123", "-m=32768"})
	s.NoError(err)
	s.Contains(output, "Digest: $argon2id$v=19$m=32768,t=3,p=4$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "argon2", "--password=apple123", "-m", "32768", "-v=argon2i"})
	s.NoError(err)
	s.Contains(output, "Digest: $argon2i$v=19$m=32768,t=3,p=4$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "argon2", "--password=apple123", "-m=32768", "-v=argon2d"})
	s.NoError(err)
	s.Contains(output, "Digest: $argon2d$v=19$m=32768,t=3,p=4$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "argon2", "--random", "-m=32"})
	s.NoError(err)
	s.Contains(output, "Random Password: ")
	s.Contains(output, "Digest: $argon2id$v=19$m=32,t=3,p=4$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "argon2", "--password=apple123", "-p=1"})
	s.NoError(err)
	s.Contains(output, "Digest: $argon2id$v=19$m=65536,t=3,p=1$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "argon2", "--password=apple123", "-i=1"})
	s.NoError(err)
	s.Contains(output, "Digest: $argon2id$v=19$m=65536,t=1,p=4$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "argon2", "--password=apple123", "-s=64"})
	s.NoError(err)
	s.Contains(output, "Digest: $argon2id$v=19$m=65536,t=3,p=4$")
	s.GreaterOrEqual(len(output), 169)

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "argon2", "--password=apple123", "-k=128"})
	s.NoError(err)
	s.Contains(output, "Digest: $argon2id$v=19$m=65536,t=3,p=4$")
	s.GreaterOrEqual(len(output), 233)
}

func (s *CLISuite) TestShouldHashPasswordSHA2Crypt() {
	var (
		output string
		err    error
	)

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "sha2crypt", "--password=apple123", "-v=sha256"})
	s.NoError(err)
	s.Contains(output, "Digest: $5$rounds=50000$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "sha2crypt", "--password=apple123", "-v=sha512"})
	s.NoError(err)
	s.Contains(output, "Digest: $6$rounds=50000$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "sha2crypt", "--random", "-s=8"})
	s.NoError(err)
	s.Contains(output, "Digest: $6$rounds=50000$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "sha2crypt", "--password=apple123", "-i=10000"})
	s.NoError(err)
	s.Contains(output, "Digest: $6$rounds=10000$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "sha2crypt", "--password=apple123", "-s=20"})
	s.NotNil(err)
	s.Contains(output, "Error: errors occurred validating the password configuration: authentication_backend: file: password: sha2crypt: option 'salt_length' is configured as '20' but must be less than or equal to '16'")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "sha2crypt", "--password=apple123", "-i=20"})
	s.NotNil(err)
	s.Contains(output, "Error: errors occurred validating the password configuration: authentication_backend: file: password: sha2crypt: option 'iterations' is configured as '20' but must be greater than or equal to '1000'")
}

func (s *CLISuite) TestShouldHashPasswordSHA2CryptSHA512() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "sha2crypt", "--password=apple123", "-v=sha512"})
	s.NoError(err)
	s.Contains(output, "Digest: $6$rounds=50000$")
}

func (s *CLISuite) TestShouldHashPasswordPBKDF2() {
	var (
		output string
		err    error
	)

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "pbkdf2", "--password=apple123", "-v=sha1"})
	s.NoError(err)
	s.Contains(output, "Digest: $pbkdf2$310000$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "pbkdf2", "--random", "-v=sha256", "-i=100000"})
	s.NoError(err)
	s.Contains(output, "Random Password: ")
	s.Contains(output, "Digest: $pbkdf2-sha256$100000$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "pbkdf2", "--password=apple123", "-v=sha512", "-i=100000"})
	s.NoError(err)
	s.Contains(output, "Digest: $pbkdf2-sha512$100000$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "pbkdf2", "--password=apple123", "-v=sha224", "-i=100000"})
	s.NoError(err)
	s.Contains(output, "Digest: $pbkdf2-sha224$100000$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "pbkdf2", "--password=apple123", "-v=sha384", "-i=100000"})
	s.NoError(err)
	s.Contains(output, "Digest: $pbkdf2-sha384$100000$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "pbkdf2", "--password=apple123", "-s=32", "-i=100000"})
	s.NoError(err)
	s.Contains(output, "Digest: $pbkdf2-sha512$100000$")
}

func (s *CLISuite) TestShouldHashPasswordBcrypt() {
	var (
		output string
		err    error
	)

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "bcrypt", "--password=apple123"})
	s.NoError(err)
	s.Contains(output, "Digest: $2b$12$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "bcrypt", "--random", "-i=10"})
	s.NoError(err)
	s.Contains(output, "Random Password: ")
	s.Contains(output, "Digest: $2b$10$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "bcrypt", "--password=apple123", "-v=sha256"})
	s.NoError(err)
	s.Contains(output, "Digest: $bcrypt-sha256$v=2,t=2b,r=12$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "bcrypt", "--random", "-v=sha256", "-i=10"})
	s.NoError(err)
	s.Contains(output, "Random Password: ")
	s.Contains(output, "Digest: $bcrypt-sha256$v=2,t=2b,r=10$")
}

func (s *CLISuite) TestShouldHashPasswordScrypt() {
	var (
		output string
		err    error
	)

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "scrypt", "--password=apple123"})
	s.NoError(err)
	s.Contains(output, "Digest: $scrypt$ln=16,r=8,p=1$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "scrypt", "--random"})
	s.NoError(err)
	s.Contains(output, "Random Password: ")
	s.Contains(output, "Digest: $scrypt$ln=16,r=8,p=1$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "scrypt", "--password=apple123", "-i=1"})
	s.NoError(err)
	s.Contains(output, "Digest: $scrypt$ln=1,r=8,p=1$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "scrypt", "--password=apple123", "-i=1", "-p=2"})
	s.NoError(err)
	s.Contains(output, "Digest: $scrypt$ln=1,r=8,p=2$")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "hash", "generate", "scrypt", "--password=apple123", "-i=1", "-r=2"})
	s.NoError(err)
	s.Contains(output, "Digest: $scrypt$ln=1,r=2,p=1$")
}

func (s *CLISuite) TestShouldGenerateRSACertificateRequest() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "request", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate Request")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tSignature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate Request: request.csr")
}

func (s *CLISuite) TestShouldGenerateECDSACurveP224CertificateRequest() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "ecdsa", "request", "--curve=P224", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate Request")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tSignature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-224")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate Request: request.csr")
}

func (s *CLISuite) TestShouldGenerateECDSACurveP256CertificateRequest() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "ecdsa", "request", "--curve=P256", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate Request")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tSignature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-256")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate Request: request.csr")
}

func (s *CLISuite) TestShouldGenerateECDSACurveP384CertificateRequest() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "ecdsa", "request", "--curve=P384", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate Request")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tSignature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-384")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate Request: request.csr")
}

func (s *CLISuite) TestShouldGenerateECDSACurveP521CertificateRequest() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "ecdsa", "request", "--curve=P521", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate Request")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tSignature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-521")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate Request: request.csr")
}

func (s *CLISuite) TestShouldGenerateEd25519CertificateRequest() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "ed25519", "request", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate Request")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tSignature Algorithm: Ed25519, Public Key Algorithm: Ed25519")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate Request: request.csr")
}

func (s *CLISuite) TestShouldGenerateCertificateRSA() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tSelf-Signed")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tCA: false, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate: public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateRSAWithIPAddress() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--common-name=example.com", "--sans", "*.example.com,127.0.0.1", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tSelf-Signed")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tCA: false, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com, IP.1:127.0.0.1")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate: public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateRSAWithNotBefore() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--common-name=example.com", "--sans='*.example.com'", "--not-before", "'Jan 1 15:04:05 2011'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tSelf-Signed")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tNot Before: 2011-01-01T15:04:05Z, Not After: 2012-01-01T15:04:05Z")
	s.Contains(output, "\tCA: false, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate: public.crt")
}

func (s *CLISuite) TestShouldFailGenerateCertificateRSAWithInvalidNotBefore() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--common-name=example.com", "--sans='*.example.com'", "--not-before", "Jan", "--directory=/tmp/"})
	s.NotNil(err)
	s.Contains(output, "Error: failed to parse not before: failed to find a suitable time layout for time 'Jan'")
}

func (s *CLISuite) TestShouldGenerateCertificateRSAWith4096Bits() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--bits=4096", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tSelf-Signed")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tCA: false, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 4096")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate: public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateWithCustomizedSubject() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--common-name=example.com", "--sans='*.example.com'", "--country=Australia", "--organization='Acme Co.'", "--organizational-unit=Tech", "--province=QLD", "--street-address='123 Smith St'", "--postcode=4000", "--locality=Internet", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tSelf-Signed")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Acme Co.], Organizational Unit: [Tech]")
	s.Contains(output, "\tCountry: [Australia], Province: [QLD], Street Address: [123 Smith St], Postal Code: [4000], Locality: [Internet]")
	s.Contains(output, "\tCA: false, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate: public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateCA() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--common-name='Authelia Standalone Root Certificate Authority'", "--ca", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tSelf-Signed")

	s.Contains(output, "\tCommon Name: Authelia Standalone Root Certificate Authority, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tCA: true, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Contains(output, "\tSubject Alternative Names: ")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: ca.private.pem")
	s.Contains(output, "\tCertificate: ca.public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateCAAndSignCertificate() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--common-name='Authelia Standalone Root Certificate Authority'", "--ca", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tSelf-Signed")

	s.Contains(output, "\tCommon Name: Authelia Standalone Root Certificate Authority, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tCA: true, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Contains(output, "\tSubject Alternative Names: ")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: ca.private.pem")
	s.Contains(output, "\tCertificate: ca.public.crt")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--common-name=example.com", "--sans='*.example.com'", "--path.ca", "/tmp/", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tAuthelia Standalone Root Certificate Authority")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, ", Expires: ")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tCA: false, CSR: false, Signature Algorithm: SHA256-RSA, Public Key Algorithm: RSA, Bits: 2048")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate: public.crt")

	// Check the certificates look fine.
	privateKeyData, err := os.ReadFile("/tmp/private.pem")
	s.NoError(err)

	certificateData, err := os.ReadFile("/tmp/public.crt")
	s.NoError(err)

	privateKeyCAData, err := os.ReadFile("/tmp/ca.private.pem")
	s.NoError(err)

	certificateCAData, err := os.ReadFile("/tmp/ca.public.crt")
	s.NoError(err)

	s.False(bytes.Equal(privateKeyData, privateKeyCAData))
	s.False(bytes.Equal(certificateData, certificateCAData))

	privateKey, err := utils.ParseX509FromPEM(privateKeyData)
	s.NoError(err)
	s.True(utils.IsX509PrivateKey(privateKey))

	privateCAKey, err := utils.ParseX509FromPEM(privateKeyCAData)
	s.NoError(err)
	s.True(utils.IsX509PrivateKey(privateCAKey))

	c, err := utils.ParseX509FromPEM(certificateData)
	s.NoError(err)
	s.False(utils.IsX509PrivateKey(c))

	cCA, err := utils.ParseX509FromPEM(certificateCAData)
	s.NoError(err)
	s.False(utils.IsX509PrivateKey(cCA))

	certificate, ok := utils.AssertToX509Certificate(c)
	s.True(ok)

	certificateCA, ok := utils.AssertToX509Certificate(cCA)
	s.True(ok)

	s.Require().NotNil(certificate)
	s.Require().NotNil(certificateCA)

	err = certificate.CheckSignatureFrom(certificateCA)

	s.NoError(err)
}

func (s *CLISuite) TestShouldGenerateCertificateEd25519() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "ed25519", "generate", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tSelf-Signed")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tCA: false, CSR: false, Signature Algorithm: Ed25519, Public Key Algorithm: Ed25519")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate: public.crt")
}

func (s *CLISuite) TestShouldFailGenerateCertificateParseNotBefore() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "ecdsa", "generate", "--not-before=invalid", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NotNil(err)
	s.Contains(output, "Error: failed to parse not before: failed to find a suitable time layout for time 'invalid'")
}

func (s *CLISuite) TestShouldFailGenerateCertificateECDSA() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "ecdsa", "generate", "--curve=invalid", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NotNil(err)
	s.Contains(output, "Error: invalid curve 'invalid' was specified: curve must be P224, P256, P384, or P521")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSACurveP224() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "ecdsa", "generate", "--curve=P224", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tSelf-Signed")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tCA: false, CSR: false, Signature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-224")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate: public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSACurveP256() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "ecdsa", "generate", "--curve=P256", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tSelf-Signed")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tCA: false, CSR: false, Signature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-256")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate: public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSACurveP384() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "ecdsa", "generate", "--curve=P384", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tSelf-Signed")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tCA: false, CSR: false, Signature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-384")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate: public.crt")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSACurveP521() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "ecdsa", "generate", "--curve=P521", "--common-name=example.com", "--sans='*.example.com'", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating Certificate")
	s.Contains(output, "\tSerial: ")
	s.Contains(output, "Signed By:\n\tSelf-Signed")

	s.Contains(output, "\tCommon Name: example.com, Organization: [Authelia], Organizational Unit: []")
	s.Contains(output, "\tCountry: [], Province: [], Street Address: [], Postal Code: [], Locality: []")
	s.Contains(output, "\tCA: false, CSR: false, Signature Algorithm: ECDSA-SHA256, Public Key Algorithm: ECDSA, Elliptic Curve: P-521")
	s.Contains(output, "\tSubject Alternative Names: DNS.1:*.example.com")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tCertificate: public.crt")
}

func (s *CLISuite) TestShouldGenerateRSAKeyPair() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "pair", "rsa", "generate", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating key pair")

	s.Contains(output, "Algorithm: RSA-256 2048 bits\n\n")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tPublic Key: public.pem")
}

func (s *CLISuite) TestShouldGenerateRSAKeyPairWith4069Bits() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "pair", "rsa", "generate", "--bits=4096", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating key pair")

	s.Contains(output, "Algorithm: RSA-512 4096 bits\n\n")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tPublic Key: public.pem")
}

func (s *CLISuite) TestShouldGenerateECDSAKeyPair() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "pair", "ecdsa", "generate", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating key pair")

	s.Contains(output, "Algorithm: ECDSA Curve P-256\n\n")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tPublic Key: public.pem")
}

func (s *CLISuite) TestShouldGenerateECDSAKeyPairCurveP224() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "pair", "ecdsa", "generate", "--curve=P224", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating key pair")

	s.Contains(output, "Algorithm: ECDSA Curve P-224\n\n")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tPublic Key: public.pem")
}

func (s *CLISuite) TestShouldGenerateECDSAKeyPairCurveP256() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "pair", "ecdsa", "generate", "--curve=P256", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating key pair")

	s.Contains(output, "Algorithm: ECDSA Curve P-256\n\n")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tPublic Key: public.pem")
}

func (s *CLISuite) TestShouldGenerateECDSAKeyPairCurveP384() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "pair", "ecdsa", "generate", "--curve=P384", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating key pair")

	s.Contains(output, "Algorithm: ECDSA Curve P-384\n\n")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tPublic Key: public.pem")
}

func (s *CLISuite) TestShouldGenerateECDSAKeyPairCurveP521() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "pair", "ecdsa", "generate", "--curve=P521", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating key pair")

	s.Contains(output, "Algorithm: ECDSA Curve P-521\n\n")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tPublic Key: public.pem")
}

func (s *CLISuite) TestShouldGenerateEd25519KeyPair() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "pair", "ed25519", "generate", "--directory=/tmp/"})
	s.NoError(err)
	s.Contains(output, "Generating key pair")

	s.Contains(output, "Algorithm: Ed25519\n\n")

	s.Contains(output, "Output Paths:")
	s.Contains(output, "\tDirectory: /tmp")
	s.Contains(output, "\tPrivate Key: private.pem")
	s.Contains(output, "\tPublic Key: public.pem")
}

func (s *CLISuite) TestShouldNotGenerateECDSAKeyPairCurveInvalid() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "pair", "ecdsa", "generate", "--curve=invalid", "--directory=/tmp/"})
	s.NotNil(err)
	s.Contains(output, "Error: invalid curve 'invalid' was specified: curve must be P224, P256, P384, or P521")
}

func (s *CLISuite) TestShouldNotGenerateRSAWithBadCAPath() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--path.ca=/tmp/invalid", "--directory=/tmp/"})
	s.NotNil(err)
	s.Contains(output, "Error: could not read private key file '/tmp/invalid/ca.private.pem': open /tmp/invalid/ca.private.pem: no such file or directory\n")
}

func (s *CLISuite) TestShouldNotGenerateRSAWithBadCAFileNames() {
	var (
		err    error
		output string
	)

	_, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--common-name='Authelia Standalone Root Certificate Authority'", "--ca", "--directory=/tmp/"})
	s.NoError(err)

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--path.ca=/tmp/", "--file.ca-private-key=invalid.pem", "--directory=/tmp/"})
	s.NotNil(err)
	s.Contains(output, "Error: could not read private key file '/tmp/invalid.pem': open /tmp/invalid.pem: no such file or directory\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--path.ca=/tmp/", "--file.ca-certificate=invalid.crt", "--directory=/tmp/"})
	s.NotNil(err)
	s.Contains(output, "Error: could not read certificate file '/tmp/invalid.crt': open /tmp/invalid.crt: no such file or directory\n")
}

func (s *CLISuite) TestShouldNotGenerateRSAWithBadCAFileContent() {
	var (
		err    error
		output string
	)

	_, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--common-name='Authelia Standalone Root Certificate Authority'", "--ca", "--directory=/tmp/"})
	s.NoError(err)

	s.Require().NoError(os.WriteFile("/tmp/ca.private.bad.pem", []byte("INVALID"), 0600)) //nolint:gosec
	s.Require().NoError(os.WriteFile("/tmp/ca.public.bad.crt", []byte("INVALID"), 0600))  //nolint:gosec

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--path.ca=/tmp/", "--file.ca-private-key=ca.private.bad.pem", "--directory=/tmp/"})
	s.NotNil(err)
	s.Contains(output, "Error: could not parse private key from file '/tmp/ca.private.bad.pem': error occurred attempting to parse PEM block: either no PEM block was supplied or it was malformed\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", "crypto", "certificate", "rsa", "generate", "--path.ca=/tmp/", "--file.ca-certificate=ca.public.bad.crt", "--directory=/tmp/"})
	s.NotNil(err)
	s.Contains(output, "Error: could not parse certificate from file '/tmp/ca.public.bad.crt': error occurred attempting to parse PEM block: either no PEM block was supplied or it was malformed\n")
}

func (s *CLISuite) TestStorage00ShouldShowCorrectPreInitInformation() {
	_ = os.Remove("/tmp/db.sqlite3")

	output, err := s.Exec("authelia-backend", []string{"authelia", "storage", "schema-info", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	pattern := regexp.MustCompile(`^Schema Version: N/A\nSchema Upgrade Available: yes - version \d+\nSchema Tables: N/A\nSchema Encryption Key: unsupported \(schema version\)`)

	s.Regexp(pattern, output)

	patternOutdated := regexp.MustCompile(`Error: command requires the use of a up to date schema version: storage schema outdated: version \d+ is outdated please migrate to version \d+ in order to use this command or use an older binary`)
	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "totp", "export", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Regexp(patternOutdated, output)

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "encryption", "change-key", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Regexp(patternOutdated, output)

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "encryption", "check", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Regexp(regexp.MustCompile(`^Error: command requires the use of a up to date schema version: storage schema outdated: version 0 is outdated please migrate to version \d+ in order to use this command or use an older binary\n`), output)

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "migrate", "down", "--target=0", "--destroy-data", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: schema migration target version 0 is the same current version 0")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "migrate", "up", "--target=2147483640", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: schema up migration target version 2147483640 is greater then the latest version ")
	s.Contains(output, " which indicates it doesn't exist")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "migrate", "history", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Contains(output, "No migration history is available for schemas that are not version 1 or above.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "migrate", "list-up", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Contains(output, "Storage Schema Migration List (Up)\n\nVersion")
	s.Regexp(regexp.MustCompile(`Version\s+Description\n`), output)
	s.Regexp(regexp.MustCompile(`1\s+Initial Schema\n`), output)

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "migrate", "list-down", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Contains(output, "Storage Schema Migration List (Down)\n\nNo Migrations Available\n")
}

func (s *CLISuite) TestStorage01ShouldMigrateUp() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "storage", "migrate", "up", "--config=/config/configuration.storage.yml"})
	s.Require().NoError(err)

	pattern0 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is being attempted"`)
	pattern1 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is complete"`)

	s.Regexp(pattern0, output)
	s.Regexp(pattern1, output)

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "migrate", "up", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")

	s.Contains(output, "Error: schema already up to date\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "migrate", "history", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Contains(output, "Migration History:\n\nID")
	s.Regexp(regexp.MustCompile(`ID\s+Date\s+Before\s+After\s+Authelia Version`), output)

	s.Regexp(regexp.MustCompile(`0\s+1\s+\w+`), output)
	s.Regexp(regexp.MustCompile(`1\s+2\s+\w+`), output)
	s.Regexp(regexp.MustCompile(`3\s+4\s+\w+`), output)
	s.Regexp(regexp.MustCompile(`4\s+5\s+\w+`), output)
	s.Regexp(regexp.MustCompile(`5\s+6\s+\w+`), output)
	s.Regexp(regexp.MustCompile(`7\s+8\s+\w+`), output)
	s.Regexp(regexp.MustCompile(`8\s+9\s+\w+`), output)
	s.Regexp(regexp.MustCompile(`9\s+10\s+\w+`), output)
	s.Regexp(regexp.MustCompile(`10\s+11\s+\w+`), output)
	s.Regexp(regexp.MustCompile(`11\s+12\s+\w+`), output)
	s.Regexp(regexp.MustCompile(`12\s+13\s+\w+`), output)
	s.Regexp(regexp.MustCompile(`13\s+14\s+\w+`), output)

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "migrate", "list-up", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Contains(output, "Storage Schema Migration List (Up)\n\nNo Migrations Available")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "migrate", "list-down", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Contains(output, "Storage Schema Migration List (Down)\n\nVersion")
	s.Regexp(regexp.MustCompile(`Version\s+Description\n`), output)
	s.Regexp(regexp.MustCompile(`1\s+Initial Schema\n`), output)
}

func (s *CLISuite) TestStorage02ShouldShowSchemaInfo() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "storage", "schema-info", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Regexp(regexp.MustCompile(`Schema Version: \d+`), output)
	s.Contains(output, "authentication_logs")
	s.Contains(output, "identity_verification")
	s.Contains(output, "duo_devices")
	s.Contains(output, "user_preferences")
	s.Contains(output, "migrations")
	s.Contains(output, "encryption")
	s.Contains(output, "encryption")
	s.Contains(output, "webauthn_credentials")
	s.Contains(output, "totp_configurations")
	s.Contains(output, "one_time_code")
	s.Contains(output, "totp_history")
	s.Contains(output, "user_opaque_identifier")
	s.Contains(output, "webauthn_users")
	s.Contains(output, "oauth2_blacklisted_jti")
	s.Contains(output, "oauth2_consent_session")
	s.Contains(output, "oauth2_consent_preconfiguration")
	s.Contains(output, "oauth2_access_token_session")
	s.Contains(output, "oauth2_authorization_code_session")
	s.Contains(output, "oauth2_openid_connect_session")
	s.Contains(output, "oauth2_par_context")
	s.Contains(output, "oauth2_pkce_request_session")
	s.Contains(output, "oauth2_refresh_token_session")
	s.Contains(output, "Schema Encryption Key: valid")
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
				Algorithm: SHA1,
			},
		},
		{
			config: model.TOTPConfiguration{
				Username:  "mary",
				Period:    45,
				Digits:    6,
				Algorithm: SHA1,
			},
		},
		{
			config: model.TOTPConfiguration{
				Username:  "fred",
				Period:    30,
				Digits:    8,
				Algorithm: SHA1,
			},
		},
		{
			config: model.TOTPConfiguration{
				Username:  "jone",
				Period:    30,
				Digits:    6,
				Algorithm: SHA512,
			},
			png: true,
		},
	}

	var (
		config   *model.TOTPConfiguration
		fileInfo os.FileInfo
	)

	dir := s.T().TempDir()

	qr := filepath.Join(dir, "qr.png")

	for _, testCase := range testCases {
		if testCase.png {
			output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "totp", "generate", testCase.config.Username, "--period", strconv.FormatUint(uint64(testCase.config.Period), 10), "--algorithm", testCase.config.Algorithm, "--digits", strconv.Itoa(int(testCase.config.Digits)), "--path", qr, "--config=/config/configuration.storage.yml"})
			s.NoError(err)
			s.Contains(output, fmt.Sprintf(" and saved it as a PNG image at the path '%s'", qr))

			fileInfo, err = os.Stat(qr)
			s.NoError(err)
			s.Require().NotNil(fileInfo)
			s.False(fileInfo.IsDir())
			s.Greater(fileInfo.Size(), int64(1000))
		} else {
			output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "totp", "generate", testCase.config.Username, "--period", strconv.FormatUint(uint64(testCase.config.Period), 10), "--algorithm", testCase.config.Algorithm, "--digits", strconv.Itoa(int(testCase.config.Digits)), "--config=/config/configuration.storage.yml"})
			s.NoError(err)
		}

		config, err = storageProvider.LoadTOTPConfiguration(ctx, testCase.config.Username)
		s.NoError(err)

		s.Contains(output, config.URI())

		expectedLinesCSV = append(expectedLinesCSV, fmt.Sprintf("%s,%s,%s,%d,%d,%s", "Authelia", config.Username, config.Algorithm, config.Digits, config.Period, string(config.Secret)))
		expectedLines = append(expectedLines, config.URI())
	}

	yml := filepath.Join(dir, "authelia.export.totp.yml")
	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "totp", "export", "--file", yml, "--config=/config/configuration.storage.yml"})
	s.NoError(err)
	s.Contains(output, fmt.Sprintf("Successfully exported %d TOTP configurations as YAML to the '%s' file\n", len(expectedLines), yml))

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "totp", "export", "uri", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	for _, expectedLine := range expectedLines {
		s.Contains(output, expectedLine)
	}

	csv := filepath.Join(dir, "authelia.export.totp.csv")
	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "totp", "export", "csv", "--file", csv, "--config=/config/configuration.storage.yml"})
	s.NoError(err)
	s.Contains(output, fmt.Sprintf("Successfully exported %d TOTP configurations as CSV to the '%s' file\n", len(expectedLines), csv))

	var data []byte

	data, err = os.ReadFile(csv)
	s.NoError(err)

	content := string(data)
	for _, expectedLine := range expectedLinesCSV {
		s.Contains(content, expectedLine)
	}

	pngs := filepath.Join(dir, "png-qr-codes")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "totp", "export", "png", "--directory", pngs, "--config=/config/configuration.storage.yml"})
	s.NoError(err)
	s.Contains(output, fmt.Sprintf("Successfully exported %d TOTP configuration as QR codes in PNG format to the '%s' directory\n", len(expectedLines), pngs))

	for _, testCase := range testCases {
		fileInfo, err = os.Stat(filepath.Join(pngs, fmt.Sprintf("%s.png", testCase.config.Username)))

		s.NoError(err)
		s.Require().NotNil(fileInfo)

		s.False(fileInfo.IsDir())
		s.Greater(fileInfo.Size(), int64(1000))
	}

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "totp", "generate", "test", "--period=30", "--algorithm=SHA1", "--digits=6", "--path", qr, "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: image output filepath already exists")
}

func (s *CLISuite) TestStorage04ShouldManageUniqueID() {
	dir := s.T().TempDir()

	output, err := s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "export", "--file=out.yml", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: no data to export")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "add", "john", "--service=webauthn", "--sector=''", "--identifier=1097c8f8-83f2-4506-8138-5f40e83a1285", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: the service name 'webauthn' is invalid, the valid values are: 'openid'")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector=''", "--identifier=1097c8f8-83f2-4506-8138-5f40e83a1285", "--config=/config/configuration.storage.yml"})
	s.NoError(err)
	s.Contains(output, "Added User Opaque Identifier:\n\tService: openid\n\tSector: \n\tUsername: john\n\tIdentifier: 1097c8f8-83f2-4506-8138-5f40e83a1285\n\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "export", "--file=/a/no/path/fileout.yml", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: error occurred writing to file '/a/no/path/fileout.yml': open /a/no/path/fileout.yml: no such file or directory")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "export", "--file=out.yml", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: error occurred writing to file 'out.yml': open out.yml: permission denied")

	out1 := filepath.Join(dir, "1.yml")
	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "export", "--file", out1, "--config=/config/configuration.storage.yml"})
	s.NoError(err)
	s.Contains(output, fmt.Sprintf("Successfully exported %d User Opaque Identifiers as YAML to the '%s' file\n", 1, out1))

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "export", "--file", out1, "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, fmt.Sprintf("Error: must specify a file that doesn't exist but '%s' exists", out1))

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector=''", "--identifier=1097c8f8-83f2-4506-8138-5f40e83a1285", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: error inserting user opaque id for user 'john' with opaque id '1097c8f8-83f2-4506-8138-5f40e83a1285':")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector=''", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: error inserting user opaque id for user 'john' with opaque id")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector='openidconnect.com'", "--identifier=1097c8f8-83f2-4506-8138-5f40e83a1285", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: error inserting user opaque id for user 'john' with opaque id '1097c8f8-83f2-4506-8138-5f40e83a1285':")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector='openidconnect.net'", "--identifier=b0e17f48-933c-4cba-8509-ee9bfadf8ce5", "--config=/config/configuration.storage.yml"})
	s.NoError(err)
	s.Contains(output, "Added User Opaque Identifier:\n\tService: openid\n\tSector: openidconnect.net\n\tUsername: john\n\tIdentifier: b0e17f48-933c-4cba-8509-ee9bfadf8ce5\n\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector='bad-uuid.com'", "--identifier=d49564dc-b7a1-11ec-8429-fcaa147128ea", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: the identifier provided 'd49564dc-b7a1-11ec-8429-fcaa147128ea' is a version 1 UUID but only version 4 UUID's accepted as identifiers")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector='bad-uuid.com'", "--identifier=asdmklasdm", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: the identifier provided 'asdmklasdm' is invalid as it must be a version 4 UUID but parsing it had an error: invalid UUID length: 10")

	data, err := os.ReadFile(out1)
	s.NoError(err)

	var export model.UserOpaqueIdentifiersExport

	s.NoError(yaml.Unmarshal(data, &export))

	s.Require().Len(export.Identifiers, 1)

	s.Equal("1097c8f8-83f2-4506-8138-5f40e83a1285", export.Identifiers[0].Identifier.String())
	s.Equal("john", export.Identifiers[0].Username)
	s.Equal("", export.Identifiers[0].SectorID)
	s.Equal("openid", export.Identifiers[0].Service)

	out2 := filepath.Join(dir, "2.yml")
	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "user", "identifiers", "export", "--file", out2, "--config=/config/configuration.storage.yml"})
	s.NoError(err)
	s.Contains(output, fmt.Sprintf("Successfully exported %d User Opaque Identifiers as YAML to the '%s' file\n", 2, out2))

	export = model.UserOpaqueIdentifiersExport{}

	data, err = os.ReadFile(out2)
	s.NoError(err)

	s.NoError(yaml.Unmarshal(data, &export))

	s.Require().Len(export.Identifiers, 2)

	s.Equal("1097c8f8-83f2-4506-8138-5f40e83a1285", export.Identifiers[0].Identifier.String())
	s.Equal("john", export.Identifiers[0].Username)
	s.Equal("", export.Identifiers[0].SectorID)
	s.Equal("openid", export.Identifiers[0].Service)

	s.Equal("b0e17f48-933c-4cba-8509-ee9bfadf8ce5", export.Identifiers[1].Identifier.String())
	s.Equal("john", export.Identifiers[1].Username)
	s.Equal("openidconnect.net", export.Identifiers[1].SectorID)
	s.Equal("openid", export.Identifiers[1].Service)
}

func (s *CLISuite) TestStorage05ShouldChangeEncryptionKey() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "storage", "encryption", "change-key", "--new-encryption-key=apple-apple-apple-apple", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Contains(output, "Completed the encryption key change. Please adjust your configuration to use the new key.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "schema-info", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Contains(output, "Schema Version: ")
	s.Contains(output, "authentication_logs")
	s.Contains(output, "identity_verification")
	s.Contains(output, "duo_devices")
	s.Contains(output, "user_preferences")
	s.Contains(output, "migrations")
	s.Contains(output, "encryption")
	s.Contains(output, "encryption")
	s.Contains(output, "webauthn_credentials")
	s.Contains(output, "totp_configurations")
	s.Contains(output, "Schema Encryption Key: invalid")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "encryption", "check", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Contains(output, "Storage Encryption Key Validation: FAILURE\n\n\tCause: the configured encryption key does not appear to be valid for this database which may occur if the encryption key was changed in the configuration without using the cli to change it in the database.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "encryption", "check", "--verbose", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Contains(output, "Storage Encryption Key Validation: FAILURE\n\n\tCause: the configured encryption key does not appear to be valid for this database which may occur if the encryption key was changed in the configuration without using the cli to change it in the database.\n\nTables:\n\n")
	s.Contains(output, "\n\n\tTable (oauth2_access_token_session): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")
	s.Contains(output, "\n\n\tTable (oauth2_authorization_code_session): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")
	s.Contains(output, "\n\n\tTable (oauth2_openid_connect_session): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")
	s.Contains(output, "\n\n\tTable (oauth2_pkce_request_session): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")
	s.Contains(output, "\n\n\tTable (oauth2_refresh_token_session): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")
	s.Contains(output, "\n\n\tTable (oauth2_par_context): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")
	s.Contains(output, "\n\n\tTable (totp_configurations): FAILURE\n\t\tInvalid Rows: 4\n\t\tTotal Rows: 4\n")
	s.Contains(output, "\n\n\tTable (webauthn_credentials): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "encryption", "check", "--encryption-key=apple-apple-apple-apple", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Contains(output, "Storage Encryption Key Validation: SUCCESS\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "encryption", "check", "--verbose", "--encryption-key=apple-apple-apple-apple", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	s.Contains(output, "Storage Encryption Key Validation: SUCCESS\n\nTables:\n\n")
	s.Contains(output, "\n\n\tTable (oauth2_access_token_session): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")
	s.Contains(output, "\n\n\tTable (oauth2_authorization_code_session): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")
	s.Contains(output, "\n\n\tTable (oauth2_openid_connect_session): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")
	s.Contains(output, "\n\n\tTable (oauth2_pkce_request_session): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")
	s.Contains(output, "\n\n\tTable (oauth2_refresh_token_session): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")
	s.Contains(output, "\n\n\tTable (oauth2_par_context): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")
	s.Contains(output, "\n\n\tTable (totp_configurations): SUCCESS\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 4\n")
	s.Contains(output, "\n\n\tTable (webauthn_credentials): N/A\n\t\tInvalid Rows: 0\n\t\tTotal Rows: 0\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "encryption", "change-key", "--encryption-key=apple-apple-apple-apple", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")

	s.Contains(output, "Error: you must either use an interactive terminal or use the --new-encryption-key flag\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "encryption", "change-key", "--encryption-key=apple-apple-apple-apple", "--new-encryption-key=abc", "--config=/config/configuration.storage.yml"})
	s.EqualError(err, "exit status 1")

	s.Contains(output, "Error: the new encryption key must be at least 20 characters\n")
}

func (s *CLISuite) TestStorage06ShouldMigrateDown() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "storage", "migrate", "down", "--target=0", "--destroy-data", "--config=/config/configuration.storage.yml"})
	s.NoError(err)

	pattern0 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is being attempted"`)
	pattern1 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is complete"`)

	s.Regexp(pattern0, output)
	s.Regexp(pattern1, output)
}

func (s *CLISuite) TestStorage07CacheMDS3() {
	var (
		output string
		err    error
	)

	dir := s.T().TempDir()

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "--help"})
	s.NoError(err)
	s.Contains(output, "Manage WebAuthn MDS3 cache storage.")
	s.Contains(output, "  delete ")
	s.Contains(output, "  dump ")
	s.Contains(output, "  status ")
	s.Contains(output, "  update ")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "status"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: webauthn metadata is disabled")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "delete"})
	s.NoError(err)
	s.Contains(output, "Successfully deleted cached MDS3 data.")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "dump", "--path=" + filepath.Join(dir, "data.mds3")})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: webauthn metadata is disabled")

	output, err = s.Exec("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "update"})
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: webauthn metadata is disabled")

	env := map[string]string{"AUTHELIA_WEBAUTHN_METADATA_ENABLED": "true"}

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "status"}, env)
	s.NoError(err)
	s.Contains(output, "WebAuthn MDS3 Cache Status:\n\n\tValid: true\n\tInitialized: false\n\tOutdated: false\n")

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "delete"}, env)
	s.NoError(err)
	s.Contains(output, "Successfully deleted cached MDS3 data.")

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "dump", "--path=" + filepath.Join(dir, "data.mds3")}, env)
	s.EqualError(err, "exit status 1")
	s.Contains(output, "Error: error dumping metadata: no metadata is in the cache")

	reUpdated := regexp.MustCompile(`^WebAuthn MDS3 cache data updated to version (\d+) and is due for update on ([A-Za-z]+ \d{1,2}, \d{4}).`)
	reAlreadyUpToDate := regexp.MustCompile(`^WebAuthn MDS3 cache data with version (\d+) due for update on ([A-Za-z]+ \d{1,2}, \d{4}) does not require an update.`)

	var updateArgs []string

	mds3 := filepath.Join("/buildkite/.cache/fido/", "mds.jwt")

	_, err = os.Stat(mds3)
	if err == nil {
		updateArgs = append(updateArgs, "--path="+strings.Replace(mds3, "/buildkite/.cache/fido/", "/tmp/", 1))
	}

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "update"}, env)
	s.NoError(err)
	s.Regexp(reUpdated, output)

	matches := reUpdated.FindStringSubmatch(output)

	version := matches[1]
	date := matches[2]

	output, err = s.ExecWithEnv("authelia-backend", append([]string{"authelia", "storage", "cache", "mds3", "update"}, updateArgs...), env)
	s.NoError(err)
	s.Regexp(reAlreadyUpToDate, output)
	s.Contains(output, version)
	s.Contains(output, date)

	output, err = s.ExecWithEnv("authelia-backend", append([]string{"authelia", "storage", "cache", "mds3", "update", "-f"}, updateArgs...), env)
	s.NoError(err)
	s.Regexp(reUpdated, output)

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "status"}, env)
	s.NoError(err)
	s.Contains(output, fmt.Sprintf("WebAuthn MDS3 Cache Status:\n\n\tValid: true\n\tInitialized: true\n\tOutdated: false\n\tVersion: %s\n\tNext Update: %s\n", version, date))

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "dump", "--path=" + filepath.Join(dir, "data.mds3")}, env)
	s.NoError(err)
	s.Contains(output, fmt.Sprintf("Successfully dumped WebAuthn MDS3 data with version %s from cache to file '%s'.", version, filepath.Join(dir, "data.mds3")))

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "update", "--path=" + filepath.Join(dir, "data.mds3")}, env)
	s.NoError(err)
	s.Regexp(reAlreadyUpToDate, output)

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "update", "--path=" + filepath.Join(dir, "data.mds3"), "-f"}, env)
	s.NoError(err)
	s.Regexp(reUpdated, output)

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "delete"}, env)
	s.NoError(err)
	s.Contains(output, "Successfully deleted cached MDS3 data.")

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "status"}, env)
	s.NoError(err)
	s.Contains(output, "WebAuthn MDS3 Cache Status:\n\n\tValid: true\n\tInitialized: false\n\tOutdated: false\n")

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "update", "--path=" + filepath.Join(dir, "data.mds3")}, env)
	s.NoError(err)
	s.Regexp(reUpdated, output)

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "status"}, env)
	s.NoError(err)
	s.Contains(output, fmt.Sprintf("WebAuthn MDS3 Cache Status:\n\n\tValid: true\n\tInitialized: true\n\tOutdated: false\n\tVersion: %s\n\tNext Update: %s\n", version, date))

	output, err = s.ExecWithEnv("authelia-backend", []string{"authelia", "storage", "cache", "mds3", "update"}, env)
	s.NoError(err)
	s.Regexp(reAlreadyUpToDate, output)
	s.Contains(output, version)
	s.Contains(output, date)
}

func (s *CLISuite) TestACLPolicyCheckVerbose() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "access-control", "check-policy", "--url=https://public.example.com", "--verbose", "--config=/config/configuration.yml"})
	s.NoError(err)

	// This is an example of `authelia access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://public.example.com --verbose`.
	s.Contains(output, "Performing policy check for request to 'https://public.example.com' method 'GET'.\n\n")
	s.Regexp(regexp.MustCompile(`#\s+Domain\s+Resource\s+Query\s+Method\s+Network\s+Subject\n`), output)
	s.Regexp(regexp.MustCompile(`\* 1\s+hit\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  2\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  3\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  4\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  5\s+miss\s+miss\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  6\s+miss\s+hit\s+hit\s+miss\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  7\s+miss\s+hit\s+hit\s+hit\s+miss\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  8\s+miss\s+hit\s+hit\s+hit\s+hit\s+may\n`), output)
	s.Regexp(regexp.MustCompile(`  9\s+miss\s+hit\s+hit\s+hit\s+hit\s+may\n`), output)
	s.Contains(output, "The policy 'bypass' from rule #1 will be applied to this request.")

	output, err = s.Exec("authelia-backend", []string{"authelia", "access-control", "check-policy", "--url=https://admin.example.com", "--method=HEAD", "--username=tom", "--groups=basic,test", "--ip=192.168.2.3", "--verbose", "--config=/config/configuration.yml"})
	s.NoError(err)

	// This is an example of `authelia access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://admin.example.com --method=HEAD --username=tom --groups=basic,test --ip=192.168.2.3 --verbose`.
	s.Contains(output, "Performing policy check for request to 'https://admin.example.com' method 'HEAD' username 'tom' groups 'basic,test' from IP '192.168.2.3'.\n\n")
	s.Regexp(regexp.MustCompile(`#\s+Domain\s+Resource\s+Query\s+Method\s+Network\s+Subject\n`), output)
	s.Regexp(regexp.MustCompile(`  1\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`\* 2\s+hit\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  3\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  4\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  5\s+miss\s+miss\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  6\s+miss\s+hit\s+hit\s+miss\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  7\s+miss\s+hit\s+hit\s+hit\s+miss\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  8\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  9\s+miss\s+hit\s+hit\s+hit\s+hit\s+miss\n`), output)
	s.Contains(output, "The policy 'two_factor' from rule #2 will be applied to this request.")

	output, err = s.Exec("authelia-backend", []string{"authelia", "access-control", "check-policy", "--url=https://resources.example.com/resources/test", "--method=POST", "--username=john", "--groups=admin,test", "--ip=192.168.1.3", "--verbose", "--config=/config/configuration.yml"})
	s.NoError(err)

	// This is an example of `authelia access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://resources.example.com/resources/test --method=POST --username=john --groups=admin,test --ip=192.168.1.3 --verbose`.
	s.Contains(output, "Performing policy check for request to 'https://resources.example.com/resources/test' method 'POST' username 'john' groups 'admin,test' from IP '192.168.1.3'.\n\n")
	s.Regexp(regexp.MustCompile(`#\s+Domain\s+Resource\s+Query\s+Method\s+Network\s+Subject\n`), output)
	s.Regexp(regexp.MustCompile(`  1\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  2\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  3\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  4\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`\* 5\s+hit\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  6\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  7\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  8\s+miss\s+hit\s+hit\s+hit\s+hit\s+miss\n`), output)
	s.Regexp(regexp.MustCompile(`  9\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Contains(output, "The policy 'one_factor' from rule #5 will be applied to this request.")

	output, err = s.Exec("authelia-backend", []string{"authelia", "access-control", "check-policy", "--url=https://user.example.com/resources/test", "--method=HEAD", "--username=john", "--groups=admin,test", "--ip=192.168.1.3", "--verbose", "--config=/config/configuration.yml"})
	s.NoError(err)

	// This is an example of `access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://user.example.com --method=HEAD --username=john --groups=admin,test --ip=192.168.1.3 --verbose`.
	s.Contains(output, "Performing policy check for request to 'https://user.example.com/resources/test' method 'HEAD' username 'john' groups 'admin,test' from IP '192.168.1.3'.\n\n")
	s.Regexp(regexp.MustCompile(`#\s+Domain\s+Resource\s+Query\s+Method\s+Network\s+Subject\n`), output)
	s.Regexp(regexp.MustCompile(`  1\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  2\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  3\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  4\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  5\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  6\s+miss\s+hit\s+hit\s+miss\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  7\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  8\s+miss\s+hit\s+hit\s+hit\s+hit\s+miss\n`), output)
	s.Regexp(regexp.MustCompile(`\* 9\s+hit\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Contains(output, "The policy 'one_factor' from rule #9 will be applied to this request.")

	output, err = s.Exec("authelia-backend", []string{"authelia", "access-control", "check-policy", "--url=https://user.example.com", "--method=HEAD", "--ip=192.168.1.3", "--verbose", "--config=/config/configuration.yml"})
	s.NoError(err)

	// This is an example of `authelia access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://user.example.com --method=HEAD --ip=192.168.1.3 --verbose`.
	s.Contains(output, "Performing policy check for request to 'https://user.example.com' method 'HEAD' from IP '192.168.1.3'.\n\n")
	s.Regexp(regexp.MustCompile(`#\s+Domain\s+Resource\s+Query\s+Method\s+Network\s+Subject\n`), output)
	s.Regexp(regexp.MustCompile(`  1\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  2\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  3\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  4\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  5\s+miss\s+miss\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  6\s+miss\s+hit\s+hit\s+miss\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  7\s+miss\s+hit\s+hit\s+hit\s+hit\s+hit\n`), output)
	s.Regexp(regexp.MustCompile(`  8\s+miss\s+hit\s+hit\s+hit\s+hit\s+may\n`), output)
	s.Regexp(regexp.MustCompile(`~ 9\s+hit\s+hit\s+hit\s+hit\s+hit\s+may\n`), output)
	s.Contains(output, "The policy 'one_factor' from rule #9 will potentially be applied to this request. Otherwise the policy 'bypass' from the default policy will be.")
}

func (s *CLISuite) TestDebugTLS() {
	output, err := s.Exec("authelia-backend", []string{"authelia", "debug", "tls", "tcp://secure.example.com:8080"})
	s.NoError(err)

	s.Contains(output, "General Information:\n\tServer Name: secure.example.com\n\tRemote Address: 192.168.240.100:8080\n\tNegotiated Protocol: \n\tTLS Version: TLS 1.3\n\tCipher Suite: TLS_AES_128_GCM_SHA256 (Supported)")
	s.Contains(output, "\t\tSerial Number: 280859886455442560996590795939870170263\n\t\tValid: true\n\t\tValid (System): false\n\t\tHostname Verification: pass")
	s.Contains(output, "\t\tSerial Number: 331626108752148202137556363956074982580\n\t\tValid: true\n\t\tValid (System): false")
	s.Contains(output, "\tCertificate Trusted: true\n\tCertificate Matches Hostname: true")

	output, err = s.Exec("authelia-backend", []string{"authelia", "debug", "tls", "tcp://secure.example.com:8080", "--hostname", "notsecure.notexample.com"})
	s.NoError(err)

	s.Contains(output, "General Information:\n\tServer Name: notsecure.notexample.com\n\tRemote Address: 192.168.240.100:8080\n\tNegotiated Protocol: \n\tTLS Version: TLS 1.3\n\tCipher Suite: TLS_AES_128_GCM_SHA256 (Supported)")
	s.Contains(output, "\t\tSerial Number: 280859886455442560996590795939870170263\n\t\tValid: true\n\t\tValid (System): false\n\t\tHostname Verification: fail\n\t\tHostname Verification Error: x509: certificate is valid for *.example.com, example.com, *.example1.com, example1.com, *.example2.com, example2.com, *.example3.com, example3.com, not notsecure.notexample.com")
	s.Contains(output, "\t\tSerial Number: 331626108752148202137556363956074982580\n\t\tValid: true\n\t\tValid (System): false")
	s.Contains(output, "\tCertificate Trusted: true\n\tCertificate Matches Hostname: false")

	output, err = s.Exec("authelia-backend", []string{"authelia", "--config", "/config/configuration.nocerts.yml", "debug", "tls", "tcp://secure.example.com:8080"})
	s.NoError(err)

	s.Contains(output, "General Information:\n\tServer Name: secure.example.com\n\tRemote Address: 192.168.240.100:8080\n\tNegotiated Protocol: \n\tTLS Version: TLS 1.3\n\tCipher Suite: TLS_AES_128_GCM_SHA256 (Supported)")
	s.Contains(output, "\t\tSerial Number: 280859886455442560996590795939870170263\n\t\tValid: false\n\t\tValid (System): false\n\t\tValidation Hint: Certificate signed by unknown authority\n\t\tValidation Error: x509: certificate signed by unknown authority\n\t\tHostname Verification: pass")
	s.Contains(output, "\t\tSerial Number: 331626108752148202137556363956074982580\n\t\tValid: false\n\t\tValid (System): false\n\t\tValidation Hint: Certificate signed by unknown authority\n\t\tValidation Error: x509: certificate signed by unknown authority")
	s.Contains(output, "\tCertificate Trusted: false\n\tCertificate Matches Hostname: true")
	s.Contains(output, "\nWARNING: The certificate is not valid for one reason or another. You may need to configure Authelia to trust certificate below.\n\n")
	s.Contains(output, "-----BEGIN CERTIFICATE-----")
	s.Contains(output, "-----END CERTIFICATE-----")

	output, err = s.Exec("authelia-backend", []string{"authelia", "debug", "tls", "tcp://secure.example.com:8081"})
	s.NoError(err)

	s.Contains(output, "General Information:\n\tFailure: Did not receive a TLS handshake from secure.example.com:8081")
}

func TestCLISuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewCLISuite())
}
