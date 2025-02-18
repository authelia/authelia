package commands

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm"
	"github.com/spf13/cobra"

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

func newCryptoHashGenerateCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	defaults := map[string]any{
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
		prefixFilePassword + suffixBCryptVariant:       schema.DefaultPasswordConfig.BCrypt.Variant,
		prefixFilePassword + suffixBCryptCost:          schema.DefaultPasswordConfig.BCrypt.Cost,
		prefixFilePassword + suffixSCryptIterations:    schema.DefaultPasswordConfig.SCrypt.Iterations,
		prefixFilePassword + suffixSCryptBlockSize:     schema.DefaultPasswordConfig.SCrypt.BlockSize,
		prefixFilePassword + suffixSCryptParallelism:   schema.DefaultPasswordConfig.SCrypt.Parallelism,
		prefixFilePassword + suffixSCryptKeyLength:     schema.DefaultPasswordConfig.SCrypt.KeyLength,
		prefixFilePassword + suffixSCryptSaltLength:    schema.DefaultPasswordConfig.SCrypt.SaltLength,
	}

	cmd = &cobra.Command{
		Use:     cmdUseGenerate,
		Short:   cmdAutheliaCryptoHashGenerateShort,
		Long:    cmdAutheliaCryptoHashGenerateLong,
		Example: cmdAutheliaCryptoHashGenerateExample,
		Args:    cobra.NoArgs,
		PreRunE: ctx.ChainRunE(
			ctx.HelperConfigSetDefaultsRunE(defaults),
			ctx.CryptoHashGenerateMapFlagsRunE,
			ctx.HelperConfigLoadRunE,
			ctx.ConfigValidateSectionPasswordRunE,
		),
		RunE: ctx.CryptoHashGenerateRunE,

		DisableAutoGenTag: true,
	}

	cmdFlagPassword(cmd, true)
	cmdFlagRandomPassword(cmd)

	for _, use := range []string{cmdUseHashArgon2, cmdUseHashSHA2Crypt, cmdUseHashPBKDF2, cmdUseHashBCrypt, cmdUseHashSCrypt} {
		cmd.AddCommand(newCryptoHashGenerateSubCmd(ctx, use))
	}

	return cmd
}

