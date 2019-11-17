package suites

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/clems4ever/authelia/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tebeka/selenium"
)

// SeleniumSuite is a selenium suite
type SeleniumSuite struct {
	suite.Suite

	*WebDriverSession
}

// WebDriver return the webdriver of the suite
func (s *SeleniumSuite) WebDriver() selenium.WebDriver {
	return s.WebDriverSession.WebDriver
}

// Wait wait until condition holds true
func (s *SeleniumSuite) Wait(ctx context.Context, condition selenium.Condition) error {
	done := make(chan error, 1)
	go func() {
		done <- s.WebDriverSession.WebDriver.Wait(condition)
	}()

	select {
	case <-ctx.Done():
		return errors.New("waiting timeout reached")
	case err := <-done:
		return err
	}
}

func rootPath() string {
	rootPath := os.Getenv("ROOT_PATH")

	// If env variable is not provided, use relative path.
	if rootPath == "" {
		rootPath = "../.."
	}
	return rootPath
}

func relativePath(path string) string {
	return fmt.Sprintf("%s/%s", rootPath(), path)
}

// RunTypescriptSuite run the tests of the typescript suite
func RunTypescriptSuite(t *testing.T, suite string) {
	forbidFlags := ""
	if os.Getenv("ONLY_FORBIDDEN") == "true" {
		forbidFlags = "--forbid-only --forbid-pending"
	}

	cmdline := "./node_modules/.bin/mocha" +
		" --exit --require ts-node/register " + forbidFlags + " " +
		fmt.Sprintf("test/suites/%s/test.ts", suite)

	command := utils.CommandWithStdout("bash", "-c", cmdline)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Dir = rootPath()
	command.Env = append(
		os.Environ(),
		"ENVIRONMENT=dev",
		fmt.Sprintf("TS_NODE_PROJECT=%s", "test/tsconfig.json"))

	assert.NoError(t, command.Run())
}
