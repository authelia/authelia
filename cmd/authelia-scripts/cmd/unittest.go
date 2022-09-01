package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func newUnitTestCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "unittest",
		Short:   cmdUnitTestShort,
		Long:    cmdUnitTestLong,
		Example: cmdUnitTestExample,
		Args:    cobra.NoArgs,
		Run:     cmdUnitTestRun,

		DisableAutoGenTag: true,
	}

	return cmd
}

func cmdUnitTestRun(_ *cobra.Command, _ []string) {
	log.SetLevel(log.TraceLevel)

	if err := utils.Shell("go test -coverprofile=coverage.txt -covermode=atomic $(go list ./... | grep -v suites)").Run(); err != nil {
		log.Fatal(err)
	}

	cmd := utils.Shell("pnpm test")
	cmd.Dir = webDirectory

	cmd.Env = append(os.Environ(), "CI=true")

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
