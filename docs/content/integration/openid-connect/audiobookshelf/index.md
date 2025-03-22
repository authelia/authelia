---
title: "audiobookshelf"
description: "Integrating audiobookshelf with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-03-22T03:16:02+00:00
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
  * [v4.39.0](https://github.com/authelia/authelia/releases/tag/v4.39.0)
* [audiobookshelf]
  * [v2.20.0](https://github.com/advplyr/audiobookshelf/releases/tag/v2.20.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://audiobookshelf.{{< sitevar name="domain" nojs="example.com" >}}/`
  * This option determines the redirect URI in the format of
        `https://audiobookshelf.{{< sitevar name="domain" nojs="example.com" >}}/login`.
        This means if you change this value, you need to update the redirect URI.
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `audiobookshelf`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [audiobookshelf] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    clients:
      - client_id: 'audiobookshelf-client-id'
        client_name: 'audiobookshelf'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://audiobookshelf.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid/callback'
          - 'https://audiobookshelf.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid/mobile-redirect'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
```

```yaml {title="users_database.yml"}
---
users:
  administrator:
    displayname: "administrator"
    groups:
      # for audiobookshelf group claim
      - admin
  non_administrator:
    displayname: "non_administrator"
    groups:
      # for audiobookshelf group claim
      - user
```

### Application

Add the following [audiobookshelf] "settings" -> "authentication" or adapt the existing one:

{{< figure src="audiobookshelf_1.png" alt="audiobookshelf_1" width="300" >}}
{{< figure src="audiobookshelf_2.png" alt="audiobookshelf_2" width="300" >}}

## See Also

* [audiobookshelf Authenticating With an OpenID Provider Documentation](https://www.audiobookshelf.org/guides/oidc_authentication/)

[Authelia]: https://www.authelia.com
[audiobookshelf]: https://www.audiobookshelf.org/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
