package commands

import (
	"fmt"

	"github.com/simia-tech/crypt"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
)

func newHashPasswordCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "hash-password [flags] -- <password>",
		Short:   cmdAutheliaHashPasswordShort,
		Long:    cmdAutheliaHashPasswordLong,
		Example: cmdAutheliaHashPasswordExample,
		Args:    cobra.MinimumNArgs(1),
		RunE:    cmdHashPasswordRunE,

		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolP("sha512", "z", false, fmt.Sprintf("use sha512 as the algorithm (changes iterations to %d, change with -i)", schema.DefaultPasswordSHA512Configuration.Iterations))
	cmd.Flags().IntP("iterations", "i", schema.DefaultPasswordConfiguration.Iterations, "set the number of hashing iterations")
	cmd.Flags().StringP("salt", "s", "", "set the salt string")
	cmd.Flags().IntP("memory", "m", schema.DefaultPasswordConfiguration.Memory, "[argon2id] set the amount of memory param (in MB)")
	cmd.Flags().IntP("parallelism", "p", schema.DefaultPasswordConfiguration.Parallelism, "[argon2id] set the parallelism param")
	cmd.Flags().IntP("key-length", "k", schema.DefaultPasswordConfiguration.KeyLength, "[argon2id] set the key length param")
	cmd.Flags().IntP("salt-length", "l", schema.DefaultPasswordConfiguration.SaltLength, "set the auto-generated salt length")
	cmd.Flags().StringSliceP("config", "c", []string{}, "Configuration files")

	return cmd
}

func cmdHashPasswordRunE(cmd *cobra.Command, args []string) (err error) {
	salt, _ := cmd.Flags().GetString("salt")
	sha512, _ := cmd.Flags().GetBool("sha512")
	configs, _ := cmd.Flags().GetStringSlice("config")

	mapDefaults := map[string]any{
		"authentication_backend.file.password.algorithm":   schema.DefaultPasswordConfiguration.Algorithm,
		"authentication_backend.file.password.iterations":  schema.DefaultPasswordConfiguration.Iterations,
		"authentication_backend.file.password.key_length":  schema.DefaultPasswordConfiguration.KeyLength,
		"authentication_backend.file.password.salt_length": schema.DefaultPasswordConfiguration.SaltLength,
		"authentication_backend.file.password.parallelism": schema.DefaultPasswordConfiguration.Parallelism,
		"authentication_backend.file.password.memory":      schema.DefaultPasswordConfiguration.Memory,
	}

	if sha512 {
		mapDefaults["authentication_backend.file.password.algorithm"] = schema.DefaultPasswordSHA512Configuration.Algorithm
		mapDefaults["authentication_backend.file.password.iterations"] = schema.DefaultPasswordSHA512Configuration.Iterations
		mapDefaults["authentication_backend.file.password.salt_length"] = schema.DefaultPasswordSHA512Configuration.SaltLength
	}

	mapCLI := map[string]string{
		"iterations":  "authentication_backend.file.password.iterations",
		"key-length":  "authentication_backend.file.password.key_length",
		"salt-length": "authentication_backend.file.password.salt_length",
		"parallelism": "authentication_backend.file.password.parallelism",
		"memory":      "authentication_backend.file.password.memory",
	}

	sources := configuration.NewDefaultSourcesWithDefaults(configs,
		configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter,
		configuration.NewMapSource(mapDefaults),
		configuration.NewCommandLineSourceWithMapping(cmd.Flags(), mapCLI, false, false),
	)

	val := schema.NewStructValidator()

	if _, config, err = configuration.Load(val, sources...); err != nil {
		return fmt.Errorf("error occurred loading configuration: %w", err)
	}

	var (
		hash      string
		algorithm authentication.CryptAlgo
	)

	p := config.AuthenticationBackend.File.Password

	switch p.Algorithm {
	case "sha512":
		algorithm = authentication.HashingAlgorithmSHA512
	default:
		algorithm = authentication.HashingAlgorithmArgon2id
	}

	validator.ValidatePasswordConfiguration(p, val)

	errs := val.Errors()

	if len(errs) != 0 {
		for i, e := range errs {
			if i == 0 {
				err = e
				continue
			}

			err = fmt.Errorf("%v, %w", err, e)
		}

		return fmt.Errorf("errors occurred validating the password configuration: %w", err)
	}

	if salt != "" {
		salt = crypt.Base64Encoding.EncodeToString([]byte(salt))
	}

	if hash, err = authentication.HashPassword(args[0], salt, algorithm, p.Iterations, p.Memory*1024, p.Parallelism, p.KeyLength, p.SaltLength); err != nil {
		return fmt.Errorf("error during password hashing: %w", err)
	}

	fmt.Printf("Password hash: %s\n", hash)

	return nil
}
