---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Organizr"
description: "A guide on integrating Organizr with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Organizr | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Organizr with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Organizr]
  - [v2.1.5000](https://github.com/causefx/Organizr/commits/v2-master)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://organizr.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://organizr.{{< sitevar name="domain" nojs="example.com" >}}/api/v2/oidc/pocketid/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `organizr`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Organizr] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'organizr'
        client_name: 'Organizr'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://organizr.{{< sitevar name="domain" nojs="example.com" >}}/api/v2/oidc/pocketid/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Organizr] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
[Organizr] does not have a generic [OpenID Connect 1.0] provider option. Each of the named provider options uses
discovery and works with any standards compliant provider, so this guide uses the `PocketID` provider option and
changes its display name. The `PocketID` option is used as it does not attempt RP-Initiated Logout.
{{< /callout >}}

To configure [Organizr] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Organizr] as an administrator.
2. Navigate to `Settings` > `System Settings` > `SSO` > `OIDC Global`.
3. Configure the following options:
   - Enable OIDC Authentication: Enabled
   - Default Group: Select the [Organizr] group which should be assigned to new users when no group mapping matches.
   - Group Claim Name: `groups`
4. Navigate to `Settings` > `System Settings` > `SSO` > `OIDC: PocketID`.
5. Configure the following options:
   - Enable PocketID: Enabled
   - Display Name: `Authelia`
   - Discovery URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Client ID: `organizr`
   - Client Secret: `insecure_secret`
   - Scopes: `openid,profile,email,groups`
6. Click `Save`.

## See Also

- [Organizr Documentation](https://docs.organizr.app/)

[Authelia]: https://www.authelia.com
[Organizr]: https://github.com/causefx/Organizr
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
