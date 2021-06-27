package configuration

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultSourcesWithoutFiles(t *testing.T) {
	sources := NewDefaultSources([]string{}, nil)

	require.Len(t, sources, 2)

	assert.Equal(t, "environment", sources[0].Name)
	assert.Equal(t, "secrets", sources[1].Name)
}

func TestNewDefaultSourcesWithFiles(t *testing.T) {
	config := filepath.FromSlash("./test_resources/config.yml")
	configalt := filepath.FromSlash("./test_resources/config_alt.yml")

	sources := NewDefaultSources([]string{config, configalt}, nil)

	require.Len(t, sources, 4)

	assert.Equal(t, fmt.Sprintf("file:%s", config), sources[0].Name)
	assert.NotNil(t, sources[0].Parser)

	assert.Equal(t, fmt.Sprintf("file:%s", configalt), sources[1].Name)
	assert.NotNil(t, sources[1].Parser)

	assert.Equal(t, "environment", sources[2].Name)
	assert.Nil(t, sources[2].Parser)

	assert.Equal(t, "secrets", sources[3].Name)
	assert.Nil(t, sources[3].Parser)
}
