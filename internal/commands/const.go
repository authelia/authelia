package commands

import (
	"errors"
)

const cmdAutheliaExample = `authelia --config /etc/authelia/config.yml --config /etc/authelia/access-control.yml
authelia --config /etc/authelia/config.yml,/etc/authelia/access-control.yml
authelia --config /etc/authelia/config/
`

const fmtAutheliaLong = `authelia %s

An open-source authentication and authorization server providing 
two-factor authentication and single sign-on (SSO) for your 
applications via a web portal.

Documentation is available at: https://www.authelia.com/docs
`

const fmtAutheliaBuild = `Last Tag: %s
State: %s
Branch: %s
Commit: %s
Build Number: %s
Build OS: %s
Build Arch: %s
Build Date: %s
Extra: %s
`

const buildLong = `Show the build information of Authelia

This outputs detailed version information about the specific version
of the Authelia binary. This information is embedded into Authelia
by the continuous integration.

This could be vital in debugging if you're not using a particular
tagged build of Authelia. It's suggested to provide it along with
your issue.
`

const completionLong = `To load completions:

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
`

const accessControlPolicyCheckLong = `
Checks a request against the access control rules to determine what policy would be applied.

Legend:

	#		The rule position in the configuration.
	*		The first fully matched rule.
	~		Potential match i.e. if the user was authenticated they may match this rule.
	hit     The criteria in this column is a match to the request.
	miss    The criteria in this column is not match to the request.
	may     The criteria in this column is potentially a match to the request.

Notes:

	A rule that potentially matches a request will cause a redirection to occur in order to perform one-factor
	authentication. This is so Authelia can adequately determine if the rule actually matches.
`
const (
	storageMigrateDirectionUp   = "up"
	storageMigrateDirectionDown = "down"
)

const (
	storageExportFormatCSV = "csv"
	storageExportFormatURI = "uri"
	storageExportFormatPNG = "png"
)

var (
	errNoStorageProvider = errors.New("no storage provider configured")
)
