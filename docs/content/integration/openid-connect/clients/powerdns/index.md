---
title: "PowerDNS Admin"
description: "Integrating PowerDNS Admin with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/powerdns/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "PowerDNS Admin | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring PowerDNS Admin with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [PowerDNS Admin]
  - [v0.4.1](https://github.com/PowerDNS-Admin/PowerDNS-Admin/releases/tag/v0.4.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://powerdns.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `powerdns`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [PowerDNS Admin] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'powerdns'
        client_name: 'PowerDNS Admin'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://powerdns.{{< sitevar name="domain" nojs="example.com" >}}/oidc/authorized'
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

To configure [PowerDNS Admin] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [PowerDNS Admin] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit Settings
2. Visit Authentication
3. Visit OpenID Connect OAuth
4. Configure the following options:
   - Enable OpenID Connect OAuth: Enabled
   - Client ID: `powerdns`
   - Client Secret: `insecure_secret`
   - Scopes: `openid profile groups email`
   - API URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - Enable OIDC OAuth Auto-Configuration: Enabled
   - Metadata URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Logout URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/logout`
   - Username: `preferred_username`
   - Email: `email`
   - First Name: `given_name`
   - Last Name: `family_name`
   - Autoprovision Account Name property: `preferred_username`
   - Autoprovision Account Description property: `name`

{{< figure src="powerdns.png" alt="PowerDNS Admin" width="736" style="padding-right: 10px" >}}

## See Also

- [Portainer OAuth Documentation](https://docs.portainer.io/admin/settings/authentication/oauth)

[Authelia]: https://www.authelia.com
[PowerDNS Admin]: https://github.com/PowerDNS/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
