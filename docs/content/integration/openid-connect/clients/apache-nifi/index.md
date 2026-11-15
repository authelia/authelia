---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Apache NiFi"
description: "A guide on integrating Apache NiFi with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Apache NiFi | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Apache NiFi with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Apache NiFi]
  - [v2.12.0](https://github.com/apache/nifi/releases/tag/rel/nifi-2.12.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://nifi.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://nifi.{{< sitevar name="domain" nojs="example.com" >}}/nifi-api/access/oidc/callback/consumer`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `nifi`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Apache NiFi] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'nifi'
        client_name: 'Apache NiFi'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://nifi.{{< sitevar name="domain" nojs="example.com" >}}/nifi-api/access/oidc/callback/consumer'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
          - 'offline_access'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Apache NiFi] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `nifi.properties`.
{{< /callout >}}

To configure [Apache NiFi] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```properties {title="nifi.properties"}
nifi.security.user.oidc.discovery.url=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
nifi.security.user.oidc.client.id=nifi
nifi.security.user.oidc.client.secret=insecure_secret
nifi.security.user.oidc.preferred.jwsalgorithm=RS256
nifi.security.user.oidc.additional.scopes=profile,groups,offline_access
nifi.security.user.oidc.claim.identifying.user=email
nifi.security.user.oidc.fallback.claims.identifying.user=preferred_username
nifi.security.user.oidc.claim.groups=groups
```

## See Also

- [Apache NiFi System Administrator's Guide: OpenID Connect](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#openid_connect)

[Authelia]: https://www.authelia.com
[Apache NiFi]: https://nifi.apache.org
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
