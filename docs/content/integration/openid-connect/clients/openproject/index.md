---
title: "OpenProject"
description: "Integrating OpenProject with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T18:35:57+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/openproject/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "OpenProject | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring OpenProject with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [OpenProject]
  - [v15.4.2](https://www.openproject.org/docs/release-notes/#1550)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://openproject.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `openproject`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [OpenProject] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'openproject'
        client_name: 'OpenProject'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://openproject.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc-authelia/callback'
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

To configure [OpenProject] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [OpenProject] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [OpenProject].
2. Navigate to Administration by clicking on your profile and selecting Administration.
3. Navigate to Authentication.
4. Navigate to OpenID providers.
5. Click the `+ OpenID Provider` button.
6. Select `Custom`.
7. Configure the following options:
   - Display Name: `Authelia`
   - I have a disovery endpoint URL: Selected
   - Endpoint URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
8. Click the `Continue` button.
9. Configure the following options:
   - Client ID: `openproject`
   - Client secret: `insecure_secret`
   - Post Logout Redirect URI: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/logout`
10. Click the `Continue` button.
11. Click the `Attribute mapping` edit button.
12. Configure the following options:
    - Mapping for Username: `preferred_username`
    - Mapping for Email: `email`
    - Mapping for First Name: `given_name`
    - Mapping for Last Name: `family_name`
13. Click the `Finish setup` button.

## See Also

- [OpenProject OpenID Providers Guide](https://www.openproject.org/docs/system-admin-guide/authentication/openid-providers/)

[OpenProject]: https://www.openproject.org
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
