---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "ReadMeABook"
description: "A guide on integrating ReadMeABook with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "ReadMeABook | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring ReadMeABook with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [ReadMeABook]
  - [v1.3.0](https://github.com/kikootwo/ReadMeABook/releases/tag/v1.3.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://readmeabook.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://readmeabook.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `readmeabook`
- **Client Secret:** `insecure_secret`
- **Administrator Group:** `readmeabook-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [ReadMeABook] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'readmeabook'
        client_name: 'ReadMeABook'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://readmeabook.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
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

To configure [ReadMeABook] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
The `Authentication` settings tab is only shown when [ReadMeABook] is configured to use Audiobookshelf as its library
backend. The `PUBLIC_URL` environment variable should be set to
`https://readmeabook.{{< sitevar name="domain" nojs="example.com" >}}` so the correct redirect URI is used.
{{< /callout >}}

To configure [ReadMeABook] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [ReadMeABook] as an administrator.
2. Navigate to the settings dashboard by going to `https://readmeabook.{{< sitevar name="domain" nojs="example.com" >}}/admin/settings`.
3. Navigate to `Authentication`.
4. Check `Enable OIDC Authentication`.
5. Configure the following options:
   - Provider Name: `Authelia`
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `readmeabook`
   - Client Secret: `insecure_secret`
6. Optionally configure the following options to grant administrator access to members of a group:
   - Enable Admin Role Mapping: Checked
   - Admin Claim Name: `groups`
   - Admin Claim Value: `readmeabook-admins`
7. Click `Save Settings`.

## See Also

- [ReadMeABook GitHub Repository](https://github.com/kikootwo/ReadMeABook)

[Authelia]: https://www.authelia.com
[ReadMeABook]: https://github.com/kikootwo/ReadMeABook
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
