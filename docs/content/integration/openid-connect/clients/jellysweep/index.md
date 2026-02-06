---
title: "Jellysweep"
description: "Integrating Jellysweep with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2026-01-09T12:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Jellysweep | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Jellysweep with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: ""
  noindex: false
---

## Tested Versions

- [Authelia]
  - [v4.38.17](https://github.com/authelia/authelia/releases/tag/v4.38.17)
- [Jellysweep]
  - [v0.14.0](https://github.com/jon4hz/jellysweep/releases/tag/v0.14.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://jellysweep.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `jellysweep`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Jellysweep] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'jellysweep'
        client_name: 'Jellysweep'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://jellysweep.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback'
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
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Jellysweep] to utilize Authelia as an [OpenID Connect 1.0] Provider, configure the following environment
variables in your deployment:

```yaml {title="docker-compose.yml"}
services:
  jellysweep:
    image: ghcr.io/jon4hz/jellysweep:v0.14.0
    container_name: jellysweep
    environment:
      - JELLYSWEEP_SERVER_URL=https://jellysweep.{{< sitevar name="domain" nojs="example.com" >}}
      - JELLYSWEEP_AUTH_OIDC_ENABLED=true
      - JELLYSWEEP_AUTH_OIDC_NAME=Authelia
      - JELLYSWEEP_AUTH_OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
      - JELLYSWEEP_AUTH_OIDC_CLIENT_ID=jellysweep
      - JELLYSWEEP_AUTH_OIDC_CLIENT_SECRET=insecure_secret
      - JELLYSWEEP_AUTH_OIDC_REDIRECT_URL=https://jellysweep.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback
      - JELLYSWEEP_AUTH_OIDC_USE_PKCE=true
      - JELLYSWEEP_AUTH_OIDC_ADMIN_GROUP=jellyfin-admins
    ports:
      - "3002:3002"
    volumes:
      - ./config.yml:/app/config.yml:ro
      - ./data:/app/data
```

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This configuration assumes [Jellysweep](https://github.com/jon4hz/jellysweep) administrators are part of the `jellyfin-admins` group.
Depending on your specific group configuration, you will have to adapt the `JELLYSWEEP_AUTH_OIDC_ADMIN_GROUP` variable.
Alternatively you may elect to create a new authorization policy in [provider authorization policies](../../../configuration/identity-providers/openid-connect/provider.md#authorization_policies)
then utilize that policy as the [client authorization policy](../../../configuration/identity-providers/openid-connect/clients.md#authorization_policy).
{{< /callout >}}

## See Also

- [Jellysweep OIDC Documentation](https://github.com/jon4hz/jellysweep/blob/main/README.md#oidcsso-authentication)

[Authelia]: https://www.authelia.com
[Jellysweep]: https://github.com/jon4hz/jellysweep
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
