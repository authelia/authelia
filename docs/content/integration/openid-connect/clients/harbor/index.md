---
title: "Harbor"
description: "Integrating Harbor with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/harbor/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Harbor | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Harbor with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.15](https://github.com/authelia/authelia/releases/tag/v4.38.15)
- [Harbor]
  - [v2.11.2](https://github.com/goharbor/harbor/releases/tag/v2.11.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://harbor.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `harbor`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Harbor] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'harbor'
        client_name: 'Harbor'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://harbor.{{< sitevar name="domain" nojs="example.com" >}}/c/oidc/callback'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'profile'
          - 'groups'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Harbor] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Harbor] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit Administration
2. Visit Configuration
3. Visit Authentication
4. Select `OIDC` from the `Auth Mode` drop down
5. Configure the following options:
   - OIDC Provider Name: `Authelia`
   - OIDC Provider Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - OIDC Client ID: `harbor`
   - OIDC Client Secret: `insecure_secret`
   - Group Claim Name: `groups`
   - For OIDC Admin Group you can specify a group name that matches your authentication backend.
   - OIDC Scope: `openid,profile,email,groups,offline_access`
   - Verify Certificate: Enabled.
   - Automatic onboarding: Enabled if you want users to be created by default.
   - Username Claim: `preferred_username`
6. Click `Test OIDC Server`
7. Click `Save`

## See Also

- [Harbor OpenID Connect Provider Documentation](https://goharbor.io/docs/2.11.0/administration/configure-authentication/oidc-auth/)

[Authelia]: https://www.authelia.com
[Harbor]: https://goharbor.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
