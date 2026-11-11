---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "NetBox"
description: "A guide on integrating NetBox with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "NetBox | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring NetBox with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [NetBox]
  - [v4.7.5](https://github.com/netbox-community/netbox/releases/tag/v4.7.5)

{{% oidc-common bugs="client-credentials-encoding" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://netbox.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://netbox.{{< sitevar name="domain" nojs="example.com" >}}/oauth/complete/oidc/`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `netbox`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [NetBox] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'netbox'
        client_name: 'NetBox'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://netbox.{{< sitevar name="domain" nojs="example.com" >}}/oauth/complete/oidc/'
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

To configure [NetBox] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [NetBox] to utilize Authelia as an [OpenID Connect 1.0] Provider, add the following to the NetBox
`configuration.py` file. When using [NetBox Docker](https://github.com/netbox-community/netbox-docker) these settings
can be placed in a Python file in the mounted `configuration/` directory such as `configuration/extra.py`:

```python {title="configuration.py"}
REMOTE_AUTH_ENABLED = True
REMOTE_AUTH_BACKEND = 'social_core.backends.open_id_connect.OpenIdConnectAuth'
SOCIAL_AUTH_OIDC_OIDC_ENDPOINT = 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
SOCIAL_AUTH_OIDC_KEY = 'netbox'
SOCIAL_AUTH_OIDC_SECRET = 'insecure_secret'
SOCIAL_AUTH_OIDC_SCOPE = ['openid', 'email', 'profile']
SOCIAL_AUTH_OIDC_USE_PKCE = True
SOCIAL_AUTH_BACKEND_ATTRS = {
    'oidc': ('Authelia', 'login'),
}
```

Restart [NetBox] after making these changes. An `Authelia` button is then displayed on the NetBox login page.

## See Also

- [NetBox Authentication Documentation](https://netboxlabs.com/docs/netbox/administration/authentication/overview/)
- [Python Social Auth OpenID Connect Documentation](https://python-social-auth.readthedocs.io/en/latest/backends/oidc.html)

[Authelia]: https://www.authelia.com
[NetBox]: https://netboxlabs.com/products/netbox/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
