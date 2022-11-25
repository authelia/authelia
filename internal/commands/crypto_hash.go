package commands

import (
	"errors"
	"fmt"
	"strings"
	"syscall"

	"github.com/go-crypt/crypt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
)

func newHashPasswordCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseHashPassword,
		Short:   cmdAutheliaHashPasswordShort,
		Long:    cmdAutheliaHashPasswordLong,
		Example: cmdAutheliaHashPasswordExample,
		Args:    cobra.MaximumNArgs(1),
		RunE:    cmdHashPasswordRunE,

		DisableAutoGenTag: true,
	}

	cmdFlagConfig(cmd)

	cmd.Flags().BoolP(cmdFlagNameSHA512, "z", false, fmt.Sprintf("use sha512 as the algorithm (changes iterations to %d, change with -i)", schema.DefaultPasswordConfig.SHA2Crypt.Iterations))
	cmd.Flags().IntP(cmdFlagNameIterations, "i", schema.DefaultPasswordConfig.Argon2.Iterations, "set the number of hashing iterations")
	cmd.Flags().IntP(cmdFlagNameMemory, "m", schema.DefaultPasswordConfig.Argon2.Memory, "[argon2id] set the amount of memory param (in MB)")
	cmd.Flags().IntP(cmdFlagNameParallelism, "p", schema.DefaultPasswordConfig.Argon2.Parallelism, "[argon2id] set the parallelism param")
	cmd.Flags().IntP("key-length", "k", schema.DefaultPasswordConfig.Argon2.KeyLength, "[argon2id] set the key length param")
	cmd.Flags().IntP("salt-length", "l", schema.DefaultPasswordConfig.Argon2.SaltLength, "set the auto-generated salt length")
	cmd.Flags().Bool(cmdFlagNameNoConfirm, false, "skip the password confirmation prompt")

	return cmd
}

func cmdHashPasswordRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		flagsMap map[string]string
		sha512   bool
	)

	if sha512, err = cmd.Flags().GetBool(cmdFlagNameSHA512); err != nil {
		return err
	}

	switch {
	case sha512:
		flagsMap = map[string]string{
			cmdFlagNameIterations: prefixFilePassword + ".sha2crypt.iterations",
			"salt-length":         prefixFilePassword + ".sha2crypt.salt_length",
		}
	default:
		flagsMap = map[string]string{
			cmdFlagNameIterations:  prefixFilePassword + ".argon2.iterations",
			"key-length":           prefixFilePassword + ".argon2.key_length",
			"salt-length":          prefixFilePassword + ".argon2.salt_length",
			cmdFlagNameParallelism: prefixFilePassword + ".argon2.parallelism",
			cmdFlagNameMemory:      prefixFilePassword + ".argon2.memory",
		}
	}

	return cmdCryptoHashGenerateFinish(cmd, args, flagsMap)
}

func newCryptoHashCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseHash,
		Short:   cmdAutheliaCryptoHashShort,
		Long:    cmdAutheliaCryptoHashLong,
		Example: cmdAutheliaCryptoHashExample,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newCryptoHashValidateCmd(),
		newCryptoHashGenerateCmd(),
	)

	return cmd
}

func newCryptoHashGenerateCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseGenerate,
		Short:   cmdAutheliaCryptoHashGenerateShort,
		Long:    cmdAutheliaCryptoHashGenerateLong,
		Example: cmdAutheliaCryptoHashGenerateExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdCryptoHashGenerateFinish(cmd, args, map[string]string{})
		},

		DisableAutoGenTag: true,
	}

	cmdFlagConfig(cmd)
	cmdFlagPassword(cmd, true)
	cmdFlagRandomPassword(cmd)

	for _, use := range []string{cmdUseHashArgon2, cmdUseHashSHA2Crypt, cmdUseHashPBKDF2, cmdUseHashBCrypt, cmdUseHashSCrypt} {
		cmd.AddCommand(newCryptoHashGenerateSubCmd(use))
	}

	return cmd
}

