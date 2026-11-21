---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "JupyterHub"
description: "A guide on integrating JupyterHub with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "JupyterHub | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring JupyterHub with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [JupyterHub]
  - [v6.2.0](https://github.com/jupyterhub/jupyterhub/releases/tag/6.2.0)
- [OAuthenticator]
  - [v17.5.0](https://github.com/jupyterhub/oauthenticator/releases/tag/17.5.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://jupyterhub.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://jupyterhub.{{< sitevar name="domain" nojs="example.com" >}}/hub/oauth_callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `jupyterhub`
- **Client Secret:** `insecure_secret`
- **Allowed Group:** `jupyterhub-users`
- **Admin Group:** `jupyterhub-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [JupyterHub] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'jupyterhub'
        client_name: 'JupyterHub'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://jupyterhub.{{< sitevar name="domain" nojs="example.com" >}}/hub/oauth_callback'
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

To configure [JupyterHub] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [JupyterHub] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration
which uses the `OIDCOAuthenticator` from [OAuthenticator]:

```python {title="jupyterhub_config.py"}
c.JupyterHub.authenticator_class = "oidc"

c.OIDCOAuthenticator.openid_provider_url = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"
c.OIDCOAuthenticator.oauth_callback_url = "https://jupyterhub.{{< sitevar name="domain" nojs="example.com" >}}/hub/oauth_callback"
c.OIDCOAuthenticator.client_id = "jupyterhub"
c.OIDCOAuthenticator.client_secret = "insecure_secret"
c.OIDCOAuthenticator.login_service = "Authelia"
c.OIDCOAuthenticator.scope = ["openid", "profile", "email", "groups"]
c.OIDCOAuthenticator.username_claim = "preferred_username"
c.OIDCOAuthenticator.enable_pkce = True

c.OIDCOAuthenticator.manage_groups = True
c.OIDCOAuthenticator.auth_state_groups_key = "oauth_user.groups"
c.OIDCOAuthenticator.allowed_groups = {"jupyterhub-users"}
c.OIDCOAuthenticator.admin_groups = {"jupyterhub-admins"}
```

## See Also

- [OAuthenticator OpenID Connect (OIDC) Setup Documentation](https://oauthenticator.readthedocs.io/en/latest/tutorials/provider-specific-setup/providers/oidc.html)
- [OAuthenticator GenericOAuthenticator Setup Documentation](https://oauthenticator.readthedocs.io/en/latest/tutorials/provider-specific-setup/providers/generic.html)

[Authelia]: https://www.authelia.com
[JupyterHub]: https://jupyter.org/hub
[OAuthenticator]: https://github.com/jupyterhub/oauthenticator
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
