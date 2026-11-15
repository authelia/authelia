---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Apache Airflow"
description: "A guide on integrating Apache Airflow with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Apache Airflow | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Apache Airflow with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Apache Airflow]
  - [3.4.0](https://github.com/apache/airflow/releases/tag/3.4.0)
  - [FAB Provider 3.11.0](https://pypi.org/project/apache-airflow-providers-fab/3.11.0/)

{{% oidc-common bugs="client-credentials-encoding" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://airflow.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://airflow.{{< sitevar name="domain" nojs="example.com" >}}/auth/oauth-authorized/authelia`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `airflow`
- **Client Secret:** `insecure_secret`
- **Groups:** The `airflow-admins` and `airflow-users` groups are mapped to the `Admin` and `User` [Apache Airflow]
  roles respectively.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Apache Airflow] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'airflow'
        client_name: 'Apache Airflow'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://airflow.{{< sitevar name="domain" nojs="example.com" >}}/auth/oauth-authorized/authelia'
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

To configure [Apache Airflow] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="note" title="Auth Manager" icon="outline/info-circle" >}}
This configuration requires the Flask AppBuilder (FAB) auth manager provided by the `apache-airflow-providers-fab`
package. On [Apache Airflow] 3 this auth manager must be enabled by setting the
`AIRFLOW__CORE__AUTH_MANAGER` environment variable (or the `auth_manager` option in the `core` section of
`airflow.cfg`) to `airflow.providers.fab.auth_manager.fab_auth_manager.FabAuthManager`.
{{< /callout >}}

To configure [Apache Airflow] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration
in the `webserver_config.py` file located in the `AIRFLOW_HOME` directory:

```python {title="webserver_config.py"}
from flask_appbuilder.const import AUTH_OAUTH

AUTH_TYPE = AUTH_OAUTH
AUTH_USER_REGISTRATION = True
AUTH_USER_REGISTRATION_ROLE = "Viewer"
AUTH_ROLES_SYNC_AT_LOGIN = True
AUTH_ROLES_MAPPING = {
    "airflow-admins": ["Admin"],
    "airflow-users": ["User"],
}

OAUTH_PROVIDERS = [
    {
        "name": "authelia",
        "icon": "fa-key",
        "token_key": "access_token",
        "remote_app": {
            "client_id": "airflow",
            "client_secret": "insecure_secret",
            "server_metadata_url": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration",
            "api_base_url": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/",
            "client_kwargs": {
                "scope": "openid email profile groups",
                "code_challenge_method": "S256",
            },
        },
    },
]
```

## See Also

- [Apache Airflow FAB Auth Manager Webserver Authentication Documentation](https://airflow.apache.org/docs/apache-airflow-providers-fab/stable/auth-manager/webserver-authentication.html)

[Authelia]: https://www.authelia.com
[Apache Airflow]: https://airflow.apache.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
