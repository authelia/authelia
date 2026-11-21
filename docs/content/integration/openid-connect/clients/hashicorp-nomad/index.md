---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "HashiCorp Nomad"
description: "A guide on integrating HashiCorp Nomad with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "HashiCorp Nomad | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring HashiCorp Nomad with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [HashiCorp Nomad]
  - [v2.1.1](https://github.com/hashicorp/nomad/releases/tag/v2.1.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://nomad.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://nomad.{{< sitevar name="domain" nojs="example.com" >}}/ui/settings/tokens`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `nomad`
- **Client Secret:** `insecure_secret`
- **Auth Method Name:** `authelia`
- **Groups:** Users in the `nomad-admins` group are granted the existing `admin` ACL role.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [HashiCorp Nomad]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'nomad'
        client_name: 'HashiCorp Nomad'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://nomad.{{< sitevar name="domain" nojs="example.com" >}}/ui/settings/tokens'
          - 'http://localhost:4649/oidc/callback'
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

To configure [HashiCorp Nomad] there is one method, using the [CLI](#cli).

#### CLI

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
The `http://localhost:4649/oidc/callback` redirect URI is only required if you intend to login using the
`nomad login` command.
{{< /callout >}}

To configure [HashiCorp Nomad] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Create the auth method configuration file:

   ```json {title="config.json"}
   {
     "OIDCDiscoveryURL": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
     "OIDCClientID": "nomad",
     "OIDCClientSecret": "insecure_secret",
     "OIDCEnablePKCE": true,
     "OIDCScopes": ["profile", "email", "groups"],
     "BoundAudiences": ["nomad"],
     "AllowedRedirectURIs": [
       "https://nomad.{{< sitevar name="domain" nojs="example.com" >}}/ui/settings/tokens",
       "http://localhost:4649/oidc/callback"
     ],
     "ClaimMappings": {
       "preferred_username": "username",
       "email": "email"
     },
     "ListClaimMappings": {
       "groups": "groups"
     }
   }
   ```

2. Create the auth method:

   ```shell
   nomad acl auth-method create -name=authelia -type=OIDC -max-token-ttl=1h -token-locality=global -default=true -config=@config.json
   ```

3. Create a binding rule which grants users in the `nomad-admins` group the `admin` ACL role:

   ```shell
   nomad acl binding-rule create -auth-method=authelia -bind-type=role -bind-name=admin -selector='"nomad-admins" in list.groups'
   ```

## See Also

- [HashiCorp Nomad OIDC Auth Method Documentation](https://developer.hashicorp.com/nomad/docs/secure/authentication/oidc)
- [HashiCorp Nomad `acl auth-method create` Command Documentation](https://developer.hashicorp.com/nomad/commands/acl/auth-method/create)
- [HashiCorp Nomad `acl binding-rule create` Command Documentation](https://developer.hashicorp.com/nomad/commands/acl/binding-rule/create)

[Authelia]: https://www.authelia.com
[HashiCorp Nomad]: https://www.nomadproject.io
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
