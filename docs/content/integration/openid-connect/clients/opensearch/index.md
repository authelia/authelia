---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "OpenSearch"
description: "A guide on integrating OpenSearch with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "OpenSearch | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring OpenSearch with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [OpenSearch]
  - [v3.9.0](https://github.com/opensearch-project/OpenSearch/releases/tag/3.9.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://opensearch-dashboards.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://opensearch-dashboards.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid/login`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `opensearch`
- **Client Secret:** `insecure_secret`
- **Administrator Group:** `opensearch-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [OpenSearch] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    claims_policies:
      opensearch:
        id_token: ['email', 'name', 'preferred_username', 'groups']
    clients:
      - client_id: 'opensearch'
        client_name: 'OpenSearch'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        claims_policy: 'opensearch'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://opensearch-dashboards.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid/login'
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
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [OpenSearch] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Single sign-on requires the OpenSearch Security plugin on OpenSearch and the Security Dashboards plugin on OpenSearch
Dashboards, with TLS enabled on the OpenSearch REST layer. Both plugins are included in the default distributions.
{{< /callout >}}

To configure [OpenSearch] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

##### OpenSearch Security

Add an `openid` authentication domain to the Security plugin configuration. OpenSearch Dashboards still signs in to
OpenSearch with its internal server user, so keep the basic authentication domain and set its `challenge` to `false`.

```yaml {title="config/opensearch-security/config.yml"}
config:
  dynamic:
    authc:
      basic_internal_auth_domain:
        http_enabled: true
        transport_enabled: false
        order: 0
        http_authenticator:
          type: 'basic'
          challenge: false
        authentication_backend:
          type: 'intern'
      openid_auth_domain:
        http_enabled: true
        transport_enabled: false
        order: 1
        http_authenticator:
          type: 'openid'
          challenge: false
          config:
            openid_connect_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
            required_audience: 'opensearch'
            subject_key: 'preferred_username'
            roles_key: 'groups'
        authentication_backend:
          type: 'noop'
```

Map the Authelia groups to OpenSearch roles as backend roles. For example, the following mapping grants full access to
members of the `opensearch-admins` group:

```yaml {title="config/opensearch-security/roles_mapping.yml"}
all_access:
  backend_roles:
    - 'opensearch-admins'
security_rest_api_access:
  backend_roles:
    - 'opensearch-admins'
```

Apply the changed files with `securityadmin.sh` as described in the
[Applying changes to configuration files](https://docs.opensearch.org/latest/security/configuration/security-admin/)
documentation.

##### OpenSearch Dashboards

```yaml {title="config/opensearch_dashboards.yml"}
opensearch_security.auth.type: 'openid'
opensearch_security.openid.connect_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
opensearch_security.openid.client_id: 'opensearch'
opensearch_security.openid.client_secret: 'insecure_secret'
opensearch_security.openid.scope: 'openid profile email groups'
opensearch_security.openid.base_redirect_url: 'https://opensearch-dashboards.{{< sitevar name="domain" nojs="example.com" >}}'
opensearch_security.cookie.secure: true
```

## See Also

- [OpenSearch OpenID Connect Documentation](https://docs.opensearch.org/latest/security/authentication-backends/openid-connect/)
- [OpenSearch Defining Users and Roles Documentation](https://docs.opensearch.org/latest/security/access-control/users-roles/)

[Authelia]: https://www.authelia.com
[OpenSearch]: https://opensearch.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
