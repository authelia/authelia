---
title: "Writefreely"
description: "Integrating Writefreely with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-25T10:04:53+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/writefreely/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Writefreely | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Writefreely with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.8](https://github.com/authelia/authelia/releases/tag/v4.38.8)
- [Writefreely]
  - [v0.15.1](https://github.com/writefreely/writefreely/releases/tag/v0.15.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://writefreely.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
        `https://writefreely.{{< sitevar name="domain" nojs="example.com" >}}/login`.
        This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `writefreely`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Writefreely] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'writefreely'
        client_name: 'Writefreely'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://writefreely.{{< sitevar name="domain" nojs="example.com" >}}/oauth/callback/generic'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Writefreely] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `config.ini`.
{{< /callout >}}

To configure [Writefreely] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```ini {title="config.ini"}
[app]
disable_password_auth = true

[oauth.generic]
client_id          = writefreely
client_secret      = insecure_secret
host               = https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
display_name       = Authelia
token_endpoint     = /api/oidc/token
inspect_endpoint   = /api/oidc/userinfo
auth_endpoint      = /api/oidc/authorization
scope              = openid email profile
allow_disconnect   = false
map_user_id        = sub
map_username       = preferred_username
map_display_name   = name
map_email          = email
```

## See Also

- [Writefreely OAuth Configuration Documentation](https://writefreely.org/docs/main/admin/config#oauth)

[Authelia]: https://www.authelia.com
[Writefreely]: https://writefreely.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
