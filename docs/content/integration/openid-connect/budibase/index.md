---
title: "Budibase"
description: "Integrating Budibase with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2023-11-16T06:16:54+11:00
draft: false
images: []
weight: 620
toc: true
community: true
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
  - 2.13.9

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://budibase.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `budibase`
* __Client Secret:__ `insecure_secret`

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
          - 'https://budibase.example.com/api/global/auth/oidc/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'offline_access'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

#### Organization configuration

Go on the builder main page: **Settings > Organization** or url : https://budibase.example.com/builder/portal/settings/organisation

{{< figure src="budibase_org.png" alt="Budibase" width="300" >}}

- Org. name: example.com
- Platform URL: https://budibase.example.com

> ⚠️ **Don't forget to press save.**

#### Auth configuration

Go the builder main page: **Settings > Auth > OpenID Connect** or url : https://budibase.example.com/builder/portal/settings/auth

{{< figure src="budibase_auth.png" alt="Budibase" width="300" >}}

- Config URL: https://auth.example.com/.well-known/openid-configuration
- Client ID: budibase
- Client Secret: myclientsecret
- Name: Authelia
- Icon: authelia.svg (Upload your own here [authelia branding](https://www.authelia.com/reference/guides/branding/))

> ⚠️ **Don't forget to press save.**

## See Also

- [Budibase OpenID Connect Documentation](https://docs.budibase.com/docs/openid-connect)

[Authelia]: https://www.authelia.com
[Budibase]: https://budibase.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
