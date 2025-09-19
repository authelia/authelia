---
title: "MinIO"
description: "Integrating MinIO with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/minio/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "MinIO | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring MinIO with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.10](https://github.com/authelia/authelia/releases/tag/v4.39.10)
- [MinIO]
  - [2025-04-22T22-12-26Z](https://github.com/minio/minio/releases/tag/RELEASE.2025-04-22T22-12-26Z)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://minio.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `minio`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [MinIO] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'minio'
        client_name: 'MinIO'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://minio.{{< sitevar name="domain" nojs="example.com" >}}/oauth_callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="minio" %}}

### Application

To configure [MinIO] there are two methods, using [Environment Variables](#environment-variables), or using the
[Web GUI](#web-gui).

#### Environment Variables

To configure [MinIO] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
MINIO_IDENTITY_OPENID_CONFIG_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
MINIO_IDENTITY_OPENID_CLIENT_ID=minio
MINIO_IDENTITY_OPENID_CLIENT_SECRET=insecure_secret
MINIO_IDENTITY_OPENID_SCOPES=openid,profile,email,groups
MINIO_IDENTITY_OPENID_REDIRECT_URI=https://minio.{{< sitevar name="domain" nojs="example.com" >}}/oauth_callback
MINIO_IDENTITY_OPENID_REDIRECT_URI_DYNAMIC=off
MINIO_IDENTITY_OPENID_DISPLAY_NAME=Authelia
MINIO_IDENTITY_OPENID_CLAIM_NAME=groups
MINIO_IDENTITY_OPENID_CLAIM_USERINFO=on
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  minio:
    environment:
      MINIO_IDENTITY_OPENID_CONFIG_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      MINIO_IDENTITY_OPENID_CLIENT_ID: 'minio'
      MINIO_IDENTITY_OPENID_CLIENT_SECRET: 'insecure_secret'
      MINIO_IDENTITY_OPENID_SCOPES: 'openid,profile,email,groups'
      MINIO_IDENTITY_OPENID_REDIRECT_URI: 'https://minio.{{< sitevar name="domain" nojs="example.com" >}}/oauth_callback'
      MINIO_IDENTITY_OPENID_REDIRECT_URI_DYNAMIC: 'off'
      MINIO_IDENTITY_OPENID_DISPLAY_NAME: 'Authelia'
      MINIO_IDENTITY_OPENID_CLAIM_NAME: 'groups'
      MINIO_IDENTITY_OPENID_CLAIM_USERINFO: 'on'
```

#### Web GUI

To configure [MinIO] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [MinIO]
2. On the left hand menu, go to `Identity`, then `OpenID`
3. On the top right, click `Create Configuration`
4. Configure the following options:
   - Name: `authelia`
   - Config URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Client ID: `minio`
   - Client Secret: `insecure_secret`
   - Claim Name: `groups`
   - Display Name: `Authelia`
   - Claim Prefix: Leave Empty
   - Scopes: `openid,profile,email,groups`
   - Redirect URI: `https://minio.{{< sitevar name="domain" nojs="example.com" >}}/oauth_callback`
   - Role Policy: Leave Empty
   - Claim User Info: Enabled
   - Redirect URI Dynamic: Disabled
5. Press `Save` at the bottom
6. Accept the offer of a server restart at the top
   - Refresh the page and sign out if not done so automatically
7. Add your user to an authelia group that matches the policy name in MinIO. There are select [default policies](https://min.io/docs/minio/linux/administration/identity-access-management/policy-based-access-control.html#built-in-policies) that can be used. (The group name and policy name must match.)
8. When the login screen appears again, click the `Other Authentication Methods` open, then select `Authelia` from the list.
9. Login

#### Additional Steps

You may also want to consider adding a
[default policy](https://min.io/docs/minio/linux/administration/identity-access-management/policy-based-access-control.html#built-in-policies)
to your user groups in Authelia.

## See Also

- [MinIO OpenID Identity Management](https://min.io/docs/minio/linux/reference/minio-server/minio-server.html#minio-server-envvar-external-identity-management-openid)

[MinIO]: https://min.io/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
