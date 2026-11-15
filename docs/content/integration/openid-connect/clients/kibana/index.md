---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Kibana"
description: "A guide on integrating Kibana with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
summary: ""
date: 2026-11-12T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Kibana | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Kibana with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Kibana]
  - [v9.5.6](https://github.com/elastic/kibana/releases/tag/v9.5.6)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://kibana.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://kibana.{{< sitevar name="domain" nojs="example.com" >}}/api/security/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `kibana`
- **Client Secret:** `insecure_secret`
- **Elasticsearch Realm Name:** `authelia`
- **Admin Group Name:** `kibana-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Kibana] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'kibana'
        client_name: 'Kibana'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://kibana.{{< sitevar name="domain" nojs="example.com" >}}/api/security/oidc/callback'
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

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The Elasticsearch [OpenID Connect 1.0] realm requires an appropriate Elastic subscription (or an active trial) and
TLS on the Elasticsearch HTTP layer.
{{< /callout >}}

To configure [Kibana] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [Kibana] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions.

##### Elasticsearch

Add the following realm to the Elasticsearch configuration on every node, ensuring the `order` value does not
conflict with your existing realms:

```yaml {title="config/elasticsearch.yml"}
xpack.security.authc.token.enabled: true
xpack.security.authc.realms.oidc.authelia:
  order: 2
  rp.client_id: 'kibana'
  rp.response_type: 'code'
  rp.client_auth_method: 'client_secret_basic'
  rp.redirect_uri: 'https://kibana.{{< sitevar name="domain" nojs="example.com" >}}/api/security/oidc/callback'
  rp.requested_scopes: ['openid', 'profile', 'email', 'groups']
  op.issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
  op.authorization_endpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
  op.token_endpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
  op.userinfo_endpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
  op.jwkset_path: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json'
  claims.principal: 'preferred_username'
  claims.groups: 'groups'
  claims.mail: 'email'
  claims.name: 'name'
```

Add the client secret to the Elasticsearch keystore on every node, then restart Elasticsearch:

```shell
bin/elasticsearch-keystore add xpack.security.authc.realms.oidc.authelia.rp.client_secret
```

##### Role Mapping

Create a role mapping which grants Elasticsearch roles to members of the `kibana-admins` group, for example using
Dev Tools in [Kibana] as a user with the `manage_security` privilege:

```http
PUT /_security/role_mapping/authelia-kibana-admins
{
  "roles": ["superuser"],
  "enabled": true,
  "rules": {
    "all": [
      { "field": { "realm.name": "authelia" } },
      { "field": { "groups": "kibana-admins" } }
    ]
  }
}
```

##### Kibana

Add the following provider to the [Kibana] configuration and restart [Kibana]:

```yaml {title="config/kibana.yml"}
server.publicBaseUrl: 'https://kibana.{{< sitevar name="domain" nojs="example.com" >}}'
xpack.security.authc.providers:
  oidc.authelia:
    order: 0
    realm: 'authelia'
    description: 'Log in with Authelia'
  basic.basic1:
    order: 1
```

Users who successfully authenticate but do not match any role mapping are not granted any roles and are refused
by [Kibana].

## See Also

- [Elastic OpenID Connect Authentication Documentation](https://www.elastic.co/docs/deploy-manage/users-roles/cluster-or-deployment-auth/openid-connect)
- [Elastic Mapping Users and Groups to Roles Documentation](https://www.elastic.co/docs/deploy-manage/users-roles/cluster-or-deployment-auth/mapping-users-groups-to-roles)

[Authelia]: https://www.authelia.com
[Kibana]: https://www.elastic.co/kibana
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
