package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/suites"
	"github.com/authelia/authelia/v4/internal/utils"
)

var (
	externalFailfast        bool
	externalHeadless        bool
	externalUpdateSnapshots bool
)

// externalSuiteTestEntrypoints maps a registered external suite name to its Go test entry
// function.
var externalSuiteTestEntrypoints = map[string]string{
	"docs":      "TestDocsSuite",
	"templates": "TestTemplatesSuite",
}

func newSuitesExternalCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "external",
		Short:   cmdSuitesExternalShort,
		Long:    cmdSuitesExternalLong,
		Example: cmdSuitesExternalExample,
		Run:     cmdSuitesExternalListRun,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newSuitesExternalListCmd(), newSuitesExternalTestCmd())

	return cmd
}

func newSuitesExternalListCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "list",
		Short:   cmdSuitesExternalListShort,
		Long:    cmdSuitesExternalListLong,
		Example: cmdSuitesExternalListExample,
		Run:     cmdSuitesExternalListRun,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newSuitesExternalTestCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "test [suite]",
		Short:   cmdSuitesExternalTestShort,
		Long:    cmdSuitesExternalTestLong,
		Example: cmdSuitesExternalTestExample,
		Run:     cmdSuitesExternalTestRun,
		Args:    cobra.ExactArgs(1),

		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&externalFailfast, "failfast", false, "Stops tests on first failure")
	cmd.Flags().BoolVar(&externalHeadless, "headless", false, "Run tests in headless mode")
	cmd.Flags().BoolVar(&externalUpdateSnapshots, "update-snapshots", false, "Overwrite visual snapshot baselines with the output of the current run")

	return cmd
}

func cmdSuitesExternalListRun(_ *cobra.Command, _ []string) {
	names := suites.ExternalGlobalRegistry.Suites()
	sort.Strings(names)
	fmt.Println(strings.Join(names, "\n"))
}

func cmdSuitesExternalTestRun(_ *cobra.Command, args []string) {
	suiteName := args[0]
	checkExternalSuiteAvailable(suiteName)

	entrypoint, ok := externalSuiteTestEntrypoints[suiteName]
	if !ok {
		log.Fatalf("No test entrypoint registered for external suite %s", suiteName)
	}

	externalSuite := suites.ExternalGlobalRegistry.Get(suiteName)

	// Swallow SIGINT/SIGTERM so cobra waits for the test subprocess to run its own TestMain
	// signal handler (which stops the dev server and exits 130) before returning.
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	interrupted := false

	go func() {
		<-signalChannel

		interrupted = true

		log.Warn("Interrupt received - waiting for the test subprocess to clean up the dev server")
	}()

	timeout := externalSuite.TestTimeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	failfast := ""
	if externalFailfast {
		failfast = "-failfast "
	}

	testCmdLine := fmt.Sprintf(
		"go test -count=1 -v -tags=externalsuites ./internal/suites -timeout %s %s-run '^(%s)$'",
		timeout, failfast, entrypoint,
	)

	log.Infof("Running external suite %s: %s", suiteName, testCmdLine)

	cmd := utils.CommandWithStdout("bash", "-c", testCmdLine)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SUITES_LOG_LEVEL="+log.GetLevel().String())

	if externalHeadless {
		cmd.Env = append(cmd.Env, "HEADLESS=y")
	}

	if externalUpdateSnapshots {
		cmd.Env = append(cmd.Env, "UPDATE_SNAPSHOTS=1")
	}

	testErr := cmd.Run()

	switch {
	case interrupted:
		os.Exit(130)
	case testErr != nil:
		log.Fatalf("external suite %s failed: %v", suiteName, testErr)
	}
}

func checkExternalSuiteAvailable(name string) {
	for _, registered := range suites.ExternalGlobalRegistry.Suites() {
		if registered == name {
			return
		}
	}

	log.Fatal(errors.New("external suite named " + name + " does not exist"))
}
