package commands

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewCryptoHashCmd(t *testing.T) {
	cmd := newCryptoHashCmd(&CmdCtx{})
	assert.NotNil(t, cmd)
	assert.Equal(t, cmdUseHash, cmd.Use)
}

func TestNewCryptoHashDefaults(t *testing.T) {
	defaults := newCryptoHashDefaults()

	assert.NotEmpty(t, defaults)
	assert.Contains(t, defaults, prefixFilePassword+suffixAlgorithm)
	assert.Contains(t, defaults, prefixFilePassword+suffixArgon2Variant)
	assert.Contains(t, defaults, prefixFilePassword+suffixBcryptVariant)
	assert.Contains(t, defaults, prefixFilePassword+suffixScryptVariant)
	assert.Contains(t, defaults, prefixFilePassword+suffixPBKDF2Variant)
	assert.Contains(t, defaults, prefixFilePassword+suffixSHA2CryptVariant)
}

func TestRunCryptoHashGenerate(t *testing.T) {
	testCases := []struct {
		name     string
		use      string
		config   *schema.Configuration
		password string
		err      string
		expected string
	}{
		{
			"ShouldSucceedArgon2Default",
			cmdUseGenerate,
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{
						Password: schema.DefaultPasswordConfig,
					},
				},
			},
			"password123",
			"",
			"Digest: $argon2id$",
		},
		{
			"ShouldSucceedArgon2SubCmd",
			cmdUseHashArgon2,
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{
						Password: schema.DefaultPasswordConfig,
					},
				},
			},
			"password123",
			"",
			"Digest: $argon2id$",
		},
		{
			"ShouldSucceedSHA2Crypt",
			cmdUseHashSHA2Crypt,
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{
						Password: schema.DefaultPasswordConfig,
					},
				},
			},
			"password123",
			"",
			"Digest: $",
		},
		{
			"ShouldSucceedBcrypt",
			cmdUseHashBcrypt,
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{
						Password: schema.DefaultPasswordConfig,
					},
				},
			},
			"password123",
			"",
			"Digest: $2b$",
		},
		{
			"ShouldSucceedPBKDF2",
			cmdUseHashPBKDF2,
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{
						Password: schema.DefaultPasswordConfig,
					},
				},
			},
			"password123",
			"",
			"Digest: $pbkdf2",
		},
		{
			"ShouldSucceedScrypt",
			cmdUseHashScrypt,
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{
						Password: schema.DefaultPasswordConfig,
					},
				},
			},
			"password123",
			"",
			"Digest: $",
		},
		{
			"ShouldErrNoFileConfig",
			cmdUseGenerate,
			&schema.Configuration{},
			"password123",
			"authentication backend file is not configured",
			"",
		},
		{
			"ShouldErrNoPasswordTerminal",
			cmdUseGenerate,
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{
						Password: schema.DefaultPasswordConfig,
					},
				},
			},
			"",
			"you must either use an interactive terminal or use the --password flag",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.String(cmdFlagNamePassword, tc.password, "")
			flags.Bool(cmdFlagNameNoConfirm, true, "")
			flags.Bool(cmdFlagNameRandom, false, "")
			flags.String(cmdFlagNameRandomCharSet, cmdFlagValueCharSet, "")
			flags.String(cmdFlagNameRandomCharacters, "", "")
			flags.Int(cmdFlagNameRandomLength, 72, "")

			if tc.password != "" {
				require.NoError(t, flags.Set(cmdFlagNamePassword, tc.password))
			}

			buf := new(bytes.Buffer)

			err := runCryptoHashGenerate(buf, flags, tc.use, nil, tc.config)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestRunCryptoHashGenerateRandom(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String(cmdFlagNamePassword, "", "")
	flags.Bool(cmdFlagNameNoConfirm, true, "")
	flags.Bool(cmdFlagNameRandom, true, "")
	flags.String(cmdFlagNameRandomCharSet, cmdFlagValueCharSet, "")
	flags.String(cmdFlagNameRandomCharacters, "", "")
	flags.Int(cmdFlagNameRandomLength, 72, "")

	require.NoError(t, flags.Set(cmdFlagNameRandom, "true"))

	config := &schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			File: &schema.AuthenticationBackendFile{
				Password: schema.DefaultPasswordConfig,
			},
		},
	}

	buf := new(bytes.Buffer)

	err := runCryptoHashGenerate(buf, flags, cmdUseGenerate, nil, config)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Random Password:")
	assert.Contains(t, buf.String(), "Digest:")
}

