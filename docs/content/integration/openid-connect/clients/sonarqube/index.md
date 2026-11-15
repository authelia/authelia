---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "SonarQube"
description: "A guide on integrating SonarQube with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "SonarQube | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring SonarQube with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [SonarQube]
  - [26.11.0.134175](https://github.com/SonarSource/sonarqube/releases/tag/26.11.0.134175)
- [sonar-auth-oidc]
  - [v3.0.0](https://github.com/sonar-auth-oidc/sonar-auth-oidc/releases/tag/v3.0.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://sonarqube.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://sonarqube.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback/oidc`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `sonarqube`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [SonarQube] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'sonarqube'
        client_name: 'SonarQube'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://sonarqube.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback/oidc'
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

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
[SonarQube] Community Build requires the [sonar-auth-oidc] plugin to use [OpenID Connect 1.0]. Install it from the
`Administration` > `Marketplace` page, or copy the plugin jar into the `extensions/plugins` directory of
[SonarQube], and then restart [SonarQube].
{{< /callout >}}

To configure [SonarQube] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [SonarQube] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [SonarQube] as an administrator.
2. Navigate to `Administration` > `Configuration` > `General Settings` > `General`.
3. Set `Server base URL` to `https://sonarqube.{{< sitevar name="domain" nojs="example.com" >}}` without a trailing
   slash.
4. Navigate to `Administration` > `Configuration` > `General Settings` > `Security` > `OpenID Connect`.
5. Configure the following options:
   - Enabled: Checked
   - Issuer URI: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `sonarqube`
   - Client secret: `insecure_secret`
   - Scopes: `openid profile email groups`
   - Allow users to sign-up: Checked
   - Login generation strategy: `Preferred username`
   - Synchronize groups: Checked
   - Groups claim name: `groups`
   - Login button text: `Authelia`
6. Click `Save` for each option.

When `Synchronize groups` is enabled, users are added to the existing [SonarQube] groups which have the same name as
their Authelia groups, for example members of the `sonar-administrators` group in Authelia will be added to the
`sonar-administrators` group in [SonarQube].

## See Also

- [sonar-auth-oidc Configuration Documentation](https://github.com/sonar-auth-oidc/sonar-auth-oidc#configuration)

[Authelia]: https://www.authelia.com
[SonarQube]: https://www.sonarsource.com/products/sonarqube/
[sonar-auth-oidc]: https://github.com/sonar-auth-oidc/sonar-auth-oidc
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
