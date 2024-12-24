---
title: "WordPress"
description: "Integrating WordPress with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 720
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

- [Authelia]
  - [v4.38.18](https://github.com/authelia/authelia/releases/tag/v4.38.18)
- [Wordpress]
  - [v6.7.1](https://core.svn.wordpress.org/tags/6.7.1/)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://wordpress.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `wordpress`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

The following example uses the [OpenID Connect Generic Client Plugin] which is assumed to be installed when following
this section of the guide.

To install the [OpenID Connect Generic Client Plugin] for [WordPress] via the Web GUI:

1. Visit `Plugins`.
2. Visit `Add New`.
3. Install `OpenID Connect Generic Client` by `daggerhart`.

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
        redirect_uris:
          - 'https://wordpress.{{< sitevar name="domain" nojs="example.com" >}}/wp-admin/admin-ajax.php?action=openid-connect-authorize'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'offline_access'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [WordPress] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [WordPress] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit `Settings`.
2. Visit `OpenID Connect Client`.
3. Select the `OpenID Connect button on login form` option from `Login Type`.
4. Enter the following values into the corresponding fields:
   1. Client ID: `wordpress`
   2. Client Secret Key: `insecure_secret`
   3. OpenID Scope: `openid profile email offline_access`
   4. Login Endpoint URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   5. Userinfo Endpoint URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   6. Token Validation Endpoint URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   7. Identity Key: `sub`
   8. Disable SSL Verify: Not Checked
   9. Nickname Key: `preferred_username`
   10. Email Formatting: `{email}`
   11. Display Name Formatting: `{name}`
   12. Identify with User Name: Not Checked
   13. Enable Refresh Token: Checked
   14. Link Existing Users: Checked if you want to automatically link existing users, or Unchecked if you want to
       create new ones.

## See Also

- [WordPress OpenID Connect Generic Client Documentation](https://wordpress.org/plugins/daggerhart-openid-connect-generic/)

[WordPress]: https://en-au.wordpress.org/
[OpenID Connect Generic Client Plugin]: https://wordpress.com/plugins/daggerhart-openid-connect-generic
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
