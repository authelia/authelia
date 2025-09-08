package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.yaml.in/yaml/v4"
	"golang.org/x/term"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/utils"
)

func flagsGetUserIdentifiersGenerateOptions(flags *pflag.FlagSet) (users, services, sectors []string, err error) {
	if users, err = flags.GetStringSlice(cmdFlagNameUsers); err != nil {
		return nil, nil, nil, err
	}

	if services, err = flags.GetStringSlice(cmdFlagNameServices); err != nil {
		return nil, nil, nil, err
	}

	if sectors, err = flags.GetStringSlice(cmdFlagNameSectors); err != nil {
		return nil, nil, nil, err
	}

	return users, services, sectors, nil
}

//nolint:gocyclo
func flagsGetRandomCharacters(flags *pflag.FlagSet, flagNameLength, flagNameCharSet, flagNameCharacters string) (r string, err error) {
	var (
		n       int
		charset string
	)

	if n, err = flags.GetInt(flagNameLength); err != nil {
		return "", err
	}

	if n < 1 {
		return "", fmt.Errorf("flag --%s with value '%d' is invalid: must be at least 1", flagNameLength, n)
	}

	useCharSet, useCharacters := flags.Changed(flagNameCharSet), flags.Changed(flagNameCharacters)

	if useCharSet && useCharacters {
		return "", fmt.Errorf("flag --%s and flag --%s are mutually exclusive, only one may be used", flagNameCharSet, flagNameCharacters)
	}

	switch {
	case useCharSet, !useCharacters:
		var c string

		if c, err = flags.GetString(flagNameCharSet); err != nil {
			return "", err
		}

		switch c {
		case "ascii":
			charset = random.CharSetASCII
		case "alphanumeric":
			charset = random.CharSetAlphaNumeric
		case "alphanumeric-lower":
			charset = random.CharSetAlphabeticLower + random.CharSetNumeric
		case "alphanumeric-upper":
			charset = random.CharSetAlphabeticUpper + random.CharSetNumeric
		case "alphabetic":
			charset = random.CharSetAlphabetic
		case "alphabetic-lower":
			charset = random.CharSetAlphabeticLower
		case "alphabetic-upper":
			charset = random.CharSetAlphabeticUpper
		case "numeric-hex":
			charset = random.CharSetNumericHex
		case "numeric":
			charset = random.CharSetNumeric
		case "rfc3986":
			charset = random.CharSetRFC3986Unreserved
		case "rfc3986-lower":
			charset = random.CharSetAlphabeticLower + random.CharSetNumeric + random.CharSetSymbolicRFC3986Unreserved
		case "rfc3986-upper":
			charset = random.CharSetAlphabeticUpper + random.CharSetNumeric + random.CharSetSymbolicRFC3986Unreserved
		default:
			return "", fmt.Errorf("flag '--%s' with value '%s' is invalid, must be one of 'ascii', 'alphanumeric', 'alphabetic', 'numeric', 'numeric-hex', or 'rfc3986'", flagNameCharSet, c)
		}
	default:
		if charset, err = flags.GetString(flagNameCharacters); err != nil {
			return "", err
		}
	}

	rand := random.New()

	return rand.StringCustom(n, charset), nil
}

func flagParseFileMode(name string, flags *pflag.FlagSet) (mode os.FileMode, err error) {
	var (
		value string
		octal uint64
	)

	if value, err = flags.GetString(name); err != nil {
		return mode, err
	}

	if octal, err = strconv.ParseUint(value, 8, 32); err != nil {
		return mode, err
	}

	return os.FileMode(octal), nil
}

func termReadConfirmation(prompt, confirmation string) (confirmed bool, err error) {
	terminal, fd, state, err := getTerminal(prompt)
	if err != nil {
		return false, err
	}

	defer func(fd int, oldState *term.State) {
		_ = term.Restore(fd, oldState)
	}(fd, state)

	var input string

	if input, err = terminal.ReadLine(); err != nil {
		return false, fmt.Errorf("failed to read from the terminal: %w", err)
	}

	if input != confirmation {
		return false, nil
	}

	return true, nil
}

