package commands

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestX509SystemCertPoolFactory struct {
	pool *x509.CertPool
	err  error
}

func (f *TestX509SystemCertPoolFactory) SystemCertPool() (*x509.CertPool, error) {
	return f.pool, f.err
}

func TestLoadXEnvCLIStringSliceValue(t *testing.T) {
	testCases := []struct {
		name                        string
		envKey, envValue, flagValue string
		flagDefault                 []string
		flag                        *pflag.Flag
		expected                    []string
		expectedResult              XEnvCLIResult
		expectedErr                 string
	}{
		{
			"ShouldParseFromEnv",
			"EXAMPLE_ONE", "abc",
			"example-one", []string{"flagdef"}, &pflag.Flag{Name: "example-one", Changed: false},
			[]string{"abc"}, XEnvCLIResultEnvironment, "",
		},
		{
			"ShouldParseMultipleFromEnv",
			"EXAMPLE_ONE", "abc,123",
			"example-one", []string{"flagdef"}, &pflag.Flag{Name: "example-one", Changed: false},
			[]string{"abc", "123"}, XEnvCLIResultEnvironment, "",
		},
		{
			"ShouldParseCLIExplicit",
			"EXAMPLE_ONE", "abc,123",
			"example-from-flag,123", []string{"flagdef"}, &pflag.Flag{Name: "example-one", Changed: true},
			[]string{"example-from-flag", "123"}, XEnvCLIResultCLIExplicit, "",
		},
		{
			"ShouldParseCLIImplicit",
			"EXAMPLE_ONE", "",
			"example-one", []string{"example-from-flag-default", "123"}, &pflag.Flag{Name: "example-one", Changed: false},
			[]string{"example-from-flag-default", "123"}, XEnvCLIResultCLIImplicit, "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}

			if tc.flag != nil {
				cmd.Flags().StringSlice(tc.flag.Name, tc.flagDefault, "")

				if tc.flag.Changed {
					require.NoError(t, cmd.Flags().Set(tc.flag.Name, tc.flagValue))
				}
			}

			if tc.envValue != "" {
				t.Setenv(tc.envKey, tc.envValue)
			}

			actual, actualResult, actualErr := loadXEnvCLIStringSliceValue(cmd, tc.envKey, tc.flag.Name)

			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.expectedResult, actualResult)

			if tc.expectedErr == "" {
				assert.NoError(t, actualErr)
			} else {
				assert.EqualError(t, actualErr, tc.expectedErr)
			}
		})
	}
}

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
	dyml := filepath.Join(otherdir, "d.yml")

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
		haveX          XEnvCLIResult
		have, expected []string
		expectedErr    string
	}{
		{"ShouldAllowFiles",
			XEnvCLIResultCLIImplicit, []string{ayml},
			[]string{ayml},
			"",
		},
		{"ShouldSkipFilesNotExistImplicit",
			XEnvCLIResultCLIImplicit, []string{dyml},
			[]string(nil),
			"",
		},
		{"ShouldNotErrFilesNotExistExplicit",
			XEnvCLIResultCLIExplicit, []string{dyml},
			[]string{dyml},
			"",
		},
		{"ShouldAllowDirectories",
			XEnvCLIResultCLIImplicit, []string{configdir},
			[]string{configdir},
			"",
		},
		{"ShouldAllowFilesDirectories",
			XEnvCLIResultCLIImplicit, []string{ayml, otherdir},
			[]string{ayml, otherdir},
			"",
		},
		{"ShouldRaiseErrOnOverlappingFilesDirectories",
			XEnvCLIResultCLIImplicit, []string{ayml, configdir},
			nil, fmt.Sprintf("failed to load config directory '%s': the config file '%s' is in that directory which is not supported", configdir, ayml),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, actualErr := loadXNormalizedPaths(tc.have, tc.haveX)

			assert.Equal(t, tc.expected, actual)

			if tc.expectedErr == "" {
				assert.NoError(t, actualErr)
			} else {
				assert.EqualError(t, actualErr, tc.expectedErr)
			}
		})
	}
}
