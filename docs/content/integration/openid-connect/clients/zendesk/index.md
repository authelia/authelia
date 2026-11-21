---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Zendesk"
description: "A guide on integrating Zendesk with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Zendesk | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Zendesk with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Zendesk]
  - Not Applicable

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Zendesk Subdomain:** `example`
  - This option determines the redirect URI in the format of `https://example.zendesk.com/access/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `zendesk`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Zendesk] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    claims_policies:
      zendesk:
        id_token: ['email', 'email_verified', 'name']
    clients:
      - client_id: 'zendesk'
        client_name: 'Zendesk'
        public: true
        authorization_policy: 'two_factor'
        claims_policy: 'zendesk'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://example.zendesk.com/access/oidc/callback'
        scopes: ['openid', 'email', 'profile']
        response_types: ['code']
        grant_types: ['authorization_code']
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
```

Notes:

- [Zendesk] identifies users by their email address, so every Authelia user who signs in must have an email address
  which matches their [Zendesk] user.

### Application

To configure [Zendesk] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="tip" title="Help Wanted" icon="outline/rocket" >}}
We would love screenshots of this configuration!
{{< /callout >}}

To configure [Zendesk] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Zendesk] as an administrator.
2. Navigate to the Admin Center.
3. Click `Account` in the sidebar, then navigate to `Security` and then `Single sign-on`.
4. Click `Create SSO configuration`, then select `OpenID Connect`.
5. Configure the following options:
   - Configuration name: `Authelia`
   - Client ID: `zendesk`
   - Client secret: leave this empty
   - Scopes: `openid profile email`
   - Turn on auto discovery: Enabled
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Authentication Mode: `PKCE`
   - Show button when users sign in: Enabled
   - Button name: `Authelia`
6. Click `Save`.
7. Click `Account` in the sidebar, then navigate to `Security` and then either `Team member authentication` or
   `End user authentication` depending on which users should sign in with Authelia.
8. Select `External authentication`.
9. Select the `Authelia` SSO configuration.
10. Select either `Let them choose` or `Redirect to SSO`.
11. Click `Save`.

## See Also

- [Zendesk Setting up single sign-on with OpenID Connect (OIDC) Documentation](https://support.zendesk.com/hc/en-us/articles/7957465432474-Setting-up-single-sign-on-with-OpenID-Connect-OIDC)
- [Zendesk Managing single sign-on (SSO) configurations Documentation](https://support.zendesk.com/hc/en-us/articles/4408882188570-Managing-single-sign-on-SSO-configurations)

[Authelia]: https://www.authelia.com
[Zendesk]: https://www.zendesk.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
