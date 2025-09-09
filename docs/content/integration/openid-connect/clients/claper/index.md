---
title: "Claper"
description: "Integrating Claper with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-13T08:35:59+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/claper/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Claper | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Claper with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.9](https://github.com/authelia/authelia/releases/tag/v4.39.9)
- [Claper]
  - [v2.3.3](https://github.com/ClaperCo/Claper/releases/tag/v2.3.3)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://claper.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `claper`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

{{< callout context="note" title="Security Note" icon="outline/info-circle" >}}
This client uses a plaintext client secret. While this is generally discouraged, in this context it is required. This is
because the client signs a request object using the client secret rather than a regular request. This helps ensure the
request can't be tampered with in any way even using a Man-in-the-Middle attack.

In addition due to the enforcement of both PKCE using SHA256 and Pushed Authorization Requests this client is incredibly
secure.
{{< /callout >}}

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Claper] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'claper'
        client_name: 'Claper'
        client_secret: '$plaintext$insecure_secret'
        public: false
        authorization_policy: 'two_factor'
        require_pushed_authorization_requests: true
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://claper.{{< sitevar name="domain" nojs="example.com" >}}/users/oidc/callback'
        scopes:
          - 'openid'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        request_object_signing_alg: 'HS256'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="claper" claims="email" %}}

### Application

To configure [Claper] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Claper] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
OIDC_PROVIDER_NAME=Authelia
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=claper
OIDC_CLIENT_SECRET=insecure_secret
OIDC_SCOPES=openid email
OIDC_LOGO_URL=https://www.authelia.com/images/branding/logo-cropped.png
OIDC_AUTO_REDIRECT_LOGIN=false
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  claper:
    environment:
      OIDC_PROVIDER_NAME: 'Authelia'
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'claper'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_SCOPES: 'openid email profile'
      OIDC_LOGO_URL: 'https://www.authelia.com/images/branding/logo-cropped.png'
      OIDC_AUTO_REDIRECT_LOGIN: 'false'
```

## See Also

- [Claper OIDC Documentation](https://docs.claper.co/integration/oidc.html)
- [Claper OIDC Configuration Documentation](https://docs.claper.co/self-hosting/configuration.html#openid-connect)

[Authelia]: https://www.authelia.com
[Claper]: https://claper.co
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
