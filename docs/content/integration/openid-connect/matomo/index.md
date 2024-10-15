---
title: "Matomo"
description: "Integrating Matomo with the Authelia OpenID Connect 1.0 Provider."
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

* [Authelia]
  * [v4.38.14](https://github.com/authelia/authelia/releases/tag/v4.38.14)
* [Matomo]
  * [v5.1.2](https://github.com/matomo-org/matomo/releases/tag/5.1.2)
* [LoginOIDC]
  * [v5.0.0](https://github.com/dominik-th/matomo-plugin-LoginOIDC/releases/tag/5.0.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://matomo.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `matomo`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Matomo] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'matomo'
        client_name: 'Matomo'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://matomo.{{< sitevar name="domain" nojs="example.com" >}}/index.php?module=LoginOIDC&action=callback&provider=oidc'
        scopes:
          - 'openid'
          - 'groups'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Matomo] to utilize [Authelia] as an [OpenID Connect 1.0] Provider:

1. Install the Plugin:
   1. Visit the [Matomo] `Administration` page.
   2. Click `Plugins`.
   3. Click `Manage Plugins`.
   4. Click `installing plugins from the Marketplace`.
   5. Install `Login OIDC` by `dominik-th`.
2. Configure the Plugin:
   1. Click `System`.
   2. Click `General settings`.
   3. Click `Login OIDC`.
   4. Enter `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization` in the `Authorize URL` field.
   5. Enter `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token` in the `Token URL` field.
   6. Enter `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo` in the `Userinfo URL` field.
   7. Enter `sub` in the `Userinfo ID` field.
   8. Enter `matomo` in the `Client ID` field.
   9. Enter `insecure_secret` in the `Client Secret` field.
   10. Enter `openid email` in the `OAuth Scope` field.

## See Also

- [Matomo Login OIDC FAQ](https://plugins.matomo.org/LoginOIDC/#faq)

[Matomo]: https://matomo.org/
[Authelia]: https://www.authelia.com
[LoginOIDC]: https://plugins.matomo.org/LoginOIDC/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
