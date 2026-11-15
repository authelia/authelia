---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Sourcegraph"
description: "A guide on integrating Sourcegraph with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Sourcegraph | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Sourcegraph with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Sourcegraph]
  - [v8.2.0](https://sourcegraph.com/changelog/releases/8.2.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://sourcegraph.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option is the `externalURL` and determines the redirect URI in the format of
    `https://sourcegraph.{{< sitevar name="domain" nojs="example.com" >}}/.auth/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `sourcegraph`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Sourcegraph] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'sourcegraph'
        client_name: 'Sourcegraph'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://sourcegraph.{{< sitevar name="domain" nojs="example.com" >}}/.auth/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Sourcegraph] there is one method, using the [Site Configuration](#site-configuration).

#### Site Configuration

To configure [Sourcegraph] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Sourcegraph] as a site administrator.
2. Navigate to `Site admin` > `Configuration` > `Site configuration`.
3. Add the following configuration, merging it with any existing `auth.providers` entries:

   ```json {title="site-config.json"}
   {
     "externalURL": "https://sourcegraph.{{< sitevar name="domain" nojs="example.com" >}}",
     "auth.providers": [
       {
         "type": "openidconnect",
         "configID": "authelia",
         "displayName": "Authelia",
         "issuer": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
         "clientID": "sourcegraph",
         "clientSecret": "insecure_secret",
         "allowSignup": true
       }
     ]
   }
   ```

4. Click `Save`.

## See Also

- [Sourcegraph OpenID Connect Documentation](https://sourcegraph.com/docs/admin/auth#openid-connect)

[Authelia]: https://www.authelia.com
[Sourcegraph]: https://sourcegraph.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