func TestCmdFlagsCryptoHashPasswordRandom(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(flags *pflag.FlagSet)
		expected bool
	}{
		{
			"ShouldReturnFalseByDefault",
			func(flags *pflag.FlagSet) {},
			false,
		},
		{
			"ShouldReturnTrueWhenRandomSet",
			func(flags *pflag.FlagSet) {
				require.NoError(nil, flags.Set(cmdFlagNameRandom, "true"))
			},
			true,
		},
		{
			"ShouldReturnTrueWhenSetterChanged",
			func(flags *pflag.FlagSet) {
				require.NoError(nil, flags.Set(cmdFlagNameRandomCharSet, "numeric"))
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.Bool(cmdFlagNameRandom, false, "")
			flags.String(cmdFlagNameRandomCharSet, cmdFlagValueCharSet, "")
			flags.String(cmdFlagNameRandomCharacters, "", "")
			flags.Int(cmdFlagNameRandomLength, 72, "")

			tc.setup(flags)

			random, err := cmdFlagsCryptoHashPasswordRandom(flags, cmdFlagNameRandom, cmdFlagNameRandomCharSet, cmdFlagNameRandomCharacters, cmdFlagNameRandomLength)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, random)
		})
	}
}

func TestCmdFlagsCryptoHashGetPassword(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func(flags *pflag.FlagSet)
		use        string
		args       []string
		useArgs    bool
		useRandom  bool
		err        string
		expectedPW string
		expectRand bool
	}{
		{
			"ShouldReturnPasswordFromFlag",
			func(flags *pflag.FlagSet) {
				require.NoError(nil, flags.Set(cmdFlagNamePassword, "mypassword"))
			},
			cmdUseGenerate,
			nil,
			false,
			false,
			"",
			"mypassword",
			false,
		},
		{
			"ShouldReturnRandomPassword",
			func(flags *pflag.FlagSet) {
				require.NoError(nil, flags.Set(cmdFlagNameRandom, "true"))
			},
			cmdUseGenerate,
			nil,
			false,
			true,
			"",
			"",
			true,
		},
		{
			"ShouldReturnPasswordFromArgs",
			func(flags *pflag.FlagSet) {},
			cmdUseGenerate,
			[]string{"argpassword"},
			true,
			false,
			"",
			"argpassword",
			false,
		},
		{
			"ShouldReturnJoinedPasswordFromMultipleArgs",
			func(flags *pflag.FlagSet) {},
			cmdUseGenerate,
			[]string{"arg", "password", "here"},
			true,
			false,
			"",
			"arg password here",
			false,
		},
		{
			"ShouldPreferPasswordFlagOverArgs",
			func(flags *pflag.FlagSet) {
				require.NoError(nil, flags.Set(cmdFlagNamePassword, "flagpassword"))
			},
			cmdUseGenerate,
			[]string{"argpassword"},
			true,
			false,
			"",
			"flagpassword",
			false,
		},
		{
			"ShouldPreferRandomOverArgs",
			func(flags *pflag.FlagSet) {
				require.NoError(nil, flags.Set(cmdFlagNameRandom, "true"))
			},
			cmdUseGenerate,
			[]string{"argpassword"},
			true,
			true,
			"",
			"",
			true,
		},
		{
			"ShouldErrNoTerminalWhenUseArgsButNoArgs",
			func(flags *pflag.FlagSet) {},
			cmdUseGenerate,
			nil,
			true,
			false,
			"failed to read the password from the terminal:",
			"",
			false,
		},
		{
			"ShouldErrNoTerminal",
			func(flags *pflag.FlagSet) {},
			cmdUseGenerate,
			nil,
			false,
			false,
			"failed to read the password from the terminal:",
			"",
			false,
		},
		{
			"ShouldErrNoTerminalValidateUse",
			func(flags *pflag.FlagSet) {},
			fmt.Sprintf(cmdUseFmtValidate, cmdUseValidate),
			nil,
			false,
			false,
			"failed to read the password from the terminal:",
			"",
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.String(cmdFlagNamePassword, "", "")
			flags.Bool(cmdFlagNameNoConfirm, true, "")
			flags.Bool(cmdFlagNameRandom, false, "")
			flags.String(cmdFlagNameRandomCharSet, cmdFlagValueCharSet, "")
			flags.String(cmdFlagNameRandomCharacters, "", "")
			flags.Int(cmdFlagNameRandomLength, 72, "")

			tc.setup(flags)

			buf := new(bytes.Buffer)

			password, random, err := cmdFlagsCryptoHashGetPassword(buf, flags, tc.use, tc.args, tc.useArgs, tc.useRandom)

			if tc.err == "" {
				assert.NoError(t, err)

				if tc.expectRand {
					assert.True(t, random)
					assert.NotEmpty(t, password)
				} else {
					assert.Equal(t, tc.expectedPW, password)
				}
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestCryptoHashValidateRunE(t *testing.T) {
	genFlags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	genFlags.String(cmdFlagNamePassword, "testpassword", "")
	genFlags.Bool(cmdFlagNameNoConfirm, true, "")
	genFlags.Bool(cmdFlagNameRandom, false, "")
	genFlags.String(cmdFlagNameRandomCharSet, cmdFlagValueCharSet, "")
	genFlags.String(cmdFlagNameRandomCharacters, "", "")
	genFlags.Int(cmdFlagNameRandomLength, 72, "")

	require.NoError(t, genFlags.Set(cmdFlagNamePassword, "testpassword"))

	config := &schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			File: &schema.AuthenticationBackendFile{
				Password: schema.DefaultCIPasswordConfig,
			},
		},
	}

	genBuf := new(bytes.Buffer)

	require.NoError(t, runCryptoHashGenerate(genBuf, genFlags, cmdUseGenerate, nil, config))

	output := genBuf.String()
	digestLine := ""

	for _, line := range splitLines(output) {
		if len(line) > 8 && line[:8] == "Digest: " {
			digestLine = line[8:]

			break
		}
	}

	require.NotEmpty(t, digestLine, "should have extracted a digest")

	testCases := []struct {
		name     string
		password string
		digest   string
		err      string
		expected string
	}{
		{
			"ShouldMatchPassword",
			"testpassword",
			digestLine,
			"",
			"The password matches the digest.",
		},
		{
			"ShouldNotMatchPassword",
			"wrongpassword",
			digestLine,
			"",
			"The password does not match the digest.",
		},
		{
			"ShouldErrInvalidDigest",
			"testpassword",
			"notavaliddigest",
			"error occurred trying to validate the password against the digest:",
			"",
		},
		{
			"ShouldErrEmptyPasswordTerminal",
			"",
			digestLine,
			"you must either use an interactive terminal or use the --password flag",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := NewCmdCtx()

			cmd := &cobra.Command{
				Use: fmt.Sprintf(cmdUseFmtValidate, cmdUseValidate),
			}

			buf := new(bytes.Buffer)

			cmd.SetOut(buf)

			cmd.Flags().String(cmdFlagNamePassword, "", "")

			if tc.password != "" {
				require.NoError(t, cmd.Flags().Set(cmdFlagNamePassword, tc.password))
			}

			err := cmdCtx.CryptoHashValidateRunE(cmd, []string{tc.digest})

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestCryptoHashGenerateMapFlagsRunE(t *testing.T) {
	testCases := []struct {
		name string
		use  string
	}{
		{
			"ShouldMapArgon2Flags",
			cmdUseHashArgon2,
		},
		{
			"ShouldMapSHA2CryptFlags",
			cmdUseHashSHA2Crypt,
		},
		{
			"ShouldMapPBKDF2Flags",
			cmdUseHashPBKDF2,
		},
		{
			"ShouldMapBcryptFlags",
			cmdUseHashBcrypt,
		},
		{
			"ShouldMapScryptFlags",
			cmdUseHashScrypt,
		},
		{
			"ShouldHandleUnknownUse",
			"unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := NewCmdCtx()
			cmdCtx.cconfig = NewCmdCtxConfig()

			cmd := &cobra.Command{Use: tc.use}

			err := cmdCtx.CryptoHashGenerateMapFlagsRunE(cmd, nil)

			assert.NoError(t, err)
		})
	}
}

func TestCryptoHashGenerateRunE(t *testing.T) {
	cmdCtx := NewCmdCtx()

	cmdCtx.config = &schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			File: &schema.AuthenticationBackendFile{
				Password: schema.DefaultPasswordConfig,
			},
		},
	}

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String(cmdFlagNamePassword, "", "")
	flags.Bool(cmdFlagNameNoConfirm, true, "")
	flags.Bool(cmdFlagNameRandom, false, "")
	flags.String(cmdFlagNameRandomCharSet, cmdFlagValueCharSet, "")
	flags.String(cmdFlagNameRandomCharacters, "", "")
	flags.Int(cmdFlagNameRandomLength, 72, "")

	require.NoError(t, flags.Set(cmdFlagNamePassword, "testpass"))

	buf := new(bytes.Buffer)

	err := runCryptoHashGenerate(buf, flags, cmdUseGenerate, nil, cmdCtx.config)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Digest:")
}

func TestNewCryptoHashGenerateSubCmd(t *testing.T) {
	testCases := []struct {
		name string
		use  string
	}{
		{
			"ShouldCreateArgon2SubCmd",
			cmdUseHashArgon2,
		},
		{
			"ShouldCreateSHA2CryptSubCmd",
			cmdUseHashSHA2Crypt,
		},
		{
			"ShouldCreatePBKDF2SubCmd",
			cmdUseHashPBKDF2,
		},
		{
			"ShouldCreateBcryptSubCmd",
			cmdUseHashBcrypt,
		},
		{
			"ShouldCreateScryptSubCmd",
			cmdUseHashScrypt,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newCryptoHashGenerateSubCmd(&CmdCtx{}, tc.use)

			assert.NotNil(t, cmd)
			assert.Equal(t, tc.use, cmd.Use)
		})
	}
}

func splitLines(s string) (lines []string) {
	for _, line := range bytes.Split([]byte(s), []byte("\n")) {
		lines = append(lines, string(line))
	}

	return lines
}
