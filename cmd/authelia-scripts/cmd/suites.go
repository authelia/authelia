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
	runningSuiteFile   = ".suite"
	failfast, headless bool
	testPattern        string
)

func newSuitesCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "suites",
		Short:   cmdSuitesShort,
		Long:    cmdSuitesLong,
		Example: cmdSuitesExample,
		Run:     cmdSuitesListRun,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newSuitesListCmd(), newSuitesSetupCmd(), newSuitesTestCmd(), newSuitesTeardownCmd())

	return cmd
}

func newSuitesListCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "list",
		Short:   cmdSuitesListShort,
		Long:    cmdSuitesListLong,
		Example: cmdSuitesListExample,
		Run:     cmdSuitesListRun,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newSuitesSetupCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "setup [suite]",
		Short:   cmdSuitesSetupShort,
		Long:    cmdSuitesSetupLong,
		Example: cmdSuitesSetupExample,
		Run:     cmdSuitesSetupRun,
		Args:    cobra.MaximumNArgs(1),

		DisableAutoGenTag: true,
	}

	return cmd
}

func newSuitesTeardownCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "teardown [suite]",
		Short:   cmdSuitesTeardownShort,
		Long:    cmdSuitesTeardownLong,
		Example: cmdSuitesTeardownExample,
		Run:     cmdSuitesTeardownRun,
		Args:    cobra.MaximumNArgs(1),

		DisableAutoGenTag: true,
	}

	return cmd
}

func newSuitesTestCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "test [suite]",
		Short:   cmdSuitesTestShort,
		Long:    cmdSuitesTestLong,
		Example: cmdSuitesTestExample,
		Run:     cmdSuitesTestRun,
		Args:    cobra.MaximumNArgs(1),

		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&failfast, "failfast", false, "Stops tests on first failure")
	cmd.Flags().BoolVar(&headless, "headless", false, "Run tests in headless mode")
	cmd.Flags().StringVar(&testPattern, "test", "", "The single test to run")

	return cmd
}

func cmdSuitesListRun(_ *cobra.Command, _ []string) {
	fmt.Println(strings.Join(listSuites(), "\n"))
}

func cmdSuitesSetupRun(_ *cobra.Command, args []string) {
	providedSuite := args[0]

	runningSuite, err := getRunningSuite()
	if err != nil {
		log.Fatal(err)
	}

	if runningSuite != "" && runningSuite != providedSuite {
		log.Fatal("A suite is already running")
	}

	if err := setupSuite(providedSuite); err != nil {
		log.Fatal(err)
	}
}

func cmdSuitesTeardownRun(_ *cobra.Command, args []string) {
	var suiteName string
	if len(args) == 1 {
		suiteName = args[0]
	} else {
		runningSuite, err := getRunningSuite()
		if err != nil {
			log.Fatal(err)
		}

		if runningSuite == "" {
			log.Fatal(ErrNoRunningSuite)
		}

		suiteName = runningSuite
	}

	if err := teardownSuite(suiteName); err != nil {
		log.Fatal(err)
	}
}

