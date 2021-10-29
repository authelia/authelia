package suites

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/poy/onpar"
	"gopkg.in/yaml.v3"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestCLISuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) (*testing.T, *CommandSuite) {
		s := setupCLITest()
		return t, s
	})

	o.Spec("TestShouldPrintBuildInformation", func(t *testing.T, s *CommandSuite) {
		if os.Getenv("CI") != "true" {
			t.Skip("Skipping testing in dev environment")
		}

		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "build-info"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Last Tag: "))
		is.True(strings.Contains(output, "State: "))
		is.True(strings.Contains(output, "Branch: "))
		is.True(strings.Contains(output, "Build Number: "))
		is.True(strings.Contains(output, "Build OS: "))
		is.True(strings.Contains(output, "Build Arch: "))
		is.True(strings.Contains(output, "Build Date: "))

		r := regexp.MustCompile(`^Last Tag: (v\d+\.\d+\.\d+|unknown)\nState: (tagged|untagged) (clean|dirty)\nBranch: [^\s\n]+\nCommit: ([0-9a-f]{40}|unknown)\nBuild Number: \d+\nBuild OS: (linux|darwin|windows|freebsd)\nBuild Arch: (amd64|arm|arm64)\nBuild Date: ((Sun|Mon|Tue|Wed|Thu|Fri|Sat), \d{2} (Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d{4} \d{2}:\d{2}:\d{2} [+-]\d{4})?\nExtra: \n`)
		is.True(r.MatchString(output))
	})

	o.Spec("TestShouldPrintVersion", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "--version"})
		is.NoErr(err)
		is.True(strings.Contains(output, "authelia version"))
	})

	o.Spec("TestShouldValidateConfig", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "validate-config", "--config=/config/configuration.yml"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Configuration parsed and loaded successfully without errors."))
	})

	o.Spec("TestShouldFailValidateConfig", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "validate-config", "--config=/config/invalid.yml"})
		is.NoErr(err)
		is.True(strings.Contains(output, "failed to load configuration from yaml file(/config/invalid.yml) source: open /config/invalid.yml: no such file or directory"))
	})

	o.Spec("TestShouldHashPasswordArgon2id", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "hash-password", "test", "-m", "32", "-s", "test1234"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Password hash: $argon2id$v=19$m=32768,t=3,p=4"))
	})

	o.Spec("TestShouldHashPasswordSHA512", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "hash-password", "test", "-z"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Password hash: $6$rounds=50000"))
	})

	o.Spec("TestShouldGenerateCertificateRSA", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host=*.example.com", "--dir=/tmp/"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Certificate written to /tmp/cert.pem"))
		is.True(strings.Contains(output, "Private Key written to /tmp/key.pem"))
	})

	o.Spec("TestShouldGenerateCertificateRSAWithIPAddress", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host=127.0.0.1", "--dir=/tmp/"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Certificate written to /tmp/cert.pem"))
		is.True(strings.Contains(output, "Private Key written to /tmp/key.pem"))
	})

	o.Spec("TestShouldGenerateCertificateRSAWithStartDate", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host=*.example.com", "--dir=/tmp/", "--start-date='Jan 1 15:04:05 2011'"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Certificate written to /tmp/cert.pem"))
		is.True(strings.Contains(output, "Private Key written to /tmp/key.pem"))
	})

	o.Spec("TestShouldFailGenerateCertificateRSAWithStartDate", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host=*.example.com", "--dir=/tmp/", "--start-date=Jan"})
		is.True(err != nil)
		is.True(strings.Contains(output, "Failed to parse start date: parsing time \"Jan\" as \"Jan 2 15:04:05 2006\": cannot parse \"\" as \"2\""))
	})

	o.Spec("TestShouldGenerateCertificateCA", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host=*.example.com", "--dir=/tmp/", "--ca"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Certificate written to /tmp/cert.pem"))
		is.True(strings.Contains(output, "Private Key written to /tmp/key.pem"))
	})

	o.Spec("TestShouldGenerateCertificateEd25519", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host=*.example.com", "--dir=/tmp/", "--ed25519"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Certificate written to /tmp/cert.pem"))
		is.True(strings.Contains(output, "Private Key written to /tmp/key.pem"))
	})

	o.Spec("TestShouldFailGenerateCertificateECDSA", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host=*.example.com", "--dir=/tmp/", "--ecdsa-curve=invalid"})
		is.True(err != nil)
		is.True(strings.Contains(output, "Failed to generate private key: unrecognized elliptic curve: \"invalid\""))
	})

	o.Spec("TestShouldGenerateCertificateECDSAP224", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host=*.example.com", "--dir=/tmp/", "--ecdsa-curve=P224"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Certificate written to /tmp/cert.pem"))
		is.True(strings.Contains(output, "Private Key written to /tmp/key.pem"))
	})

	o.Spec("TestShouldGenerateCertificateECDSAP256", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host=*.example.com", "--dir=/tmp/", "--ecdsa-curve=P256"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Certificate written to /tmp/cert.pem"))
		is.True(strings.Contains(output, "Private Key written to /tmp/key.pem"))
	})

	o.Spec("TestShouldGenerateCertificateECDSAP384", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host=*.example.com", "--dir=/tmp/", "--ecdsa-curve=P384"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Certificate written to /tmp/cert.pem"))
		is.True(strings.Contains(output, "Private Key written to /tmp/key.pem"))
	})

	o.Spec("TestShouldGenerateCertificateECDSAP521", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host=*.example.com", "--dir=/tmp/", "--ecdsa-curve=P521"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Certificate written to /tmp/cert.pem"))
		is.True(strings.Contains(output, "Private Key written to /tmp/key.pem"))
	})

	o.Spec("TestStorageShouldShowErrWithoutConfig", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "schema-info"})
		is.Equal(err.Error(), "exit status 1")
		is.True(strings.Contains(output, "Error: storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided, storage: option 'encryption_key' must is required\n"))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "history"})
		is.Equal(err.Error(), "exit status 1")
		is.True(strings.Contains(output, "Error: storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided, storage: option 'encryption_key' must is required\n"))
	})

	o.Spec("TestStorage00ShouldShowCorrectPreInitInformation", func(t *testing.T, s *CommandSuite) {
		_ = os.Remove("/tmp/db.sqlite3")
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "schema-info", "--config=/config/configuration.storage.yml"})
		is.NoErr(err)
		pattern := regexp.MustCompile(`^Schema Version: N/A\nSchema Upgrade Available: yes - version \d+\nSchema Tables: N/A\nSchema Encryption Key: unsupported \(schema version\)`)
		is.True(pattern.MatchString(output))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "export", "--config=/config/configuration.storage.yml"})
		is.Equal(err.Error(), "exit status 1")
		patternOutdated := regexp.MustCompile(`Error: schema is version \d+ which is outdated please migrate to version \d+ in order to use this command or use an older binary`)
		is.True(patternOutdated.MatchString(output))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "change-key", "--config=/config/configuration.storage.yml"})
		is.Equal(err.Error(), "exit status 1")
		is.True(patternOutdated.MatchString(output))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--config=/config/configuration.storage.yml"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Could not check encryption key for validity. The schema version doesn't support encryption."))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "down", "--target=0", "--destroy-data", "--config=/config/configuration.storage.yml"})
		is.Equal(err.Error(), "exit status 1")
		is.True(strings.Contains(output, "Error: schema migration target version 0 is the same current version 0"))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "up", "--target=2147483640", "--config=/config/configuration.storage.yml"})
		is.Equal(err.Error(), "exit status 1")
		is.True(strings.Contains(output, "Error: schema up migration target version 2147483640 is greater then the latest version "))
		is.True(strings.Contains(output, " which indicates it doesn't exist"))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "history", "--config=/config/configuration.storage.yml"})
		is.NoErr(err)
		is.True(strings.Contains(output, "No migration history is available for schemas that not version 1 or above.\n"))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "list-up", "--config=/config/configuration.storage.yml"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Storage Schema Migration List (Up)\n\nVersion\t\tDescription\n1\t\tInitial Schema\n"))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "list-down", "--config=/config/configuration.storage.yml"})
		is.NoErr(err)
		is.True(strings.Contains(output, "Storage Schema Migration List (Down)\n\nNo Migrations Available\n"))
	})

	o.Spec("TestACLPolicyCheckVerbose", func(t *testing.T, s *CommandSuite) {
		is := is.New(t)
		output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "access-control", "check-policy", "--url=https://public.example.com", "--verbose", "--config=/config/configuration.yml"})
		is.NoErr(err)

		// This is an example of `authelia access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://public.example.com --verbose`.
		is.True(strings.Contains(output, "Performing policy check for request to 'https://public.example.com' method 'GET'.\n\n"))
		is.True(strings.Contains(output, "  #\tDomain\tResource\tMethod\tNetwork\tSubject\n"))
		is.True(strings.Contains(output, "* 1\thit\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  2\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  4\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  5\tmiss\tmiss\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  6\tmiss\thit\t\tmiss\thit\thit\n"))
		is.True(strings.Contains(output, "  7\tmiss\thit\t\thit\tmiss\thit\n"))
		is.True(strings.Contains(output, "  8\tmiss\thit\t\thit\thit\tmay\n"))
		is.True(strings.Contains(output, "  9\tmiss\thit\t\thit\thit\tmay\n"))
		is.True(strings.Contains(output, "The policy 'bypass' from rule #1 will be applied to this request."))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "access-control", "check-policy", "--url=https://admin.example.com", "--method=HEAD", "--username=tom", "--groups=basic,test", "--ip=192.168.2.3", "--verbose", "--config=/config/configuration.yml"})
		is.NoErr(err)

		// This is an example of `authelia access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://admin.example.com --method=HEAD --username=tom --groups=basic,test --ip=192.168.2.3 --verbose`.
		is.True(strings.Contains(output, "Performing policy check for request to 'https://admin.example.com' method 'HEAD' username 'tom' groups 'basic,test' from IP '192.168.2.3'.\n\n"))
		is.True(strings.Contains(output, "  #\tDomain\tResource\tMethod\tNetwork\tSubject\n"))
		is.True(strings.Contains(output, "  #\tDomain\tResource\tMethod\tNetwork\tSubject\n"))
		is.True(strings.Contains(output, "  1\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "* 2\thit\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  3\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  4\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  5\tmiss\tmiss\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  6\tmiss\thit\t\tmiss\thit\thit\n"))
		is.True(strings.Contains(output, "  7\tmiss\thit\t\thit\tmiss\thit\n"))
		is.True(strings.Contains(output, "  8\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  9\tmiss\thit\t\thit\thit\tmiss\n"))
		is.True(strings.Contains(output, "The policy 'two_factor' from rule #2 will be applied to this request."))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "access-control", "check-policy", "--url=https://resources.example.com/resources/test", "--method=POST", "--username=john", "--groups=admin,test", "--ip=192.168.1.3", "--verbose", "--config=/config/configuration.yml"})
		is.NoErr(err)

		// This is an example of `authelia access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://resources.example.com/resources/test --method=POST --username=john --groups=admin,test --ip=192.168.1.3 --verbose`.
		is.True(strings.Contains(output, "Performing policy check for request to 'https://resources.example.com/resources/test' method 'POST' username 'john' groups 'admin,test' from IP '192.168.1.3'.\n\n"))
		is.True(strings.Contains(output, "  #\tDomain\tResource\tMethod\tNetwork\tSubject\n"))
		is.True(strings.Contains(output, "  1\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  2\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  3\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  4\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "* 5\thit\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  6\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  7\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  8\tmiss\thit\t\thit\thit\tmiss\n"))
		is.True(strings.Contains(output, "  9\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "The policy 'one_factor' from rule #5 will be applied to this request."))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "access-control", "check-policy", "--url=https://user.example.com/resources/test", "--method=HEAD", "--username=john", "--groups=admin,test", "--ip=192.168.1.3", "--verbose", "--config=/config/configuration.yml"})
		is.NoErr(err)

		// This is an example of `access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://user.example.com --method=HEAD --username=john --groups=admin,test --ip=192.168.1.3 --verbose`.
		is.True(strings.Contains(output, "Performing policy check for request to 'https://user.example.com/resources/test' method 'HEAD' username 'john' groups 'admin,test' from IP '192.168.1.3'.\n\n"))
		is.True(strings.Contains(output, "  #\tDomain\tResource\tMethod\tNetwork\tSubject\n"))
		is.True(strings.Contains(output, "  1\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  2\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  3\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  4\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  5\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  6\tmiss\thit\t\tmiss\thit\thit\n"))
		is.True(strings.Contains(output, "  7\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  8\tmiss\thit\t\thit\thit\tmiss\n"))
		is.True(strings.Contains(output, "* 9\thit\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "The policy 'one_factor' from rule #9 will be applied to this request."))

		output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "access-control", "check-policy", "--url=https://user.example.com", "--method=HEAD", "--ip=192.168.1.3", "--verbose", "--config=/config/configuration.yml"})
		is.NoErr(err)

		// This is an example of `authelia access-control check-policy --config .\internal\suites\CLI\configuration.yml --url=https://user.example.com --method=HEAD --ip=192.168.1.3 --verbose`.
		is.True(strings.Contains(output, "Performing policy check for request to 'https://user.example.com' method 'HEAD' from IP '192.168.1.3'.\n\n"))
		is.True(strings.Contains(output, "  #\tDomain\tResource\tMethod\tNetwork\tSubject\n"))
		is.True(strings.Contains(output, "  1\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  2\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  3\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  4\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  5\tmiss\tmiss\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  6\tmiss\thit\t\tmiss\thit\thit\n"))
		is.True(strings.Contains(output, "  7\tmiss\thit\t\thit\thit\thit\n"))
		is.True(strings.Contains(output, "  8\tmiss\thit\t\thit\thit\tmay\n"))
		is.True(strings.Contains(output, "~ 9\thit\thit\t\thit\thit\tmay\n"))
		is.True(strings.Contains(output, "The policy 'one_factor' from rule #9 will potentially be applied to this request. Otherwise the policy 'bypass' from the default policy will be."))
	})

	t.Run("TestStorage01ShouldMigrateUp", TestStorage01ShouldMigrateUp)
	t.Run("TestStorage02ShouldShowSchemaInfo", TestStorage02ShouldShowSchemaInfo)
	t.Run("TestStorage03ShouldExportTOTP", TestStorage03ShouldExportTOTP)
	t.Run("TestStorage04ShouldManageUniqueID", TestStorage04ShouldManageUniqueID)
	t.Run("TestStorage05ShouldChangeEncryptionKey", TestStorage05ShouldChangeEncryptionKey)
	t.Run("TestStorage06ShouldMigrateDown", TestStorage06ShouldMigrateDown)
}

func TestStorage01ShouldMigrateUp(t *testing.T) {
	s := setupCLITest()
	is := is.New(t)
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "up", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)

	pattern0 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is being attempted"`)
	is.True(pattern0.MatchString(output))

	pattern1 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is complete"`)
	is.True(pattern1.MatchString(output))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "up", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: schema already up to date\n"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "history", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)
	is.True(strings.Contains(output, "Migration History:\n\nID\tDate\t\t\t\tBefore\tAfter\tAuthelia Version\n"))
	is.True(strings.Contains(output, "0\t1"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "list-up", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)
	is.True(strings.Contains(output, "Storage Schema Migration List (Up)\n\nNo Migrations Available"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "list-down", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)
	is.True(strings.Contains(output, "Storage Schema Migration List (Down)\n\nVersion\t\tDescription\n"))
	is.True(strings.Contains(output, "1\t\tInitial Schema"))
}

func TestStorage02ShouldShowSchemaInfo(t *testing.T) {
	s := setupCLITest()
	is := is.New(t)
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "schema-info", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)

	is.True(strings.Contains(output, "Schema Version: "))
	is.True(strings.Contains(output, "authentication_logs"))
	is.True(strings.Contains(output, "identity_verification"))
	is.True(strings.Contains(output, "duo_devices"))
	is.True(strings.Contains(output, "user_preferences"))
	is.True(strings.Contains(output, "migrations"))
	is.True(strings.Contains(output, "encryption"))
	is.True(strings.Contains(output, "webauthn_devices"))
	is.True(strings.Contains(output, "totp_configurations"))
	is.True(strings.Contains(output, "Schema Encryption Key: valid"))
}

func TestStorage03ShouldExportTOTP(t *testing.T) {
	s := setupCLITest()
	is := is.New(t)
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
			is.NoErr(err)
			is.True(strings.Contains(output, " and saved it as a PNG image at the path '/tmp/qr.png'"))

			fileInfo, err = os.Stat("/tmp/qr.png")
			is.NoErr(err)
			is.True(fileInfo != nil)
			is.True(!fileInfo.IsDir())
			is.True(fileInfo.Size() > int64(1000))
		} else {
			output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "generate", testCase.config.Username, "--period", strconv.Itoa(int(testCase.config.Period)), "--algorithm", testCase.config.Algorithm, "--digits", strconv.Itoa(int(testCase.config.Digits)), "--config=/config/configuration.storage.yml"})
			is.NoErr(err)
		}

		config, err = storageProvider.LoadTOTPConfiguration(ctx, testCase.config.Username)
		is.NoErr(err)
		is.True(strings.Contains(output, config.URI()))

		expectedLinesCSV = append(expectedLinesCSV, fmt.Sprintf("%s,%s,%s,%d,%d,%s", "Authelia", config.Username, config.Algorithm, config.Digits, config.Period, string(config.Secret)))
		expectedLines = append(expectedLines, config.URI())
	}

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "export", "--format=uri", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)

	for _, expectedLine := range expectedLines {
		is.True(strings.Contains(output, expectedLine))
	}

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "export", "--format=csv", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)

	for _, expectedLine := range expectedLinesCSV {
		is.True(strings.Contains(output, expectedLine))
	}

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "export", "--format=wrong", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: format must be csv, uri, or png"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "export", "--format=png", "--dir=/tmp/qr", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)
	is.True(strings.Contains(output, "Exported TOTP QR codes in PNG format in the '/tmp/qr' directory"))

	for _, testCase := range testCases {
		fileInfo, err = os.Stat(fmt.Sprintf("/tmp/qr/%s.png", testCase.config.Username))

		is.NoErr(err)
		is.True(fileInfo != nil)
		is.True(!fileInfo.IsDir())
		is.True(fileInfo.Size() > int64(1000))
	}

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "totp", "generate", "test", "--period=30", "--algorithm=SHA1", "--digits=6", "--path=/tmp/qr.png", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: image output filepath already exists"))
}

