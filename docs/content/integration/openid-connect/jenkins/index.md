---
title: "Jenkins"
description: "Integrating Jenkins with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-13T13:46:05+10:00
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
* [Jenkins]
  * [v2.453](https://www.jenkins.io/changelog/2.453/)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://jenkins.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `jenkins`
* __Client Secret:__ `insecure_secret`

## Configuration

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
          - 'https://jenkins.example.com/accounts/authelia/login/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

#### Installation

The plugin required to use [OpenID Connect 1.0] can either be installed and configured via the GUI or via [Jenkins]
Configuration as Code.

##### Via the UI

To install the [Jenkins] plugin for [OpenID Connect 1.0] via the UI:

1. Visit `Manage Jenkins`.

2. Visit `Plugins`.

3. Visit `Available Plugins`.

4. Search for `oic-auth`.

5. Install.

6. Restart [Jenkins].

7. Proceed to the [Configuration](#configuration-1) step.

##### Via Jenkins Configuration as Code

Ensure the plugin is installed before running the Jenkins Configuration as Code:

```bash
jenkins-plugin-cli --plugins oic-auth
```

Add this to your Jenkins Configuration as Code:

```yaml
jenkins:
  systemMessage: "This Jenkins instance was configured using the Authelia example Configuration as Code, thanks Authelia!"
  securityRealm:
    oic:
      automanualconfigure: auto
      wellKnownOpenIDConfigurationUrl: https://auth.example.com/.well-known/openid-configuration
      clientId: jenkins
      clientSecret: insecure_secret
      tokenAuthMethod: client_secret_basic
      scopes: openid profile email groups
      userNameField: preferred_username
      groupsFieldName: groups
      fullNameFieldName: name
      emailFieldName: email
      pkceEnabled: true
      # escapeHatchEnabled: <boolean>
      # escapeHatchUsername: escapeHatchUsername
      # escapeHatchSecret: <string:secret>
      # escapeHatchGroup: <string>
```

#### Configuration

To configure [Jenkins] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Visit `Manage Jenkins`.
2. Visit `Security`.
3. Select `Login with Openid Connect` in the Security Realm.
4. Enter `jenkins` in the `Client id` field.
5. Enter `insecure_secret` in the `Client secret` field.
6. Select `Automatic configuration` from the configuration mode.
7. Enter `https://auth.example.com/.well-known/openid-configuration` in the `Well-known configuration endpoint` field.
8. Select `Override scopes`.
9. Enter `openid profile email groups` in the `Scopes` field.
10. Expand `Advanced`.
11. Enter `preferred_username` into the `User name field name` field.
12. Enter `name` into the `Full name field name` field.
13. Enter `email` into the `Email field name` field.
14. Enter `groups` into the `Groups field name` field.
15. Select `Enable Proof Key for Code Exchange`.
16. Consider using the `Configure 'escape hatch' for when the OpenID Provider is unavailable` to prevent login issues.

## See Also

- [Jenkins OpenID Connect Documentation](https://plugins.jenkins.io/oic-auth/)
- [Jenkins OpenID JCasC Documentation](https://github.com/jenkinsci/oic-auth-plugin/blob/master/docs/configuration/README.md)

[Jenkins]: https://www.jenkins.io/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
