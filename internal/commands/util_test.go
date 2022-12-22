package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadXNormalizedPaths(t *testing.T) {
	root := t.TempDir()

	configdir := filepath.Join(root, "config")
	otherdir := filepath.Join(root, "other")

	require.NoError(t, os.Mkdir(configdir, 0700))
	require.NoError(t, os.Mkdir(otherdir, 0700))

	var (
		info os.FileInfo
		file *os.File
		err  error
	)

	ayml := filepath.Join(configdir, "a.yml")
	byml := filepath.Join(configdir, "b.yml")
	cyml := filepath.Join(otherdir, "c.yml")

	file, err = os.Create(ayml)

	require.NoError(t, err)

	require.NoError(t, file.Close())

	file, err = os.Create(byml)

	require.NoError(t, err)

	require.NoError(t, file.Close())

	file, err = os.Create(cyml)

	require.NoError(t, err)

	require.NoError(t, file.Close())

	info, err = os.Stat(configdir)

	require.NoError(t, err)
	require.True(t, info.IsDir())

	info, err = os.Stat(otherdir)

	require.NoError(t, err)
	require.True(t, info.IsDir())

	info, err = os.Stat(ayml)

	require.NoError(t, err)
	require.False(t, info.IsDir())

	info, err = os.Stat(byml)

	require.NoError(t, err)
	require.False(t, info.IsDir())

	info, err = os.Stat(cyml)

	require.NoError(t, err)
	require.False(t, info.IsDir())

	testCases := []struct {
		name           string
		have, expected []string
		expectedErr    string
	}{
		{"ShouldAllowFiles",
			[]string{ayml},
			[]string{ayml}, "",
		},
		{"ShouldAllowDirectories",
			[]string{configdir},
			[]string{configdir}, "",
		},
		{"ShouldAllowFilesDirectories",
			[]string{ayml, otherdir},
			[]string{ayml, otherdir}, "",
		},
		{"ShouldRaiseErrOnOverlappingFilesDirectories",
			[]string{ayml, configdir},
			nil, fmt.Sprintf("failed to load config directory '%s': the config file '%s' is in that directory which is not supported", configdir, ayml),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, actualErr := loadXNormalizedPaths(tc.have)

			assert.Equal(t, tc.expected, actual)

			if tc.expectedErr == "" {
				assert.NoError(t, actualErr)
			} else {
				assert.EqualError(t, actualErr, tc.expectedErr)
			}
		})
	}
}
