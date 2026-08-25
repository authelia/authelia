---
title: "2FAuth"
description: "Integrating 2FAuth with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2026-08-22T20:00:00+02:00
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
  title: "2FAuth | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring 2FAuth with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.20](https://github.com/authelia/authelia/releases/tag/v4.39.20)
- [2FAuth]
  - [v8.0.1](https://github.com/Bubka/2FAuth/releases/tag/v8.0.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://{{< sitevar name="subdomain-2fauth" nojs="2fauth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Provider ID:** `{{< sitevar name="provider-id" nojs="authelia" >}}`
- **Client ID:** `{{< sitevar name="client-id" nojs="2fauth" >}}`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [2FAuth] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: '{{< sitevar name="client-id" nojs="2fauth" >}}'
        client_name: '2FAuth'
        client_secret: '$pbkdf2-sha512$310000$yTn6OQHuQRyif.WYB6/eJw$Plhbt5sn5800fNj4lMr5/gRAjtbur6nue6tbavbgJ4IyxISw.4i6py409653cv35gbvzWA'
        public: false
        authorization_policy: 'one_factor'
        require_pkce: false
        redirect_uris:
          - 'https://{{< sitevar name="subdomain-2fauth" nojs="2fauth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/socialite/callback/openid'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [2FAuth] there is one method, using the [environmental variables][2FAuth SSO Documentation].

#### Environmental Variables

To configure [2FAuth] to utilize [Authelia] as an [OpenID Connect 1.0] Provider, set the following configuration:

```text
OPENID_AUTHORIZE_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
OPENID_TOKEN_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
OPENID_USERINFO_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
OPENID_CLIENT_ID={{< sitevar name="client-id" nojs="2fauth" >}}
OPENID_CLIENT_SECRET=insecure_secret
OPENID_HTTP_VERIFY_SSL_PEER=true
```

If you use a self-signed certificate, set *OPENID_HTTP_VERIFY_SSL_PEER* to the filepath of that certificate or to *false* to disable SSL verification completely.

## See Also

- [2FAuth SSO Documentation]

[Authelia]: https://www.authelia.com
[2FAuth]: https://docs.2fauth.app/
[2FAuth SSO Documentation]: https://docs.2fauth.app/security/authentication/sso/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
