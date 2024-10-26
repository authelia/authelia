---
title: "Rocket.Chat"
description: "Integrating Rocket.Chat with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-09-28T23:18:03+10:00
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
  * [v4.38.12](https://github.com/authelia/authelia/releases/tag/v4.38.12)
* [Rocket.Chat]
  * [v6.11.1](https://github.com/RocketChat/Rocket.Chat/releases/tag/6.11.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://rocket-chat.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `rocket-chat`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Rocket.Chat] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'rocket-chat'
        client_name: 'Rocket.Chat'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://rocket-chat.{{< sitevar name="domain" nojs="example.com" >}}/_oauth/authelia'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        grant_types:
          - 'authorization_code'
        token_endpoint_auth_method: 'client_secret_post'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [Rocket.Chat] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Visit the [Rocket.Chat] `Administration` page.
2. Click `OAuth`.
3. Click `Add`.
4. Enter `authelia` as the unique name.
5. Click `Enable`.
6. Enter the following values:
   - URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Token Path: `/api/oidc/token`
   - Token sent via: `Payload`
   - Identity Token Sent Via: `Same as "Token Sent Via"`
   - Identity Path: `/api/oidc/userinfo`
   - Authorize Path: `/api/oidc/authorization`
   - Scope: `openid profile email groups`
   - Param Name for Access Token: `access_token`
   - Id: `rocket-chat`
   - Secret: `insecure_secret`
   - Login Style: `Redirect`
   - Button Text: `Login with Authelia`
   - Key Field: `Username`
   - Username field: `preferred_username`
   - Email field: `email`
   - Name field: `name`
   - Roles/Groups field name: `groups`
   - Roles/Groups field for channel mapping: `groups`
   - Merge users: On
   - Show Button on Login Page: On

#### Group Mapping

[Rocket.Chat] has a means of mapping identity provider groups or roles to internal roles. For this option to take effect
you must enable the `Map Roles/Groups to channels` option and fill in the `OAuth Group Channel Map` field with a JSON
object. The key for this object is the Authelia group name, and the value is a JSON array of [Rocket.Chat] room names.

The following example shows matching the Authelia group `admins` to the groups `administration` and `moderators`, and
the `users` group to the `community` room.

```json
{
  "admins": ["administration", "moderators"],
  "users": ["community"]
}

```

## See Also

* [Rocket.Chat OpenID Connect Documentation](https://docs.rocket.chat/docs/openid-connect)

[Authelia]: https://www.authelia.com
[Rocket.Chat]: https://www.rocket.chat
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
