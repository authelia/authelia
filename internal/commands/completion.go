package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  $ source <(authelia completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ authelia completion bash > /etc/bash_completion.d/authelia
  # macOS:
  $ authelia completion bash > /usr/local/etc/bash_completion.d/authelia

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ authelia completion zsh > "${fpath[1]}/_authelia"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ authelia completion fish | source

  # To load completions for each session, execute once:
  $ authelia completion fish > ~/.config/fish/completions/authelia.fish

PowerShell:

  PS> authelia completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> authelia completion powershell > authelia.ps1
  # and source this file from your PowerShell profile.
`,
		Args:                  cobra.ExactValidArgs(1),
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		DisableFlagsInUseLine: true,
		RunE:                  cmdCompletionRunE,
	}

	return cmd
}

func cmdCompletionRunE(cmd *cobra.Command, args []string) (err error) {
	switch args[0] {
	case "bash":
		return cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	}

	return fmt.Errorf("Invalid shell provided for completion command")
}
