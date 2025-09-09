---
title: "SFTPGo"
description: "Integrating SFTPGo with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T11:03:16+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/sftpgo/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "SFTPGo | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring SFTPGo with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.9](https://github.com/authelia/authelia/releases/tag/v4.39.9)
- [SFTPGo]
  - [v2.6.6](https://github.com/drakkan/sftpgo/releases/tag/v2.6.6)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://sftpgo.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `sftpgo`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
The `sftpgo_role` user attribute renders the value `admin` if the user is in the `sftpgo_admins` group within Authelia,
renders the value `manager` if they are in the `sftpgo_managers` group, otherwise it renders `user`. You can adjust this
to your preference to assign a role to the appropriate user groups.
{{< /callout >}}

The following YAML configuration is an example __Authelia__ [client configuration] for use with [SFTPGo] which
will operate with the application example:

```yaml {title="configuration.yml"}
definitions:
  user_attributes:
    sftpgo_role:
      expression: '"sftpgo_admins" in groups ? "admin" : "sftpgo_managers" in groups ? "manager" : "user"'
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    claims_policies:
      sftpgo:
        id_token: ['preferred_username', 'sftpgo_role']
        custom_claims:
          sftpgo_role: {}
    scopes:
      sftpgo:
        claims:
          - 'sftpgo_role'
    clients:
      - client_id: 'sftpgo'
        client_name: 'SFTPGo'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        claims_policy: 'sftpgo'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://sftpgo.{{< sitevar name="domain" nojs="example.com" >}}/web/oidc/redirect'
          - 'https://sftpgo.{{< sitevar name="domain" nojs="example.com" >}}/web/oauth2/redirect'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'sftpgo'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [SFTPGo] there are two methods, using the [Configuration File](#configuration-file), or using the
[Environment Variables](#environment-variables).

#### Configuration File

To configure [SFTPGo] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```json
{
  "oidc": {
    "client_id": "sftpgo",
    "client_secret": "insecure_secret",
    "config_url": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
    "redirect_base_url": "https://sftpgo.{{< sitevar name="domain" nojs="example.com" >}}",
    "scopes": [
      "openid",
      "profile",
      "email",
      "sftpgo"
    ],
    "username_field": "preferred_username",
    "role_field": "sftpgo_role",
    "implicit_roles": false,
    "custom_fields": []
  }
}
```

#### Environment Variables

To configure [SFTPGo] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
SFTPGO_HTTPD__BINDINGS__0__OIDC__CLIENT_ID=sftpgo
SFTPGO_HTTPD__BINDINGS__0__OIDC__CLIENT_SECRET=insecure_secret
SFTPGO_HTTPD__BINDINGS__0__OIDC__CONFIG_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
SFTPGO_HTTPD__BINDINGS__0__OIDC__REDIRECT_BASE_URL=https://sftpgo.{{< sitevar name="domain" nojs="example.com" >}}
SFTPGO_HTTPD__BINDINGS__0__OIDC__SCOPES=openid,profile,email,sftpgo
SFTPGO_HTTPD__BINDINGS__0__OIDC__USERNAME_FIELD=preferred_username
SFTPGO_HTTPD__BINDINGS__0__OIDC__ROLE_FIELD=sftpgo_role
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  sftpgo:
    environment:
      SFTPGO_HTTPD__BINDINGS__0__OIDC__CLIENT_ID: 'sftpgo'
      SFTPGO_HTTPD__BINDINGS__0__OIDC__CLIENT_SECRET: 'insecure_secret'
      SFTPGO_HTTPD__BINDINGS__0__OIDC__CONFIG_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      SFTPGO_HTTPD__BINDINGS__0__OIDC__REDIRECT_BASE_URL: 'https://sftpgo.{{< sitevar name="domain" nojs="example.com" >}}'
      SFTPGO_HTTPD__BINDINGS__0__OIDC__SCOPES: 'openid,profile,email'
      SFTPGO_HTTPD__BINDINGS__0__OIDC__USERNAME_FIELD: 'preferred_username'
      SFTPGO_HTTPD__BINDINGS__0__OIDC__ROLE_FIELD: 'sftpgo_role'
```

## See Also

- [SFTPGo OpenID Connect Documentation](https://docs.sftpgo.com/2.6/oidc/)

[Authelia]: https://www.authelia.com
[SFTPGo]: https://sftpgo.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
