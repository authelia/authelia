package suites

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/models"
	"github.com/authelia/authelia/v4/internal/storage"
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
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "validate-config", "/config/configuration.yml"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Configuration parsed successfully without errors")
}

func (s *CLISuite) TestShouldFailValidateConfig() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "validate-config", "/config/invalid.yml"})
	s.Assert().NotNil(err)
	s.Assert().Contains(output, "Error Loading Configuration: stat /config/invalid.yml: no such file or directory")
}

func (s *CLISuite) TestShouldHashPasswordArgon2id() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "hash-password", "test", "-m", "32", "-s", "test1234"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Password hash: $argon2id$v=19$m=32768,t=1,p=8")
}

func (s *CLISuite) TestShouldHashPasswordSHA512() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "hash-password", "test", "-z"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Password hash: $6$rounds=50000")
}

func (s *CLISuite) TestShouldGenerateCertificateRSA() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldGenerateCertificateRSAWithIPAddress() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "127.0.0.1", "--dir", "/tmp/"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldGenerateCertificateRSAWithStartDate() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--start-date", "'Jan 1 15:04:05 2011'"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldFailGenerateCertificateRSAWithStartDate() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--start-date", "Jan"})
	s.Assert().NotNil(err)
	s.Assert().Contains(output, "Failed to parse start date: parsing time \"Jan\" as \"Jan 2 15:04:05 2006\": cannot parse \"\" as \"2\"")
}

func (s *CLISuite) TestShouldGenerateCertificateCA() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ca"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldGenerateCertificateEd25519() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ed25519"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldFailGenerateCertificateECDSA() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ecdsa-curve", "invalid"})
	s.Assert().NotNil(err)
	s.Assert().Contains(output, "Failed to generate private key: unrecognized elliptic curve: \"invalid\"")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSAP224() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ecdsa-curve", "P224"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSAP256() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ecdsa-curve", "P256"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSAP384() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ecdsa-curve", "P384"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSAP521() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ecdsa-curve", "P521"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestStorageShouldShowErrWithoutConfig() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "schema-info"})
	s.Assert().EqualError(err, "exit status 1")

	s.Assert().Contains(output, "Error: storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided, storage: 'encryption_key' configuration option must be provided\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "history"})
	s.Assert().EqualError(err, "exit status 1")

	s.Assert().Contains(output, "Error: storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided, storage: 'encryption_key' configuration option must be provided\n")
}

func (s *CLISuite) TestStorage00ShouldShowCorrectPreInitInformation() {
	_ = os.Remove("/tmp/db.sqlite3")

	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "schema-info", "--config", "/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	pattern := regexp.MustCompile(`^Schema Version: N/A\nSchema Upgrade Available: yes - version \d+\nSchema Tables: N/A\nSchema Encryption Key: unsupported \(schema version\)`)

	s.Assert().Regexp(pattern, output)

	patternOutdated := regexp.MustCompile(`Error: schema is version \d+ which is outdated please migrate to version \d+ in order to use this command or use an older binary`)

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "totp", "export", "--config", "/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Regexp(patternOutdated, output)

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "change-key", "--config", "/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Regexp(patternOutdated, output)

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--config", "/config/configuration.storage.yml"})
	s.Assert().NoError(err)
	s.Assert().Contains(output, "Could not check encryption key for validity. The schema version doesn't support encryption.")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "down", "--target", "0", "--destroy-data", "--config", "/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: schema migration target version 0 is the same current version 0")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "up", "--target", "2147483640", "--config", "/config/configuration.storage.yml"})
	s.Assert().EqualError(err, "exit status 1")
	s.Assert().Contains(output, "Error: schema up migration target version 2147483640 is greater then the latest version ")
	s.Assert().Contains(output, " which indicates it doesn't exist")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "--config", "/config/configuration.storage.yml", "migrate", "history"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "No migration history is available for schemas that not version 1 or above.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "--config", "/config/configuration.storage.yml", "migrate", "list-up"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Storage Schema Migration List (Up)\n\nVersion\t\tDescription\n1\t\tInitial Schema\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "--config", "/config/configuration.storage.yml", "migrate", "list-down"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Storage Schema Migration List (Down)\n\nNo Migrations Available\n")
}

