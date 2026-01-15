package utils

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShouldExecCommandOnAutheliaRootPath(t *testing.T) {
	cmd := Command("pwd")
	result, err := cmd.CombinedOutput()
	assert.NoError(t, err, "")

	str := strings.Trim(string(result), "\n")

	assert.NoError(t, err, "")
	// Check that the working directory ends with either "authelia" or "user-management".
	assert.True(t, strings.HasSuffix(str, "authelia") || strings.HasSuffix(str, "user-management"),
		"expected directory to end with 'authelia' or 'user-management', got: %s", str)
}

func TestCommandShouldOutputResult(t *testing.T) {
	output, exitcode, err := RunCommandAndReturnOutput("echo hello")

	assert.NoError(t, err)
	assert.Equal(t, 0, exitcode)
	assert.Equal(t, "hello", output)
}

func TestShouldWaitUntilCommandEnds(t *testing.T) {
	cmd := Command("sleep", "2")

	err := RunCommandWithTimeout(cmd, 3*time.Second)
	assert.NoError(t, err, "")
}

func TestShouldTimeoutWaitingCommand(t *testing.T) {
	cmd := Command("sleep", "3")

	err := RunCommandWithTimeout(cmd, 2*time.Second)
	assert.Error(t, err)
}

func TestShouldRunFuncUntilNoError(t *testing.T) {
	counter := 0

	err := RunFuncWithRetry(3, 500*time.Millisecond, func() error {
		counter++
		if counter < 3 {
			return errors.New("not ready")
		}

		return nil
	})
	assert.NoError(t, err, "")
}

func TestShouldFailAfterMaxAttemps(t *testing.T) {
	counter := 0

	err := RunFuncWithRetry(3, 500*time.Millisecond, func() error {
		counter++
		if counter < 4 {
			return errors.New("not ready")
		}

		return nil
	})
	assert.ErrorContains(t, err, "not ready")
}
