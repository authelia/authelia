package commands

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newCryptoHashCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseHash,
		Short:   cmdAutheliaCryptoHashShort,
		Long:    cmdAutheliaCryptoHashLong,
		Example: cmdAutheliaCryptoHashExample,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newCryptoHashValidateCmd(ctx),
		newCryptoHashGenerateCmd(ctx),
	)

	return cmd
}

func newCryptoHashDefaults() (defaults map[string]any) {
	return map[string]any{
		prefixFilePassword + suffixAlgorithm:           schema.DefaultPasswordConfig.Algorithm,
		prefixFilePassword + suffixArgon2Variant:       schema.DefaultPasswordConfig.Argon2.Variant,
		prefixFilePassword + suffixArgon2Iterations:    schema.DefaultPasswordConfig.Argon2.Iterations,
		prefixFilePassword + suffixArgon2Memory:        schema.DefaultPasswordConfig.Argon2.Memory,
		prefixFilePassword + suffixArgon2Parallelism:   schema.DefaultPasswordConfig.Argon2.Parallelism,
		prefixFilePassword + suffixArgon2KeyLength:     schema.DefaultPasswordConfig.Argon2.KeyLength,
		prefixFilePassword + suffixArgon2SaltLength:    schema.DefaultPasswordConfig.Argon2.SaltLength,
		prefixFilePassword + suffixSHA2CryptVariant:    schema.DefaultPasswordConfig.SHA2Crypt.Variant,
		prefixFilePassword + suffixSHA2CryptIterations: schema.DefaultPasswordConfig.SHA2Crypt.Iterations,
		prefixFilePassword + suffixSHA2CryptSaltLength: schema.DefaultPasswordConfig.SHA2Crypt.SaltLength,
		prefixFilePassword + suffixPBKDF2Variant:       schema.DefaultPasswordConfig.PBKDF2.Variant,
		prefixFilePassword + suffixPBKDF2Iterations:    schema.DefaultPasswordConfig.PBKDF2.Iterations,
		prefixFilePassword + suffixPBKDF2SaltLength:    schema.DefaultPasswordConfig.PBKDF2.SaltLength,
		prefixFilePassword + suffixBcryptVariant:       schema.DefaultPasswordConfig.Bcrypt.Variant,
		prefixFilePassword + suffixBcryptCost:          schema.DefaultPasswordConfig.Bcrypt.Cost,
		prefixFilePassword + suffixScryptVariant:       schema.DefaultPasswordConfig.Scrypt.Variant,
		prefixFilePassword + suffixScryptIterations:    schema.DefaultPasswordConfig.Scrypt.Iterations,
		prefixFilePassword + suffixScryptBlockSize:     schema.DefaultPasswordConfig.Scrypt.BlockSize,
		prefixFilePassword + suffixScryptParallelism:   schema.DefaultPasswordConfig.Scrypt.Parallelism,
		prefixFilePassword + suffixScryptKeyLength:     schema.DefaultPasswordConfig.Scrypt.KeyLength,
		prefixFilePassword + suffixScryptSaltLength:    schema.DefaultPasswordConfig.Scrypt.SaltLength,
	}
}

func newCryptoHashGenerateCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseGenerate,
		Short:   cmdAutheliaCryptoHashGenerateShort,
		Long:    cmdAutheliaCryptoHashGenerateLong,
		Example: cmdAutheliaCryptoHashGenerateExample,
		Args:    cobra.NoArgs,
		PreRunE: ctx.ChainRunE(
			ctx.HelperConfigSetDefaultsRunE(newCryptoHashDefaults()),
			ctx.CryptoHashGenerateMapFlagsRunE,
			ctx.HelperConfigLoadRunE,
			ctx.ConfigValidateSectionPasswordRunE,
		),
		RunE: ctx.CryptoHashGenerateRunE,

		DisableAutoGenTag: true,
	}

	cmdFlagPassword(cmd, true)
	cmdFlagRandomPassword(cmd)

	for _, use := range []string{cmdUseHashArgon2, cmdUseHashSHA2Crypt, cmdUseHashPBKDF2, cmdUseHashBcrypt, cmdUseHashScrypt} {
		cmd.AddCommand(newCryptoHashGenerateSubCmd(ctx, use))
	}

	return cmd
}

