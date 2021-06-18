package commands

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
