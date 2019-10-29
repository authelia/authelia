package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const dockerPullCommandLine = "docker-compose -f docker-compose.yml " +
	"-f example/compose/mongo/docker-compose.yml " +
	"-f example/compose/redis/docker-compose.yml " +
	"-f example/compose/nginx/portal/docker-compose.yml " +
	"-f example/compose/smtp/docker-compose.yml " +
	"-f example/compose/httpbin/docker-compose.yml " +
	"-f example/compose/ldap/docker-compose.admin.yml " +
	"-f example/compose/ldap/docker-compose.yml " +
	"pull"

// RunCI run the CI scripts
func RunCI(cmd *cobra.Command, args []string) {
	command := CommandWithStdout("bash", "-c", dockerPullCommandLine)
	err := command.Run()

	if err != nil {
		panic(err)
	}

	fmt.Println("===== Build stage =====")
	command = CommandWithStdout("authelia-scripts", "build")
	err = command.Run()

	if err != nil {
		panic(err)
	}

	fmt.Println("===== Unit testing stage =====")
	command = CommandWithStdout("authelia-scripts", "unittest")
	err = command.Run()

	if err != nil {
		panic(err)
	}

	fmt.Println("===== Docker image build stage =====")
	command = CommandWithStdout("authelia-scripts", "docker", "build")
	err = command.Run()

	if err != nil {
		panic(err)
	}

	fmt.Println("===== End-to-end testing stage =====")
	command = CommandWithStdout("authelia-scripts", "suites", "test", "--headless", "--only-forbidden")
	err = command.Run()

	if err != nil {
		panic(err)
	}
}
