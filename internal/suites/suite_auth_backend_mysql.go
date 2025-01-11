package suites

import (
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
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
		if err != nil {
			return err
		}

		storageMySQLTmpConfig.Storage.MySQL.StorageSQL.Address = &schema.AddressTCP{Address: *address}
		storageProvider := storage.NewMySQLProvider(&storageMySQLTmpConfig, nil)
		userProvider := authentication.NewDBUserProvider(&schema.DefaultDBAuthenticationBackendConfig, storageProvider)

		if err = userProvider.StartupCheck(); err != nil {
			return err
		}

		userList := []struct {
			Username    string
			Email       string
			DisplayName string
			Groups      []string
			Password    string
		}{
			{
				Username:    "john",
				Email:       "john.doe@authelia.com",
				DisplayName: "John Doe",
				Groups:      []string{"admins", "dev"},
				Password:    "password",
			},
			{
				Username:    "harry",
				Email:       "harry.potter@authelia.com",
				DisplayName: "Harry Potter",
				Password:    "password",
			},
			{
				Username:    "bob",
				Email:       "bob.dylan@authelia.com",
				DisplayName: "Bob Dylan",
				Groups:      []string{"dev"},
				Password:    "password",
			},
			{
				Username:    "james",
				Email:       "james.dean@authelia.com",
				DisplayName: "James Dean",
				Password:    "password",
			},
		}

		for _, user := range userList {
			if err := userProvider.AddUser(
				user.Username,
				user.DisplayName,
				user.Password,
				authentication.WithEmail(user.Email),
				authentication.WithGroups(user.Groups)); err != nil {
				log.Warnf("failed to create demo user '%s': %s.\n", user.Username, err)
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
