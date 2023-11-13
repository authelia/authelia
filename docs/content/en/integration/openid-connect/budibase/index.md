---
title: "Budibase"
description: "Integrating Budibase with the Authelia OpenID Connect 1.0 Provider."
lead: ""
date: 2023-11-12T15:50:35+00:00
draft: false
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 620
toc: true
community: true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Budibase]
  - 2.13.9

## Before You Begin

{{% oidc-common %}}

### Authelia

Authelia configuration.yml

```yaml
identity_providers:
  oidc:
    clients:
      - id: budibase
        secret: mysecret
        authorization_policy: two_factor
        redirect_uris:
          - https://budibase.example.com/api/global/auth/oidc/callback
        scopes:
          - openid
          - profile
          - email
          - offline_access
```

## Budibase

### Organization configuration

Go on the builder main page: **Settings > Organization** or url : https://budibase.example.com/builder/portal/settings/organisation

{{< figure src="budibase_org.png" alt="Budibase" width="300" >}}

- Org. name: example.com
- Platform URL: https://budibase.example.com

### Auth configuration

Go the builder main page: **Settings > Auth > OpenID Connect** or url : https://budibase.example.com/builder/portal/settings/auth

{{< figure src="budibase_auth.png" alt="Budibase" width="300" >}}

- Config URL: https://auth.example.com/.well-known/openid-configuration
- Client ID: budibase
- Client Secret: mysecret
- Name: Authelia
- Icon: authelia.svg (Upload your own here [authelia branding](https://www.authelia.com/reference/guides/branding/))

## See Also

- [Budibase OpenID Connect Documentation](https://docs.budibase.com/docs/openid-connect)

[Authelia]: https://www.authelia.com
[Budibase]: https://budibase.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
