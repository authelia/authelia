package suites

import (
	"os"
	"time"
)

var pamSuiteName = "PAM"

const pamTOTPSecret = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP" //nolint:gosec // Test fixture TOTP secret, not a real credential.

var pamDockerEnvironment = NewDockerEnvironment([]string{
	"internal/suites/compose.yml",
	"internal/suites/PAM/compose.yml",
	"internal/suites/example/compose/authelia/compose.backend.{}.yml",
	"internal/suites/example/compose/authelia/compose.frontend.{}.yml",
	"internal/suites/example/compose/nginx/portal/compose.yml",
	"internal/suites/example/compose/pam/compose.yml",
})

func init() {
	setup := func(suitePath string) (err error) {
		if err = pamDockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(pamDockerEnvironment, pamSuiteName); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return pamDockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend", "pam")
	}

	teardown := func(suitePath string) error {
		err := pamDockerEnvironment.Down()
		_ = os.Remove("/tmp/db.sqlite3")

		return err
	}

	GlobalRegistry.Register(pamSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TestTimeout:     3 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
		Description:     "PAM module integration tests for SSH authentication via pam_authelia",
	})
}
