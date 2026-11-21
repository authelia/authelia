---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "OpenCode"
description: "A guide on integrating OpenCode with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "OpenCode | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring OpenCode with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [OpenCode]
  - Not Applicable

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `opencode`
- **Client Secret:** `insecure_secret`
- **Email Domain:** `{{< sitevar name="domain" nojs="example.com" >}}`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [OpenCode] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    claims_policies:
      opencode:
        id_token: ['email', 'email_verified', 'name']
    clients:
      - client_id: 'opencode'
        client_name: 'OpenCode'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        claims_policy: 'opencode'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - ''  # See step 6 below.
        scopes: ['openid', 'email', 'profile']
        response_types: ['code']
        grant_types: ['authorization_code']
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [OpenCode] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="tip" title="Help Wanted" icon="outline/rocket" >}}
We would love screenshots of this configuration!
{{< /callout >}}

To configure [OpenCode] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to the [OpenCode Console](https://opencode.ai/console) as a workspace owner.
2. Navigate to `Settings` and then `Security`.
3. Click `Configure SSO`, then select `OpenID Connect` and the generic provider.
4. Configure the following options:
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `opencode`
   - Client Secret: `insecure_secret`
   - Display Name: `Authelia`
5. Click `Save`.
6. Copy the `Callback URL` which will be in the format of
   `https://opencode.ai/console/auth/sso/ssoconn_.../callback` and configure the Authelia `redirect_uris` option.
7. Under `Email domains`, click `Add domain` and enter `{{< sitevar name="domain" nojs="example.com" >}}`.
8. Publish the `TXT` record displayed with your DNS provider.
9. Click `Verify Domain`.

Notes:

- [OpenCode] only redirects users to Authelia when their email address belongs to a verified email domain.
- Workspace owners are still able to sign in without SSO so that a misconfigured identity provider can not lock them
  out, so a workspace admin or member account should be used to test the configuration.

## See Also

- [OpenCode SSO Documentation](https://opencode.ai/docs/console/sso/)

[Authelia]: https://www.authelia.com
[OpenCode]: https://opencode.ai/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
