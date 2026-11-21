---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Personio"
description: "A guide on integrating Personio with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Personio | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Personio with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Personio]
  - Not Applicable

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `personio`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Personio] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'personio'
        client_name: 'Personio'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://login.personio.com/login/callback'  # See step 4 below.
        scopes:
          - 'openid'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Personio] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

[Personio] does not create employees using single sign-on. Each employee must already have an invited and activated
[Personio] profile, and the email address returned by Authelia must match the email address of the corresponding
[Personio] employee.

To configure [Personio] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Personio] as an administrator.
2. Navigate to `Settings` > `Security & integrations` > `Security & authentication`.
3. From the list of login methods navigate to `Open ID Connect (OIDC)` and click `Configure`.
4. Copy the `Callback URLs/Redirect URIs` value displayed under `Provider settings` and ensure it is included in the
   redirect URIs configured in Authelia.
5. Configure the following options:
   - Button Display Text: `Continue with Authelia`
   - Issuer: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Authorization URI: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URI: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Userinfo URI: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - JSON Web Key Set URI: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json`
   - Scopes: `openid,email`
   - Client ID: `personio`
   - Client Secret: `insecure_secret`
   - Secret expiry date: The date you plan to rotate the client secret
   - Claim Field: `email`
6. Submit the changes and enable the connection.

The `Perform a configuration test` option on the [Personio] OIDC configuration screen can be used to validate the
connection without logging out.

## See Also

- [Personio Set up your company's authentication method Documentation](https://support.personio.de/hc/en-us/articles/360000019129-Set-up-your-company-s-authentication-method)
- [Personio Set up single sign-on (SSO) in Personio with an external identity provider Documentation](https://support.personio.de/hc/en-us/articles/36170323038621-Set-up-single-sign-on-SSO-in-Personio-with-an-external-identity-provider)
- [Personio Troubleshoot SSO login issues Documentation](https://support.personio.de/hc/en-us/articles/21947067413149-Troubleshoot-SSO-login-issues)

[Authelia]: https://www.authelia.com
[Personio]: https://www.personio.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
