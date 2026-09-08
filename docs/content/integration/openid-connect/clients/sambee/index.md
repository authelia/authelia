---
title: "Sambee"
description: "Integrating Sambee with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2026-08-18
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
  title: "Sambee | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Sambee with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.20](https://github.com/authelia/authelia/releases/tag/v4.39.20)
- [Sambee]
  - [v0.9.37](https://github.com/helgeklein/sambee/releases/tag/v0.9.37)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://sambee.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `sambee`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Sambee] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    lifespans:
      custom:
        sambee:
          # Maximum useful length: Sambee's interactive sign-in interval + 1 day
          refresh_token: '31d'
    clients:
      - client_id: 'sambee'
        client_name: 'Sambee'
        # The digest of 'insecure_secret'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://sambee.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        response_types:
          - 'code'
        lifespan: 'sambee'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'offline_access'
          # Add groups when using group admission or group-based role mappings.
          - 'groups'
```

### Application

To configure [Sambee] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Sambee] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to your Sambee as an administrator.
2. Access `Settings`.
3. Select `Authentication`.
4. Select one of these values as **Authentication mode**:

- `OIDC or password`
- `OIDC only`

5. Select **Configure OIDC**.
6. In the OIDC configuration dialog, enter the following values:

- Provider name: `Authelia`
- Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
- Client ID: `sambee`
- Client secret: `insecure_secret`
- Scopes: `openid, offline_access, profile, email, groups`

7. As **Admission**, select `All authenticated users`.
8. As **Role assignment**, select `All users are assigned to the same role`.
9. As **Assigned role**, select `Editor`.
10. Select **Connect and test**.
11. Select **Activate configuration**.

## See Also

- [Sambee: OpenID Connect Authentication Setup](https://sambee.net/docs/admin-guide/authentication/openid-connect-authentication-setup/)
- [Helge Klein: Samba & SMB Web Access Through Sambee With Automatic HTTPS & OIDC SSO](https://helgeklein.com/blog/samba-smb-web-access-through-sambee-with-automatic-https-oidc-sso/)

[Authelia]: https://www.authelia.com
[Sambee]: https://sambee.net/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
