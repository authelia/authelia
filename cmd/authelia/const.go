package main

const fmtAutheliaLong = `authelia %s

Authelia is an open-source authentication and authorization server providing 2-factor authentication and 
single sign-on (SSO) for your applications via a web portal. It acts as a companion of reverse proxies like 
nginx, Traefik or HAProxy to let them know whether queries should pass through.Unauthenticated users are 
redirected to Authelia Sign-in portal instead.

Documentation is available at https://www.authelia.com/docs`

const fmtAutheliaVersionAll = `Branch: %s
Last Tag: %s
Commit: %s
Build Number: %s
Build Arch: %s
Build Date: %s
State Tag: %s
State Extra: %s`
