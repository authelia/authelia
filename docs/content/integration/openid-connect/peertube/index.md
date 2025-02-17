---
title: "PeerTube"
description: "Integrating PeerTube with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-01-12T15:26:39+01:00
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
* [PeerTube]
  * [v7.0.1](https://github.com/Chocobozzz/PeerTube/releases/tag/v7.0.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://peertube.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `peertube`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

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
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Users will not be able to log in using this provider if you do not set the `Allowed group` parameter.
{{< /callout >}}

1. Install the Plugin:
   1. Visit `Settings` under `Administration`.
   2. Visit `Plugins/Themes`.
   3. Visit `Search plugins`.
   4. Install the official `auth-openid-connect` plugin.
2. Configure the Plugin:
   1. Visit `Settings` under `Administration`.
   2. Visit `Plugins/Themes`.
   3. Visit `Installed plugins`.
   4. Click the `Settings` button of the installed plugin.
   5. Enter the following configuration:
    - Auth display name: `Authelia`
    - Discover URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
    - Client ID: `peertube`
    - Client secret: `insecure_secret`
    - Scope: `openid email profile groups`
    - Username property: `preferred_username`
    - Email property: `email`
    - Display name property: `name`
    - Group property: `groups`
    - Allowed group: Authelia's group allowed to log in using this provider.
    - Token signature algorithm: `RS256`

{{< figure src="peertube.png" alt="Peertube" width="736" style="padding-right: 10px" >}}

## See Also

- [PeerTube Auth OpenID Connect Documentation](https://framagit.org/framasoft/peertube/official-plugins/tree/master/peertube-plugin-auth-openid-connect)

[PeerTube]: https://joinpeertube.org
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
