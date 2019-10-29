package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

func listDirectories(path string) ([]string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0)

	for _, f := range files {
		if f.IsDir() {
			dirs = append(dirs, f.Name())
		}
	}

	return dirs, nil
}

func listSuites() ([]string, error) {
	return listDirectories("./test/suites/")
}

func suiteAvailable(suite string, suites []string) (bool, error) {
	suites, err := listSuites()

	if err != nil {
		return false, err
	}

	for _, s := range suites {
		if s == suite {
			return true, nil
		}
	}
	return false, nil
}

// SuitesListCmd Command for listing the available suites
var SuitesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available suites.",
	Run: func(cmd *cobra.Command, args []string) {
		suites, err := listSuites()

		if err != nil {
			panic(err)
		}

		fmt.Println(strings.Join(suites, "\n"))
	},
	Args: cobra.ExactArgs(0),
}

// SuitesCleanCmd Command for cleaning suite environments
var SuitesCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean suite environments.",
	Run: func(cmd *cobra.Command, args []string) {
		command := CommandWithStdout("bash", "-c",
			"./node_modules/.bin/ts-node -P test/tsconfig.json -- ./scripts/clean-environment.ts")
		err := command.Run()

		if err != nil {
			panic(err)
		}
	},
	Args: cobra.ExactArgs(0),
}

// SuitesStartCmd Command for starting a suite
var SuitesStartCmd = &cobra.Command{
	Use:   "start [suite]",
	Short: "Start a suite. Suites can be listed using the list command.",
	Run: func(cmd *cobra.Command, args []string) {
		suites, err := listSuites()

		if err != nil {
			panic(err)
		}

		selectedSuite := args[0]

		available, err := suiteAvailable(selectedSuite, suites)

		if err != nil {
			panic(err)
		}

		if !available {
			panic(errors.New("Suite named " + selectedSuite + " does not exist"))
		}

		err = ioutil.WriteFile(RunningSuiteFile, []byte(selectedSuite), 0644)

		if err != nil {
			panic(err)
		}

		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		cmdline := "./node_modules/.bin/ts-node -P test/tsconfig.json -- ./scripts/run-environment.ts " + selectedSuite
		command := CommandWithStdout("bash", "-c", cmdline)
		command.Env = append(os.Environ(), "ENVIRONMENT=dev")

		err = command.Run()

		if err != nil {
			panic(err)
		}

		err = os.Remove(RunningSuiteFile)

		if err != nil {
			panic(err)
		}
	},
	Args: cobra.ExactArgs(1),
}

// SuitesTestCmd Command for testing a suite
var SuitesTestCmd = &cobra.Command{
	Use:   "test [suite]",
	Short: "Test a suite. Suites can be listed using the list command.",
	Run: func(cmd *cobra.Command, args []string) {
		runningSuite, err := getRunningSuite()
		if err != nil {
			panic(err)
		}

		if len(args) == 1 {
			suite := args[0]

			if runningSuite != "" && suite != runningSuite {
				panic(errors.New("Running suite (" + runningSuite + ") is different than suite to be tested (" + suite + "). Shutdown running suite and retry"))
			}

			runSuiteTests(suite, runningSuite == "")
		} else {
			if runningSuite != "" {
				panic(errors.New("Cannot run all tests while a suite is currently running. Shutdown running suite and retry"))
			}
			fmt.Println("No suite provided therefore all suites will be tested")
			runAllSuites()
		}
	},
	Args: cobra.MaximumNArgs(1),
}

func getRunningSuite() (string, error) {
	exist, err := FileExists(RunningSuiteFile)

	if err != nil {
		return "", err
	}

	if !exist {
		return "", nil
	}

	b, err := ioutil.ReadFile(RunningSuiteFile)
	return string(b), err
}

func runSuiteTests(suite string, withEnv bool) {
	mochaArgs := []string{"--exit", "--colors", "--require", "ts-node/register", "test/suites/" + suite + "/test.ts"}
	if onlyForbidden {
		mochaArgs = append(mochaArgs, "--forbid-only", "--forbid-pending")
	}
	mochaCmdLine := "./node_modules/.bin/mocha " + strings.Join(mochaArgs, " ")

	fmt.Println(mochaCmdLine)

	headlessValue := "n"
	if headless {
		headlessValue = "y"
	}

	var cmd *exec.Cmd

	if withEnv {
		fmt.Println("No running suite detected, setting up an environment for running the tests")
		cmd = CommandWithStdout("bash", "-c",
			"./node_modules/.bin/ts-node ./scripts/run-environment.ts "+suite+" '"+mochaCmdLine+"'")
	} else {
		fmt.Println("Running suite detected. Running tests...")
		cmd = CommandWithStdout("bash", "-c", mochaCmdLine)
	}

	cmd.Env = append(os.Environ(),
		"TS_NODE_PROJECT=test/tsconfig.json",
		"HEADLESS="+headlessValue,
		"ENVIRONMENT=dev")

	err := cmd.Run()

	if err != nil {
		os.Exit(1)
	}
}

func runAllSuites() {
	suites, err := listSuites()

	if err != nil {
		panic(err)
	}

	for _, s := range suites {
		runSuiteTests(s, true)
	}
}

var headless bool
var onlyForbidden bool

func init() {
	SuitesTestCmd.Flags().BoolVar(&headless, "headless", false, "Run tests in headless mode")
	SuitesTestCmd.Flags().BoolVar(&onlyForbidden, "only-forbidden", false, "Mocha 'only' filters are forbidden")
}