func newCryptoHashGenerateSubCmd(ctx *CmdCtx, use string) (cmd *cobra.Command) {
	useFmt := fmtCryptoHashUse(use)

	cmd = &cobra.Command{
		Use:     use,
		Short:   fmt.Sprintf(fmtCmdAutheliaCryptoHashGenerateSubShort, useFmt),
		Long:    fmt.Sprintf(fmtCmdAutheliaCryptoHashGenerateSubLong, useFmt, useFmt),
		Example: fmt.Sprintf(fmtCmdAutheliaCryptoHashGenerateSubExample, use),
		Args:    cobra.NoArgs,
		PersistentPreRunE: ctx.ChainRunE(
			ctx.HelperConfigSetDefaultsRunE(newCryptoHashDefaults()),
			ctx.CryptoHashGenerateMapFlagsRunE,
			ctx.HelperConfigLoadRunE,
			ctx.ConfigValidateSectionPasswordRunE,
		),
		RunE: ctx.CryptoHashGenerateRunE,

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
	case cmdUseHashSHA2Crypt:
		cmdFlagIterations(cmd, schema.DefaultPasswordConfig.SHA2Crypt.Iterations)
		cmdFlagSaltSize(cmd, schema.DefaultPasswordConfig.SHA2Crypt.SaltLength)

		cmd.Flags().StringP(cmdFlagNameVariant, "v", schema.DefaultPasswordConfig.SHA2Crypt.Variant, "variant, options are sha256 and sha512")
		cmd.PreRunE = ctx.ChainRunE()
	case cmdUseHashPBKDF2:
		cmdFlagIterationsWithUsage(cmd, 0, "number of iterations (default is determined by the variant)")
		cmdFlagSaltSize(cmd, schema.DefaultPasswordConfig.PBKDF2.SaltLength)

		cmd.Flags().StringP(cmdFlagNameVariant, "v", schema.DefaultPasswordConfig.PBKDF2.Variant, "variant, options are 'sha1', 'sha224', 'sha256', 'sha384', and 'sha512'")
	case cmdUseHashBcrypt:
		cmd.Flags().StringP(cmdFlagNameVariant, "v", schema.DefaultPasswordConfig.Bcrypt.Variant, "variant, options are 'standard' and 'sha256'")
		cmd.Flags().IntP(cmdFlagNameCost, "i", schema.DefaultPasswordConfig.Bcrypt.Cost, "hashing cost")
	case cmdUseHashScrypt:
		cmdFlagIterations(cmd, schema.DefaultPasswordConfig.Scrypt.Iterations)
		cmdFlagKeySize(cmd, schema.DefaultPasswordConfig.Scrypt.KeyLength)
		cmdFlagSaltSize(cmd, schema.DefaultPasswordConfig.Scrypt.SaltLength)
		cmdFlagParallelism(cmd, schema.DefaultPasswordConfig.Scrypt.Parallelism)

		cmd.Flags().StringP(cmdFlagNameVariant, "v", schema.DefaultPasswordConfig.Scrypt.Variant, "variant, options are 'scrypt', and 'yescrypt'")
		cmd.Flags().IntP(cmdFlagNameBlockSize, "r", schema.DefaultPasswordConfig.Scrypt.BlockSize, "block size")
	}

	return cmd
}

func newCryptoHashValidateCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     fmt.Sprintf(cmdUseFmtValidate, cmdUseValidate),
		Short:   cmdAutheliaCryptoHashValidateShort,
		Long:    cmdAutheliaCryptoHashValidateLong,
		Example: cmdAutheliaCryptoHashValidateExample,
		Args:    cobra.ExactArgs(1),
		RunE:    ctx.CryptoHashValidateRunE,

		DisableAutoGenTag: true,
	}

	cmdFlagPassword(cmd, false)

	return cmd
}

// CryptoHashValidateRunE is the RunE for the authelia crypto hash validate command.
func (ctx *CmdCtx) CryptoHashValidateRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		password string
		valid    bool
	)

	if password, _, err = cmdFlagsCryptoHashGetPassword(cmd.OutOrStdout(), cmd.Flags(), cmd.Use, args, false, false); err != nil {
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
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "The password matches the digest.\n")
	default:
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "The password does not match the digest.\n")
	}

	return nil
}

