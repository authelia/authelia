---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Moodle"
description: "A guide on integrating Moodle with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
summary: ""
date: 2026-11-12T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Moodle | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Moodle with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Moodle]
  - [v5.3.1](https://github.com/moodle/moodle/releases/tag/v5.3.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://moodle.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://moodle.{{< sitevar name="domain" nojs="example.com" >}}/admin/oauth2callback.php`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `moodle`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Moodle] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'moodle'
        client_name: 'Moodle'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://moodle.{{< sitevar name="domain" nojs="example.com" >}}/admin/oauth2callback.php'
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
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Moodle] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Moodle] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Moodle] as an administrator.
2. Navigate to `Site administration`.
3. Navigate to `Server`.
4. Navigate to `OAuth 2 services`.
5. Click `Custom`.
6. Configure the following options:
   - Name: `Authelia`
   - Client ID: `moodle`
   - Client secret: `insecure_secret`
   - Authenticate token requests via HTTP headers: Unchecked
   - Service base URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Scopes included in a login request: `openid profile email`
   - This service will be used: `Login page and internal services`
   - Name displayed on the login page: `Authelia`
7. Click `Save changes`.
8. Click the `Configure user field mappings` icon for the `Authelia` service and ensure the `given_name`,
   `family_name`, and `email` external fields are mapped to the `firstname`, `lastname`, and `email` internal fields
   respectively.
9. Navigate to `Site administration`.
10. Navigate to `Plugins`.
11. Navigate to `Authentication`.
12. Navigate to `Manage authentication`.
13. Enable the `OAuth 2` authentication plugin.

## See Also

- [Moodle OAuth 2 services Documentation](https://docs.moodle.org/en/OAuth_2_services)
- [Moodle OAuth 2 authentication Documentation](https://docs.moodle.org/en/OAuth_2_authentication)

[Authelia]: https://www.authelia.com
[Moodle]: https://moodle.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
