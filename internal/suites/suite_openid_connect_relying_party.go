package suites

import (
	"time"
)

var openIDConnectRelyingPartySuiteName = "OpenIDConnectRelyingParty"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/compose.yml",
		"internal/suites/OpenIDConnectRelyingParty/compose.yml",
		"internal/suites/OpenIDConnectRelyingParty/compose.upstream.{}.yml",
		"internal/suites/example/compose/authelia/compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/compose.yml",
		"internal/suites/example/compose/nginx/portal/compose.yml",
		"internal/suites/example/compose/smtp/compose.yml",
	})

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		// The external provider is waited on first because the instance under test discovers it while starting and
		// restarts until that discovery succeeds.
		if err = waitUntilAutheliaUpstreamBackendIsReady(dockerEnvironment); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, openIDConnectRelyingPartySuiteName); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, dockerEnvironment)
	}

	displayAutheliaLogs := func() error {
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend", "authelia-upstream-backend")
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down()
		return err
	}

	GlobalRegistry.Register(openIDConnectRelyingPartySuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TestTimeout:     2 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