// CryptoHashGenerateMapFlagsRunE is the RunE which configures the flags map configuration source for the
// authelia crypto hash generate commands.
func (ctx *CmdCtx) CryptoHashGenerateMapFlagsRunE(cmd *cobra.Command, args []string) (err error) {
	flagsMap := getCryptoHashGenerateMapFlagsFromUse(cmd.Use)

	if flagsMap != nil {
		ctx.cconfig.sources = append(ctx.cconfig.sources, configuration.NewCommandLineSourceWithMapping(cmd.Flags(), flagsMap, false, false))
	}

	return nil
}

// CryptoHashGenerateRunE is the RunE for the authelia crypto hash generate commands.
func (ctx *CmdCtx) CryptoHashGenerateRunE(cmd *cobra.Command, args []string) (err error) {
	return runCryptoHashGenerate(cmd.OutOrStdout(), cmd.Flags(), cmd.Use, args, ctx.config)
}

func runCryptoHashGenerate(w io.Writer, flags *pflag.FlagSet, use string, args []string, config *schema.Configuration) (err error) {
	var (
		hash     algorithm.Hash
		digest   algorithm.Digest
		password string
		random   bool
	)

	if config.AuthenticationBackend.File == nil {
		return fmt.Errorf("authentication backend file is not configured")
	}

	if password, random, err = cmdFlagsCryptoHashGetPassword(w, flags, use, args, false, true); err != nil {
		return err
	}

	if len(password) == 0 {
		return fmt.Errorf("no password provided")
	}

	switch use {
	case cmdUseGenerate:
		break
	default:
		config.AuthenticationBackend.File.Password.Algorithm = use
	}

	if hash, err = authentication.NewFileCryptoHashFromConfig(config.AuthenticationBackend.File.Password); err != nil {
		return err
	}

	if digest, err = hash.Hash(password); err != nil {
		return err
	}

	if random {
		_, _ = fmt.Fprintf(w, "Random Password: %s\n", password)

		if value := url.QueryEscape(password); password != value {
			_, _ = fmt.Fprintf(w, "Random Password (URL Encoded): %s\n", value)
		}
	}

	_, _ = fmt.Fprintf(w, "Digest: %s\n", digest.Encode())

	return nil
}

func cmdFlagsCryptoHashPasswordRandom(flags *pflag.FlagSet, flagNameRandom string, flagsSetters ...string) (random bool, err error) {
	if random, err = flags.GetBool(flagNameRandom); err != nil {
		return false, err
	}

	if random {
		return true, nil
	}

	for _, setter := range flagsSetters {
		if flags.Changed(setter) {
			return true, nil
		}
	}

	return false, nil
}

func cmdFlagsCryptoHashGetPassword(w io.Writer, flags *pflag.FlagSet, use string, args []string, useArgs, useRandom bool) (password string, random bool, err error) {
	if useRandom {
		if random, err = cmdFlagsCryptoHashPasswordRandom(flags, cmdFlagNameRandom, cmdFlagNameRandomCharSet, cmdFlagNameRandomCharacters, cmdFlagNameRandomLength); err != nil {
			return
		}
	}

	switch {
	case random:
		password, err = flagsGetRandomCharacters(flags, cmdFlagNameRandomLength, cmdFlagNameRandomCharSet, cmdFlagNameRandomCharacters)

		return
	case flags.Changed(cmdFlagNamePassword):
		password, err = flags.GetString(cmdFlagNamePassword)

		return
	case useArgs && len(args) != 0:
		password, err = strings.Join(args, " "), nil

		return
	}

	var (
		noConfirm bool
	)

	if password, err = termReadPasswordWithPrompt("Enter Password: ", "password"); err != nil {
		err = fmt.Errorf("failed to read the password from the terminal: %w", err)

		return
	}

	if use == fmt.Sprintf(cmdUseFmtValidate, cmdUseValidate) {
		_, _ = fmt.Fprintln(w)

		return
	}

	if noConfirm, err = flags.GetBool(cmdFlagNameNoConfirm); err == nil && !noConfirm {
		var confirm string

		if confirm, err = termReadPasswordWithPrompt("Confirm Password: ", ""); err != nil {
			return
		}

		if password != confirm {
			_, _ = fmt.Fprintln(w)

			err = fmt.Errorf("the password did not match the confirmation password")

			return
		}
	}

	_, _ = fmt.Fprintln(w)

	return
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
	cmdFlagIterationsWithUsage(cmd, value, "number of iterations")
}

func cmdFlagIterationsWithUsage(cmd *cobra.Command, value int, usage string) {
	cmd.Flags().IntP(cmdFlagNameIterations, "i", value, usage)
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