func (s *CLISuite) TestStorage01ShouldMigrateUp() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "--config", "/config/configuration.storage.yml", "migrate", "up"})
	s.Require().NoError(err)

	pattern0 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is being attempted"`)
	pattern1 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is complete"`)

	s.Regexp(pattern0, output)
	s.Regexp(pattern1, output)

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "--config", "/config/configuration.storage.yml", "migrate", "up"})
	s.Assert().EqualError(err, "exit status 1")

	s.Assert().Contains(output, "Error: schema already up to date\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "--config", "/config/configuration.storage.yml", "migrate", "history"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Migration History:\n\nID\tDate\t\t\t\tBefore\tAfter\tAuthelia Version\n")
	s.Assert().Contains(output, "0\t1")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "--config", "/config/configuration.storage.yml", "migrate", "list-up"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Storage Schema Migration List (Up)\n\nNo Migrations Available")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "--config", "/config/configuration.storage.yml", "migrate", "list-down"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Storage Schema Migration List (Down)\n\nVersion\t\tDescription\n")
	s.Assert().Contains(output, "1\t\tInitial Schema")
}

func (s *CLISuite) TestStorage02ShouldShowSchemaInfo() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "schema-info", "--config", "/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	pattern := regexp.MustCompile(`^Schema Version: \d+\nSchema Upgrade Available: no\nSchema Tables: authentication_logs, identity_verification, totp_configurations, u2f_devices, duo_devices, user_preferences, migrations, encryption\nSchema Encryption Key: valid`)

	s.Assert().Regexp(pattern, output)
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

	configs := []*models.TOTPConfiguration{
		{
			Username:  "john",
			Period:    30,
			Digits:    6,
			Algorithm: "SHA1",
		},
		{
			Username:  "mary",
			Period:    45,
			Digits:    6,
			Algorithm: "SHA1",
		},
		{
			Username:  "fred",
			Period:    30,
			Digits:    8,
			Algorithm: "SHA1",
		},
		{
			Username:  "jone",
			Period:    30,
			Digits:    6,
			Algorithm: "SHA512",
		},
	}

	for _, config := range configs {
		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "totp", "generate", config.Username, "--period", strconv.Itoa(int(config.Period)), "--algorithm", config.Algorithm, "--digits", strconv.Itoa(int(config.Digits)), "--config", "/config/configuration.storage.yml"})
		s.Assert().NoError(err)

		config, err = storageProvider.LoadTOTPConfiguration(ctx, config.Username)
		s.Assert().NoError(err)
		s.Assert().Contains(output, config.URI())

		expectedLinesCSV = append(expectedLinesCSV, fmt.Sprintf("%s,%s,%s,%d,%d,%s", "Authelia", config.Username, config.Algorithm, config.Digits, config.Period, string(config.Secret)))
		expectedLines = append(expectedLines, config.URI())
	}

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "totp", "export", "--format", "uri", "--config", "/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	for _, expectedLine := range expectedLines {
		s.Assert().Contains(output, expectedLine)
	}

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "totp", "export", "--format", "csv", "--config", "/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	for _, expectedLine := range expectedLinesCSV {
		s.Assert().Contains(output, expectedLine)
	}
}

func (s *CLISuite) TestStorage04ShouldChangeEncryptionKey() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "change-key", "--new-encryption-key", "apple-apple-apple-apple", "--config", "/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Completed the encryption key change. Please adjust your configuration to use the new key.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "schema-info", "--config", "/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	pattern := regexp.MustCompile(`Schema Version: \d+\nSchema Upgrade Available: no\nSchema Tables: authentication_logs, identity_verification, totp_configurations, u2f_devices, duo_devices, user_preferences, migrations, encryption\nSchema Encryption Key: invalid`)
	s.Assert().Regexp(pattern, output)

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--config", "/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Encryption key validation: failed.\n\nError: the encryption key is not valid against the schema check value.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--verbose", "--config", "/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Encryption key validation: failed.\n\nError: the encryption key is not valid against the schema check value, 4 of 4 total TOTP secrets were invalid.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--encryption-key", "apple-apple-apple-apple", "--config", "/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Encryption key validation: success.\n")

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--verbose", "--encryption-key", "apple-apple-apple-apple", "--config", "/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	s.Assert().Contains(output, "Encryption key validation: success.\n")
}

func (s *CLISuite) TestStorage05ShouldMigrateDown() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "down", "--target", "0", "--destroy-data", "--config", "/config/configuration.storage.yml"})
	s.Assert().NoError(err)

	pattern0 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is being attempted"`)
	pattern1 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is complete"`)

	s.Regexp(pattern0, output)
	s.Regexp(pattern1, output)
}

func TestCLISuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewCLISuite())
}