func getTerminal(prompt string) (terminal *term.Terminal, fd int, state *term.State, err error) {
	fd = int(syscall.Stdin) //nolint:unconvert,nolintlint

	if !term.IsTerminal(fd) {
		return nil, -1, nil, ErrStdinIsNotTerminal
	}

	var width, height int

	if width, height, err = term.GetSize(int(syscall.Stdout)); err != nil { //nolint:unconvert,nolintlint
		return nil, -1, nil, fmt.Errorf("failed to get terminal size: %w", err)
	}

	state, err = term.MakeRaw(fd)
	if err != nil {
		return nil, -1, nil, fmt.Errorf("failed to get terminal state: %w", err)
	}

	c := struct {
		io.Reader
		io.Writer
	}{
		os.Stdin,
		os.Stdout,
	}

	terminal = term.NewTerminal(c, prompt)

	if err = terminal.SetSize(width, height); err != nil {
		return nil, -1, nil, fmt.Errorf("failed to set terminal size: %w", err)
	}

	return terminal, fd, state, nil
}

func termReadPasswordWithPrompt(prompt, flag string) (password string, err error) {
	terminal, fd, state, err := getTerminal("")
	if err != nil {
		if errors.Is(err, ErrStdinIsNotTerminal) {
			switch len(flag) {
			case 0:
				return "", err
			case 1:
				return "", fmt.Errorf("you must either use an interactive terminal or use the -%s flag", flag)
			default:
				return "", fmt.Errorf("you must either use an interactive terminal or use the --%s flag", flag)
			}
		}

		return "", err
	}

	defer func(fd int, oldState *term.State) {
		_ = term.Restore(fd, oldState)
	}(fd, state)

	if password, err = terminal.ReadPassword(prompt); err != nil {
		return "", fmt.Errorf("failed to read the input from the terminal: %w", err)
	}

	return password, nil
}

type XEnvCLIResult int

const (
	XEnvCLIResultCLIExplicit XEnvCLIResult = iota
	XEnvCLIResultCLIImplicit
	XEnvCLIResultEnvironment
)

func loadXEnvCLIConfigValues(cmd *cobra.Command) (configs []string, filters []configuration.BytesFilter, err error) {
	var (
		filterNames []string
		result      XEnvCLIResult
	)

	if configs, result, err = loadXEnvCLIStringSliceValue(cmd, cmdFlagEnvNameConfig, cmdFlagNameConfig); err != nil {
		return nil, nil, err
	}

	if configs, err = loadXNormalizedPaths(configs, result); err != nil {
		return nil, nil, err
	}

	if filterNames, _, err = loadXEnvCLIStringSliceValue(cmd, cmdFlagEnvNameConfigFilters, cmdFlagNameConfigExpFilters); err != nil {
		return nil, nil, err
	}

	if filters, err = configuration.NewFileFilters(filterNames); err != nil {
		return nil, nil, fmt.Errorf("error occurred loading configuration: flag '--%s' is invalid: %w", cmdFlagNameConfigExpFilters, err)
	}

	return
}

func loadXNormalizedPaths(paths []string, result XEnvCLIResult) ([]string, error) {
	var (
		configs, files, dirs []string
		err                  error
	)

	var stat os.FileInfo

	for _, path := range paths {
		if path, err = filepath.Abs(path); err != nil {
			return nil, fmt.Errorf("failed to determine absolute path for '%s': %w", path, err)
		}

		switch stat, err = os.Stat(path); {
		case err == nil && stat.IsDir():
			configs = append(configs, path)
			dirs = append(dirs, path)
		case err == nil:
			configs = append(configs, path)
			files = append(files, path)
		default:
			if os.IsNotExist(err) {
				switch result {
				case XEnvCLIResultCLIImplicit:
					continue
				default:
					configs = append(configs, path)
					files = append(files, path)

					continue
				}
			}

			return nil, fmt.Errorf("error occurred stating file at path '%s': %w", path, err)
		}
	}

	for i, file := range files {
		if file, err = filepath.Abs(file); err != nil {
			return nil, fmt.Errorf("failed to determine absolute path for '%s': %w", files[i], err)
		}

		if len(dirs) != 0 {
			filedir := filepath.Dir(file)

			for _, dir := range dirs {
				if filedir == dir {
					return nil, fmt.Errorf("failed to load config directory '%s': the config file '%s' is in that directory which is not supported", dir, file)
				}
			}
		}
	}

	return configs, nil
}

