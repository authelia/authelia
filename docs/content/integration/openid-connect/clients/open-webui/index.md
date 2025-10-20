---
title: "Open WebUI"
description: "Integrating Open WebUI with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-25T00:03:43+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/open-webui/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Open WebUI | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Open WebUI with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.13](https://github.com/authelia/authelia/releases/tag/v4.39.13)
- [Open WebUI]
  - [v0.6.13](https://github.com/open-webui/open-webui/releases/tag/v0.6.13)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://ai.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `open-webui`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Open WebUI] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'open-webui'
        client_name: 'Open WebUI'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://ai.{{< sitevar name="domain" nojs="example.com" >}}/oauth/oidc/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Open WebUI] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
This configuration limits who can log in to those with the `openwebui` or `openwebui-admin` groups. This is configured
via the `OAUTH_ALLOWED_ROLES` environment variable. Anyone with the `openwebui-admin` group will be an admin for the
application. This is configured via the `OAUTH_ADMIN_ROLES` environment variable.
{{< /callout >}}

To configure [Open WebUI] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
WEBUI_URL=https://ai.{{< sitevar name="domain" nojs="example.com" >}}
ENABLE_OAUTH_SIGNUP=true
OAUTH_MERGE_ACCOUNTS_BY_EMAIL=true
OAUTH_CLIENT_ID=open-webui
OAUTH_CLIENT_SECRET=insecure_secret
OPENID_PROVIDER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
OAUTH_PROVIDER_NAME=Authelia
OAUTH_SCOPES=openid email profile groups
ENABLE_OAUTH_ROLE_MANAGEMENT=true
OAUTH_ALLOWED_ROLES=openwebui,openwebui-admin
OAUTH_ADMIN_ROLES=openwebui-admin
OAUTH_ROLES_CLAIM=groups
```

###### Docker Compose

```yaml {title="comppse.yml"}
services:
  open-webui:
    environment:
      WEBUI_URL: 'https://ai.{{< sitevar name="domain" nojs="example.com" >}}'
      ENABLE_OAUTH_SIGNUP: 'true'
      OAUTH_MERGE_ACCOUNTS_BY_EMAIL: 'true'
      OAUTH_CLIENT_ID: 'open-webui'
      OAUTH_CLIENT_SECRET: 'insecure_secret'
      OPENID_PROVIDER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      OAUTH_PROVIDER_NAME: 'Authelia'
      OAUTH_SCOPES: 'openid email profile groups'
      ENABLE_OAUTH_ROLE_MANAGEMENT: 'true'
      OAUTH_ALLOWED_ROLES: 'openwebui,openwebui-admin'
      OAUTH_ADMIN_ROLES: 'openwebui-admin'
      OAUTH_ROLES_CLAIM: 'groups'
```

## See Also

- [Open WebUI OAuth Documentation](https://docs.openwebui.com/features/sso)
- [Open WebUI OAuth Role Management](https://docs.openwebui.com/features/sso#oauth-role-management)

[Authelia]: https://www.authelia.com
[Open WebUI]: https://docs.openwebui.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
