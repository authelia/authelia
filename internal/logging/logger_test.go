package logging

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestShouldWriteLogsToFile(t *testing.T) {
	dir := t.TempDir()

	path := fmt.Sprintf("%s/authelia.log", dir)
	err := InitializeLogger(schema.Log{Format: "text", FilePath: path, KeepStdout: false}, false)
	require.NoError(t, err)

	Logger().Info("This is a test")

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	require.NoError(t, err)

	b, err := io.ReadAll(f)
	require.NoError(t, err)

	assert.Contains(t, string(b), "level=info msg=\"This is a test\"\n")
}

func TestShouldWriteLogsToFileAndStdout(t *testing.T) {
	dir := t.TempDir()

	path := fmt.Sprintf("%s/authelia.log", dir)
	err := InitializeLogger(schema.Log{Format: "text", FilePath: path, KeepStdout: true}, false)
	require.NoError(t, err)

	Logger().Info("This is a test")

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	require.NoError(t, err)

	b, err := io.ReadAll(f)
	require.NoError(t, err)

	assert.Contains(t, string(b), "level=info msg=\"This is a test\"\n")
}

func TestShouldFormatLogsAsJSON(t *testing.T) {
	dir := t.TempDir()

	path := fmt.Sprintf("%s/authelia.log", dir)
	err := InitializeLogger(schema.Log{Format: "json", FilePath: path, KeepStdout: false}, false)
	require.NoError(t, err)

	Logger().Info("This is a test")

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	require.NoError(t, err)

	b, err := io.ReadAll(f)
	require.NoError(t, err)

	assert.Contains(t, string(b), "{\"level\":\"info\",\"msg\":\"This is a test\",")
}

func TestShouldRaiseErrorOnInvalidFile(t *testing.T) {
	err := InitializeLogger(schema.Log{FilePath: "/not/a/valid/path/to.log"}, false)

	switch runtime.GOOS {
	case "windows":
		assert.EqualError(t, err, "open /not/a/valid/path/to.log: The system cannot find the path specified.")
	default:
		assert.EqualError(t, err, "open /not/a/valid/path/to.log: no such file or directory")
	}
}

func TestSetLevels(t *testing.T) {
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())

	setLevelStr("error", false)
	assert.Equal(t, logrus.ErrorLevel, logrus.GetLevel())

	setLevelStr("warn", false)
	assert.Equal(t, logrus.WarnLevel, logrus.GetLevel())

	setLevelStr("info", false)
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())

	setLevelStr("debug", false)
	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())

	setLevelStr("trace", false)
	assert.Equal(t, logrus.TraceLevel, logrus.GetLevel())
}
