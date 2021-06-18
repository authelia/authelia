package main

const (
	cmdAutheliaExample = `authelia --config /etc/authelia/config.yml --config /etc/authelia/access-control.yml
authelia --config /etc/authelia/config.yml,/etc/authelia/access-control.yml
authelia --config /etc/authelia/config/
`
	cmdAutheliaLong = `authelia

An open-source authentication and authorization server
providing two-factor authentication and single sign-on (SSO)
for your applications via a web portal
`
)
