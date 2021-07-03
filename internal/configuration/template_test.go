package configuration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/internal/utils"
)

func TestShouldGenerateConfiguration(t *testing.T) {
	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	cfg := filepath.Join(dir, "config.yml")

	created, err := EnsureConfigurationExists([]string{cfg})
	assert.NoError(t, err)
	assert.True(t, created)

	_, err = os.Stat(cfg)
	assert.NoError(t, err)
}

func TestShouldNotGenerateConfigurationOnFSAccessDenied(t *testing.T) {
	if runtime.GOOS == constWindows {
		t.Skip("skipping test due to being on windows")
	}

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	assert.NoError(t, os.Mkdir(filepath.Join(dir, "zero"), 0000))

	cfg := filepath.Join(dir, "zero", "config.yml")

	created, err := EnsureConfigurationExists([]string{cfg})
	assert.EqualError(t, err, fmt.Sprintf("error occurred generating Configuration: stat %s: permission denied", cfg))
	assert.False(t, created)
}

func TestShouldNotGenerateMultipleConfigurations(t *testing.T) {
	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	cfgOne := filepath.Join(dir, "config.yml")
	cfgTwo := filepath.Join(dir, "config-acl.yml")

	created, err := EnsureConfigurationExists([]string{cfgOne, cfgTwo})
	assert.NoError(t, err)
	assert.False(t, created)

	_, err = os.Stat(cfgOne)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(cfgTwo)
	assert.True(t, os.IsNotExist(err))
}

func TestShouldNotGenerateConfiguration(t *testing.T) {
	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	cfg := filepath.Join(dir, "..", "not-a-dir", "config.yml")

	created, err := EnsureConfigurationExists([]string{cfg})

	expectedErr := fmt.Sprintf(utils.GetExpectedErrTxt("pathnotfound"), cfg)

	assert.EqualError(t, err, fmt.Sprintf(errFmtGenerateConfiguration, expectedErr))
	assert.False(t, created)
}
