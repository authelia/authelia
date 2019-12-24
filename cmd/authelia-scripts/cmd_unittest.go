package main

import (
	"github.com/authelia/authelia/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

// RunUnitTest run the unit tests
func RunUnitTest(cobraCmd *cobra.Command, args []string) {
	log.SetLevel(log.TraceLevel)
	if err := utils.Shell("go test $(go list ./... | grep -v suites)").Run(); err != nil {
		log.Fatal(err)
	}

	cmd := utils.Shell("npm run test")
	cmd.Dir = "web"
	cmd.Env = append(os.Environ(), "CI=true")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
