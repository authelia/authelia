---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "ChatGPT"
description: "A guide on integrating ChatGPT with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "ChatGPT | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring ChatGPT with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [ChatGPT]

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[ChatGPT] single sign-on requires a ChatGPT Business, Enterprise, or Edu workspace, a Global Admin or workspace owner,
and a verified domain. Domain verification is outside the scope of this guide. The Authelia instance must be reachable
from the public internet.
{{< /callout >}}

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `chatgpt`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [ChatGPT] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'chatgpt'
        client_name: 'ChatGPT'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - ''  # See step 4 below.
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

{{% oidc-escape-hatch-claims-hydration client_id="chatgpt" claims="email,email_verified,name,given_name,family_name" %}}

### Application

To configure [ChatGPT] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [ChatGPT] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to the [OpenAI Identity page](https://admin.openai.com/settings/identity) as a Global Admin.
2. In the `Single Sign-On` section click `Set up SSO`.
3. Select `Custom OIDC` as the identity provider.
4. Copy the `Login redirect URI` displayed on the `Create an Application` step and configure the Authelia
   `redirect_uris` option.
5. Configure the following options:
   - Identity Provider Name: `Authelia`
   - Client ID: `chatgpt`
   - Client Secret: `insecure_secret`
   - Discovery endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
6. Run the `Test Single Sign-On` step to verify the configuration.

## See Also

- [OpenAI Configuring SSO Documentation](https://help.openai.com/en/articles/9534785-configuring-sso)
- [OpenAI SSO in ChatGPT Business Documentation](https://help.openai.com/en/articles/10468051-sso-in-chatgpt-business)

[Authelia]: https://www.authelia.com
[ChatGPT]: https://chatgpt.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
