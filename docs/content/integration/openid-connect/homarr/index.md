---
title: "Homarr"
description: "Integrating Homarr with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 720
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
  - [v4.38.19](https://github.com/authelia/authelia/releases/tag/v4.38.19)
- [Homarr]
  - [1.7.0](https://github.com/homarr-labs/homarr/releases/tag/v1.7.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://homarr.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `homarr`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Homarr] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'homarr'
        client_name: 'Homarr'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://homarr.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/oidc'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Homarr] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Homarr] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
AUTH_PROVIDERS=oidc
AUTH_OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
AUTH_OIDC_CLIENT_ID=homarr
AUTH_OIDC_CLIENT_SECRET=insecure_secret
AUTH_OIDC_CLIENT_NAME=Authelia
AUTH_OIDC_SCOPE_OVERWRITE=openid email profile groups
AUTH_OIDC_GROUPS_ATTRIBUTE=groups
AUTH_LOGOUT_REDIRECT_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/logout
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  homarr:
    environment:
      AUTH_PROVIDERS: 'oidc'
      AUTH_OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      AUTH_OIDC_CLIENT_ID: 'homarr'
      AUTH_OIDC_CLIENT_SECRET: 'insecure_secret'
      AUTH_OIDC_CLIENT_NAME: 'Authelia'
      AUTH_OIDC_SCOPE_OVERWRITE: 'openid email profile groups'
      AUTH_OIDC_GROUPS_ATTRIBUTE: 'groups'
      AUTH_LOGOUT_REDIRECT_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/logout'
```

### Groups

To assign users to Homarr groups, refer to the [Homarr] SSO Documentation on their [permission system](https://homarr.dev/docs/advanced/single-sign-on/#permission-system).

## See Also

- [Homarr SSO Documentation](https://homarr.dev/docs/advanced/single-sign-on/)

[Authelia]: https://www.authelia.com
[Homarr]: https://homarr.dev
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
