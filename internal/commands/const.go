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

const (
	storageMigrateDirectionUp   = "up"
	storageMigrateDirectionDown = "down"
)

const (
	errFmtStorageMigrateNegativeVersion    = "can't migrate %s to version %d as versions below 0 are not valid"
	errFmtStorageMigrateWrongDirection     = "can't migrate %s to version %d as it's %s than the current version %d"
	errFmtStorageMigrateSame               = "can't migrate %s to version %d as it's the same as the current version"
	errFmtStorageMigrateUpHigherThanLatest = "can't migrate up to version %d as it's a higher version than the latest version available %d"
)

var (
	errStorageMigrateUpToVersion0           = errors.New("can't migrate up to version 0 as it's not an actual version and just represents an empty schema")
	errStorageMigrateAlreadyOnLatestVersion = errors.New("already on latest version")
	errStorageMigrateDownMissingTargetFlag  = errors.New("you must set the --target flag in order to migrate down")
	errStorageMigrateDownWhenZero           = errors.New("can't migrate down when the current version is 0")
)
