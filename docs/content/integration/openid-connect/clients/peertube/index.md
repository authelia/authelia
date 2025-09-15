---
title: "PeerTube"
description: "Integrating PeerTube with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-24T23:57:05+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/peertube/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "PeerTube | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring PeerTube with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.4](https://github.com/authelia/authelia/releases/tag/v4.39.4)
- [PeerTube]
  - [v7.2.1](https://github.com/Chocobozzz/PeerTube/releases/tag/v7.2.1)
- [OpenID Connect Plugin]
  - v1.0.2

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://peertube.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `peertube`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

The following example uses the [OpenID Connect Plugin] which is assumed to be installed when following this
section of the guide.

To install the [OpenID Connect Plugin] for [PeerTube] via the Web GUI:

1. Visit `Settings` under `Administration`.
2. Visit `Plugins/Themes`.
3. Visit `Search plugins`.
4. Install the official `auth-openid-connect` plugin.

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [PeerTube] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'peertube'
        client_name: 'PeerTube'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://peertube.{{< sitevar name="domain" nojs="example.com" >}}/plugins/auth-openid-connect/router/code-cb'
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

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The following example assumes the `peertube-users` group was setup for users who should be able to access this app. The
configuration of a group is not optional, but it can be any group of users you wish.
{{< /callout >}}

To configure [PeerTube] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [PeerTube] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit `Settings` under `Administration`.
2. Visit `Plugins/Themes`.
3. Visit `Installed plugins`.
4. Click the `Settings` button of the installed [OpenID Connect Plugin].
5. Configure the following options:
   - Auth display name: `Authelia`
   - Discover URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Client ID: `peertube`
   - Client secret: `insecure_secret`
   - Scope: `openid email profile groups`
   - Username property: `preferred_username`
   - Email property: `email`
   - Display name property: `name`
   - Group property: `groups`
   - Allowed group: `peertube-users`
   - Token signature algorithm: `RS256`
6. Save.

{{< figure src="peertube.png" alt="Peertube" width="736" style="padding-right: 10px" >}}

## See Also

- [PeerTube Auth OpenID Connect Documentation](https://framagit.org/framasoft/peertube/official-plugins/tree/master/peertube-plugin-auth-openid-connect)

[PeerTube]: https://joinpeertube.org
[OpenID Connect Plugin]: https://framagit.org/framasoft/peertube/official-plugins/-/tree/master/peertube-plugin-auth-openid-connect
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
