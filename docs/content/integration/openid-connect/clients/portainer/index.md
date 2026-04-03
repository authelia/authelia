---
title: "Portainer"
description: "Integrating Portainer with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/portainer/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Portainer | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Portainer with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Portainer] CE and EE
  - [v2.21.4](https://docs.portainer.io/release-notes#release-2.21.4)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://portainer.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `portainer`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Portainer] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'portainer'
        client_name: 'Portainer'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://portainer.{{< sitevar name="domain" nojs="example.com" >}}'
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
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Portainer] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Portainer] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit Settings
2. Visit Authentication
3. Configure the following options:
   - Authentication Method: `OAuth`
   - Provider: `Custom`
   - Automatic User Provision: Enable if you want users to automatically be created in [Portainer].
   - Client ID: `portainer`
   - Client Secret: `insecure_secret`
   - Authorization URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Access Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Resource URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - Redirect URL: `https://portainer.{{< sitevar name="domain" nojs="example.com" >}}`
   - User Identifier: `preferred_username`
   - Scopes: `openid profile groups email`
   - Auth Style: `In Params`

{{< figure src="portainer.png" alt="Portainer" width="736" style="padding-right: 10px" >}}

## See Also

- [Portainer OAuth Documentation](https://docs.portainer.io/admin/settings/authentication/oauth)

[Authelia]: https://www.authelia.com
[Portainer]: https://www.portainer.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
