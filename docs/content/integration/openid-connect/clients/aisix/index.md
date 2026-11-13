---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "AISIX AI Gateway"
description: "A guide on integrating AISIX AI Gateway with the Authelia OpenID Connect 1.0 Provider for SSO, with configuration examples and available options."
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
  title: "AISIX AI Gateway | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring AISIX AI Gateway with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [AISIX AI Gateway]
  - [v1.10.0](https://github.com/api7/aisix/releases/tag/v1.10.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://aisix.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the audience of the Access Tokens in the format of
    `https://aisix.{{< sitevar name="domain" nojs="example.com" >}}`.
    This means if you change this value, you need to update the audience.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `aisix`
- **Client Secret:** `insecure_secret`
- **AISIX Model Alias:** `my-model`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [AISIX AI Gateway]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'aisix'
        client_name: 'AISIX AI Gateway'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'one_factor'
        require_pkce: false
        pkce_challenge_method: ''
        requested_audience_mode: 'implicit'
        audience:
          - 'https://aisix.{{< sitevar name="domain" nojs="example.com" >}}'
        grant_types: ['client_credentials']
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

Notes:

- This client uses the Client Credentials Flow, which is intended for machine-to-machine workloads calling the
  [AISIX AI Gateway], and is always treated as `one_factor` which is why the `authorization_policy` is `one_factor`.
- The `access_token_signed_response_alg` option is required as [AISIX AI Gateway] verifies the Access Token as a JWT
  using the Authelia JSON Web Key Set.
- The `requested_audience_mode` of `implicit` automatically grants the configured `audience`, which [AISIX AI Gateway]
  requires to be present in the `aud` claim of the Access Token.
- The `sub` claim of Access Tokens issued via the Client Credentials Flow is the client id, i.e. `aisix`.

### Application

To configure [AISIX AI Gateway] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [AISIX AI Gateway] to utilize Authelia as an [OpenID Connect 1.0] Provider, add the following entries to
the existing resources file:

```yaml {title="resources.yaml"}
api_keys:
  - display_name: 'authelia-workload'
    key_env: 'WORKLOAD_FALLBACK_API_KEY'
    allowed_models: ['my-model']
    jwt_subject: 'aisix'
    jwt_provider: 'authelia'
oidc_providers:
  - name: 'authelia'
    issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
    audiences: ['https://aisix.{{< sitevar name="domain" nojs="example.com" >}}']
    jwks_uri: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json'
```

Notes:

- The `WORKLOAD_FALLBACK_API_KEY` environment variable must be set to a randomly generated value in the
  [AISIX AI Gateway] environment. This value is a valid long-lived credential and should be kept private.
- The workload obtains an Access Token from
  `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
  using the `client_credentials` grant with the Client ID `aisix` and the Client Secret `insecure_secret`, and uses it
  as the bearer token when making requests to the [AISIX AI Gateway].

## See Also

- [AISIX AI Gateway JWT Authentication Documentation](https://docs.api7.ai/ai-gateway/traffic-controls/jwt-authentication)
- [AISIX AI Gateway Resources File Documentation](https://docs.api7.ai/ai-gateway/reference/resources-file)

[Authelia]: https://www.authelia.com
[AISIX AI Gateway]: https://api7.ai/ai-gateway
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
