---
title: "Node-RED"
description: "Integrating Node-RED with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-08-12T14:36:35+10:00
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
  * [v4.38.10](https://github.com/authelia/authelia/releases/tag/v4.38.10)
* [Node-RED]
  * [v4.0.2](https://github.com/node-red/node-red/releases/tag/4.0.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://node-red.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `node-red`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Node-RED] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'node-red'
        client_name: 'Node-RED'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://node-red.{{< sitevar name="domain" nojs="example.com" >}}/auth/strategy/callback/'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Node-RED] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Install the `passport-openidconnect` npm package.
2. Use the following `settings.js` configuration:

```js
adminAuth: {
    type: 'strategy',
    strategy: {
        name: 'openidconnect',
        label: 'Sign in with Authelia',
        icon: 'fa-openid',
        strategy: require('passport-openidconnect').Strategy,
        options: {
            issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}',
            authorizationURL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization',
            tokenURL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token',
            userInfoURL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo',
            clientID: 'node-red',
            clientSecret: 'insecure_secret',
            callbackURL: 'https://node-red.{{< sitevar name="domain" nojs="example.com" >}}/auth/strategy/callback/',
            scope: ['openid', 'email', 'profile', 'groups'],
            proxy: true,
            verify: function(issuer, profile, done) {
                done(null, profile)
            }
        }
    },
    users: function(user) {
        return Promise.resolve({ username: user, permissions: "*" });
    }
},
```

## See Also

- [Node-RED OAuth/OpenID based authentication Documentation](https://nodered.org/docs/user-guide/runtime/securing-node-red#oauthopenid-based-authentication)

[Node-RED]: https://nodered.org/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
