---
title: "Pangolin"
description: "Integrating Pangolin with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-05-07T10:26:55+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/pangolin/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Pangolin | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Pangolin with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.12](https://github.com/authelia/authelia/releases/tag/v4.39.12)
- [Pangolin]
  - [v1.3.1](https://github.com/fosrl/pangolin/releases/tag/1.3.1)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `pangolin`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Pangolin] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'pangolin'
        client_name: 'Pangolin'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - '<provided by pangolin>'
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

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="pangolin" %}}

### Application

To configure [Pangolin] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Unless you have a Pangolin licence, you will need to manually create the user in Access Control > Users before attempting login.
{{< /callout >}}

To configure [Pangolin] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
instructions:

1. Visit your [Pangolin Web GUI]
2. Visit `Server Admin`
3. Visit `Identity Providers`
4. Select `Add Identity Provider`
5. Select `OpenID Connect`
6. Configure the following options:
   - Name: `Authelia`
   - Provider Type: `OAuth2/OIDC`
   - Client ID: `pangolin`
   - Client Secret: `insecure_secret`
   - Auth URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Identifier Path: `sub`
   - Email Path: `email`
   - Name Path: `name`
   - Scopes: `openid profile email`
7. Click `Create Identity Provider`.
8. On page refresh, note the Redirection URL, and enter it into your Authelia config under `redirect_uris`.

## See Also

- [Pangolin OIDC Documentation](https://docs.fossorial.io/Pangolin/Identity%20Providers/configuring-identity-providers)

[Authelia]: https://www.authelia.com
[Pangolin]: https://fossorial.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
