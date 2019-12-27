package suites

import (
	"fmt"
	//TODO(nightah): Remove when turning off Travis
	"os"
	"time"
)

func init() {
	//TODO(nightah): Remove when turning off Travis
	travis := os.Getenv("TRAVIS")
	backend := ""
	if travis == "true" {
		backend = "example/compose/authelia/docker-compose.backend-dist-travis.yml"
	} else {
		backend = "example/compose/authelia/docker-compose.backend-dist.yml"
	}
	//TODO(nightah): Remove when turning off Travis

	dockerEnvironment := NewDockerEnvironment([]string{
		"docker-compose.yml",
		"internal/suites/Docker/docker-compose.yml",
		//TODO(nightah): Change to "example/compose/authelia/docker-compose.backend-dist.yml" when removing Travis
		backend,
		"example/compose/authelia/docker-compose.frontend-dist.yml",
		"example/compose/nginx/backend/docker-compose.yml",
		"example/compose/nginx/portal/docker-compose.yml",
		"example/compose/smtp/docker-compose.yml",
	})

	setup := func(suitePath string) error {
		if err := dockerEnvironment.Up(); err != nil {
			return err
		}

		return waitUntilAutheliaBackendIsReady(dockerEnvironment)
	}

	onSetupTimeout := func() error {
		backendLogs, err := dockerEnvironment.Logs("authelia-backend", nil)
		if err != nil {
			return err
		}
		fmt.Println(backendLogs)

		frontendLogs, err := dockerEnvironment.Logs("authelia-frontend", nil)
		if err != nil {
			return err
		}
		fmt.Println(frontendLogs)
		return nil
	}

	teardown := func(suitePath string) error {
		return dockerEnvironment.Down()
	}

	GlobalRegistry.Register("Docker", Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  onSetupTimeout,
		TestTimeout:     1 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,

		Description: `This suite has been created to test the distributable version of Authelia
It's often useful to test this one before the Kube one.`,
	})
}
