package commands

import (
	"fmt"
	"github.com/authelia/authelia/internal/authentication"

	"github.com/spf13/cobra"
)

var HashPasswordCmd = &cobra.Command{
	Use:   "hash-password [password]",
	Short: "Hash a password to be used in file-based users database",
	Run: func(cobraCmd *cobra.Command, args []string) {
		var err error
		var hash string
		sha512, _ := cobraCmd.Flags().GetBool("sha512")
		times, _ := cobraCmd.Flags().GetInt("times")
		salt, _ := cobraCmd.Flags().GetString("salt")
		saltLength, _ := cobraCmd.Flags().GetInt("salt_length")
		memory, _ := cobraCmd.Flags().GetInt("memory")
		parallelism, _ := cobraCmd.Flags().GetInt("parallelism")

		if sha512 {
			hash, err = authentication.HashPassword(args[0], salt, authentication.HashingAlgorithmSHA512, times, memory, parallelism, saltLength)
		} else {
			hash, err = authentication.HashPassword(args[0], salt, authentication.HashingAlgorithmArgon2id, times, memory, parallelism, saltLength)
		}

		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(hash)
		}
	},
	Args: cobra.MinimumNArgs(1),
}
