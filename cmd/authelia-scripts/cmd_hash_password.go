package main

import (
	"fmt"

	"github.com/clems4ever/authelia/authentication"
	"github.com/spf13/cobra"
)

// HashPassword hash the provided password with crypt sha256 hash function
func HashPassword(cobraCmd *cobra.Command, args []string) {
	fmt.Println(authentication.HashPassword(args[0], nil))
}
