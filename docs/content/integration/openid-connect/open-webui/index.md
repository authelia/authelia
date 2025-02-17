---
title: "Open WebUI"
description: "Integrating Open WebUI with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-16T14:55:47-08:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
aliases:
  - /docs/community/oidc-integrations/open-webui.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.18](https://github.com/authelia/authelia/releases/tag/v4.38.18)
* [Open WebUI]
  * [0.5.4](https://github.com/open-webui/open-webui/releases/tag/v0.5.4)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://ai.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `open-webui`
* __Client Secret:__ `insecure_secret`

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
        redirect_uris:
          - 'https://ai.{{< sitevar name="domain" nojs="example.com" >}}/oauth/oidc/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        userinfo_signed_response_alg: 'RS256'
```

### Application

To configure [Open WebUI] to utilize Authelia as an [OpenID Connect 1.0] Provider, specify the below environment variables.

```yaml {title="configuration.yml"}
environment:
  - 'ENABLE_OAUTH_SIGNUP=true'
  - 'OAUTH_MERGE_ACCOUNTS_BY_EMAIL=true'
  - 'OAUTH_CLIENT_ID=open-webui'
  - 'OAUTH_CLIENT_SECRET=insecure_secret'
  - 'OPENID_PROVIDER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
  - 'OAUTH_PROVIDER_NAME=Authelia'
  - 'OAUTH_SCOPES=openid email profile groups'
  - 'ENABLE_OAUTH_ROLE_MANAGEMENT=true'
  - 'OAUTH_ALLOWED_ROLES=openwebui,openwebui-admin'
  - 'OAUTH_ADMIN_ROLES=openwebui-admin'
  - 'OAUTH_ROLES_CLAIM=groups'
```

This configuration limits who can log in to [Open WebUI] to those with either the `openwebui` or `openwebui-admin` groups. Anyone with the `openwebui-admin` group, will be an admin in [Open WebUI].

## See Also

* [Open WebUI OAuth Documentation](https://docs.openwebui.com/features/sso)
* [Open WebUI OAuth Role Management](https://docs.openwebui.com/features/sso#oauth-role-management)

[Authelia]: https://www.authelia.com
[Open WebUI]: https://docs.openwebui.com/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
