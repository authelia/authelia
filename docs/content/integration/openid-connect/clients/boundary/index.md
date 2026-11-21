---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Boundary"
description: "A guide on integrating Boundary with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Boundary | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Boundary with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Boundary]
  - [v0.21.3](https://github.com/hashicorp/boundary/releases/tag/v0.21.3)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://boundary.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://boundary.{{< sitevar name="domain" nojs="example.com" >}}/v1/auth-methods/oidc:authenticate:callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `boundary`
- **Client Secret:** `insecure_secret`
- **Scope ID:** `global`
  - The [Boundary] scope the auth method is created in.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Boundary] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'boundary'
        client_name: 'Boundary'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://boundary.{{< sitevar name="domain" nojs="example.com" >}}/v1/auth-methods/oidc:authenticate:callback'
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

To configure [Boundary] there is one method, using the [CLI](#cli).

#### CLI

To configure [Boundary] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Authenticate to [Boundary] with the CLI as an administrator.
2. Create the auth method:

   ```shell
   boundary auth-methods create oidc \
     -scope-id global \
     -name "authelia" \
     -issuer "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}" \
     -client-id "boundary" \
     -client-secret "insecure_secret" \
     -signing-algorithm RS256 \
     -api-url-prefix "https://boundary.{{< sitevar name="domain" nojs="example.com" >}}" \
     -claims-scopes email \
     -claims-scopes profile \
     -claims-scopes groups
   ```

3. Take note of the auth method ID in the output (for example `amoidc_1234567890`).
4. Activate the auth method:

   ```shell
   boundary auth-methods change-state oidc \
     -id amoidc_1234567890 \
     -state active-public
   ```

5. Optionally make the auth method the primary auth method for the scope so that users are automatically created on
   their first login:

   ```shell
   boundary scopes update \
     -id global \
     -primary-auth-method-id amoidc_1234567890
   ```

6. Optionally create a managed group to assign roles based on the Authelia groups of the user:

   ```shell
   boundary managed-groups create oidc \
     -auth-method-id amoidc_1234567890 \
     -name "boundary-admins" \
     -filter '"boundary-admins" in "/userinfo/groups"'
   ```

## See Also

- [Boundary OIDC Auth Method Documentation](https://developer.hashicorp.com/boundary/docs/commands/auth-methods/create)
- [Boundary Managed Groups Documentation](https://developer.hashicorp.com/boundary/docs/concepts/filtering/managed-groups)

[Authelia]: https://www.authelia.com
[Boundary]: https://www.boundaryproject.io
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
