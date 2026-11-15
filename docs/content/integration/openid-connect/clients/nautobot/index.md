---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Nautobot"
description: "A guide on integrating Nautobot with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Nautobot | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Nautobot with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Nautobot]
  - [v3.2.9](https://github.com/nautobot/nautobot/releases/tag/v3.2.9)

{{% oidc-common bugs="client-credentials-encoding" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://nautobot.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://nautobot.{{< sitevar name="domain" nojs="example.com" >}}/complete/oidc/`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `nautobot`
- **Client Secret:** `insecure_secret`
- **Groups:** Members of the `nautobot-admins` group are made superusers, and members of the `nautobot-staff` group
  are made staff.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Nautobot] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'nautobot'
        client_name: 'Nautobot'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://nautobot.{{< sitevar name="domain" nojs="example.com" >}}/complete/oidc/'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Nautobot] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [Nautobot] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration in the
`nautobot_config.py` file:

```python {title="nautobot_config.py"}
AUTHENTICATION_BACKENDS = [
    "social_core.backends.open_id_connect.OpenIdConnectAuth",
    "nautobot.core.authentication.ObjectPermissionBackend",
]

SOCIAL_AUTH_OIDC_OIDC_ENDPOINT = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"
SOCIAL_AUTH_OIDC_KEY = "nautobot"
SOCIAL_AUTH_OIDC_SECRET = "insecure_secret"
SOCIAL_AUTH_OIDC_SCOPE = ["groups"]
SOCIAL_AUTH_OIDC_USE_PKCE = True
SOCIAL_AUTH_OIDC_USERNAME_KEY = "preferred_username"
SOCIAL_AUTH_REDIRECT_IS_HTTPS = True
```

To synchronize the [Nautobot] group memberships and the staff and superuser flags from the `groups` claim, set the
following environment variables:

##### Standard

```shell {title=".env"}
NAUTOBOT_SSO_ENABLE_GROUP_SYNC=true
NAUTOBOT_SSO_CLAIMS_GROUP=groups
NAUTOBOT_SSO_STAFF_GROUPS=nautobot-admins,nautobot-staff
NAUTOBOT_SSO_SUPERUSER_GROUPS=nautobot-admins
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  nautobot:
    environment:
      NAUTOBOT_SSO_ENABLE_GROUP_SYNC: 'true'
      NAUTOBOT_SSO_CLAIMS_GROUP: 'groups'
      NAUTOBOT_SSO_STAFF_GROUPS: 'nautobot-admins,nautobot-staff'
      NAUTOBOT_SSO_SUPERUSER_GROUPS: 'nautobot-admins'
```

## See Also

- [Nautobot Single Sign On Documentation](https://docs.nautobot.com/projects/core/en/stable/user-guide/administration/configuration/authentication/sso/)
- [Python Social Auth OpenID Connect Documentation](https://python-social-auth.readthedocs.io/en/latest/backends/oidc.html)

[Authelia]: https://www.authelia.com
[Nautobot]: https://networktocode.com/nautobot/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
