---
title: "Drupal"
description: "Integrating Drupal with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T11:03:16+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/drupal/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Drupal | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Drupal with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [Drupal]
  - [v10.4.0](https://www.drupal.org/project/drupal/releases/10.4.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://drupal.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `drupal`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Drupal] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'drupal'
        client_name: 'Drupal'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://drupal.{{< sitevar name="domain" nojs="example.com" >}}/openid-connect/generic'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Drupal] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Drupal] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit `https://drupal.{{< sitevar name="domain" nojs="example.com" >}}/admin/config/services/openid-connect`.
2. Configure the following options:
   - Enabled OpenID Connect clients: `Generic`
   - Client ID: `drupal`
   - Client Secret: `insecure_secret`
   - Authorization Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - UserInfo Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
3. Visit `https://drupal.{{< sitevar name="domain" nojs="example.com" >}}/admin/config/people/accounts`.
4. Configure the following options:
   - Override registration settings: Enabled

## See Also

- [Drupal OpenID Connect Generic Client Documentation](https://www.drupal.org/node/2274339#toc-5)

[Authelia]: https://www.authelia.com
[Drupal]: https://new.drupal.org/home
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
