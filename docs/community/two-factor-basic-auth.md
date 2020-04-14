---
layout: default
title: 2FA through basic auth
parent: Community
nav_order: 1
---

The following project allows you to use Authelia's one-time password (OTP) 2-factor authentication (2FA) through only
[basic auth](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication)
and a custom credentials format described below.
This allows you to use 2FA on clients and scenarios
that demand basic auth, e.g. [webdav](https://en.wikipedia.org/wiki/WebDAV) network streaming.
More information:

## [authelia-basic-2fa](https://github.com/ViRb3/authelia-basic-2fa)

## Jira auto-login with authelia HTTP Headers

You can make Jira auto-login to the user that is currently logged in to authelia.
I say "auto-login" as I couldn't find any plugin to actually be authentication
provider through HTTP headers only - LDAP though seems to have support.

So this guide is targeted to authelia users that don't use any other authentication
backend.

I'm using traefik with docker as an example, but any proxy that can forward
authelia `Remote-User` header is fine.

First of all, users should exist on both authelia and Jira AND have the same
username for this to work. Also you will have to [pay for a plugin](https://marketplace.atlassian.com/apps/1212581/easy-sso-jira-kerberos-ntlm-saml?hosting=server&tab=overview).

After both steps are done:
  - Add `traefik.http.middlewares.authelia.forwardauth.authResponseHeaders=Remote-User` in the labels of authelia
  - Add `traefik.http.routers.jira.middlewares=authelia@docker` in the labels of Jira (to actually enable authelia for the jira instance)
  - Install EasySSO in Jira
  - Go to EasySSO preferences and add the "Remote-User" header under HTTP and tick the "Username" checkbox.
  - Save
