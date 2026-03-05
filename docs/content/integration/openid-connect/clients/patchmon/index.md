---
title: "PatchMon"
description: "Integrating PatchMon with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2026-02-14T22:22:00+01:00
draft: false
images: []
weight: 620
toc: true
aliases: []
support:
  level: community
  versions: true
  integration: true
seo:
  title: "PatchMon | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring PatchMon with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [PatchMon]
  - [v1.4.1](https://github.com/PatchMon/PatchMon/releases/tag/v1.4.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://patchmon.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `patchmon`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [PatchMon] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'patchmon'
        client_name: 'PatchMon'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://patchmon.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/auth/oidc/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
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

To configure [PatchMon] consult the corresponding [documentation](https://docs.patchmon.net/books/patchmon-application-documentation/page/setting-up-oidc-sso-single-sign-on-integration).

#### Docker Configuration

To configure [PatchMon] to utilize Authelia as an [OpenID Connect 1.0] Provider, configure the following environment variables:

| Environment Variable | Value |
| --- | --- |
| `OIDC_ENABLED` | `true` |
| `OIDC_ISSUER_URL` | `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` |
| `OIDC_CLIENT_ID` | `patchmon` |
| `OIDC_CLIENT_SECRET` | `insecure_secret` |
| `OIDC_REDIRECT_URI` | `https://patchmon.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/auth/oidc/callback` |
| `OIDC_POST_LOGOUT_URI` | `https://patchmon.{{< sitevar name="domain" nojs="example.com" >}}` |

## See Also

- [PatchMon OIDC Documentation](https://docs.patchmon.net/books/patchmon-application-documentation/page/setting-up-oidc-sso-single-sign-on-integration)

[PatchMon]: https://patchmon.net/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
