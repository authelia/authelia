package cmd

import (
	"os"
	"strings"

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

func cmdUnitTestRun(cmd *cobra.Command, _ []string) {
	log.SetLevel(log.TraceLevel)

	goTestCmd := "go test -coverprofile=coverage.txt -covermode=atomic"

	if buildkite, _ := cmd.Flags().GetBool("buildkite"); buildkite && strings.HasPrefix(os.Getenv("BUILDKITE_BRANCH"), "gh-readonly-queue/master/") {
		goTestCmd += " -race"
	}

	goTestCmd += " $(go list ./... | grep -v suites)"

	if err := utils.Shell(goTestCmd).Run(); err != nil {
		log.Fatal(err)
	}

	pnpmCmd := utils.Shell("pnpm test")
	pnpmCmd.Dir = webDirectory

	pnpmCmd.Env = append(os.Environ(), "CI=true")

	if err := pnpmCmd.Run(); err != nil {
		log.Fatal(err)
	}
}
