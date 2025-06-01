---
title: "ROM Manager (RomM)"
description: "Integrating ROM Manager (RomM) with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
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
  - [v4.38.17](https://github.com/authelia/authelia/releases/tag/v4.38.17)
- [ROM Manager]
  - [v3.9.0](https://github.com/rommapp/romm/releases/tag/3.9.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://romm.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `romm`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

{{% oidc-conformance-claims claims="email,email_verified,alt_emails,preferred_username,name" %}}

The following YAML configuration is an example __Authelia__ [client configuration] for use with [ROM Manager] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'romm'
        client_name: 'ROM Manager'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://romm.{{< sitevar name="domain" nojs="example.com" >}}/api/oauth/openid'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [ROM Manager] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [ROM Manager] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

##### Standard

```shell {title=".env"}
OIDC_ENABLED=true
OIDC_PROVIDER=authelia
OIDC_CLIENT_ID=romm
OIDC_CLIENT_SECRET=insecure_secret
OIDC_REDIRECT_URI=https://romm.{{< sitevar name="domain" nojs="example.com" >}}/api/oauth/openid
OIDC_SERVER_APPLICATION_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  romm:
    environment:
      OIDC_ENABLED: 'true'
      OIDC_PROVIDER: 'authelia'
      OIDC_CLIENT_ID: 'romm'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_REDIRECT_URI: 'https://romm.{{< sitevar name="domain" nojs="example.com" >}}/api/oauth/openid'
      OIDC_SERVER_APPLICATION_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
```

## See Also

- [ROM Manager OIDC Setup With Authelia Documentation](https://docs.romm.app/latest/OIDC-Guides/OIDC-Setup-With-Authelia/)

[Authelia]: https://www.authelia.com
[ROM Manager]: https://romm.app/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