func newCryptoHashGenerateSubCmd(use string) (cmd *cobra.Command) {
	useFmt := fmtCryptoHashUse(use)

	cmd = &cobra.Command{
		Use:     use,
		Short:   fmt.Sprintf(fmtCmdAutheliaCryptoHashGenerateSubShort, useFmt),
		Long:    fmt.Sprintf(fmtCmdAutheliaCryptoHashGenerateSubLong, useFmt, useFmt),
		Example: fmt.Sprintf(fmtCmdAutheliaCryptoHashGenerateSubExample, use),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},

		DisableAutoGenTag: true,
	}

	switch use {
	case cmdUseHashArgon2:
		cmdFlagIterations(cmd, schema.DefaultPasswordConfig.Argon2.Iterations)
		cmdFlagParallelism(cmd, schema.DefaultPasswordConfig.Argon2.Parallelism)
		cmdFlagKeySize(cmd, schema.DefaultPasswordConfig.Argon2.KeyLength)
		cmdFlagSaltSize(cmd, schema.DefaultPasswordConfig.Argon2.SaltLength)

		cmd.Flags().StringP(cmdFlagNameVariant, "v", schema.DefaultPasswordConfig.Argon2.Variant, "variant, options are 'argon2id', 'argon2i', and 'argon2d'")
		cmd.Flags().IntP(cmdFlagNameMemory, "m", schema.DefaultPasswordConfig.Argon2.Memory, "memory in kibibytes")
		cmd.Flags().String(cmdFlagNameProfile, "", "profile to use, options are low-memory and recommended")

		cmd.RunE = cryptoHashGenerateArgon2RunE
	case cmdUseHashSHA2Crypt:
		cmdFlagIterations(cmd, schema.DefaultPasswordConfig.SHA2Crypt.Iterations)
		cmdFlagSaltSize(cmd, schema.DefaultPasswordConfig.SHA2Crypt.SaltLength)

		cmd.Flags().StringP(cmdFlagNameVariant, "v", schema.DefaultPasswordConfig.SHA2Crypt.Variant, "variant, options are sha256 and sha512")

		cmd.RunE = cryptoHashGenerateSHA2CryptRunE
	case cmdUseHashPBKDF2:
		cmdFlagIterations(cmd, schema.DefaultPasswordConfig.PBKDF2.Iterations)
		cmdFlagSaltSize(cmd, schema.DefaultPasswordConfig.PBKDF2.SaltLength)

		cmd.Flags().StringP(cmdFlagNameVariant, "v", schema.DefaultPasswordConfig.PBKDF2.Variant, "variant, options are 'sha1', 'sha224', 'sha256', 'sha384', and 'sha512'")

		cmd.RunE = cryptoHashGeneratePBKDF2RunE
	case cmdUseHashBCrypt:
		cmd.Flags().StringP(cmdFlagNameVariant, "v", schema.DefaultPasswordConfig.BCrypt.Variant, "variant, options are 'standard' and 'sha256'")
		cmd.Flags().IntP(cmdFlagNameCost, "i", schema.DefaultPasswordConfig.BCrypt.Cost, "hashing cost")

		cmd.RunE = cryptoHashGenerateBCryptRunE
	case cmdUseHashSCrypt:
		cmdFlagIterations(cmd, schema.DefaultPasswordConfig.SCrypt.Iterations)
		cmdFlagKeySize(cmd, schema.DefaultPasswordConfig.SCrypt.KeyLength)
		cmdFlagSaltSize(cmd, schema.DefaultPasswordConfig.SCrypt.SaltLength)
		cmdFlagParallelism(cmd, schema.DefaultPasswordConfig.SCrypt.Parallelism)

		cmd.Flags().IntP(cmdFlagNameBlockSize, "r", schema.DefaultPasswordConfig.SCrypt.BlockSize, "block size")

		cmd.RunE = cryptoHashGenerateSCryptRunE
	}

	return cmd
}

func cryptoHashGenerateArgon2RunE(cmd *cobra.Command, args []string) (err error) {
	flagsMap := map[string]string{
		cmdFlagNameVariant:     prefixFilePassword + ".argon2.variant",
		cmdFlagNameIterations:  prefixFilePassword + ".argon2.iterations",
		cmdFlagNameMemory:      prefixFilePassword + ".argon2.memory",
		cmdFlagNameParallelism: prefixFilePassword + ".argon2.parallelism",
		cmdFlagNameKeySize:     prefixFilePassword + ".argon2.key_length",
		cmdFlagNameSaltSize:    prefixFilePassword + ".argon2.salt_length",
	}

	return cmdCryptoHashGenerateFinish(cmd, args, flagsMap)
}