func loadXEnvCLIStringSliceValue(cmd *cobra.Command, envKey, flagName string) (value []string, result XEnvCLIResult, err error) {
	if cmd.Flags().Changed(flagName) {
		value, err = cmd.Flags().GetStringSlice(flagName)

		return value, XEnvCLIResultCLIExplicit, err
	}

	var (
		env string
		ok  bool
	)

	if envKey != "" {
		env, ok = os.LookupEnv(envKey)
	}

	switch {
	case ok && env != "":
		return strings.Split(env, ","), XEnvCLIResultEnvironment, nil
	default:
		value, err = cmd.Flags().GetStringSlice(flagName)

		return value, XEnvCLIResultCLIImplicit, err
	}
}

func newHelpTopic(topic, short, body string) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   topic,
		Short: short,
	}

	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		cmdHelpTopic(cmd, body, topic)

		return nil
	})

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmdHelpTopic(cmd, body, topic)
	})

	return cmd
}

func cmdHelpTopic(cmd *cobra.Command, body, topic string) {
	_ = cmd.Parent().Help()

	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Help Topic: %s\n\n", topic)
	_, _ = fmt.Fprint(cmd.OutOrStdout(), body)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n\n")
}

func exportYAMLWithJSONSchema(w io.Writer, name string, v any) (err error) {
	var (
		semver *model.SemanticVersion
	)

	version := "latest"

	if semver, err = model.NewSemanticVersion(utils.BuildTag); err == nil {
		version = fmt.Sprintf("v%d.%d", semver.Major, semver.Minor+1)
	}

	if _, err = fmt.Fprintf(w, model.FormatJSONSchemaYAMLLanguageServer, version, name); err != nil {
		return err
	}

	if _, err = fmt.Fprintf(w, "\n\n"); err != nil {
		return err
	}

	encoder := yaml.NewEncoder(w)

	if err = encoder.Encode(v); err != nil {
		return fmt.Errorf("error occurred marshalling data to YAML: %w", err)
	}

	return nil
}

func getCryptoHashGenerateMapFlagsFromUse(use string) (flags map[string]string) {
	switch use {
	case cmdUseHashArgon2:
		return map[string]string{
			cmdFlagNameVariant:     prefixFilePassword + suffixArgon2Variant,
			cmdFlagNameIterations:  prefixFilePassword + suffixArgon2Iterations,
			cmdFlagNameMemory:      prefixFilePassword + suffixArgon2Memory,
			cmdFlagNameParallelism: prefixFilePassword + suffixArgon2Parallelism,
			cmdFlagNameKeySize:     prefixFilePassword + suffixArgon2KeyLength,
			cmdFlagNameSaltSize:    prefixFilePassword + suffixArgon2SaltLength,
		}
	case cmdUseHashSHA2Crypt:
		return map[string]string{
			cmdFlagNameVariant:    prefixFilePassword + suffixSHA2CryptVariant,
			cmdFlagNameIterations: prefixFilePassword + suffixSHA2CryptIterations,
			cmdFlagNameSaltSize:   prefixFilePassword + suffixSHA2CryptSaltLength,
		}
	case cmdUseHashPBKDF2:
		return map[string]string{
			cmdFlagNameVariant:    prefixFilePassword + suffixPBKDF2Variant,
			cmdFlagNameIterations: prefixFilePassword + suffixPBKDF2Iterations,
			cmdFlagNameSaltSize:   prefixFilePassword + suffixPBKDF2SaltLength,
		}
	case cmdUseHashBcrypt:
		return map[string]string{
			cmdFlagNameVariant: prefixFilePassword + suffixBcryptVariant,
			cmdFlagNameCost:    prefixFilePassword + suffixBcryptCost,
		}
	case cmdUseHashScrypt:
		return map[string]string{
			cmdFlagNameVariant:     prefixFilePassword + suffixScryptVariant,
			cmdFlagNameIterations:  prefixFilePassword + suffixScryptIterations,
			cmdFlagNameBlockSize:   prefixFilePassword + suffixScryptBlockSize,
			cmdFlagNameParallelism: prefixFilePassword + suffixScryptParallelism,
			cmdFlagNameKeySize:     prefixFilePassword + suffixScryptKeyLength,
			cmdFlagNameSaltSize:    prefixFilePassword + suffixScryptSaltLength,
		}
	default:
		return nil
	}
}
