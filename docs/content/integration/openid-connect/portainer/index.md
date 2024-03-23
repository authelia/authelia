---
title: "Portainer"
description: "Integrating Portainer with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 620
toc: true
community: true
aliases:
  - /docs/community/oidc-integrations/portainer.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [Portainer] CE and EE
  * 2.12.2

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://portainer.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `portainer`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/openid-connect/clients.md) for use with [Portainer]
which will operate with the above example:

```yaml
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'portainer'
        client_name: 'Portainer'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://portainer.example.com'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [Portainer] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Visit Settings
2. Visit Authentication
3. Set the following values:
   1. Authentication Method: OAuth
   2. Provider: Custom
   3. Enable *Automatic User Provision* if you want users to automatically be created in [Portainer].
   4. Client ID: `portainer`
   5. Client Secret: `insecure_secret`
   6. Authorization URL: `https://auth.example.com/api/oidc/authorization`
   7. Access Token URL: `https://auth.example.com/api/oidc/token`
   8. Resource URL: `https://auth.example.com/api/oidc/userinfo`
   9. Redirect URL: `https://portainer.example.com`
   10. User Identifier: `preferred_username`
   11. Scopes: `openid profile groups email`

{{< figure src="portainer.png" alt="Portainer" width="736" style="padding-right: 10px" >}}

## See Also

* [Portainer OAuth Documentation](https://docs.portainer.io/admin/settings/authentication/oauth)

[Authelia]: https://www.authelia.com
[Portainer]: https://www.portainer.io/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
