---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "DFIR-IRIS"
description: "A guide on integrating DFIR-IRIS with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "DFIR-IRIS | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring DFIR-IRIS with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [DFIR-IRIS]
  - [v2.4.29](https://github.com/dfir-iris/iris-web/releases/tag/v2.4.29)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://dfir-iris.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://dfir-iris.{{< sitevar name="domain" nojs="example.com" >}}/oidc-authorize`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `dfir-iris`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [DFIR-IRIS] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'dfir-iris'
        client_name: 'DFIR-IRIS'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://dfir-iris.{{< sitevar name="domain" nojs="example.com" >}}/oidc-authorize'
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

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="dfir-iris" claims="email,preferred_username" %}}

### Application

To configure [DFIR-IRIS] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [DFIR-IRIS] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables in the `.env` file used by the [DFIR-IRIS] Docker Compose deployment:

##### Standard

```shell {title=".env"}
IRIS_AUTHENTICATION_TYPE=oidc
OIDC_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=dfir-iris
OIDC_CLIENT_SECRET=insecure_secret
OIDC_SCOPES=openid email profile
OIDC_MAPPING_USERNAME=preferred_username
OIDC_MAPPING_EMAIL=email
## Optionally create users in DFIR-IRIS the first time they log in.
IRIS_AUTHENTICATION_CREATE_USER_IF_NOT_EXIST=True
```

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
Unless `IRIS_AUTHENTICATION_CREATE_USER_IF_NOT_EXIST` is set to `True`, users must already exist in [DFIR-IRIS] with a
login matching their Authelia username before they can log in.
{{< /callout >}}

## See Also

- [DFIR-IRIS Authentication Documentation](https://docs.dfir-iris.org/latest/operations/access_control/authentication/)

[Authelia]: https://www.authelia.com
[DFIR-IRIS]: https://dfir-iris.org
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
