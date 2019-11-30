package main

import (
	"github.com/clems4ever/authelia/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const dockerPullCommandLine = "docker-compose -f docker-compose.yml " +
	"-f example/compose/mariadb/docker-compose.yml " +
	"-f example/compose/redis/docker-compose.yml " +
	"-f example/compose/nginx/portal/docker-compose.yml " +
	"-f example/compose/smtp/docker-compose.yml " +
	"-f example/compose/httpbin/docker-compose.yml " +
	"-f example/compose/ldap/docker-compose.admin.yml " +
	"-f example/compose/ldap/docker-compose.yml " +
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

	log.Info("=====> Build Docker stage <=====")
	if err := utils.CommandWithStdout("authelia-scripts", "--log-level", "debug", "docker", "build").Run(); err != nil {
		log.Fatal(err)
	}
}
