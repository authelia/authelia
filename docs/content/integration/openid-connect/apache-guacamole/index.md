---
title: "Apache Guacamole"
description: "Integrating Apache Guacamole with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-07-31T13:09:05+10:00
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
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [Apache Guacamole]
  * __UNKNOWN__

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://guacamole.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `guacamole`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Apache Guacamole] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'guacamole'
        client_name: 'Apache Guacamole'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://guacamole.{{< sitevar name="domain" nojs="example.com" >}}'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        response_types:
          - 'id_token'
        grant_types:
          - 'implicit'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [Apache Guacamole] to utilize Authelia as an [OpenID Connect 1.0] Provider use the following configuration:

```yaml
openid-client-id: guacamole
openid-scope: openid profile groups email
openid-issuer: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
openid-jwks-endpoint: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json
openid-authorization-endpoint: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization?state=1234abcedfdhf
openid-redirect-uri: https://guacamole.{{< sitevar name="domain" nojs="example.com" >}}
openid-username-claim-type: preferred_username
openid-groups-claim-type: groups
```

## See Also

* [Apache Guacamole OpenID Connect Authentication Documentation](https://guacamole.apache.org/doc/gug/openid-auth.html)

[Authelia]: https://www.authelia.com
[Apache Guacamole]: https://guacamole.apache.org/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
