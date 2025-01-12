package suites

import (
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
)

var authBackendPostgresSuiteName = "AuthBackendPostgres"

func init() {
	composeFiles := defaultComposeFiles
	composeFiles = append(composeFiles,
		"internal/suites/AuthenticationBackendPostgres/docker-compose.yml",
		"internal/suites/example/compose/postgres/docker-compose.yml",
	)

	dockerEnvironment := NewDockerEnvironment(composeFiles)

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, authBackendPostgresSuiteName); err != nil {
			return err
		}

		address, err := schema.NewAddressFromURL(&url.URL{Scheme: schema.AddressSchemeTCP, Host: "192.168.240.151:5432"})
		if err != nil {
			return err
		}

		storagePostgressTmpConfig.Storage.PostgreSQL.StorageSQL.Address = &schema.AddressTCP{Address: *address}
		storageProvider := storage.NewPostgreSQLProvider(&storagePostgressTmpConfig, nil)
		userProvider := authentication.NewDBUserProvider(&schema.DefaultDBAuthenticationBackendConfig, storageProvider)

		if err = createDemoUsers(userProvider); err != nil {
			log.Warn(err)
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend")
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down()
		return err
	}

	GlobalRegistry.Register(authBackendPostgresSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
