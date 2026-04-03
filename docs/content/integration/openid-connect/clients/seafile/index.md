---
title: "Seafile"
description: "Integrating Seafile with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/seafile/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Seafile | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Seafile with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Seafile] Server
  - [v10.0.1](https://manual.seafile.com/latest/changelog/server-changelog/#1001-2023-04-11)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://seafile.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `seafile`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

1. [Seafile] may require some dependencies such as `requests_oauthlib` to be manually installed. See the [Seafile]
   documentation in the [see also](#see-also) section for more information.

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Seafile] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'seafile'
        client_name: 'Seafile'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://seafile.{{< sitevar name="domain" nojs="example.com" >}}/oauth/callback/'
        scopes:
          - 'openid'
          - 'profile'
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

To configure [Seafile] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The [Seafile's WebDAV extension](https://manual.seafile.com/extension/webdav/)
does not [support OAuth Bearer](https://github.com/haiwen/seafdav/issues/76) at the time of this writing.
{{< /callout >}}

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `seahub_settings.py`.
{{< /callout >}}

To configure [Seafile] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```python {title="seahub_settings.py"}
ENABLE_OAUTH = True
OAUTH_ENABLE_INSECURE_TRANSPORT = False
OAUTH_CLIENT_ID = "seafile"
OAUTH_CLIENT_SECRET = "insecure_secret"
OAUTH_REDIRECT_URL = 'https://seafile.{{< sitevar name="domain" nojs="example.com" >}}/oauth/callback/'
OAUTH_PROVIDER_DOMAIN = '{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
OAUTH_AUTHORIZATION_URL = 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
OAUTH_TOKEN_URL = 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
OAUTH_USER_INFO_URL = 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
OAUTH_SCOPE = [
    "openid",
    "profile",
    "email",
]
OAUTH_ATTRIBUTE_MAP = {
    "sub": (True, "uid"),
    "email": (False, "email"),
    "name": (False, "name"),
}

# Optional
#ENABLE_WEBDAV_SECRET = True
```

### Existing Users

When using [Seafile] with external authentication you may have to perform manual steps to achieve this. 

The [See Also](#see-also) has a link to the [Seafile] `migrating from local user database to external authentication` guide which has been verified to work.  



## Additional Steps

Optionally [enable webdav secrets](https://manual.seafile.com/latest/config/seahub_settings_py/#user-management-options) so
that clients that do not support OAuth 2.0 (e.g., [davfs2](https://savannah.nongnu.org/bugs/?57589)) can login via
basic auth.

## See Also

- [Seafile OAuth Authentication Documentation](https://manual.seafile.com/latest/config/oauth/)
- [Seafile's WebDAV extension](https://manual.seafile.com/latest/extension/webdav/)
- [Migrate from local user database to OAuth](https://manual.seafile.com/11.0/deploy/auth_switch/#migrating-from-local-user-database-to-external-authentication)

[Authelia]: https://www.authelia.com
[Seafile]: https://www.seafile.com/
[Seafile's WebDAV extension]: https://manual.seafile.com/extension/webdav/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
