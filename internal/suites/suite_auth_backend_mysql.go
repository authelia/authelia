package suites

import (
	"context"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

var authBackendMysqlSuiteName = "AuthBackendMySQL"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/AuthenticationBackendMySQL/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/docker-compose.yml",
		"internal/suites/example/compose/nginx/portal/docker-compose.yml",
		"internal/suites/example/compose/smtp/docker-compose.yml",
		"internal/suites/example/compose/mysql/docker-compose.yml",
		"internal/suites/example/compose/ldap/docker-compose.yml",
	})

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, authBackendMysqlSuiteName); err != nil {
			return err
		}

		address, err := schema.NewAddressFromURL(&url.URL{Scheme: schema.AddressSchemeTCP, Host: "192.168.240.150:3306"})
		storageMySQLTmpConfig.Storage.MySQL.StorageSQL.Address = &schema.AddressTCP{Address: *address}
		provider := storage.NewMySQLProvider(&storageMySQLTmpConfig, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer func() {
			cancel()
		}()

		passwordHash := []byte("$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/")
		userList := []model.User{
			{
				Username:    "john",
				Email:       "john.doe@authelia.com",
				DisplayName: "John Doe",
				Groups:      []string{"admins", "dev"},
				Password:    passwordHash,
			},
			{
				Username:    "harry",
				Email:       "harry.potter@authelia.com",
				DisplayName: "Harry Potter",
				Password:    passwordHash,
			},
			{
				Username:    "bob",
				Email:       "bob.dylan@authelia.com",
				DisplayName: "Bob Dylan",
				Groups:      []string{"dev"},
				Password:    passwordHash,
			},
			{
				Username:    "james",
				Email:       "james.dean@authelia.com",
				DisplayName: "James Dean",
				Password:    passwordHash,
			},
		}

		for _, user := range userList {
			if err := provider.CreateUser(ctx, user); err != nil {
				log.Warnf("could not create user %s (%s).\n", user.Username, err)
			}
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

	GlobalRegistry.Register(authBackendMysqlSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
