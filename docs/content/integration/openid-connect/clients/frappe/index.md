---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Frappe"
description: "A guide on integrating Frappe with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Frappe | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Frappe with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Frappe]
  - [v16.56.0](https://github.com/frappe/frappe/releases/tag/v16.56.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://frappe.{{< sitevar name="domain" nojs="example.com" >}}/`
  - The `Provider Name` determines the redirect URI in the format of
    `https://frappe.{{< sitevar name="domain" nojs="example.com" >}}/api/method/frappe.integrations.oauth2_logins.custom/authelia`.
    This means if you change the `Provider Name`, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `frappe`
- **Client Secret:** `insecure_secret`
- **Provider Name:** `Authelia`

These instructions apply to all [Frappe] Framework applications which use the Social Login Key feature, including
ERPNext.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Frappe] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'frappe'
        client_name: 'Frappe'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://frappe.{{< sitevar name="domain" nojs="example.com" >}}/api/method/frappe.integrations.oauth2_logins.custom/authelia'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Frappe] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Frappe] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Frappe] as an administrator.
2. Navigate to `Integrations`.
3. Navigate to `Social Login Key`.
4. Click `+ New`.
5. Configure the following options:
   - Enable Social Login: Enabled
   - Social Login Provider: `Custom`
   - Provider Name: `Authelia`
   - Client ID: `frappe`
   - Client Secret: `insecure_secret`
   - Sign ups: `Allow`
   - Base URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Authorize URL: `/api/oidc/authorization`
   - Access Token URL: `/api/oidc/token`
   - Redirect URL: `/api/method/frappe.integrations.oauth2_logins.custom/authelia`
   - API Endpoint: `/api/oidc/userinfo`
   - User ID Property: `sub`
   - Auth URL Data: `{"response_type": "code", "scope": "openid email profile"}`
6. Click `Save`.

## See Also

- [Frappe OpenID Connect and Frappe Social Login Documentation](https://docs.frappe.io/framework/user/en/guides/integration/openid_connect_and_frappe_social_login)
- [Frappe Adding Social Login Provider Documentation](https://docs.frappe.io/framework/user/en/guides/app-development/adding-social-login-provider)

[Authelia]: https://www.authelia.com
[Frappe]: https://frappe.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
