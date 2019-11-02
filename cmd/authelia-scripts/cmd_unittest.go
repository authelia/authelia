package main

import (
	"github.com/clems4ever/authelia/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// RunUnitTest run the unit tests
func RunUnitTest(cobraCmd *cobra.Command, args []string) {
	err := utils.Shell("go test $(go list ./... | grep -v suites)").Run()
	if err != nil {
		log.Fatal(err)
	}
}
