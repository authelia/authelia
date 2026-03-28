package commands

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/random"
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

func TestFlagsGetUserIdentifiersGenerateOptions(t *testing.T) {
	testCases := []struct {
		name             string
		setup            func(flags *pflag.FlagSet)
		expectedUsers    []string
		expectedServices []string
		expectedSectors  []string
		err              string
	}{
		{
			"ShouldReturnDefaults",
			func(flags *pflag.FlagSet) {
				flags.StringSlice(cmdFlagNameUsers, nil, "")
				flags.StringSlice(cmdFlagNameServices, nil, "")
				flags.StringSlice(cmdFlagNameSectors, nil, "")
			},
			[]string{},
			[]string{},
			[]string{},
			"",
		},
		{
			"ShouldReturnValues",
			func(flags *pflag.FlagSet) {
				flags.StringSlice(cmdFlagNameUsers, []string{"john", "harry"}, "")
				flags.StringSlice(cmdFlagNameServices, []string{"openid"}, "")
				flags.StringSlice(cmdFlagNameSectors, []string{"example.com"}, "")
			},
			[]string{"john", "harry"},
			[]string{"openid"},
			[]string{"example.com"},
			"",
		},
		{
			"ShouldErrUsersWrongType",
			func(flags *pflag.FlagSet) {
				flags.Bool(cmdFlagNameUsers, false, "")
				flags.StringSlice(cmdFlagNameServices, nil, "")
				flags.StringSlice(cmdFlagNameSectors, nil, "")
			},
			nil,
			nil,
			nil,
			"trying to get stringSlice value of flag of type bool",
		},
		{
			"ShouldErrServicesWrongType",
			func(flags *pflag.FlagSet) {
				flags.StringSlice(cmdFlagNameUsers, nil, "")
				flags.Bool(cmdFlagNameServices, false, "")
				flags.StringSlice(cmdFlagNameSectors, nil, "")
			},
			nil,
			nil,
			nil,
			"trying to get stringSlice value of flag of type bool",
		},
		{
			"ShouldErrSectorsWrongType",
			func(flags *pflag.FlagSet) {
				flags.StringSlice(cmdFlagNameUsers, nil, "")
				flags.StringSlice(cmdFlagNameServices, nil, "")
				flags.Bool(cmdFlagNameSectors, false, "")
			},
			nil,
			nil,
			nil,
			"trying to get stringSlice value of flag of type bool",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

			tc.setup(flags)

			users, services, sectors, err := flagsGetUserIdentifiersGenerateOptions(flags)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUsers, users)
				assert.Equal(t, tc.expectedServices, services)
				assert.Equal(t, tc.expectedSectors, sectors)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestFlagsGetRandomCharacters(t *testing.T) {
	testCases := []struct {
		name      string
		length    int
		charset   string
		chars     string
		setCharS  bool
		setChars  bool
		err       string
		resultLen int
		resultSet string
	}{
		{
			"ShouldReturnASCII",
			10, "ascii", "", true, false,
			"", 10, random.CharSetASCII,
		},
		{
			"ShouldReturnAlphanumeric",
			8, "alphanumeric", "", true, false,
			"", 8, random.CharSetAlphaNumeric,
		},
		{
			"ShouldReturnAlphanumericLower",
			5, "alphanumeric-lower", "", true, false,
			"", 5, random.CharSetAlphabeticLower + random.CharSetNumeric,
		},
		{
			"ShouldReturnAlphanumericUpper",
			5, "alphanumeric-upper", "", true, false,
			"", 5, random.CharSetAlphabeticUpper + random.CharSetNumeric,
		},
		{
			"ShouldReturnAlphabetic",
			6, "alphabetic", "", true, false,
			"", 6, random.CharSetAlphabetic,
		},
		{
			"ShouldReturnAlphabeticLower",
			6, "alphabetic-lower", "", true, false,
			"", 6, random.CharSetAlphabeticLower,
		},
		{
			"ShouldReturnAlphabeticUpper",
			6, "alphabetic-upper", "", true, false,
			"", 6, random.CharSetAlphabeticUpper,
		},
		{
			"ShouldReturnNumericHex",
			12, "numeric-hex", "", true, false,
			"", 12, random.CharSetNumericHex,
		},
		{
			"ShouldReturnNumeric",
			4, "numeric", "", true, false,
			"", 4, random.CharSetNumeric,
		},
		{
			"ShouldReturnRFC3986",
			10, "rfc3986", "", true, false,
			"", 10, random.CharSetRFC3986Unreserved,
		},
		{
			"ShouldReturnRFC3986Lower",
			10, "rfc3986-lower", "", true, false,
			"", 10, random.CharSetAlphabeticLower + random.CharSetNumeric + random.CharSetSymbolicRFC3986Unreserved,
		},
		{
			"ShouldReturnRFC3986Upper",
			10, "rfc3986-upper", "", true, false,
			"", 10, random.CharSetAlphabeticUpper + random.CharSetNumeric + random.CharSetSymbolicRFC3986Unreserved,
		},
		{
			"ShouldReturnCustomCharacters",
			10, "", "abc123", false, true,
			"", 10, "",
		},
		{
			"ShouldReturnDefaultCharsetWhenNeitherSet",
			10, "alphanumeric", "", false, false,
			"", 10, random.CharSetAlphaNumeric,
		},
		{
			"ShouldErrInvalidCharSet",
			10, "invalid", "", true, false,
			"flag '--charset' with value 'invalid' is invalid", 0, "",
		},
		{
			"ShouldErrZeroLength",
			0, "ascii", "", false, false,
			"flag --length with value '0' is invalid: must be at least 1", 0, "",
		},
		{
			"ShouldErrNegativeLength",
			-1, "ascii", "", false, false,
			"flag --length with value '-1' is invalid: must be at least 1", 0, "",
		},
		{
			"ShouldErrMutuallyExclusive",
			10, "ascii", "abc", true, true,
			"flag --charset and flag --characters are mutually exclusive", 0, "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.Int("length", tc.length, "")
			flags.String("charset", tc.charset, "")
			flags.String("characters", tc.chars, "")

			if tc.setCharS {
				require.NoError(t, flags.Set("charset", tc.charset))
			}

			if tc.setChars {
				require.NoError(t, flags.Set("characters", tc.chars))
			}

			result, err := flagsGetRandomCharacters(flags, "length", "charset", "characters")

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Len(t, result, tc.resultLen)

				if tc.resultSet != "" {
					for _, c := range result {
						assert.Contains(t, tc.resultSet, string(c))
					}
				}
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestFlagParseFileMode(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		expected os.FileMode
		err      string
	}{
		{
			"ShouldParse0600",
			"0600",
			os.FileMode(0600),
			"",
		},
		{
			"ShouldParse0644",
			"0644",
			os.FileMode(0644),
			"",
		},
		{
			"ShouldParse0755",
			"0755",
			os.FileMode(0755),
			"",
		},
		{
			"ShouldParse0700",
			"0700",
			os.FileMode(0700),
			"",
		},
		{
			"ShouldErrInvalidOctal",
			"999",
			os.FileMode(0),
			"strconv.ParseUint: parsing \"999\": invalid syntax",
		},
		{
			"ShouldErrNotANumber",
			"abc",
			os.FileMode(0),
			"strconv.ParseUint: parsing \"abc\": invalid syntax",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.String("mode", tc.value, "")

			mode, err := flagParseFileMode("mode", flags)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, mode)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestTermReadPasswordWithPrompt(t *testing.T) {
	testCases := []struct {
		name string
		flag string
		err  string
	}{
		{
			"ShouldErrNotTerminalNoFlag",
			"",
			"stdin is not a terminal",
		},
		{
			"ShouldErrNotTerminalSingleCharFlag",
			"p",
			"you must either use an interactive terminal or use the -p flag",
		},
		{
			"ShouldErrNotTerminalMultiCharFlag",
			"password",
			"you must either use an interactive terminal or use the --password flag",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			password, err := termReadPasswordWithPrompt("Enter: ", tc.flag)

			assert.Empty(t, password)
			assert.EqualError(t, err, tc.err)
		})
	}
}

func TestGetTerminal(t *testing.T) {
	// In a test environment, stdin is not a terminal.
	terminal, fd, state, err := getTerminal("prompt> ")

	assert.Nil(t, terminal)
	assert.Equal(t, -1, fd)
	assert.Nil(t, state)
	assert.ErrorIs(t, err, ErrStdinIsNotTerminal)
}

func TestTermReadConfirmation(t *testing.T) {
	// In a test environment, stdin is not a terminal so getTerminal returns an error.
	confirmed, err := termReadConfirmation("Confirm: ", "YES")

	assert.False(t, confirmed)
	assert.Error(t, err)
}

func TestExportYAMLWithJSONSchema(t *testing.T) {
	testCases := []struct {
		name     string
		schemaID string
		data     any
		expected []string
	}{
		{
			"ShouldExportSimpleStruct",
			"export.test",
			map[string]string{"key": "value"},
			[]string{"yaml-language-server", "export.test", "key: value"},
		},
		{
			"ShouldExportSlice",
			"export.items",
			[]string{"a", "b"},
			[]string{"yaml-language-server", "export.items", "- a", "- b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			err := exportYAMLWithJSONSchema(buf, tc.schemaID, tc.data)

			assert.NoError(t, err)

			for _, s := range tc.expected {
				assert.Contains(t, buf.String(), s)
			}
		})
	}
}

func TestGetCryptoHashGenerateMapFlagsFromUse(t *testing.T) {
	testCases := []struct {
		name         string
		use          string
		expectedKeys []string
		expectedNil  bool
	}{
		{
			"ShouldReturnArgon2Flags",
			cmdUseHashArgon2,
			[]string{cmdFlagNameVariant, cmdFlagNameIterations, cmdFlagNameMemory, cmdFlagNameParallelism, cmdFlagNameKeySize, cmdFlagNameSaltSize},
			false,
		},
		{
			"ShouldReturnSHA2CryptFlags",
			cmdUseHashSHA2Crypt,
			[]string{cmdFlagNameVariant, cmdFlagNameIterations, cmdFlagNameSaltSize},
			false,
		},
		{
			"ShouldReturnPBKDF2Flags",
			cmdUseHashPBKDF2,
			[]string{cmdFlagNameVariant, cmdFlagNameIterations, cmdFlagNameSaltSize},
			false,
		},
		{
			"ShouldReturnBcryptFlags",
			cmdUseHashBcrypt,
			[]string{cmdFlagNameVariant, cmdFlagNameCost},
			false,
		},
		{
			"ShouldReturnScryptFlags",
			cmdUseHashScrypt,
			[]string{cmdFlagNameVariant, cmdFlagNameIterations, cmdFlagNameBlockSize, cmdFlagNameParallelism, cmdFlagNameKeySize, cmdFlagNameSaltSize},
			false,
		},
		{
			"ShouldReturnNilForUnknown",
			"unknown",
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getCryptoHashGenerateMapFlagsFromUse(tc.use)

			if tc.expectedNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)

				for _, key := range tc.expectedKeys {
					assert.Contains(t, result, key)
				}

				assert.Len(t, result, len(tc.expectedKeys))
			}
		})
	}
}

func TestNewHelpTopic(t *testing.T) {
	testCases := []struct {
		name  string
		topic string
		short string
		body  string
	}{
		{
			"ShouldCreateHelpTopic",
			"test-topic",
			"A test topic",
			"This is the body of the test topic.",
		},
		{
			"ShouldCreateAnotherHelpTopic",
			"another-topic",
			"Another topic",
			"Another body.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newHelpTopic(tc.topic, tc.short, tc.body)

			assert.Equal(t, tc.topic, cmd.Use)
			assert.Equal(t, tc.short, cmd.Short)
		})
	}
}

func TestCmdHelpTopic(t *testing.T) {
	parent := &cobra.Command{Use: "parent"}
	child := &cobra.Command{Use: "child"}

	parent.AddCommand(child)

	buf := new(bytes.Buffer)

	child.SetOut(buf)
	parent.SetOut(buf)

	cmdHelpTopic(child, "This is help body text.", "child")

	output := buf.String()

	assert.Contains(t, output, "Help Topic: child")
	assert.Contains(t, output, "This is help body text.")
}

func TestLoadXEnvCLIConfigValues(t *testing.T) {
	testCases := []struct {
		name string
		env  map[string]string
		err  string
	}{
		{
			"ShouldSucceedDefaults",
			nil,
			"",
		},
		{
			"ShouldErrInvalidFilter",
			map[string]string{cmdFlagEnvNameConfigFilters: "invalidfilter"},
			"error occurred loading configuration: flag '--config.experimental.filters' is invalid:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().StringSlice(cmdFlagNameConfig, []string{}, "")
			cmd.Flags().StringSlice(cmdFlagNameConfigExpFilters, nil, "")

			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			configs, filters, err := loadXEnvCLIConfigValues(cmd)

			if tc.err == "" {
				assert.NoError(t, err)

				_ = configs
				_ = filters
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestLoadXEnvCLIConfigValuesWithConfigFiles(t *testing.T) {
	dir := t.TempDir()

	configFile := filepath.Join(dir, "config.yml")

	require.NoError(t, os.WriteFile(configFile, []byte("---\n"), 0600))

	cmd := &cobra.Command{}
	cmd.Flags().StringSlice(cmdFlagNameConfig, nil, "")
	cmd.Flags().StringSlice(cmdFlagNameConfigExpFilters, nil, "")

	require.NoError(t, cmd.Flags().Set(cmdFlagNameConfig, configFile))

	configs, filters, err := loadXEnvCLIConfigValues(cmd)

	assert.NoError(t, err)
	assert.Len(t, configs, 1)
	assert.Contains(t, configs[0], "config.yml")
	assert.NotNil(t, filters)
}

func TestFlagParseFileModeWrongType(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.Bool("mode", false, "")

	_, err := flagParseFileMode("mode", flags)

	assert.ErrorContains(t, err, "trying to get string value of flag of type bool")
}

func TestFlagsGetRandomCharactersLengthWrongType(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("length", "10", "")
	flags.String("charset", "ascii", "")
	flags.String("characters", "", "")

	_, err := flagsGetRandomCharacters(flags, "length", "charset", "characters")

	assert.ErrorContains(t, err, "trying to get int value of flag of type string")
}
