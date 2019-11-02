package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/clems4ever/authelia/suites"
	"github.com/clems4ever/authelia/utils"
	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var tmpDirectory = "/tmp/authelia/suites/"

// runningSuiteFile name of the file containing the currently running suite
var runningSuiteFile = ".suite"

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	rootCmd := &cobra.Command{
		Use: "authelia-suites",
	}

	startCmd := &cobra.Command{
		Use:   "setup [suite]",
		Short: "Setup the suite environment",
		Run:   setupSuite,
	}

	stopCmd := &cobra.Command{
		Use:   "teardown [suite]",
		Short: "Teardown the suite environment",
		Run:   teardownSuite,
	}

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.Execute()
}

func createRunningSuiteFile(suite string) error {
	return ioutil.WriteFile(runningSuiteFile, []byte(suite), 0644)
}

func removeRunningSuiteFile() error {
	return os.Remove(runningSuiteFile)
}

func setupSuite(cmd *cobra.Command, args []string) {
	suiteName := args[0]
	s := suites.GlobalRegistry.Get(suiteName)

	cwd, err := filepath.Abs("./")

	if err != nil {
		log.Fatal(err)
	}

	suiteResourcePath := cwd + "/suites/" + suiteName

	exist, err := utils.FileExists(suiteResourcePath)

	if err != nil {
		log.Fatal(err)
	}

	suiteTmpDirectory := tmpDirectory + suiteName

	if exist {
		err := copy.Copy(suiteResourcePath, suiteTmpDirectory)

		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := os.MkdirAll(suiteTmpDirectory, 0755)

		if err != nil {
			log.Fatal(err)
		}
	}

	// Create the .suite file
	if err := createRunningSuiteFile(suiteName); err != nil {
		log.Fatal(err)
	}

	err = s.SetUp(suiteTmpDirectory)

	if err != nil {
		log.Error("Failure during environment deployment.")
		teardownSuite(nil, args)
		log.Fatal(err)
	}

	log.Info("Environment is ready!")
}

func teardownSuite(cmd *cobra.Command, args []string) {
	if os.Getenv("SKIP_TEARDOWN") != "" {
		return
	}

	s := suites.GlobalRegistry.Get(args[0])

	if err := removeRunningSuiteFile(); err != nil {
		log.Print(err)
	}

	suiteTmpDirectory := tmpDirectory + args[0]
	s.TearDown(suiteTmpDirectory)

	err := os.RemoveAll(suiteTmpDirectory)

	if err != nil {
		log.Fatal(err)
	}

	log.Info("Environment has been cleaned!")
}
