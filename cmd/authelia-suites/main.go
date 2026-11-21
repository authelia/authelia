// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

// Command authelia-suites manages the environment of an integration suite.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/suites"
	"github.com/authelia/authelia/v4/internal/utils"
)

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

func loadSuiteEnvironment(path string) (err error) {
	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("error checking environment file '%s': %w", path, err)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening environment file '%s': %w", path, err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		v := strings.SplitN(line, "=", 2)
		if len(v) != 2 {
			return fmt.Errorf("error parsing environment file '%s': line '%s' is not a comment and does not contain a '=' character", path, line)
		}

		if err = os.Setenv(v[0], v[1]); err != nil {
			return fmt.Errorf("error setting environment variable '%s' from environment file '%s': %w", v[0], path, err)
		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("error reading environment file '%s': %w", path, err)
	}

	return nil
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

	if err = loadSuiteEnvironment(suiteResourcePath + "/.env"); err != nil {
		log.Fatal(err)
	}

	suiteTmpDirectory := suites.SuiteTmpPath("authelia", "suites", suiteName)

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

	if err := createRunningSuiteFile(suiteName); err != nil {
		log.Fatal(err)
	}

	if err = s.SetUp(suiteTmpDirectory); err != nil {
		log.Error("Failure during environment deployment.")

		if s.OnError != nil {
			if errLogs := s.OnError(); errLogs != nil {
				log.Errorf("Error collecting suite logs: %v", errLogs)
			}
		}

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

	suiteTmpDirectory := suites.SuiteTmpPath("authelia", "suites", args[0])
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
