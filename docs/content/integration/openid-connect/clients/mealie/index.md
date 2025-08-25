---
title: "Mealie"
description: "Integrating Mealie with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-13T21:01:17+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/mealie/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Mealie | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Mealie with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Mealie]
  - [v2.0.0](https://github.com/mealie-recipes/mealie/releases/tag/v2.0.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://mealie.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `mealie`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Mealie] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'mealie'
        client_name: 'Mealie'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng' # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://mealie.{{< sitevar name="domain" nojs="example.com" >}}/login'
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

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This configuration assumes [Mealie](https://mealie.io/) administrators are part of the `mealie-admins` group, and
[Mealie](https://mealie.io/) users are part of the `mealie-users` group. Depending on your specific group configuration, you will have to
adapt the `OIDC_ADMIN_GROUP` and `OIDC_USER_GROUP` nodes respectively. Alternatively you may elect to create a new
authorization policy in [provider authorization policies](../../../configuration/identity-providers/openid-connect/provider.md#authorization_policies) then utilize that policy as the
[client authorization policy](./../../configuration/identity-providers/openid-connect/clients.md#authorization_policy).
{{< /callout >}}

To configure [Mealie] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Mealie] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
OIDC_AUTH_ENABLED=true
OIDC_SIGNUP_ENABLED=true
OIDC_CONFIGURATION_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
OIDC_CLIENT_ID=mealie
OIDC_CLIENT_SECRET=insecure_secret
OIDC_AUTO_REDIRECT=false
OIDC_ADMIN_GROUP=mealie-admins
OIDC_USER_GROUP=mealie-users
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  mealie:
    environment:
      OIDC_AUTH_ENABLED: 'true'
      OIDC_SIGNUP_ENABLED: 'true'
      OIDC_CONFIGURATION_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      OIDC_CLIENT_ID: 'mealie'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_AUTO_REDIRECT: 'false'
      OIDC_ADMIN_GROUP: 'mealie-admins'
      OIDC_USER_GROUP: 'mealie-users'
```

## See Also

- [Mealie OpenID Connect Documentation](https://docs.mealie.io/documentation/getting-started/authentication/oidc-v2/)
- [Mealie OpenID Connect Environment Variables Documentation](https://docs.mealie.io/documentation/getting-started/installation/backend-config/#openid-connect-oidc)

[Mealie]: https://mealie.io/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