func newCryptoHashGenerateSubCmd(ctx *CmdCtx, use string) (cmd *cobra.Command) {
	defaults := map[string]any{
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
		prefixFilePassword + suffixBCryptVariant:       schema.DefaultPasswordConfig.BCrypt.Variant,
		prefixFilePassword + suffixBCryptCost:          schema.DefaultPasswordConfig.BCrypt.Cost,
		prefixFilePassword + suffixSCryptIterations:    schema.DefaultPasswordConfig.SCrypt.Iterations,
		prefixFilePassword + suffixSCryptBlockSize:     schema.DefaultPasswordConfig.SCrypt.BlockSize,
		prefixFilePassword + suffixSCryptParallelism:   schema.DefaultPasswordConfig.SCrypt.Parallelism,
		prefixFilePassword + suffixSCryptKeyLength:     schema.DefaultPasswordConfig.SCrypt.KeyLength,
		prefixFilePassword + suffixSCryptSaltLength:    schema.DefaultPasswordConfig.SCrypt.SaltLength,
	}

	useFmt := fmtCryptoHashUse(use)

	cmd = &cobra.Command{
		Use:     use,
		Short:   fmt.Sprintf(fmtCmdAutheliaCryptoHashGenerateSubShort, useFmt),
		Long:    fmt.Sprintf(fmtCmdAutheliaCryptoHashGenerateSubLong, useFmt, useFmt),
		Example: fmt.Sprintf(fmtCmdAutheliaCryptoHashGenerateSubExample, use),
		Args:    cobra.NoArgs,
		PersistentPreRunE: ctx.ChainRunE(
			ctx.HelperConfigSetDefaultsRunE(defaults),
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
		cmdFlagIterations(cmd, schema.DefaultPasswordConfig.PBKDF2.Iterations)
		cmdFlagSaltSize(cmd, schema.DefaultPasswordConfig.PBKDF2.SaltLength)

		cmd.Flags().StringP(cmdFlagNameVariant, "v", schema.DefaultPasswordConfig.PBKDF2.Variant, "variant, options are 'sha1', 'sha224', 'sha256', 'sha384', and 'sha512'")
	case cmdUseHashBCrypt:
		cmd.Flags().StringP(cmdFlagNameVariant, "v", schema.DefaultPasswordConfig.BCrypt.Variant, "variant, options are 'standard' and 'sha256'")
		cmd.Flags().IntP(cmdFlagNameCost, "i", schema.DefaultPasswordConfig.BCrypt.Cost, "hashing cost")
	case cmdUseHashSCrypt:
		cmdFlagIterations(cmd, schema.DefaultPasswordConfig.SCrypt.Iterations)
		cmdFlagKeySize(cmd, schema.DefaultPasswordConfig.SCrypt.KeyLength)
		cmdFlagSaltSize(cmd, schema.DefaultPasswordConfig.SCrypt.SaltLength)
		cmdFlagParallelism(cmd, schema.DefaultPasswordConfig.SCrypt.Parallelism)

		cmd.Flags().IntP(cmdFlagNameBlockSize, "r", schema.DefaultPasswordConfig.SCrypt.BlockSize, "block size")
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
}

// CryptoHashGenerateMapFlagsRunE is the RunE which configures the flags map configuration source for the
// authelia crypto hash generate commands.
func (ctx *CmdCtx) CryptoHashGenerateMapFlagsRunE(cmd *cobra.Command, args []string) (err error) {
	var flagsMap map[string]string

	switch cmd.Use {
	case cmdUseHashArgon2:
		flagsMap = map[string]string{
			cmdFlagNameVariant:     prefixFilePassword + suffixArgon2Variant,
			cmdFlagNameIterations:  prefixFilePassword + suffixArgon2Iterations,
			cmdFlagNameMemory:      prefixFilePassword + suffixArgon2Memory,
			cmdFlagNameParallelism: prefixFilePassword + suffixArgon2Parallelism,
			cmdFlagNameKeySize:     prefixFilePassword + suffixArgon2KeyLength,
			cmdFlagNameSaltSize:    prefixFilePassword + suffixArgon2SaltLength,
		}
	case cmdUseHashSHA2Crypt:
		flagsMap = map[string]string{
			cmdFlagNameVariant:    prefixFilePassword + suffixSHA2CryptVariant,
			cmdFlagNameIterations: prefixFilePassword + suffixSHA2CryptIterations,
			cmdFlagNameSaltSize:   prefixFilePassword + suffixSHA2CryptSaltLength,
		}
	case cmdUseHashPBKDF2:
		flagsMap = map[string]string{
			cmdFlagNameVariant:    prefixFilePassword + suffixPBKDF2Variant,
			cmdFlagNameIterations: prefixFilePassword + suffixPBKDF2Iterations,
			cmdFlagNameKeySize:    prefixFilePassword + suffixPBKDF2KeyLength,
			cmdFlagNameSaltSize:   prefixFilePassword + suffixPBKDF2SaltLength,
		}
	case cmdUseHashBCrypt:
		flagsMap = map[string]string{
			cmdFlagNameVariant: prefixFilePassword + suffixBCryptVariant,
			cmdFlagNameCost:    prefixFilePassword + suffixBCryptCost,
		}
	case cmdUseHashSCrypt:
		flagsMap = map[string]string{
			cmdFlagNameIterations:  prefixFilePassword + suffixSCryptIterations,
			cmdFlagNameBlockSize:   prefixFilePassword + suffixSCryptBlockSize,
			cmdFlagNameParallelism: prefixFilePassword + suffixSCryptParallelism,
			cmdFlagNameKeySize:     prefixFilePassword + suffixSCryptKeyLength,
			cmdFlagNameSaltSize:    prefixFilePassword + suffixSCryptSaltLength,
		}
	}

	if flagsMap != nil {
		ctx.cconfig.sources = append(ctx.cconfig.sources, configuration.NewCommandLineSourceWithMapping(cmd.Flags(), flagsMap, false, false))
	}

	return nil
}

// CryptoHashGenerateRunE is the RunE for the authelia crypto hash generate commands.
func (ctx *CmdCtx) CryptoHashGenerateRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		hash     algorithm.Hash
		digest   algorithm.Digest
		password string
		random   bool
	)

	if password, random, err = cmdCryptoHashGetPassword(cmd, args, false, true); err != nil {
		return err
	}

	if len(password) == 0 {
		return fmt.Errorf("no password provided")
	}

	switch cmd.Use {
	case cmdUseGenerate:
		break
	default:
		ctx.config.AuthenticationBackend.File.Password.Algorithm = cmd.Use
	}

	if hash, err = authentication.NewCryptoHashFromConfig(ctx.config.AuthenticationBackend.File.Password); err != nil {
		return err
	}

	if digest, err = hash.Hash(password); err != nil {
		return err
	}

	if random {
		fmt.Printf("Random Password: %s\n", password)

		if value := url.QueryEscape(password); password != value {
			fmt.Printf("Random Password (URL Encoded): %s\n", value)
		}
	}

	fmt.Printf("Digest: %s\n", digest.Encode())

	return nil
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
		noConfirm bool
	)

	if password, err = termReadPasswordWithPrompt("Enter Password: ", "password"); err != nil {
		err = fmt.Errorf("failed to read the password from the terminal: %w", err)

		return
	}

	if cmd.Use == fmt.Sprintf(cmdUseFmtValidate, cmdUseValidate) {
		fmt.Println("")

		return
	}

	if noConfirm, err = cmd.Flags().GetBool(cmdFlagNameNoConfirm); err == nil && !noConfirm {
		var confirm string

		if confirm, err = termReadPasswordWithPrompt("Confirm Password: ", ""); err != nil {
			return
		}

		if password != confirm {
			fmt.Println("")

			err = fmt.Errorf("the password did not match the confirmation password")

			return
		}
	}

	fmt.Println("")

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
