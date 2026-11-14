---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "DigitalOcean"
description: "A guide on integrating DigitalOcean with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
summary: ""
date: 2026-11-12T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "DigitalOcean | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring DigitalOcean with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [DigitalOcean]

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `digitalocean`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [DigitalOcean] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'digitalocean'
        client_name: 'DigitalOcean'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://cloud.digitalocean.com/sessions/sso/callback'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="digitalocean" claims="email,email_verified,name,preferred_username" %}}

### Application

To configure [DigitalOcean] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="note" title="Team Roles" icon="outline/info-circle" >}}
[DigitalOcean] can assign a team role to each user via the `team_role` claim. As [DigitalOcean] does not request a
custom scope that this claim could be associated with, this guide relies on the `Default role` instead, which is
assigned to every user who does not receive a `team_role` claim. Team roles can be adjusted by a team owner after the
user has joined the team.
{{< /callout >}}

To configure [DigitalOcean] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to the [DigitalOcean Control Panel](https://cloud.digitalocean.com/) as a team owner.
2. Click the profile icon in the top right, click `Switch Teams`, and select the team you wish to configure.
3. Navigate to `Settings`.
4. In the `Single sign-on (OIDC)` section click `Enable`.
5. Configure the following options:
   - OpenID provider URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - OpenID client ID: `digitalocean`
   - OpenID client secret: `insecure_secret`
   - Default role: `Member`
6. Click `Test SSO config to continue` and complete the login with Authelia.
7. Leave `Require sign-in via SSO only` disabled while testing and click `Continue`.
8. Take note of the `SSO sign-in URL` on the `Summary` page.
9. Click `Enable SSO`.

Users can then login using the `SSO sign-in URL`. After successfully testing the configuration you may optionally
enable `Require sign-in via SSO only` by selecting `Edit SSO` from the `...` menu in the `Single sign-on (OIDC)`
section.

## See Also

- [DigitalOcean How to Configure Single Sign-On for Teams Documentation](https://docs.digitalocean.com/platform/teams/how-to/configure-sso/)
- [DigitalOcean Teams Predefined Roles Documentation](https://docs.digitalocean.com/platform/teams/roles/predefined/)

[Authelia]: https://www.authelia.com
[DigitalOcean]: https://www.digitalocean.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
