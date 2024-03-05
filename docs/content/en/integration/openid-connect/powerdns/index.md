---
title: "PowerDNS Admin"
description: "Integrating PowerDNS with the Authelia OpenID Connect 1.0 Provider."
lead: ""
date: 2024-01-16T08:47:18+11:00
draft: false
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 620
toc: true
community: true
---

## Tested Versions

* [Authelia]
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [PowerDNS Admin]
  * [v0.4.1](https://github.com/PowerDNS-Admin/PowerDNS-Admin/releases/tag/v0.4.1)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://powerdns.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `powerdns`
* __Client Secret:__ `insecure_secret`

## Configuration

### Application

To configure [PowerDNS Admin] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Visit Settings
2. Visit Authentication
3. Visit OpenID Connect OAuth
3. Set the following values:
   1. Enable *Enable OpenID Connect OAuth*
   2. Client ID: `powerdns`
   3. Client Secret: `insecure_secret`
   4. Scopes: `openid profile groups email`
   5. API URL: `https://auth.example.com/api/oidc/userinfo`
   6. Enable *Enable OIDC OAuth Auto-Configurationh*
   6. Metadata URL: `https://auth.example.com/.well-known/openid-configuration`
   7. Username: `preferred_username`
   8. Email: `email`
   9. Firstname: `preferred_username`
   10. Last Name: `name`
   11. Autoprovision Account Name property: `preferred_username`
   12. Autoprovision Account Description property : `name`

*__Note:__ Currently, Authelia only supports the preferred_username and name claims under the profile scope. However PowerDNS-Admin only supports a FirstName LastName system, where the two are separate, instead of using the name claim to fetch the full name. This means that the names in the system are incorrect. (See linked ticket(https://github.com/authelia/authelia/issues/4338))

{{< figure src="powerdns.png" alt="PowerDNS Admin" width="736" style="padding-right: 10px" >}}

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/openid-connect/clients.md) for use with [PowerDNS Admin]
which will operate with the above example:

```yaml
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - id: 'powerdns'
        description: 'PowerDNS'
        secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://powerdns.example.com/oidc/authorized'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        userinfo_signed_response_alg: 'none'
```

## See Also

* [Portainer OAuth Documentation](https://docs.portainer.io/admin/settings/authentication/oauth)

[Authelia]: https://www.authelia.com
[Portainer]: https://www.portainer.io/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