func cryptoHashGenerateSHA2CryptRunE(cmd *cobra.Command, args []string) (err error) {
	flagsMap := map[string]string{
		cmdFlagNameVariant:    prefixFilePassword + ".sha2crypt.variant",
		cmdFlagNameIterations: prefixFilePassword + ".sha2crypt.iterations",
		cmdFlagNameSaltSize:   prefixFilePassword + ".sha2crypt.salt_length",
	}

	return cmdCryptoHashGenerateFinish(cmd, args, flagsMap)
}

func cryptoHashGeneratePBKDF2RunE(cmd *cobra.Command, args []string) (err error) {
	flagsMap := map[string]string{
		cmdFlagNameVariant:    prefixFilePassword + ".pbkdf2.variant",
		cmdFlagNameIterations: prefixFilePassword + ".pbkdf2.iterations",
		cmdFlagNameKeySize:    prefixFilePassword + ".pbkdf2.key_length",
		cmdFlagNameSaltSize:   prefixFilePassword + ".pbkdf2.salt_length",
	}

	return cmdCryptoHashGenerateFinish(cmd, args, flagsMap)
}

func cryptoHashGenerateBCryptRunE(cmd *cobra.Command, args []string) (err error) {
	flagsMap := map[string]string{
		cmdFlagNameVariant: prefixFilePassword + ".bcrypt.variant",
		cmdFlagNameCost:    prefixFilePassword + ".bcrypt.cost",
	}

	return cmdCryptoHashGenerateFinish(cmd, args, flagsMap)
}

func cryptoHashGenerateSCryptRunE(cmd *cobra.Command, args []string) (err error) {
	flagsMap := map[string]string{
		cmdFlagNameIterations:  prefixFilePassword + ".scrypt.iterations",
		cmdFlagNameBlockSize:   prefixFilePassword + ".scrypt.block_size",
		cmdFlagNameParallelism: prefixFilePassword + ".scrypt.parallelism",
		cmdFlagNameKeySize:     prefixFilePassword + ".scrypt.key_length",
		cmdFlagNameSaltSize:    prefixFilePassword + ".scrypt.salt_length",
	}

	return cmdCryptoHashGenerateFinish(cmd, args, flagsMap)
}

func newCryptoHashValidateCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     fmt.Sprintf(cmdUseFmtValidate, cmdUseValidate),
		Short:   cmdAutheliaCryptoHashValidateShort,
		Long:    cmdAutheliaCryptoHashValidateLong,
		Example: cmdAutheliaCryptoHashValidateExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var (
				password string
				valid    bool
			)

			if password, _, err = cmdCryptoHashGetPassword(cmd, args, false, false); err != nil {
				return fmt.Errorf("error occurred trying to obtain the password: %w", err)
			}

			if len(password) == 0 {
				return fmt.Errorf("no password provided")
			}

			if valid, err = crypt.CheckPassword(password, args[0]); err != nil {
				return fmt.Errorf("error occurred trying to validate the password against the digest: %w", err)
			}

			switch {
			case valid:
				fmt.Println("The password matches the digest.")
			default:
				fmt.Println("The password does not match the digest.")
			}

			return nil
		},

		DisableAutoGenTag: true,
	}

	cmdFlagPassword(cmd, false)

	return cmd
}

func cmdCryptoHashGenerateFinish(cmd *cobra.Command, args []string, flagsMap map[string]string) (err error) {
	var (
		algorithm string
		configs   []string

		c schema.Password
	)

	if configs, err = cmd.Flags().GetStringSlice(cmdFlagNameConfig); err != nil {
		return err
	}

	// Skip config if the flag wasn't set and the default is non-existent.
	if !cmd.Flags().Changed(cmdFlagNameConfig) {
		configs = configFilterExisting(configs)
	}

	legacy := cmd.Use == cmdUseHashPassword

	switch {
	case cmd.Use == cmdUseGenerate:
		break
	case legacy:
		if sha512, _ := cmd.Flags().GetBool(cmdFlagNameSHA512); sha512 {
			algorithm = cmdUseHashSHA2Crypt
		} else {
			algorithm = cmdUseHashArgon2
		}
	default:
		algorithm = cmd.Use
	}

	if c, err = cmdCryptoHashGetConfig(algorithm, configs, cmd.Flags(), flagsMap); err != nil {
		return err
	}

	if legacy && algorithm == cmdUseHashArgon2 && cmd.Flags().Changed(cmdFlagNameMemory) {
		c.Argon2.Memory *= 1024
	}

	var (
		hash     crypt.Hash
		digest   crypt.Digest
		password string
		random   bool
	)

	if password, random, err = cmdCryptoHashGetPassword(cmd, args, legacy, !legacy); err != nil {
		return err
	}

	if len(password) == 0 {
		return fmt.Errorf("no password provided")
	}

	if hash, err = authentication.NewFileCryptoHashFromConfig(c); err != nil {
		return err
	}

	if digest, err = hash.Hash(password); err != nil {
		return err
	}

	if random {
		fmt.Printf("Random Password: %s\n", password)
	}

	fmt.Printf("Digest: %s\n", digest.Encode())

	return nil
}

