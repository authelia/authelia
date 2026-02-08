---
title: "Jenkins"
description: "Integrating Jenkins with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-13T13:46:05+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/jenkins/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Jenkins | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Jenkins with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.13](https://github.com/authelia/authelia/releases/tag/v4.39.13)
- [Jenkins]
  - [v2.516.3](https://www.jenkins.io/changelog/2.516)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://jenkins.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `jenkins`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

The following example uses the [OpenId Connect Authentication Plugin] which is assumed to be installed when following
this section of the guide.

To install the [OpenId Connect Authentication Plugin] for [Jenkins] via the Web GUI:

1. Visit `Manage Jenkins`.
2. Visit `Plugins`.
3. Visit `Available Plugins`.
4. Search for `oic-auth`.
5. Install.
6. Restart [Jenkins].

To install the [OpenId Connect Authentication Plugin] for [Jenkins] using the CLI:

```shell
jenkins-plugin-cli --plugins oic-auth
```

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Jenkins] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'jenkins'
        client_name: 'Jenkins'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://jenkins.{{< sitevar name="domain" nojs="example.com" >}}/accounts/authelia/login/callback'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Jenkins] there are two methods, using the [Configuration File](#configuration-file), or using the
[Web GUI](#web-gui).

#### Configuration File

To configure [Jenkins] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml
jenkins:
  systemMessage: "This Jenkins instance was configured using the Authelia example Configuration as Code, thanks Authelia!"
  securityRealm:
    oic:
      clientId: "jenkins"
      clientSecret: "insecure_secret"
      disableSslVerification: false
      emailFieldName: "email"
      fullNameFieldName: "name"
      groupIdStrategy: "caseSensitive"
      groupsFieldName: "groups"
      logoutFromOpenidProvider: false
      properties:
        - "pkce"
        - escapeHatch:
            group: "admin-users"
            secret: "escapeHatch"
            username: "escapeHatch"
      sendScopesInTokenRequest: true
      serverConfiguration:
        wellKnown:
          scopesOverride: "openid profile email groups"
          wellKnownOpenIDConfigurationUrl: "https://{{< sitevar name=\"subdomain-authelia\" nojs=\"auth\" >}}.{{< sitevar name=\"domain\" nojs=\"example.com\" >}}/.well-known/openid-configuration"
      userIdStrategy: "caseSensitive"
      userNameField: "preferred_username"
```

#### Web GUI

To configure [Jenkins] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit `Manage Jenkins`.
2. Visit `Security`.
3. Select `Login with Openid Connect` in the Security Realm.
4. Configure the following options:
   - Client id: `jenkins`
   - Client secret: `insecure_secret`
   - Configuration mode: `Discovery via well-known endpoint`
   - Well-known configuration endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
     - Under `Advanced`:
       - Override scopes: `openid profile email groups`
   - Under `Advanced configuration`:
     - Under `User fields`
     - User name field name: `preferred_username`
     - Full name field name: `name`
     - Email field name: `email`
     - Groups field name: `groups`
   - Add the following properties:
     - Enable Proof Key for Code Exchange: Enabled
     - Configure 'Escape Hatch' for when the OpenID Provider is unavailable: Consider using this setting

## See Also

- [Jenkins OpenID Connect Documentation](https://plugins.jenkins.io/oic-auth/)
- [Jenkins OpenID JCasC Documentation](https://github.com/jenkinsci/oic-auth-plugin/blob/master/docs/configuration/README.md)

[Jenkins]: https://www.jenkins.io/
[OpenId Connect Authentication Plugin]: https://plugins.jenkins.io/oic-auth/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
