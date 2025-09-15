---
title: "Misago"
description: "Integrating Misago with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/misago/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Misago | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Misago with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Misago]
  - [v0.29.1](https://github.com/tetricky/misago-image/releases/tag/v0.29.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://misago.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `misago`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Misago] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'misago'
        client_name: 'Misago'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        redirect_uris:
          - 'https://misago.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/complete/'
        response_modes:
          - 'query'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Misago] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Misago] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Sign in to the [Misago] Admin Panel.
2. Visit `Settings` and click `OAuth 2`.
3. Configure the following options:
    1. Basic settings:
        - Provider name: `authelia`
        - Client ID: `misago`
        - Client Secret: `insecure_secret`
    2. Initializing Login:
        - Login form URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
        - Scopes: `openid profile email`
    3. Retrieving access token:
        - Access token retrieval URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
        - Request method: `POST`
        - JSON path to access token: `access_token`
    4. Retrieving user data:
        - User data URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
        - Request method: `GET`
        - Access token location: `Query string`
        - Access token name: `access_token`
    5. User JSON mappings:
        - User ID path: `sub`
        - User name path: `name`
        - User e-mail path: `email`
4. Save the configuration.

{{< figure src="misago-step-2.png" alt="Settings" width="736" style="padding-right: 10px" >}}

{{< figure src="misago-step-3-1.png" alt="Basic Settings" width="736" style="padding-right: 10px" >}}

{{< figure src="misago-step-3-2.png" alt="Initializing Login" width="736" style="padding-right: 10px" >}}

{{< figure src="misago-step-3-3.png" alt="Retrieving access token" width="736" style="padding-right: 10px" >}}

{{< figure src="misago-step-3-4.png" alt="Retrieving user data" width="736" style="padding-right: 10px" >}}

{{< figure src="misago-step-3-5.png" alt="User JSON mappings" width="736" style="padding-right: 10px" >}}

---
## See Also

- [Misago] [OAuth 2 Client Configuration guide](https://misago-project.org/t/oauth-2-client-configuration-guide/1147/)

[Misago]: https://misago-project.org/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
