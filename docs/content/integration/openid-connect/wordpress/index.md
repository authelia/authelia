---
title: "WordPress"
description: "Integrating WordPress with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 620
toc: true
community: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [Wordpress]
  * [v6.5.2](https://core.svn.wordpress.org/tags/6.5.2/)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://wordpress.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `wordpress`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [WordPress] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'wordpress'
        client_name: 'WordPress'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://wordpress.example.com/wp-admin/admin-ajax.php?action=openid-connect-authorize'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

1. Install the Plugin:
   1. Visit `Plugins`.
   2. Visit `Add New`.
   3. Install `OpenID Connect Generic Client` by `daggerhart`.
2. Configure the Plugin:
   1. Visit `Settings`.
   2. Visit `OpenID Connect Client`.
   3. Select the `OpenID Connect button on login form` option from `Login Type`.
   4. Enter `wordpress` in the `Client ID` field.
   5. Enter `insecure_secret` in the `Client Secret` field.
   6. Enter `openid profile email` in the `OpenID Scope` field.
   7. Enter `https://auth.example.com/api/oidc/authorization` in the `Login Endpoint URL` field.
   8. Enter `https://auth.example.com/api/oidc/token` in the `Token Validation Endpoint URL` field.
   9. Enter `https://auth.example.com/api/oidc/userinfo` in the `Userinfo Endpoint URL` field.

## See Also

- [WordPress OpenID Connect Generic Client Documentation](https://wordpress.org/plugins/daggerhart-openid-connect-generic/)

[WordPress]: https://en-au.wordpress.org/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
