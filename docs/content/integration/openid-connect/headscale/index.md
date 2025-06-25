---
title: "Headscale"
description: "Integrating Headscale with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-06-25T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.4](https://github.com/authelia/authelia/releases/tag/v4.39.4)
- [Headscale]
  - [v0.26.1](https://github.com/juanfont/headscale/releases/tag/v0.26.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://headscale.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `headscale`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is a example __Authelia__ [client configuration] for use with [Headscale] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    claims_policies:
      headscale:
        id_token: ['groups', 'email', 'email_verified', 'alt_emails', 'preferred_username', 'name']
    clients:
      - client_id: 'headscale'
        client_name: 'Headscale'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://headscale.{{< sitevar name="domain" nojs="example.com" >}}/oidc/callback'
        claims_policy: 'headscale'
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

Note that the `claims_policy` is only necessary if you are authorizing users based on
domain, groups or email (`oidc.allowed_domains`, `oidc.allowed_groups` and
`oidc.allowed_users` in the Headscale configuration file). This is because currently
Headscale doesn't query the userinfo endpoint for these claims if they are missing from the
id token
(see [Headscale#2655](https://github.com/juanfont/headscale/issues/2655) and
[Restore functionality prior to claims parameter](../openid-connect-1.0-claims.md#restore-functionality-prior-to-claims-parameter)
for details).

### Application

To configure [Headscale] to utilize Authelia as an [OpenID Connect 1.0] provider, configure the `oidc:` section in the `config.yaml`

```yaml {title="config.yaml"}
oidc:
  only_start_if_oidc_is_available: true
  issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
  client_id: 'headscale'
  client_secret: 'insecure_secret'
  # client_secret_path: '/path/to/client_secret.txt' # Alternative to client_secret
  scope: ['openid', 'profile', 'email', 'groups']
  pkce:
    enabled: true
    method: 'S256'
```

## See Also

- [Configuring headscale to use OIDC authentication](https://headscale.net/stable/ref/oidc/)

[Authelia]: https://www.authelia.com
[Headscale]: https://headscale.net
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
