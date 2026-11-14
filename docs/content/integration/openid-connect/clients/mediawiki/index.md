---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "MediaWiki"
description: "A guide on integrating MediaWiki with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "MediaWiki | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring MediaWiki with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [MediaWiki]
  - [v1.46.2](https://www.mediawiki.org/wiki/Release_notes)
  - [OpenID Connect Extension v8.3.0](https://www.mediawiki.org/wiki/Extension:OpenID_Connect)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://mediawiki.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://mediawiki.{{< sitevar name="domain" nojs="example.com" >}}/index.php/Special:PluggableAuthLogin`.
    This means if you change this value, you need to update the redirect URI.
  - The redirect URI is the full URL of the `Special:PluggableAuthLogin` page, which depends on the
    [MediaWiki] `$wgArticlePath` setting. For example with short URLs it may be
    `https://mediawiki.{{< sitevar name="domain" nojs="example.com" >}}/wiki/Special:PluggableAuthLogin`.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `mediawiki`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [MediaWiki] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'mediawiki'
        client_name: 'MediaWiki'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://mediawiki.{{< sitevar name="domain" nojs="example.com" >}}/index.php/Special:PluggableAuthLogin'
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

To configure [MediaWiki] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [MediaWiki] to utilize Authelia as an [OpenID Connect 1.0] Provider, install the [PluggableAuth] and
[OpenID Connect][OpenID Connect Extension] extensions and add the following to the `LocalSettings.php` file:

```php {title="LocalSettings.php"}
wfLoadExtension( 'PluggableAuth' );
wfLoadExtension( 'OpenIDConnect' );

$wgPluggableAuth_Config['Authelia'] = [
	'plugin' => 'OpenIDConnect',
	'data' => [
		'providerURL' => 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}',
		'clientID' => 'mediawiki',
		'clientsecret' => 'insecure_secret',
		'scope' => [ 'openid', 'profile', 'email' ],
		'codeChallengeMethod' => 'S256',
		'authMethods' => [ 'client_secret_basic' ],
	],
];
```

After installing the extensions run the `maintenance/run.php update` script to create the required database tables.

## See Also

- [MediaWiki OpenID Connect Extension Documentation](https://www.mediawiki.org/wiki/Extension:OpenID_Connect)
- [MediaWiki PluggableAuth Extension Documentation](https://www.mediawiki.org/wiki/Extension:PluggableAuth)

[Authelia]: https://www.authelia.com
[MediaWiki]: https://www.mediawiki.org/
[PluggableAuth]: https://www.mediawiki.org/wiki/Extension:PluggableAuth
[OpenID Connect Extension]: https://www.mediawiki.org/wiki/Extension:OpenID_Connect
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
