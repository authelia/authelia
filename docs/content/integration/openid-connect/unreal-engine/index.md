---
title: "Unreal Engine"
description: "Integrating Unreal Engine with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-06-08T08:46:06+10:00
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
  - [v4.39.4](https://github.com/authelia/authelia/releases/tag/v4.39.4)
- [Unreal Engine]
  - [v5.6](https://dev.epicgames.com/documentation/en-us/unreal-engine/unreal-engine-5-6-release-notes)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://unreal-engine.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `unreal-engine`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Unreal Engine] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'unreal-engine'
        client_name: 'Unreal Engine'
        public: true
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://unreal-engine.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid_connect'
        scopes:
          - 'openid'
          - 'groups'
        response_types:
          - 'id_token'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [Unreal Engine] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [Unreal Engine] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```json
{
  "ServerUrl": "https://unreal-engine.{{< sitevar name="domain" nojs="example.com" >}}",
  "HttpsPort": "443",
  "AuthMethod": "OpenIdConnect",
  "OidcAuthority": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
  "OidcAudience": "unreal-engine",
  "OidcClientId": "unreal-engine",
  "OidcClientSecret": "insecure-secret",
  "OidcSigninRedirect": "https://unreal-engine.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid_connect",
  "OidcRequestedScopes": "openid groups",
  "OidcClaimHordeUserMapping": ["groups"],
  "AdminClaimType": "http://epicgames.com/ue/horde/role",
  "AdminClaimValue": "unreal-engine-admin"
}
```

## See Also

- [Unreal Engine > Horde Server Settings Documentation](https://dev.epicgames.com/documentation/en-us/unreal-engine/horde-settings-for-unreal-engine#serversettings)

[Authelia]: https://www.authelia.com
[Unreal Engine]: https://www.unrealengine.com/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
