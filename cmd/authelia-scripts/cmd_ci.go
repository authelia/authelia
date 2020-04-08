package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/utils"
)

const dockerPullCommandLine = "docker-compose -p authelia -f internal/suites/docker-compose.yml " +
	"-f internal/suites/example/compose/mariadb/docker-compose.yml " +
	"-f internal/suites/example/compose/redis/docker-compose.yml " +
	"-f internal/suites/example/compose/nginx/portal/docker-compose.yml " +
	"-f internal/suites/example/compose/smtp/docker-compose.yml " +
	"-f internal/suites/example/compose/httpbin/docker-compose.yml " +
	"-f internal/suites/example/compose/ldap/docker-compose.admin.yml " +
	"-f internal/suites/example/compose/ldap/docker-compose.yml " +
	"pull"

// RunCI run the CI scripts
func RunCI(cmd *cobra.Command, args []string) {
	log.Info("=====> Build stage <=====")
	if err := utils.CommandWithStdout("authelia-scripts", "--log-level", "debug", "build").Run(); err != nil {
		log.Fatal(err)
	}

	log.Info("=====> Unit testing stage <=====")
	if err := utils.CommandWithStdout("authelia-scripts", "--log-level", "debug", "unittest").Run(); err != nil {
		log.Fatal(err)
	}
}
