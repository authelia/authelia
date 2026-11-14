---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Proxmox Mail Gateway"
description: "A guide on integrating Proxmox Mail Gateway with the Authelia OpenID Connect 1.0 Provider for SSO, with configuration examples and available options."
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
  title: "Proxmox Mail Gateway | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Proxmox Mail Gateway with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Proxmox Mail Gateway]
  - [v9.1](https://pmg.proxmox.com/wiki/Roadmap)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://pmg.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://pmg.{{< sitevar name="domain" nojs="example.com" >}}`.
    This means if you change this value, you need to update the redirect URI.
  - The redirect URI is the origin of the URL used to access the [Proxmox Mail Gateway] web interface, which includes
    the port if it is not the default for the scheme. For example when accessing the web interface directly it may be
    `https://pmg.{{< sitevar name="domain" nojs="example.com" >}}:8006`.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `pmg`
- **Client Secret:** `insecure_secret`
- **Realm:** `authelia`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with
[Proxmox Mail Gateway] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'pmg'
        client_name: 'Proxmox Mail Gateway'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://pmg.{{< sitevar name="domain" nojs="example.com" >}}'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Proxmox Mail Gateway] there are two methods, using the [Web GUI](#web-gui), or using the [CLI](#cli).

#### Web GUI

To configure [Proxmox Mail Gateway] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
instructions:

1. Login to [Proxmox Mail Gateway].
2. Navigate to Configuration.
3. Navigate to User Management.
4. Navigate to Realms.
5. Click `Add` and select `OpenID Connect Server`.
6. Configure the following options:
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Realm: `authelia`
   - Client ID: `pmg`
   - Client Key: `insecure_secret`
   - Username Claim: `preferred_username`
   - Autocreate Users: Enable if you want users to automatically be created in [Proxmox Mail Gateway].
   - Source for Role Assignment: `Fixed role for all auto-created users`
   - Fixed Role: The role auto-created users should be assigned, for example `audit`.
7. Click `Add`.

#### CLI

To configure [Proxmox Mail Gateway] to utilize Authelia as an [OpenID Connect 1.0] Provider, run the following command
on the [Proxmox Mail Gateway] host:

```shell
pmgsh create /access/auth-realm --realm authelia --type oidc --issuer-url 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}' --client-id 'pmg' --client-key 'insecure_secret' --username-claim 'preferred_username' --autocreate 1 --autocreate-role-assignment 'source=fixed,fixed-role=audit'
```

## See Also

- [Proxmox Mail Gateway Administration Guide: OpenID Connect](https://pmg.proxmox.com/pmg-docs/pmg-admin-guide.html#user_oidc)

[Authelia]: https://www.authelia.com
[Proxmox Mail Gateway]: https://www.proxmox.com/en/products/proxmox-mail-gateway
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
