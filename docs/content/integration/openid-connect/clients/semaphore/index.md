---
title: "Semaphore"
description: "Integrating Semaphore with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T11:03:16+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/semaphore/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Semaphore | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Semaphore with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.8](https://github.com/authelia/authelia/releases/tag/v4.39.8)
- [Semaphore]
  - [v2.13.14](https://github.com/semaphoreui/semaphore/releases/tag/v2.13.14)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://semaphore.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `semaphore`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Semaphore] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'semaphore'
        client_name: 'Semaphore'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://semaphore.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/authelia/redirect'
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

To configure [Semaphore] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

```json {title="config.json"}
{
  "oidc_providers":  {
    "authelia": {
      "display_name": "Authelia",
      "provider_url": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
      "client_id": "semaphore",
      "client_secret": "insecure_secret",
      "redirect_url": "https://semaphore.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/authelia/redirect",
      "scopes": ["openid", "profile", "email"],
      "username_claim": "preferred_username",
      "email_claim": "email",
      "name_claim": "name",
      "order": 1
    }
  }
}
```

## See Also

- [Semaphore OpenID Setup Guide](https://docs.semaphoreui.com/administration-guide/openid/)
  - [Semaphore OpenID Setup Guide (Authelia)](https://docs.semaphoreui.com/administration-guide/openid/authelia/)

[Semaphore]: https://semaphoreui.com/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