func cmdSuitesTestRun(_ *cobra.Command, args []string) {
	runningSuite, err := getRunningSuite()
	if err != nil {
		log.Fatal(err)
	}

	// If suite(s) are provided as argument.
	if len(args) >= 1 {
		suiteArg := args[0]

		if runningSuite != "" && suiteArg != runningSuite {
			log.Fatal(errors.New(txtRunningSuite + runningSuite + ") is different than suite(s) to be tested (" + suiteArg + "). Shutdown running suite and retry"))
		}

		if err := runMultipleSuitesTests(strings.Split(suiteArg, ","), runningSuite == ""); err != nil {
			log.Fatal(err)
		}
	} else {
		if runningSuite != "" {
			fmt.Println(txtRunningSuite + runningSuite + ") detected. Run tests of that suite")

			if err := runSuiteTests(runningSuite, false); err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("No suite provided therefore all suites will be tested")

			if err := runAllSuites(); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func listSuites() []string {
	suiteNames := suites.GlobalRegistry.Suites()

	sort.Strings(suiteNames)

	return suiteNames
}

func checkSuiteAvailable(suite string) error {
	suiteNames := listSuites()

	for _, s := range suiteNames {
		if s == suite {
			return nil
		}
	}

	return ErrNotAvailableSuite
}

func runSuiteSetupTeardown(command string, suite string) error {
	selectedSuite := suite

	err := checkSuiteAvailable(selectedSuite)
	if err != nil {
		if err == ErrNotAvailableSuite {
			log.Fatal(errors.New("Suite named " + selectedSuite + " does not exist"))
		}

		log.Fatal(err)
	}

	s := suites.GlobalRegistry.Get(selectedSuite)

	if command == "teardown" {
		if _, err := os.Stat("../../web/.nyc_output"); err == nil {
			log.Infof("Generating frontend coverage reports for suite %s...", suite)

			cmd := utils.Command("pnpm", "report")
			cmd.Dir = "web"
			cmd.Env = os.Environ()

			err := cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	cmd := utils.CommandWithStdout("go", "run", "cmd/authelia-suites/main.go", command, selectedSuite)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return utils.RunCommandWithTimeout(cmd, s.SetUpTimeout)
}

func runOnSetupTimeout(suite string) error {
	cmd := utils.CommandWithStdout("go", "run", "cmd/authelia-suites/main.go", "timeout", suite)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return utils.RunCommandWithTimeout(cmd, 15*time.Second)
}

func runOnError(suite string) error {
	cmd := utils.CommandWithStdout("go", "run", "cmd/authelia-suites/main.go", "error", suite)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return utils.RunCommandWithTimeout(cmd, 15*time.Second)
}

func setupSuite(suiteName string) error {
	log.Infof("Setup environment for suite %s...", suiteName)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	interrupted := false

	go func() {
		<-signalChannel

		interrupted = true
	}()

	if errSetup := runSuiteSetupTeardown("setup", suiteName); errSetup != nil || interrupted {
		if errSetup == utils.ErrTimeoutReached {
			err := runOnSetupTimeout(suiteName)
			if err != nil {
				log.Fatal(err)
			}
		}

		err := teardownSuite(suiteName)
		if err != nil {
			log.Fatal(err)
		}

		return errSetup
	}

	return nil
}

func teardownSuite(suiteName string) error {
	log.Infof("Tear down environment for suite %s...", suiteName)
	return runSuiteSetupTeardown("teardown", suiteName)
}

func getRunningSuite() (string, error) {
	exist, err := utils.FileExists(runningSuiteFile)
	if err != nil {
		return "", err
	}

	if !exist {
		return "", nil
	}

	b, err := os.ReadFile(runningSuiteFile)

	return string(b), err
}

func runSuiteTests(suiteName string, withEnv bool) error {
	if withEnv {
		if err := setupSuite(suiteName); err != nil {
			return err
		}
	}

	suite := suites.GlobalRegistry.Get(suiteName)

	// Default value is 1 minute.
	timeout := "60s"
	if suite.TestTimeout > 0 {
		timeout = fmt.Sprintf("%ds", int64(suite.TestTimeout/time.Second))
	}

	fail := ""
	if failfast {
		fail = "-failfast"
	}

	testCmdLine := fmt.Sprintf("go test -count=1 -v ./internal/suites -timeout %s %s ", timeout, fail)

	if testPattern != "" {
		testCmdLine += fmt.Sprintf("-run '%s'", testPattern)
	} else {
		testCmdLine += fmt.Sprintf("-run '^(Test%sSuite)$'", suiteName)
	}

	log.Infof("Running tests of suite %s...", suiteName)
	log.Debugf("Running tests with command: %s", testCmdLine)

	cmd := utils.CommandWithStdout("bash", "-c", testCmdLine)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if headless {
		cmd.Env = append(cmd.Env, "HEADLESS=y")
	}

	cmd.Env = append(cmd.Env, "SUITES_LOG_LEVEL="+log.GetLevel().String())

	testErr := cmd.Run()

	// If the tests failed, run the error hook.
	if testErr != nil {
		if err := runOnError(suiteName); err != nil {
			// Do not return this error to return the test error instead
			// and not skip the teardown phase.
			log.Errorf("Error executing OnError callback: %v", err)
		}
	}

	if withEnv {
		if err := teardownSuite(suiteName); err != nil {
			// Do not return this error to return the test error instead.
			log.Errorf("Error running teardown: %v", err)
		}
	}

	return testErr
}

func runMultipleSuitesTests(suiteNames []string, withEnv bool) error {
	for _, suiteName := range suiteNames {
		if err := runSuiteTests(suiteName, withEnv); err != nil {
			return err
		}
	}

	return nil
}

func runAllSuites() error {
	log.Info("Start running all suites")

	for _, s := range listSuites() {
		if err := runSuiteTests(s, true); err != nil {
			return err
		}
	}

	log.Info("All suites passed successfully")

	return nil
}