func cmdCryptoHashGetConfig(algorithm string, configs []string, flags *pflag.FlagSet, flagsMap map[string]string) (c schema.Password, err error) {
	mapDefaults := map[string]interface{}{
		prefixFilePassword + ".algorithm":             schema.DefaultPasswordConfig.Algorithm,
		prefixFilePassword + ".argon2.variant":        schema.DefaultPasswordConfig.Argon2.Variant,
		prefixFilePassword + ".argon2.iterations":     schema.DefaultPasswordConfig.Argon2.Iterations,
		prefixFilePassword + ".argon2.memory":         schema.DefaultPasswordConfig.Argon2.Memory,
		prefixFilePassword + ".argon2.parallelism":    schema.DefaultPasswordConfig.Argon2.Parallelism,
		prefixFilePassword + ".argon2.key_length":     schema.DefaultPasswordConfig.Argon2.KeyLength,
		prefixFilePassword + ".argon2.salt_length":    schema.DefaultPasswordConfig.Argon2.SaltLength,
		prefixFilePassword + ".sha2crypt.variant":     schema.DefaultPasswordConfig.SHA2Crypt.Variant,
		prefixFilePassword + ".sha2crypt.iterations":  schema.DefaultPasswordConfig.SHA2Crypt.Iterations,
		prefixFilePassword + ".sha2crypt.salt_length": schema.DefaultPasswordConfig.SHA2Crypt.SaltLength,
		prefixFilePassword + ".pbkdf2.variant":        schema.DefaultPasswordConfig.PBKDF2.Variant,
		prefixFilePassword + ".pbkdf2.iterations":     schema.DefaultPasswordConfig.PBKDF2.Iterations,
		prefixFilePassword + ".pbkdf2.salt_length":    schema.DefaultPasswordConfig.PBKDF2.SaltLength,
		prefixFilePassword + ".bcrypt.variant":        schema.DefaultPasswordConfig.BCrypt.Variant,
		prefixFilePassword + ".bcrypt.cost":           schema.DefaultPasswordConfig.BCrypt.Cost,
		prefixFilePassword + ".scrypt.iterations":     schema.DefaultPasswordConfig.SCrypt.Iterations,
		prefixFilePassword + ".scrypt.block_size":     schema.DefaultPasswordConfig.SCrypt.BlockSize,
		prefixFilePassword + ".scrypt.parallelism":    schema.DefaultPasswordConfig.SCrypt.Parallelism,
		prefixFilePassword + ".scrypt.key_length":     schema.DefaultPasswordConfig.SCrypt.KeyLength,
		prefixFilePassword + ".scrypt.salt_length":    schema.DefaultPasswordConfig.SCrypt.SaltLength,
	}

	sources := configuration.NewDefaultSourcesWithDefaults(configs,
		configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter,
		configuration.NewMapSource(mapDefaults),
		configuration.NewCommandLineSourceWithMapping(flags, flagsMap, false, false),
	)

	if algorithm != "" {
		alg := map[string]interface{}{prefixFilePassword + ".algorithm": algorithm}

		sources = append(sources, configuration.NewMapSource(alg))
	}

	val := schema.NewStructValidator()

	if _, err = configuration.LoadAdvanced(val, prefixFilePassword, &c, sources...); err != nil {
		return schema.Password{}, fmt.Errorf("error occurred loading configuration: %w", err)
	}

	validator.ValidatePasswordConfiguration(&c, val)

	errs := val.Errors()

	if len(errs) != 0 {
		for i, e := range errs {
			if i == 0 {
				err = e
				continue
			}

			err = fmt.Errorf("%v, %w", err, e)
		}

		return schema.Password{}, fmt.Errorf("errors occurred validating the password configuration: %w", err)
	}

	return c, nil
}

