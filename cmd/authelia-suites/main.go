package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/suites"
	"github.com/authelia/authelia/v4/internal/utils"
)

var tmpDirectory = "/tmp/authelia/suites/"

// runningSuiteFile name of the file containing the currently running suite.
var runningSuiteFile = ".suite"

func init() {
	log.SetLevel(log.InfoLevel)
}

func main() {
	rootCmd := &cobra.Command{
		Use: "authelia-suites",

		DisableAutoGenTag: true,
	}

	startCmd := &cobra.Command{
		Use:   "setup [suite]",
		Short: "Setup the suite environment",
		Run:   setupSuite,

		DisableAutoGenTag: true,
	}

	setupTimeoutCmd := &cobra.Command{
		Use:   "timeout [suite]",
		Short: "Run the OnSetupTimeout callback when setup times out",
		Run:   setupTimeoutSuite,

		DisableAutoGenTag: true,
	}

	errorCmd := &cobra.Command{
		Use:   "error [suite]",
		Short: "Run the OnError callback when some tests fail",
		Run:   runErrorCallback,

		DisableAutoGenTag: true,
	}

	stopCmd := &cobra.Command{
		Use:   "teardown [suite]",
		Short: "Teardown the suite environment",
		Run:   teardownSuite,

		DisableAutoGenTag: true,
	}

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(setupTimeoutCmd)
	rootCmd.AddCommand(errorCmd)
	rootCmd.AddCommand(stopCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func createRunningSuiteFile(suite string) error {
	return os.WriteFile(runningSuiteFile, []byte(suite), 0600)
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

	suiteResourcePath := cwd + "/internal/suites/" + suiteName

	exist, err := utils.PathExists(suiteResourcePath)

	if err != nil {
		log.Fatal(err)
	}

	suiteEnv := suiteResourcePath + "/.env"

	_, err = os.Stat(suiteEnv)
	if err == nil {
		file, err := os.Open(suiteEnv)
		if err != nil {
			log.Fatal(err)
		}

		env := bufio.NewScanner(file)

		for env.Scan() {
			v := strings.Split(env.Text(), "=")

			err := os.Setenv(v[0], v[1])
			if err != nil {
				log.Fatal(err)
			}
		}
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

	// Create the .suite file.
	if err := createRunningSuiteFile(suiteName); err != nil {
		log.Fatal(err)
	}

	if err = s.SetUp(suiteTmpDirectory); err != nil {
		log.Error("Failure during environment deployment.")
		teardownSuite(nil, args)
		log.Fatal(err)
	}

	log.Info("Environment is ready!")
}

func setupTimeoutSuite(cmd *cobra.Command, args []string) {
	suiteName := args[0]
	s := suites.GlobalRegistry.Get(suiteName)

	if s.OnSetupTimeout == nil {
		return
	}

	if err := s.OnSetupTimeout(); err != nil {
		log.Fatal(err)
	}
}

func runErrorCallback(cmd *cobra.Command, args []string) {
	suiteName := args[0]
	s := suites.GlobalRegistry.Get(suiteName)

	if s.OnError == nil {
		return
	}

	if err := s.OnError(); err != nil {
		log.Fatal(err)
	}
}

func teardownSuite(cmd *cobra.Command, args []string) {
	if os.Getenv("SKIP_TEARDOWN") != "" {
		return
	}

	s := suites.GlobalRegistry.Get(args[0])

	suiteTmpDirectory := tmpDirectory + args[0]
	if err := s.TearDown(suiteTmpDirectory); err != nil {
		log.Fatal(err)
	}

	if err := os.RemoveAll(suiteTmpDirectory); err != nil {
		log.Fatal(err)
	}

	if err := removeRunningSuiteFile(); err != nil {
		log.Print(err)
	}

	log.Info("Environment has been cleaned!")
}
