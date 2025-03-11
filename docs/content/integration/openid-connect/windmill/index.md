---
title: "Windmill"
description: "Integrating Windmill with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2023-12-10T10:52:22+11:00
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
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Windmill]
  - [1.224.0](https://github.com/windmill-labs/windmill/releases/tag/v1.224.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://windmill.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `windmill`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Windmill]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'windmill'
        client_name: 'Windmill'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://windmill.{{< sitevar name="domain" nojs="example.com" >}}/user/login_callback/authelia'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        userinfo_signed_response_alg: 'none'
```

## Application

### Core configuration

**Superadmin settings > Core**

- Base Url: `https://windmill.{{< sitevar name="domain" nojs="example.com" >}}`

{{< figure src="windmill_core.png" alt="Windmill" >}}

> ⚠️ **Don't forget to press save.**

### Auth configuration

**Superadmin settings > SSO/OAuth**

- Config URL: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
- Client Id: Windmill
- Client Secret: insecure_secret

{{< figure src="windmill_sso.png" alt="Windmill" >}}

> ⚠️ **Don't forget to press save.**

## See Also

- [Windmill OpenID Connect Documentation](https://www.windmill.dev/docs/misc/setup_oauth)

[Authelia]: https://www.authelia.com
[Windmill]: https://www.windmill.dev
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
