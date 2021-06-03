package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/commands"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/utils"
)

var configPathFlag string

func main() {
	logger := logging.Logger()
	rootCmd := &cobra.Command{
		Use: "authelia",
		Run: func(cmd *cobra.Command, args []string) {
			startServer()
		},
		Short: fmt.Sprintf("authelia %s", utils.VersionShort()),
		Long: fmt.Sprintf(`authelia %s

Authelia is an open-source authentication and authorization server providing 2-factor authentication and 
single sign-on (SSO) for your applications via a web portal. It acts as a companion of reverse proxies like 
nginx, Traefik or HAProxy to let them know whether queries should pass through.Unauthenticated users are 
redirected to Authelia Sign-in portal instead.

Documentation is available at https://www.authelia.com/docs.`, utils.VersionLong()),
	}

	rootCmd.Flags().StringVar(&configPathFlag, "config", "", "Configuration file")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version of Authelia",
		Run: func(cmd *cobra.Command, args []string) {
			long, err := cmd.Flags().GetBool("long")
			if err != nil {
				logger.Fatal(fmt.Errorf("Error parsing flag: %w", err))
			}

			if long {
				fmt.Printf("Authelia version %s\n", utils.VersionLong())

				return
			}

			fmt.Printf("Authelia version %s, build %s\n", utils.VersionShort(), utils.BuildCommit)
		},
	}

	versionCmd.Flags().Bool("long", false, "Toggles the long version output")

	versionAllCmd := &cobra.Command{
		Use:   "all",
		Short: "Show all version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Branch: %s\nLast Tag: %s\nCommit: %s\nBuild Date: %s\nState Tag: %s\nState Extra: %s\n",
				utils.BuildBranch, utils.BuildTag, utils.BuildCommit, utils.BuildDate, utils.BuildStateTag, utils.BuildStateExtra)
		},
	}

	versionCmd.AddCommand(versionAllCmd)

	rootCmd.AddCommand(versionCmd, commands.HashPasswordCmd,
		commands.ValidateConfigCmd, commands.CertificatesCmd,
		commands.RSACmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}
}
