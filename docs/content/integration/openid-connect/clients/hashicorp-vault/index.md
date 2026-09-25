---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "HashiCorp Vault"
description: "Integrating HashiCorp Vault with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/hashicorp-vault/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "HashiCorp Vault | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring HashiCorp Vault with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [HashiCorp Vault]
  - [v1.8.1](https://developer.hashicorp.com/vault/docs/updates/release-notes)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://vault.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `vault`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [HashiCorp Vault] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'vault'
        client_name: 'HashiCorp Vault'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://vault.{{< sitevar name="domain" nojs="example.com" >}}/oidc/callback'
          - 'https://vault.{{< sitevar name="domain" nojs="example.com" >}}/ui/vault/auth/oidc/oidc/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [HashiCorp Vault] to utilize Authelia as an [OpenID Connect 1.0] Provider please see the links in the
[see also](#see-also) section.

Below are Terraform code snippets that describe how to configure the OIDC auth backend for Vault and a Vault role to go with it.

Terraform Vault Provider v5.11 was used with this example.

```hcl {title="vault_jwt_auth_backend resource"}
resource "vault_jwt_auth_backend" "anthelia" {
  description                   = ""
  path                          = "oidc"
  type                          = "oidc"
  oidc_discovery_url            = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"
  oidc_client_id                = "vault"
  oidc_client_secret_wo         = "insecure_secret"
  oidc_client_secret_wo_version = 1
  bound_issuer                  = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"
  default_role                  = "authelia"
  oidc_response_types           = ["code"]
}
```

```hcl {title="vault_jwt_auth_backend_role resource"}
resource "vault_jwt_auth_backend_role" "anthelia" {
  backend    = vault_jwt_auth_backend.authelia.path
  role_name  = "authelia"
  role_type  = "oidc"
  user_claim = "sub" # REQUIRED
  allowed_redirect_uris = [
    "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/oidc/callback",
    "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/ui/vault/auth/oidc/callback"
  ]
  oidc_scopes = [
    "profile",
    "email"
  ]
  token_policies = [
    "default"
  ]
  token_ttl     = 3600  # in seconds
  token_max_ttl = 28800 # in seconds
}
```

## See Also

- [HashiCorp Vault JWT/OIDC Auth Documentation](https://www.vaultproject.io/docs/auth/jwt)
- [HashiCorp Vault OpenID Connect Providers Documentation](https://www.vaultproject.io/docs/auth/jwt/oidc-providers)
- [Terraform Vault Provider Resource - `vault_jwt_auth_backend`](https://registry.terraform.io/providers/hashicorp/vault/latest/docs/resources/jwt_auth_backend)
- [Terraform Vault Provider Resource - `vault_jwt_auth_backend_role`](https://registry.terraform.io/providers/hashicorp/vault/latest/docs/resources/jwt_auth_backend_role)

[Authelia]: https://www.authelia.com
[HashiCorp Vault]: https://www.vaultproject.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
