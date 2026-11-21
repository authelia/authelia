---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Apache Superset"
description: "A guide on integrating Apache Superset with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Apache Superset | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Apache Superset with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Apache Superset]
  - [v6.1.0](https://github.com/apache/superset/releases/tag/6.1.0)

{{% oidc-common bugs="client-credentials-encoding" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://superset.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `superset`
- **Client Secret:** `insecure_secret`
- **Provider Name:** `authelia`
  - This option determines the redirect URI in the format of
    `https://superset.{{< sitevar name="domain" nojs="example.com" >}}/oauth-authorized/<Provider Name>`.
    This means if you change this value, you need to update the redirect URI.
- **Groups:**
  - `superset-admins`: Users in this group are assigned the `Admin` role.
  - `superset-users`: Users in this group are assigned the `Gamma` role.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Apache Superset]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'superset'
        client_name: 'Apache Superset'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://superset.{{< sitevar name="domain" nojs="example.com" >}}/oauth-authorized/authelia'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
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

To configure [Apache Superset] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [Apache Superset] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
configuration in the `superset_config.py` file:

```python {title="superset_config.py"}
from flask_appbuilder.security.manager import AUTH_OAUTH
from superset.security import SupersetSecurityManager


class AutheliaSecurityManager(SupersetSecurityManager):
    def get_oauth_user_info(self, provider, response):
        if provider == 'authelia':
            data = self.oauth_remotes[provider].userinfo()

            return {
                'username': data.get('preferred_username'),
                'email': data.get('email'),
                'first_name': data.get('given_name', ''),
                'last_name': data.get('family_name', ''),
                'role_keys': data.get('groups', []),
            }

        return super().get_oauth_user_info(provider, response)


CUSTOM_SECURITY_MANAGER = AutheliaSecurityManager

AUTH_TYPE = AUTH_OAUTH
AUTH_USER_REGISTRATION = True
AUTH_USER_REGISTRATION_ROLE = 'Public'
AUTH_ROLES_SYNC_AT_LOGIN = True
AUTH_ROLES_MAPPING = {
    'superset-admins': ['Admin'],
    'superset-users': ['Gamma'],
}

OAUTH_PROVIDERS = [
    {
        'name': 'authelia',
        'icon': 'fa-key',
        'token_key': 'access_token',
        'remote_app': {
            'client_id': 'superset',
            'client_secret': 'insecure_secret',
            'server_metadata_url': 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration',
            'client_kwargs': {
                'scope': 'openid profile email groups',
                'code_challenge_method': 'S256',
            },
        },
    },
]
```

## See Also

- [Apache Superset Custom OAuth2 Configuration Documentation](https://superset.apache.org/docs/configuration/configuring-superset#custom-oauth2-configuration)

[Authelia]: https://www.authelia.com
[Apache Superset]: https://superset.apache.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
