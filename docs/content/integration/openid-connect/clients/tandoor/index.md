---
title: "Tandoor"
description: "Integrating Tandoor with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T11:03:16+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/tandoor/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Tandoor | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Tandoor with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.9](https://github.com/authelia/authelia/releases/tag/v4.39.9)
- [Tandoor]
  - [v1.5.34](https://github.com/TandoorRecipes/recipes/releases/tag/1.5.34)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://tandoor.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `tandoor`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Tandoor] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'tandoor'
        client_name: 'Tandoor'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://tandoor.{{< sitevar name="domain" nojs="example.com" >}}/accounts/oidc/authelia/login/callback/'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Tandoor] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

For reference purposes, the below `SOCIALACCOUNT_PROVIDERS` environment variable examples are the minified
format of the following:

```json
{
  "openid_connect": {
    "SCOPE": ["openid", "profile", "email"],
    "OAUTH_PKCE_ENABLED": true,
    "APPS": [
      {
        "provider_id": "authelia",
        "name": "Authelia",
        "client_id": "tandoor",
        "secret": "insecure_secret",
        "settings": {
          "server_url": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration",
          "token_auth_method": "client_secret_basic"
        }
      }
    ]
  }
}
```

To configure [Tandoor] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
SOCIAL_PROVIDERS=allauth.socialaccount.providers.openid_connect
SOCIALACCOUNT_PROVIDERS={"openid_connect":{"SCOPE":["openid","profile","email"],"OAUTH_PKCE_ENABLED":true,"APPS":[{"provider_id":"authelia","name":"Authelia","client_id":"tandoor","secret":"insecure_secret","settings":{"server_url":"https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration","token_auth_method":"client_secret_basic"}}]}}
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  tandoor:
    environment:
      SOCIAL_PROVIDERS: 'allauth.socialaccount.providers.openid_connect'
      SOCIALACCOUNT_PROVIDERS: '{"openid_connect":{"SCOPE":["openid","profile","email"],"OAUTH_PKCE_ENABLED":true,"APPS":[{"provider_id":"authelia","name":"Authelia","client_id":"tandoor","secret":"insecure_secret","settings":{"server_url":"https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration","token_auth_method":"client_secret_basic"}}]}}'
```

## See Also

- [Tandoor Authentication Allauth Documentation](https://docs.tandoor.dev/features/authentication/#allauth)

[Tandoor]: https://tandoor.dev/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
