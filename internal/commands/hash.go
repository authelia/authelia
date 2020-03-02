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
		fmt.Println(authentication.HashPassword(args[0], "", authentication.HashingAlgorithmArgon2id, 3, 64*1024, 2, 16))
	},
	Args: cobra.MinimumNArgs(1),
}