func TestStorage04ShouldManageUniqueID(t *testing.T) {
	_ = os.Mkdir("/tmp/out", 0777)
	s := setupCLITest()
	is := is.New(t)
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "export", "--file=out.yml", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: no data to export"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=webauthn", "--sector=''", "--identifier=1097c8f8-83f2-4506-8138-5f40e83a1285", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: the service name 'webauthn' is invalid, the valid values are: 'openid'"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector=''", "--identifier=1097c8f8-83f2-4506-8138-5f40e83a1285", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)
	is.True(strings.Contains(output, "Added User Opaque Identifier:\n\tService: openid\n\tSector: \n\tUsername: john\n\tIdentifier: 1097c8f8-83f2-4506-8138-5f40e83a1285\n\n"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "export", "--file=/a/no/path/fileout.yml", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: error occurred writing to file '/a/no/path/fileout.yml': open /a/no/path/fileout.yml: no such file or directory"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "export", "--file=out.yml", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: error occurred writing to file 'out.yml': open out.yml: permission denied"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "export", "--file=/tmp/out/1.yml", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)
	is.True(strings.Contains(output, "Exported 1 User Opaque Identifiers to /tmp/out/1.yml"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "export", "--file=/tmp/out/1.yml", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: must specify a file that doesn't exist but '/tmp/out/1.yml' exists"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector=''", "--identifier=1097c8f8-83f2-4506-8138-5f40e83a1285", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: error inserting user opaque id for user 'john' with opaque id '1097c8f8-83f2-4506-8138-5f40e83a1285':"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector=''", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: error inserting user opaque id for user 'john' with opaque id"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector='openidconnect.com'", "--identifier=1097c8f8-83f2-4506-8138-5f40e83a1285", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: error inserting user opaque id for user 'john' with opaque id '1097c8f8-83f2-4506-8138-5f40e83a1285':"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector='openidconnect.net'", "--identifier=b0e17f48-933c-4cba-8509-ee9bfadf8ce5", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)
	is.True(strings.Contains(output, "Added User Opaque Identifier:\n\tService: openid\n\tSector: openidconnect.net\n\tUsername: john\n\tIdentifier: b0e17f48-933c-4cba-8509-ee9bfadf8ce5\n\n"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector='bad-uuid.com'", "--identifier=d49564dc-b7a1-11ec-8429-fcaa147128ea", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: the identifier providerd 'd49564dc-b7a1-11ec-8429-fcaa147128ea' is a version 1 UUID but only version 4 UUID's accepted as identifiers"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "add", "john", "--service=openid", "--sector='bad-uuid.com'", "--identifier=asdmklasdm", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: the identifier provided 'asdmklasdm' is invalid as it must be a version 4 UUID but parsing it had an error: invalid UUID length: 10"))

	data, err := os.ReadFile("/tmp/out/1.yml")
	is.NoErr(err)

	var export model.UserOpaqueIdentifiersExport

	is.NoErr(yaml.Unmarshal(data, &export))

	is.True(len(export.Identifiers) == 1)

	is.Equal(1, export.Identifiers[0].ID)
	is.Equal("1097c8f8-83f2-4506-8138-5f40e83a1285", export.Identifiers[0].Identifier.String())
	is.Equal("john", export.Identifiers[0].Username)
	is.Equal("", export.Identifiers[0].SectorID)
	is.Equal("openid", export.Identifiers[0].Service)

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "user", "identifiers", "export", "--file=/tmp/out/2.yml", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)
	is.True(strings.Contains(output, "Exported 2 User Opaque Identifiers to /tmp/out/2.yml"))

	export = model.UserOpaqueIdentifiersExport{}

	data, err = os.ReadFile("/tmp/out/2.yml")
	is.NoErr(err)

	is.NoErr(yaml.Unmarshal(data, &export))

	is.True(len(export.Identifiers) == 2)

	is.Equal(1, export.Identifiers[0].ID)
	is.Equal("1097c8f8-83f2-4506-8138-5f40e83a1285", export.Identifiers[0].Identifier.String())
	is.Equal("john", export.Identifiers[0].Username)
	is.Equal("", export.Identifiers[0].SectorID)
	is.Equal("openid", export.Identifiers[0].Service)

	is.Equal(2, export.Identifiers[1].ID)
	is.Equal("b0e17f48-933c-4cba-8509-ee9bfadf8ce5", export.Identifiers[1].Identifier.String())
	is.Equal("john", export.Identifiers[1].Username)
	is.Equal("openidconnect.net", export.Identifiers[1].SectorID)
	is.Equal("openid", export.Identifiers[1].Service)
}

func TestStorage05ShouldChangeEncryptionKey(t *testing.T) {
	s := setupCLITest()
	is := is.New(t)
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "change-key", "--new-encryption-key=apple-apple-apple-apple", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)

	is.True(strings.Contains(output, "Completed the encryption key change. Please adjust your configuration to use the new key.\n"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "schema-info", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)

	is.True(strings.Contains(output, "Schema Version: "))
	is.True(strings.Contains(output, "authentication_logs"))
	is.True(strings.Contains(output, "identity_verification"))
	is.True(strings.Contains(output, "duo_devices"))
	is.True(strings.Contains(output, "user_preferences"))
	is.True(strings.Contains(output, "migrations"))
	is.True(strings.Contains(output, "encryption"))
	is.True(strings.Contains(output, "webauthn_devices"))
	is.True(strings.Contains(output, "totp_configurations"))
	is.True(strings.Contains(output, "Schema Encryption Key: invalid"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)
	is.True(strings.Contains(output, "Encryption key validation: failed.\n\nError: the encryption key is not valid against the schema check value.\n"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--verbose", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)
	is.True(strings.Contains(output, "Encryption key validation: failed.\n\nError: the encryption key is not valid against the schema check value, 4 of 4 total TOTP secrets were invalid.\n"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--encryption-key=apple-apple-apple-apple", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)
	is.True(strings.Contains(output, "Encryption key validation: success.\n"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "check", "--verbose", "--encryption-key=apple-apple-apple-apple", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)
	is.True(strings.Contains(output, "Encryption key validation: success.\n"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "change-key", "--encryption-key=apple-apple-apple-apple", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: you must set the --new-encryption-key flag\n"))

	output, err = s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "encryption", "change-key", "--encryption-key=apple-apple-apple-apple", "--new-encryption-key=abc", "--config=/config/configuration.storage.yml"})
	is.Equal(err.Error(), "exit status 1")
	is.True(strings.Contains(output, "Error: the new encryption key must be at least 20 characters\n"))
}

func TestStorage06ShouldMigrateDown(t *testing.T) {
	s := setupCLITest()
	is := is.New(t)
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "storage", "migrate", "down", "--target=0", "--destroy-data", "--config=/config/configuration.storage.yml"})
	is.NoErr(err)

	pattern0 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is being attempted"`)
	is.True(pattern0.MatchString(output))

	pattern1 := regexp.MustCompile(`"Storage schema migration from \d+ to \d+ is complete"`)
	is.True(pattern1.MatchString(output))
}
