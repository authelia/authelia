---
title: "ZTNET"
description: "Integrating ZTNET with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2026-10-03T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "ztnet | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring ztnet with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: ""
  noindex: false
---

## Tested Versions

* [Authelia]
  * [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
* [ZTNET]
  * [v0.8.4](https://github.com/sinamics/ztnet/releases/tag/v0.8.4)
  
{{% oidc-common %}}

## Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://ztnet.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `ztnet`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [ZTNET] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'ztnet'
        client_name: 'ZTNET'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://ztnet.example.com/api/auth/callback/oauth'
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
        token_endpoint_auth_method: 'client_secret_post'
```

ZTNET authenticates to the token endpoint using the `client_secret_post` method, which is not the Authelia default
(`client_secret_basic`). The `token_endpoint_auth_method` option must therefore be explicitly set to
`client_secret_post`, otherwise the token exchange fails with an `invalid_client` error.

#### Restricting Access to a Group

Access to ZTNET can optionally be limited to members of a specific group by using a custom
[authorization policy](https://www.authelia.com/configuration/identity-providers/openid-connect/provider/#authorization_policies)
and referencing it from the client configuration:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    authorization_policies:
      ztnet:
        default_policy: 'deny'
        rules:
          - policy: 'two_factor'
            subject: 'group:ztnet_users'
    clients:
      - client_id: 'ztnet'
        authorization_policy: 'ztnet'
        ## The remainder of the client configuration is unchanged.
```

### Application

To configure [ZTNET] to utilize Authelia as an [OpenID Connect 1.0] Provider, set the following environment variables
for the ZTNET container or service:

```shell {title=".env"}
OAUTH_ID=ztnet
OAUTH_SECRET=insecure_secret
OAUTH_WELLKNOWN=https://auth.example.com/.well-known/openid-configuration
OAUTH_ALLOW_NEW_USERS=true
NEXTAUTH_URL=https://ztnet.example.com
```

Notes:

* `OAUTH_SECRET` is the plaintext secret, not the digest used in the Authelia configuration.
* `OAUTH_ALLOW_NEW_USERS` is required for users to be created on their first login via OpenID Connect.
* `NEXTAUTH_URL` must be the external URL of ZTNET. This is especially important behind a reverse proxy, as the
  redirect URI sent to Authelia is derived from it and must exactly match the registered `redirect_uris` value.
* Optionally, `OAUTH_EXCLUSIVE_LOGIN=true` disables the local email and password login. It is recommended to only set
  this after the first administrator account exists and OpenID Connect login has been confirmed to work.

## See Also

* [ZTNET OAuth Authentication Documentation](https://ztnet.network/authentication/oauth)

[Authelia]: https://www.authelia.com
[ZTNET]: https://ztnet.network/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
