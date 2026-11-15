---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Taiga"
description: "A guide on integrating Taiga with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Taiga | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Taiga with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Taiga]
  - [6.10.2](https://github.com/taigaio/taiga-back/releases/tag/6.10.2)
- [taiga-contrib-oidc-auth]
  - [0.3](https://github.com/taigaio/taiga-contrib-oidc-auth)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://taiga.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://taiga.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/callback/`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `taiga`
- **Client Secret:** `insecure_secret`
- **OIDC Mount Point:** `/api/oidc`
  - This option determines the redirect URI. Mounting the plugin inside the `/api` path means requests are routed to
    the Taiga backend by the standard reverse proxy configuration.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Taiga] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'taiga'
        client_name: 'Taiga'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://taiga.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/callback/'
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
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Taiga] there is one method, using the [Configuration File](#configuration-file) with the
[taiga-contrib-oidc-auth] plugin.

#### Configuration File

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[Taiga] does not support [OpenID Connect 1.0] natively. The [taiga-contrib-oidc-auth] plugin must be installed in
both the backend (`pip install taiga-contrib-oidc-auth`) and the frontend (build the plugin and copy the `dist`
directory to `dist/plugins/oidc-auth/`) as described in the plugin documentation.
{{< /callout >}}

To configure [Taiga] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Add the following to the backend `settings/local.py` file:

   ```python {title="settings/local.py"}
   INSTALLED_APPS += [
       "mozilla_django_oidc",
       "taiga_contrib_oidc_auth",
   ]

   AUTHENTICATION_BACKENDS = list(AUTHENTICATION_BACKENDS) + [
       "taiga_contrib_oidc_auth.oidc.TaigaOIDCAuthenticationBackend",
   ]

   ROOT_URLCONF = "settings.urls"

   OIDC_CALLBACK_CLASS = "taiga_contrib_oidc_auth.views.TaigaOIDCAuthenticationCallbackView"
   OIDC_RP_SCOPES = "openid profile email"
   OIDC_RP_SIGN_ALGO = "RS256"
   OIDC_BASE_URL = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"
   OIDC_OP_JWKS_ENDPOINT = OIDC_BASE_URL + "/jwks.json"
   OIDC_OP_AUTHORIZATION_ENDPOINT = OIDC_BASE_URL + "/api/oidc/authorization"
   OIDC_OP_TOKEN_ENDPOINT = OIDC_BASE_URL + "/api/oidc/token"
   OIDC_OP_USER_ENDPOINT = OIDC_BASE_URL + "/api/oidc/userinfo"
   OIDC_RP_CLIENT_ID = "taiga"
   OIDC_RP_CLIENT_SECRET = "insecure_secret"
   ```

2. Create the backend `settings/urls.py` file with the following content:

   ```python {title="settings/urls.py"}
   from taiga.urls import *
   urlpatterns += [
       re_path(r"^api/oidc/", include("mozilla_django_oidc.urls")),
   ]
   ```

3. Add the following options to the frontend `dist/conf.json` file:

   ```json {title="dist/conf.json"}
   {
     "contribPlugins": ["/plugins/oidc-auth/oidc-auth.json"],
     "oidcMountPoint": "/api/oidc",
     "oidcButtonText": "Authelia"
   }
   ```

## See Also

- [taiga-contrib-oidc-auth Documentation](https://github.com/taigaio/taiga-contrib-oidc-auth)
- [mozilla-django-oidc Settings Documentation](https://mozilla-django-oidc.readthedocs.io/en/stable/settings.html)

[Authelia]: https://www.authelia.com
[Taiga]: https://taiga.io
[taiga-contrib-oidc-auth]: https://github.com/taigaio/taiga-contrib-oidc-auth
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
