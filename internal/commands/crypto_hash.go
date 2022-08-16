package commands

import (
	"fmt"

	"github.com/go-crypt/crypt"
	"github.com/spf13/cobra"
)

func newCryptoHashCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseHash,
		Short:   cmdAutheliaCryptoHashShort,
		Long:    cmdAutheliaCryptoHashLong,
		Example: cmdAutheliaCryptoHashExample,
		Args:    cobra.NoArgs,
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
	}

	for _, use := range []string{cmdUseHashArgon2, cmdUseHashSHA2Crypt, cmdUseHashPBKDF2, cmdUseHashBCrypt, cmdUseHashSCrypt} {
		cmd.AddCommand(newCryptoHashGenerateSubCmd(use))
	}

	return cmd
}

func newCryptoHashGenerateSubCmd(use string) (cmd *cobra.Command) {
	var (
		useFmt string
	)

	useFmt = fmtCryptoHashUse(use)

	cmd = &cobra.Command{
		Use:     use,
		Short:   fmt.Sprintf(fmtCmdAutheliaCryptoHashGenerateSubShort, useFmt),
		Long:    fmt.Sprintf(fmtCmdAutheliaCryptoHashGenerateSubLong, useFmt, useFmt),
		Example: fmt.Sprintf(fmtCmdAutheliaCryptoHashGenerateSubExample, use),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	switch use {
	case cmdUseHashArgon2:
		cmdFlagIterations(cmd, 3)
		cmdFlagParallelism(cmd, 4)
		cmdFlagKeySize(cmd)
		cmdFlagSaltSize(cmd)

		cmd.Flags().StringP(cmdFlagNameVariant, "v", "id", "variant, options are 'id', 'i', and 'd'")
		cmd.Flags().Uint32P(cmdFlagNameMemory, "m", 65536, "memory in kibibytes")
		cmd.Flags().String(cmdFlagNameProfile, "low-memory", "profile to use, options are low-memory and recommended")

		cmd.RunE = cryptoHashGenerateArgon2RunE
	case cmdUseHashSHA2Crypt:
		cmdFlagIterations(cmd, 150000)
		cmdFlagSaltSize(cmd)

		cmd.Flags().StringP(cmdFlagNameVariant, "v", "sha512", "variant, options are sha256 and sha512")

		cmd.RunE = cryptoHashGenerateSHA2CryptRunE
	case cmdUseHashPBKDF2:
		cmdFlagIterations(cmd, 120000)
		cmdFlagKeySize(cmd)
		cmdFlagSaltSize(cmd)

		cmd.Flags().StringP(cmdFlagNameVariant, "v", "sha512", "variant, options are 'sha1', 'sha224', 'sha256', 'sha384', and 'sha512'")

		cmd.RunE = cryptoHashGeneratePBKDF2RunE
	case cmdUseHashBCrypt:
		cmd.Flags().StringP(cmdFlagNameVariant, "v", "standard", "variant, options are 'standard' and 'sha256'")
		cmd.Flags().Uint8P(cmdFlagNameCost, "c", 13, "hashing cost")

		cmd.RunE = cryptoHashGenerateBCryptRunE
	case cmdUseHashSCrypt:
		cmdFlagIterations(cmd, 16)
		cmdFlagKeySize(cmd)
		cmdFlagSaltSize(cmd)
		cmdFlagParallelism(cmd, 1)

		cmd.Flags().Uint32P(cmdFlagNameBlockSize, "r", 8, "block size")

		cmd.RunE = cryptoHashGenerateSCryptRunE
	}

	return cmd
}

func cryptoHashGenerateArgon2RunE(cmd *cobra.Command, args []string) (err error) {
	var profile, password, variant string

	password = args[0]

	hash := crypt.NewArgon2Hash()

	var t, m, p, k, s int

	if t, err = cmd.Flags().GetInt(cmdFlagNameIterations); err != nil {
		return err
	}

	if m, err = cmd.Flags().GetInt(cmdFlagNameMemory); err != nil {
		return err
	}

	if p, err = cmd.Flags().GetInt(cmdFlagNameParallelism); err != nil {
		return err
	}

	if k, err = cmd.Flags().GetInt(cmdFlagNameKeySize); err != nil {
		return err
	}

	if s, err = cmd.Flags().GetInt(cmdFlagNameSaltSize); err != nil {
		return err
	}

	if variant, err = cmd.Flags().GetString(cmdFlagNameVariant); err != nil {
		return err
	}

	switch variant {
	case "id", "i", "d":
		break
	default:
		return fmt.Errorf("variant '%s' is not valid: valid variants are 'id', 'i', and 'd'", variant)
	}

	hash.WithVariant(crypt.NewArgon2Variant("argon2" + variant))

	if profile, err = cmd.Flags().GetString(cmdFlagNameProfile); err != nil {
		return err
	}

	switch profile {
	case "low-memory":
		hash.WithProfile(crypt.Argon2ProfileRFC9106LowMemory)
	case "recommended":
		hash.WithProfile(crypt.Argon2ProfileRFC9106Recommended)
	default:
		return fmt.Errorf("profile '%s' is not valid: valid profiles are 'low-memory' and 'recommended'", profile)
	}

	if cmd.Flags().Changed(cmdFlagNameProfile) {
		if cmd.Flags().Changed(cmdFlagNameIterations) {
			hash.WithT(t)
		}

		if cmd.Flags().Changed(cmdFlagNameMemory) {
			hash.WithM(m)
		}

		if cmd.Flags().Changed(cmdFlagNameParallelism) {
			hash.WithP(p)
		}

		if cmd.Flags().Changed(cmdFlagNameKeySize) {
			hash.WithK(k)
		}

		if cmd.Flags().Changed(cmdFlagNameSaltSize) {
			hash.WithS(s)
		}
	} else {
		hash.WithT(t).WithM(m).WithP(p).WithK(k).WithS(s)
	}

	var digest crypt.Digest

	if digest, err = hash.Hash(password); err != nil {
		return err
	}

	fmt.Println(digest.Encode())

	return nil
}

func cryptoHashGenerateSHA2CryptRunE(cmd *cobra.Command, args []string) (err error) {
	var password, variant string

	password = args[0]

	hash := crypt.NewSHA2CryptHash()

	var i, s int

	if i, err = cmd.Flags().GetInt(cmdFlagNameIterations); err != nil {
		return err
	}

	if s, err = cmd.Flags().GetInt(cmdFlagNameSaltSize); err != nil {
		return err
	}

	if variant, err = cmd.Flags().GetString(cmdFlagNameVariant); err != nil {
		return err
	}

	switch variant {
	case "sha512", "sha256", "6", "5":
		break
	default:
		return fmt.Errorf("variant '%s' is not valid: valid variants are 'sha512' and 'sha256'", variant)
	}

	hash.WithVariant(crypt.NewSHA2CryptVariant(variant)).WithRounds(i).WithSaltLength(s)

	var digest crypt.Digest

	if digest, err = hash.Hash(password); err != nil {
		return err
	}

	fmt.Println(digest.Encode())

	return nil
}

func cryptoHashGeneratePBKDF2RunE(cmd *cobra.Command, args []string) (err error) {
	var password, variant string

	password = args[0]

	hash := crypt.NewPBKDF2Hash()

	var i, k, s int

	if i, err = cmd.Flags().GetInt(cmdFlagNameIterations); err != nil {
		return err
	}

	if k, err = cmd.Flags().GetInt(cmdFlagNameKeySize); err != nil {
		return err
	}

	if s, err = cmd.Flags().GetInt(cmdFlagNameSaltSize); err != nil {
		return err
	}

	if variant, err = cmd.Flags().GetString(cmdFlagNameVariant); err != nil {
		return err
	}

	switch variant {
	case "sha1", "sha224", "sha256", "sha384", "sha512":
		break
	default:
		return fmt.Errorf("variant '%s' is not valid: valid variants are 'sha1', 'sha224', 'sha256', 'sha385', and 'sha512'", variant)
	}

	hash.WithVariant(crypt.NewPBKDF2Variant(variant)).WithIterations(i).WithKeyLength(k).WithSaltLength(s)

	var digest crypt.Digest

	if digest, err = hash.Hash(password); err != nil {
		return err
	}

	fmt.Println(digest.Encode())

	return nil
}

func cryptoHashGenerateBCryptRunE(cmd *cobra.Command, args []string) (err error) {
	var password, variant string

	password = args[0]

	hash := crypt.NewBcryptHash()

	var i uint32

	if i, err = cmd.Flags().GetUint32(cmdFlagNameIterations); err != nil {
		return err
	}

	if variant, err = cmd.Flags().GetString(cmdFlagNameVariant); err != nil {
		return err
	}

	switch variant {
	case "standard", "sha256":
		break
	default:
		return fmt.Errorf("variant '%s' is not valid: valid variants are 'sha1', 'sha224', 'sha256', 'sha385', and 'sha512'", variant)
	}

	hash.WithVariant(crypt.NewBcryptVariant(variant)).WithCost(int(i))

	var digest crypt.Digest

	if digest, err = hash.Hash(password); err != nil {
		return err
	}

	fmt.Println(digest.Encode())

	return nil
}

func cryptoHashGenerateSCryptRunE(cmd *cobra.Command, args []string) (err error) {
	var password string

	password = args[0]

	hash := crypt.NewScryptHash()

	var ln, r, p, k, s int

	if ln, err = cmd.Flags().GetInt(cmdFlagNameIterations); err != nil {
		return err
	}

	if r, err = cmd.Flags().GetInt(cmdFlagNameBlockSize); err != nil {
		return err
	}

	if p, err = cmd.Flags().GetInt(cmdFlagNameParallelism); err != nil {
		return err
	}

	if k, err = cmd.Flags().GetInt(cmdFlagNameKeySize); err != nil {
		return err
	}

	if s, err = cmd.Flags().GetInt(cmdFlagNameSaltSize); err != nil {
		return err
	}

	hash.WithLN(ln).WithR(r).WithP(p).WithKeySize(k).WithSaltSize(s)

	var digest crypt.Digest

	if digest, err = hash.Hash(password); err != nil {
		return err
	}

	fmt.Println(digest.Encode())

	return nil
}

func newCryptoHashValidateCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [flags] <digest> -- <password>", cmdUseValidate),
		Short:   cmdAutheliaCryptoHashValidateShort,
		Long:    cmdAutheliaCryptoHashValidateLong,
		Example: cmdAutheliaCryptoHashValidateExample,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var valid bool

			if valid, err = crypt.CheckPassword(args[1], args[0]); err != nil {
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
	}

	return cmd
}

func cmdFlagIterations(cmd *cobra.Command, value uint32) {
	cmd.Flags().Uint32P(cmdFlagNameIterations, "i", value, "number of iterations")
}

func cmdFlagKeySize(cmd *cobra.Command) {
	cmd.Flags().Uint32P(cmdFlagNameKeySize, "k", 32, "key size in bytes")
}

func cmdFlagSaltSize(cmd *cobra.Command) {
	cmd.Flags().Uint32P(cmdFlagNameSaltSize, "s", 16, "salt size in bytes")
}

func cmdFlagParallelism(cmd *cobra.Command, value uint32) {
	cmd.Flags().Uint32P(cmdFlagNameParallelism, "p", value, "parallelism or threads")
}
