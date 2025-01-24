---
title: "Mobilizon"
description: "Integrating Mobilizon with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-21T22:32:51+11:00
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

* [Authelia]
  * [v4.38.18](https://github.com/authelia/authelia/releases/tag/v4.38.18)
* [Mobilizon]
  * [5.1.0](https://framagit.org/framasoft/mobilizon/-/releases/5.1.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://mobilizon.{{< sitevar name="domain" nojs="example.com" >}}/`
  * This option determines the redirect URI in the format of
        `https://mobilizon.{{< sitevar name="domain" nojs="example.com" >}}/login`.
        This means if you change this value, you need to update the redirect URI.
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `mobilizon`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Mobilizon] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'mobilizon'
        client_name: 'Mobilizon'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://mobilizon.{{< sitevar name="domain" nojs="example.com" >}}/auth/keycloak/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

Add the following values to [Mobilizon] `config.exs`:

```exs
config :ueberauth,
       Ueberauth,
       providers: [
         keycloak: {Ueberauth.Strategy.Keycloak, [default_scope: "openid email profile"]}
       ]

config :mobilizon, :auth,
  oauth_consumer_strategies: [
    {:keycloak, "Authelia"}
  ]

config :ueberauth, Ueberauth.Strategy.Keycloak.OAuth,
  client_id: "mobilizon",
  client_secret: "insecure_secret",
  site: "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
  authorize_url: "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization",
  token_url: "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token",
  userinfo_url: "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo",
  token_method: :post
```

## See Also

- [Mobilizon OAuth Authentication Documentation](https://docs.joinmobilizon.org/administration/configure/auth/#oauth)

[Authelia]: https://www.authelia.com
[Mobilizon]: https://joinmobilizon.org/en/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
