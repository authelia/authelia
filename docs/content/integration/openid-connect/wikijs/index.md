---
title: "Wiki.js"
description: "Integrating Wiki.js with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-18T15:25:09+10:00
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
  * [v4.38.8](https://github.com/authelia/authelia/releases/tag/v4.38.8)
* [Wiki.js]
  * [v2.5.301](https://github.com/requarks/wiki/releases/tag/v2.5.301)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://wiki.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `Wiki.js`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Wiki.js] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'wikijs'
        client_name: 'Wiki'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://wikijs.example.com/login/<UUID>/callback'  # Note this must be copied during step 7 of the Application configuration.
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Wiki.js] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Login to [Wiki.js] as an Administrator.
2. Visit Administration.
3. Select `Modules` > `Authentication`.
4. Select `+ Add Strategy`.
5. Select `Generic OpenID Connect / OAuth2`.
6. Enter the following values:
   1. Display Name: `Authelia`
   2. Client ID: `wikijs`
   3. Client Secret: `insecure_secret`
   4. Authorization Endpoint URL: `https://auth.example.com/api/oidc/authorization`
   5. Token Endpoint URL: `https://auth.example.com/api/oidc/token`
   6. User Info Endpoint URL: `https://auth.example.com/api/oidc/userinfo`
   7. Issuer URL: `https://auth.example.com`
   8. Email Claim: `email`
   9. Display Name Claim: `name`
   10. Map Groups: Disabled
   11. Groups Claim: `groups`
   12. Allow self-registration: Enabled
7. Copy the `Callback URL / Redirect URI` for the Authelia configuration.
8. Click Apply.

{{< figure src="wikijs.png" alt="Wiki.js" width="736" style="padding-right: 10px" >}}

## See Also

- [Wiki.js Authentication Guide](https://docs.requarks.io/auth)

[Wiki.js]: https://js.wiki/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