func cmdCryptoHashGetPassword(cmd *cobra.Command, args []string, useArgs, useRandom bool) (password string, random bool, err error) {
	if useRandom {
		if random, err = cmd.Flags().GetBool(cmdFlagNameRandom); err != nil {
			return
		}
	}

	switch {
	case random:
		password, err = flagsGetRandomCharacters(cmd.Flags(), cmdFlagNameRandomLength, cmdFlagNameRandomCharSet, cmdFlagNameCharacters)

		return
	case cmd.Flags().Changed(cmdFlagNamePassword):
		password, err = cmd.Flags().GetString(cmdFlagNamePassword)

		return
	case useArgs && len(args) != 0:
		password, err = strings.Join(args, " "), nil

		return
	}

	var (
		data      []byte
		noConfirm bool
	)

	if data, err = termReadPasswordWithPrompt("Enter Password: "); err != nil {
		if errors.Is(err, ErrStdinIsNotTerminal) {
			err = fmt.Errorf("you must either use an interactive terminal or use the --password flag")
		} else {
			err = fmt.Errorf("failed to read the password from the terminal: %w", err)
		}

		return
	}

	password = string(data)

	if cmd.Use == fmt.Sprintf(cmdUseFmtValidate, cmdUseValidate) {
		fmt.Println("")

		return
	}

	if noConfirm, err = cmd.Flags().GetBool(cmdFlagNameNoConfirm); err == nil && !noConfirm {
		if data, err = termReadPasswordWithPrompt("Confirm Password: "); err != nil {
			err = fmt.Errorf("failed to read the password from the terminal: %w", err)

			return
		}

		if password != string(data) {
			fmt.Println("")

			err = fmt.Errorf("the password did not match the confirmation password")

			return
		}
	}

	fmt.Println("")

	return
}

// ErrStdinIsNotTerminal is returned when Stdin is not an interactive terminal.
var ErrStdinIsNotTerminal = errors.New("stdin is not a terminal")

func termReadPasswordWithPrompt(prompt string) (data []byte, err error) {
	fd := int(syscall.Stdin) //nolint:unconvert,nolintlint

	if isTerm := term.IsTerminal(fd); !isTerm {
		return nil, ErrStdinIsNotTerminal
	}

	fmt.Print(prompt)

	if data, err = term.ReadPassword(fd); err != nil {
		return nil, err
	}

	fmt.Println("")

	return data, nil
}

func cmdFlagConfig(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceP(cmdFlagNameConfig, "c", []string{"configuration.yml"}, "configuration files to load")
}

func cmdFlagPassword(cmd *cobra.Command, noConfirm bool) {
	cmd.PersistentFlags().String(cmdFlagNamePassword, "", "manually supply the password rather than using the terminal prompt")

	if noConfirm {
		cmd.PersistentFlags().Bool(cmdFlagNameNoConfirm, false, "skip the password confirmation prompt")
	}
}

func cmdFlagRandomPassword(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(cmdFlagNameRandom, false, "uses a randomly generated password")
	cmd.PersistentFlags().String(cmdFlagNameRandomCharSet, cmdFlagValueCharSet, cmdFlagUsageCharset)
	cmd.PersistentFlags().String(cmdFlagNameRandomCharacters, "", cmdFlagUsageCharacters)
	cmd.PersistentFlags().Int(cmdFlagNameRandomLength, 72, cmdFlagUsageLength)
}

func cmdFlagIterations(cmd *cobra.Command, value int) {
	cmd.Flags().IntP(cmdFlagNameIterations, "i", value, "number of iterations")
}

func cmdFlagKeySize(cmd *cobra.Command, value int) {
	cmd.Flags().IntP(cmdFlagNameKeySize, "k", value, "key size in bytes")
}

func cmdFlagSaltSize(cmd *cobra.Command, value int) {
	cmd.Flags().IntP(cmdFlagNameSaltSize, "s", value, "salt size in bytes")
}

func cmdFlagParallelism(cmd *cobra.Command, value int) {
	cmd.Flags().IntP(cmdFlagNameParallelism, "p", value, "parallelism or threads")
}
