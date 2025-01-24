---
title: "DokuWiki"
description: "Integrating DokuWiki with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-21T22:32:51+11:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.18](https://github.com/authelia/authelia/releases/tag/v4.38.18)
* [DokuWiki]
  * [55.2](https://github.com/dokuwiki/dokuwiki/releases/tag/release-2024-02-06b)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://dokuwiki.{{< sitevar name="domain" nojs="example.com" >}}/`
  * This option determines the redirect URI in the format of
        `https://dokuwiki.{{< sitevar name="domain" nojs="example.com" >}}/login`.
        This means if you change this value, you need to update the redirect URI.
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `dokuwiki`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [DokuWiki] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'dokuwiki'
        client_name: 'DokuWiki'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://dokuwiki.{{< sitevar name="domain" nojs="example.com" >}}/doku.php'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
          - 'offline_access'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [DokuWiki] to utilize Authelia as an [OpenID Connect 1.0] Provider you must install the relevant plugins
and use the GUI to configure it.

Within [DokuWiki] visit the Administration section and then the Extension Manager. Install the following Extensions:

1. [plugin:oauth](https://www.dokuwiki.org/plugin:oauth)
2. [plugin:oauthgeneric](https://www.dokuwiki.org/plugin:oauthgeneric)

Within [DokuWiki] visit the Administration section and then the Configuration Settings.

1. In the Oauth section set the following values:
   - `plugin»oauth»register-on-auth`: enabled.
2. In the Oauthgeneric section:
   - `plugin»oauthgeneric»key`: `dokuwiki`.
   - `plugin»oauthgeneric»secret`: `insecure_secret`.
   - `plugin»oauthgeneric»authurl`: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`.
   - `plugin»oauthgeneric»tokenurl`: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`.
   - `plugin»oauthgeneric»userurl`: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`.
   - `plugin»oauthgeneric»authmethod`: `Bearer Header`.
   - `plugin»oauthgeneric»scopes`: `openid,email,profile,groups,offline_access`.
   - `plugin»oauthgeneric»needs-state`: enabled.
   - `plugin»oauthgeneric»json-user`: `preferred_username`.
   - `plugin»oauthgeneric»json-name`: `name`.
   - `plugin»oauthgeneric»json-mail`: `email`.
   - `plugin»oauthgeneric»json-grps`: `groups`.
   - `plugin»oauthgeneric»label`: `Authelia`.

[Authelia]: https://www.authelia.com
[DokuWiki]: https://www.dokuwiki.org/dokuwiki
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
