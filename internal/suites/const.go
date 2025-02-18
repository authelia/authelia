package suites

import (
	"fmt"
	"os"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// BaseDomain the base domain.
var (
	BaseDomain     = "example.com:8080"
	Example2DotCom = "example2.com:8080"
	Example3DotCom = "example3.com:8080"
)

const (
	SHA1   = "SHA1"
	SHA256 = "SHA256"
	SHA512 = "SHA512"
)

// GetPathPrefix returns the prefix/url_base of the login portal.
func GetPathPrefix() string {
	return os.Getenv("PathPrefix")
}

// LoginBaseURLFmt the base URL of the login portal for specified baseDomain.
func LoginBaseURLFmt(baseDomain string) string {
	if baseDomain == "" {
		baseDomain = BaseDomain
	}

	return fmt.Sprintf("https://login.%s", baseDomain)
}

// LoginBaseURL the base URL of the login portal.
var LoginBaseURL = LoginBaseURLFmt(BaseDomain)

// SingleFactorBaseURLFmt the base URL of the singlefactor with custom domain.
func SingleFactorBaseURLFmt(baseDomain string) string {
	if baseDomain == "" {
		baseDomain = BaseDomain
	}

	return fmt.Sprintf("https://singlefactor.%s", baseDomain)
}

// SingleFactorBaseURL the base URL of the singlefactor domain.
var SingleFactorBaseURL = SingleFactorBaseURLFmt(BaseDomain)

// AdminBaseURL the base URL of the admin domain.
var AdminBaseURL = fmt.Sprintf("https://admin.%s", BaseDomain)

// MailBaseURL the base URL of the mail domain.
var MailBaseURL = fmt.Sprintf("https://mail.%s", BaseDomain)

// HomeBaseURL the base URL of the home domain.
var HomeBaseURL = fmt.Sprintf("https://home.%s", BaseDomain)

// PublicBaseURL the base URL of the public domain.
var PublicBaseURL = fmt.Sprintf("https://public.%s", BaseDomain)

// SecureBaseURL the base URL of the secure domain.
var SecureBaseURL = fmt.Sprintf("https://secure.%s", BaseDomain)

// DenyBaseURL the base URL of the dev domain.
var DenyBaseURL = fmt.Sprintf("https://deny.%s", BaseDomain)

// DevBaseURL the base URL of the dev domain.
var DevBaseURL = fmt.Sprintf("https://dev.%s", BaseDomain)

// MX1MailBaseURL the base URL of the mx1.mail domain.
var MX1MailBaseURL = fmt.Sprintf("https://mx1.mail.%s", BaseDomain)

// MX2MailBaseURL the base URL of the mx2.mail domain.
var MX2MailBaseURL = fmt.Sprintf("https://mx2.mail.%s", BaseDomain)

// OIDCBaseURL the base URL of the oidc domain.
var OIDCBaseURL = fmt.Sprintf("https://oidc.%s", BaseDomain)

// DuoBaseURL the base URL of the Duo configuration API.
var DuoBaseURL = "https://duo.example.com"

// AutheliaBaseURL the base URL of Authelia service.
var AutheliaBaseURL = "https://authelia.example.com:9091"

const (
	t            = "true"
	testUsername = "john"
	testPassword = "password"
)

const (
	envFileProd        = "/web/.env.production"
	envFileDev         = "/web/.env.development"
	namespaceAuthelia  = "authelia"
	namespaceDashboard = "kubernetes-dashboard"
	namespaceKube      = "kube-system"
)

var (
	storageLocalTmpConfig = schema.Configuration{
		TOTP: schema.TOTP{
			Issuer:        "Authelia",
			DefaultPeriod: 6,
		},
		Storage: schema.Storage{
			EncryptionKey: "a_not_so_secure_encryption_key",
			Local: &schema.StorageLocal{
				Path: "/tmp/db.sqlite3",
			},
		},
	}

	storageMySQLTmpConfig = schema.Configuration{
		TOTP: schema.TOTP{
			Issuer:        "Authelia",
			DefaultPeriod: 6,
		},
		Storage: schema.Storage{
			EncryptionKey: "a_not_so_secure_encryption_key",
			MySQL: &schema.StorageMySQL{
				StorageSQL: schema.StorageSQL{
					// Address: schema.NewAddressFromURL(url.URL{Scheme: AddressSchemeTCP, Host: "localhost:3306"}),.
					Database: "authelia",
					Username: "admin",
					Password: "password",
				},
			},
		},
	}

	storagePostgressTmpConfig = schema.Configuration{
		TOTP: schema.TOTP{
			Issuer:        "Authelia",
			DefaultPeriod: 6,
		},
		Storage: schema.Storage{
			EncryptionKey: "a_not_so_secure_encryption_key",
			PostgreSQL: &schema.StoragePostgreSQL{
				StorageSQL: schema.StorageSQL{
					Database: "authelia",
					Username: "admin",
					Password: "password",
				},
				Schema: "public",
			},
		},
	}
)

var (
	defaultComposeFiles = []string{
		"internal/suites/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/docker-compose.yml",
		"internal/suites/example/compose/nginx/portal/docker-compose.yml",
		"internal/suites/example/compose/smtp/docker-compose.yml",
	}
)
