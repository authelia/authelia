---
title: "Stalwart"
description: "Integrating Stalwart with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T18:35:57+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/stalwart/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Stalwart | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Stalwart with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [Stalwart]
  - [v0.11.7](https://github.com/stalwartlabs/mail-server/releases/tag/v0.11.7)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://example.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `stalwart`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
This client is created as an example but [Stalwart](https://stalw.art) doesn't use this client directly, it just queries
the Introspection or User Info Endpoint given an Access Token. This means you must procure the relevant Access Token
yourself in order to use it. In this example we issue it to an application that has a URI different to
[Stalwart](https://stalw.art) which allows that application to leverage OAuth 2.0 to authenticate on a users behalf.
{{< /callout >}}

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Stalwart] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'stalwart'
        client_name: 'Stalwart'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://example.{{< sitevar name="domain" nojs="example.com" >}}'
        scopes:
          - 'openid'
          - 'profile'
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

To configure [Stalwart] there are two methods, using the [Configuration File](#configuration-file),  or using the [Web GUI](#web-gui).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `config.toml`.
{{< /callout >}}

To configure [Stalwart] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```toml {title="config.toml"}
[directory."authelia"]
type = "oidc"
timeout = "15s"
endpoint.url = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo"
endpoint.method = "userinfo"
fields.email = "email"
fields.username = "preferred_username"
fields.full-name = "name"
```

#### Web GUI

To configure [Stalwart] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Stalwart].
2. Navigate to Settings.
3. Navigate to Authentication.
4. Navigate to Directories.
5. Click Create Directory.
6. Configure the following options:
   - Directory Id: `authelia`
   - Type: `OpenID Connect`
   - URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - Type: `OpenID Connect Userinfo`
   - Timeout: `15 seconds`
   - E-mail Field: `email`
   - Username field: `preferred_username`
   - Name field: `name`
7. Press `Save & Reload` at the bottom.

## See Also

- [Stalwart OpenID Connect Directory Guide](https://stalw.art/docs/auth/backend/oidc/)

[Stalwart]: https://stalw.art/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
