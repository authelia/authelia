---
title: "Incus"
description: "Integrating Incus with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-08-13T18:39:00+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/incus/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Incus | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Incus with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.10](https://github.com/authelia/authelia/releases/tag/v4.38.10)
- [Incus]
  - [v6.0.1](https://github.com/lxc/incus/releases/tag/v6.0.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://incus.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `incus`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Incus]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'incus'
        client_name: 'Incus'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://incus.{{< sitevar name="domain" nojs="example.com" >}}/iodc/callback'
        audience:
          - 'https://incus.{{< sitevar name="domain" nojs="example.com" >}}'
        scopes:
          - 'openid'
          - 'offline_access'
        response_types:
            - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
```

## Application

To configure [Incus] there is one method, using the [CLI](#cli).

#### CLI

To configure [Incus] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Make sure Web Interface is configured and accessible from `https://incus.{{< sitevar name="domain" nojs="example.com" >}}/`.
2. Set the following configuration options, either via individual commands as shown below or via the `incus config edit` command:
   1. Set `oidc.issuer` to match the Authelia Root URL: `incus config set oidc.issuer https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`.
   2. Set `oidc.client.id` to match the `client_id` in the Authelia configuration: `incus config set oidc.client.id incus`.
   3. Set `oidc.audience` to match the Application Root URL: `incus config set oidc.audience https://incus.{{< sitevar name="domain" nojs="example.com" >}}`.
3. You should now see a `Login with SSO` button when you access [Incus] Web Interface.

Example finalized config which can be viewed using `incus config show`:

```yaml
config:
  oidc.issuer: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
  oidc.client.id: incus
  oidc.audience: https://incus.{{< sitevar name="domain" nojs="example.com" >}}
```

## See Also

- [Incus OpenID Connect Documentation](https://linuxcontainers.org/incus/docs/main/authentication/#authentication-openid)

[Authelia]: https://www.authelia.com
[Incus]: https://github.com/lxc/incus
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
