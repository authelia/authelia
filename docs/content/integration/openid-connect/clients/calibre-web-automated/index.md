---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Calibre-Web Automated"
description: "A guide on integrating Calibre-Web Automated with the Authelia OpenID Connect 1.0 Provider for SSO, with configuration examples and available options."
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
  title: "Calibre-Web Automated | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Calibre-Web Automated with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Calibre-Web Automated]
  - [v4.0.8](https://github.com/crocodilestick/Calibre-Web-Automated/releases/tag/v4.0.8)

{{% oidc-common bugs="client-credentials-encoding" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://calibre-web-automated.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://calibre-web-automated.{{< sitevar name="domain" nojs="example.com" >}}/login/generic/authorized`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `calibre-web-automated`
- **Client Secret:** `insecure_secret`
- **Admin Group Name:** `calibre-web-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with
[Calibre-Web Automated] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'calibre-web-automated'
        client_name: 'Calibre-Web Automated'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://calibre-web-automated.{{< sitevar name="domain" nojs="example.com" >}}/login/generic/authorized'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Calibre-Web Automated] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="tip" title="Help Wanted" icon="outline/rocket" >}}
We would love screenshots of this configuration!
{{< /callout >}}

To configure [Calibre-Web Automated] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
instructions:

1. Login to [Calibre-Web Automated] as an administrator.
2. Navigate to Admin.
3. Click `Edit Basic Configuration`.
4. Expand `Feature Configuration`.
5. Set `Login type` to `Use OAuth (requires HTTPS)`.
6. Configure the following options:
   - OAuth Metadata URL (for auto-discovery): `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - OAuth Scopes: `openid profile email groups`
   - OAuth Redirect Host: `https://calibre-web-automated.{{< sitevar name="domain" nojs="example.com" >}}`
   - OAuth Client Id: `calibre-web-automated`
   - OAuth Client Secret: `insecure_secret`
   - Username Field: `preferred_username`
   - Email Field: `email`
   - OAuth group for Admin: `calibre-web-admins`
   - Login Button Text: `Authelia`
7. Click `Save`.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
Members of the `OAuth group for Admin` group are only granted the admin role when the
`Enable OAuth Group-Based Admin Role Management` option is enabled.
{{< /callout >}}

## See Also

- [Calibre-Web Automated OAuth Configuration Documentation](https://github.com/crocodilestick/Calibre-Web-Automated/wiki/OAuth-Configuration)

[Authelia]: https://www.authelia.com
[Calibre-Web Automated]: https://github.com/crocodilestick/Calibre-Web-Automated
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
