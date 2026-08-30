---
title: "WG-Portal"
description: "Integrating WG-Portal with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2026-08-06T20:00:00+02:00
draft: false
images: []
weight: 620
toc: true
aliases: []
support:
  level: community
  versions: true
  integration: true
seo:
  title: "WG-Portal | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring WG-Portal with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.20](https://github.com/authelia/authelia/releases/tag/v4.39.20)
- [WG-Portal]
  - [v2.3.1](https://github.com/h44z/wg-portal/releases/tag/v2.3.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://wg-portal.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Provider ID:** `{{< sitevar name="provider-id" nojs="authelia" >}}`
- **Client ID:** `{{< sitevar name="client-id" nojs="wg-portal" >}}`
- **Client Secret:** `insecure_secret`
- **Admin Groupname:** `{{< sitevar name="admin-group" nojs="admins" >}}`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [WG-Portal] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: '{{< sitevar name="client-id" nojs="wg-portal" >}}'
        client_name: 'WG-Portal'
        client_secret: '$pbkdf2-sha512$310000$hDm3SNqIGebGH33JVGnCZA$5XA6C8mg.j2MejL1uJfHwFPEqXbW2K.xufEuIIpwN.0Jn0dKs/LLvRKbPTo7MTOjTpwoAVyZj6BsUTHeBjd3sQ'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://wg-portal.{{< sitevar name="domain" nojs="example.com" >}}/api/v0/auth/login/{{< sitevar name="provider-id" nojs="authelia" >}}/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [WG-Portal] there is one method, using the [config.yml][WG-Portal OIDC Documentation] file.

#### Config.yml

To configure [WG-Portal] to utilize [Authelia] as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="config.yml"}
auth:
  # Other authentication configurations go here
  oidc:
    - provider_name: {{< sitevar name="provider-id" nojs="authelia" >}}
      display_name: Log in with Authelia
      base_url: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
      client_id: {{< sitevar name="client-id" nojs="wg-portal" >}}
      client_secret: insecure_secret
      extra_scopes:
        - email
        - profile
        - groups
      field_map:
        user_identifier: preferred_username
        email: email
        firstname: given_name
        lastname: family_name
        user_groups: groups
      admin_mapping:
        admin_group_regex: ^{{< sitevar name="admin-group" nojs="admins" >}}$
```

If you do not want to map an admin group just omit the _admin_mapping_ section.

## See Also

- [WG-Portal OIDC Documentation]

[Authelia]: https://www.authelia.com
[WG-Portal]: https://wgportal.org
[WG-Portal OIDC Documentation]: https://wgportal.org/master/documentation/configuration/overview/#oidc
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
