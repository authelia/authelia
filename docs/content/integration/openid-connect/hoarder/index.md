---
title: "hoarder"
description: "Integrating hoarder with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-26
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
  - [v4.38.18](https://github.com/authelia/authelia/releases/tag/v4.38.18)
- [Hoarder]
  - [v0.21.0](https://github.com/hoarder-app/hoarder/releases/tag/v0.21.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://hoarder.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `hoarder`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [hoarder] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: "hoarder"
        client_name: "hoarder"
        client_secret: "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng" # The digest of 'insecure_secret'.
        public: false
        authorization_policy: "two_factor"
        redirect_uris:
          - 'https://hoarder.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/custom'
        scopes:
          - "openid"
          - "profile"
          - "email"
        userinfo_signed_response_alg: "none"
```

### Application

To configure [hoarder] to utilize Authelia as an [OpenID Connect 1.0] Provider, specify the below environment variables in hoarder provided `.env` file:

```.env
OAUTH_WELLKNOWN_URL=https://auth.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
OAUTH_CLIENT_ID=hoarder
OAUTH_CLIENT_SECRET=insecure_secret
OAUTH_PROVIDER_NAME="Authelia"
```

## See Also

- [Hoarder OAuth OIDC config](https://docs.hoarder.app/configuration#authentication--signup)
- [Hoarder GitHub Discussion Authelia Configuration](https://github.com/hoarder-app/hoarder/discussions/419)

[hoarder]: https://hoarder.app/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
