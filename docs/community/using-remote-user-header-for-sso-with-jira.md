---
layout: default
title: Using Remote-User header for SSO with Jira
parent: Community
nav_order: 2
---

# Using Remote-User header for SSO with Jira

You can make Jira auto-login to the user that is currently logged in to authelia.
I say "auto-login" as I couldn't find any plugin to actually be authentication
provider through HTTP headers only - LDAP though seems to have support.

So this guide is targeted to authelia users that don't use any other authentication
backend.

I'm using traefik with docker as an example, but any proxy that can forward
authelia `Remote-User` header is fine.

First of all, users should exist on both Authelia and Jira, and have the same
username for this to work. Also you will have to
[pay for a plugin](https://marketplace.atlassian.com/apps/1212581/easy-sso-jira-kerberos-ntlm-saml?hosting=server&tab=overview).

After both steps are done:
  - Add `traefik.http.middlewares.authelia.forwardauth.authResponseHeaders=Remote-User` in the labels of authelia
  - Add `traefik.http.routers.jira.middlewares=authelia@docker` in the labels of Jira (to actually enable Authelia for 
    the Jira instance)
  - Install EasySSO in Jira
  - Go to EasySSO preferences and add the "Remote-User" header under HTTP and tick the "Username" checkbox.
  - Save

## Other Systems

While this guide is tailored for Jira, you can use a similar method with many other services like Jenkins and Grafana.
