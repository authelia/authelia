---
title: "Budibase"
description: "Integrating Budibase with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2023-11-16T06:16:54+11:00
draft: false
images: []
weight: 720
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
- [Budibase]
  - [v2.13.9](https://github.com/Budibase/budibase/releases/tag/2.13.9)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://budibase.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `budibase`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Budibase] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'budibase'
        client_name: 'Budibase'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://budibase.{{< sitevar name="domain" nojs="example.com" >}}/api/global/auth/oidc/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'offline_access'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Budibase] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Budibase] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

You may be able to skip steps 1 to 3 by visiting the following URL: https://budibase.{{< sitevar name="domain" nojs="example.com" >}}/builder/portal/settings/organisation

1. Perform one of the following steps:
   1. Visit https://budibase.{{< sitevar name="domain" nojs="example.com" >}}/builder/portal/settings/organisation
   2. Perform all the following steps to get to the above URL:
      1. Navigate to the Builder Main Page.
      2. Navigate to Settings.
      3. Navigate to Organization.
2. Configure the following options:
   - Org. name: `{{< sitevar name="domain" nojs="example.com" >}}`
   - Platform URL: `https://budibase.{{< sitevar name="domain" nojs="example.com" >}}`
3. Click Save.
4. Perform one of the following steps:
   1. Visit https://budibase.{{< sitevar name="domain" nojs="example.com" >}}/builder/portal/settings/auth
   2. Perform all the following steps to get to the above URL:
      1. Navigate to the Builder Main Page.
      2. Navigate to Settings.
      3. Navigate to Auth.
      4. Navigate to OpenID Connect.
5. Configure the following options:
   - Config URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Client ID: `budibase`
   - Client Secret: `insecure_secret`
   - Name: `Authelia`
   - Icon: `authelia.svg` (download available on the [authelia branding](https://www.authelia.com/reference/guides/branding/) guide)
6. Click Save.

{{< figure src="budibase_org.png" alt="Budibase" width="300" >}}

{{< figure src="budibase_auth.png" alt="Budibase" width="300" >}}

## See Also

- [Budibase OpenID Connect Documentation](https://docs.budibase.com/docs/openid-connect)

[Authelia]: https://www.authelia.com
[Budibase]: https://budibase.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
