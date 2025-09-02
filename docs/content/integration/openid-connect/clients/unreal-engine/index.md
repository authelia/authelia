---
title: "Unreal Engine"
description: "Integrating Unreal Engine with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-06-08T01:14:26+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/unreal-engine/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Unreal Engine | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Unreal Engine with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.8](https://github.com/authelia/authelia/releases/tag/v4.39.8)
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
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://unreal-engine.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid_connect'
        scopes:
          - 'openid'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
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
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
