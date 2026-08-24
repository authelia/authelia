---
title: "openCloud"
description: "Integrating openCloud with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-07-17T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases: []
support:
  level: community
  versions: true
  integration: true
seo:
  title: "openCloud | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring openCloud with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.21](https://github.com/authelia/authelia/releases/tag/v4.39.21)
- [openCloud]
  - [v7.2.2](https://github.com/opencloud-eu/opencloud/releases/tag/v7.2.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://opencloud.{{< sitevar name="domain" nojs="example.com" >}}`
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
- **Client ID:**
  - Web Application: `opencloud`
  - Android App: `opencloud-android`
  - iOS App: `opencloud-ios`
  - Desktop client: `opencloud-desktop`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with
[openCloud] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    clients:
      - client_id: 'opencloud'
        client_name: 'openCloud'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'groups'
          - 'profile'
          - 'email'
        redirect_uris:
          - 'https://opencloud.{{< sitevar name="domain" nojs="example.com" >}}/'
          - 'https://opencloud.{{< sitevar name="domain" nojs="example.com" >}}/oidc-callback.html'
          - 'https://opencloud.{{< sitevar name="domain" nojs="example.com" >}}/oidc-silent-redirect.html'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
      - client_id: 'opencloud-android'
        client_name: 'openCloud'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'oc://android.opencloud.eu'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'groups'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
      - client_id: 'opencloud-desktop'
        client_name: 'openCloud'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'groups'
          - 'profile'
          - 'email'
        redirect_uris:
          - 'http://127.0.0.1'
          - 'http://localhost'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
      - client_id: 'opencloud-ios'
        client_name: 'openCloud'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'oc://ios.opencloud.eu'
          - 'oc.ios://ios.opencloud.eu'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'groups'
          - 'profile'
          - 'email'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
```

### Application

To configure [openCloud] there is one method, using the [Environment Variables](#environment-variables).

#### Limitations

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
All limitations are limitations due to the development lifecycle of the application and are not related to Authelia.
{{< /callout >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The desktop client integration is currently not production ready.

The current implementation is intended for one-time use cases, such as migrating files. It should not be considered a
stable long-term desktop client setup until the WebFinger integration is completed.
{{< /callout >}}

- The desktop client WebFinger integration is currently incomplete (pull request https://github.com/opencloud-eu/desktop/pull/847).
- Some additional customizations may be needed as noted below.

The `groups` scope must be manually added to the authorization link when setting up the desktop client.

Example:

```
https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization?response_type=code&client_id=<client_id>&redirect_uri=<redirect_uri>&code_challenge=<code_challenge>&code_challenge_method=S256&scope=<scope>&prompt=<prompt>&state=<state>
```

The default scope is:

```
scope=openid%20offline_access%20email%20profile
```

Add `groups%20` manually at the beginning of the scope:

```
scope=groups%20openid%20offline_access%20email%20profile
```

#### Environment Variables

To configure [openCloud] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
OC_OIDC_ISSUER="https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"
PROXY_OIDC_ACCESS_TOKEN_VERIFY_METHOD="jwt"
PROXY_OIDC_REWRITE_WELLKNOWN="true"
PROXY_AUTOPROVISION_ACCOUNTS="true"
PROXY_AUTOPROVISION_CLAIM_USERNAME="preferred_username"
PROXY_AUTOPROVISION_CLAIM_EMAIL="email"
PROXY_AUTOPROVISION_CLAIM_DISPLAYNAME="name"
PROXY_AUTOPROVISION_CLAIM_GROUPS="groups"
PROXY_CSP_CONFIG_FILE_LOCATION="/etc/opencloud/csp.yaml"
WEBFINGER_WEB_OIDC_CLIENT_ID="opencloud"
WEBFINGER_WEB_OIDC_CLIENT_SCOPES="openid profile email groups offline_access"
WEBFINGER_ANDROID_OIDC_CLIENT_ID="opencloud-android"
WEBFINGER_ANDROID_OIDC_CLIENT_SCOPES="openid profile email groups offline_access"
WEBFINGER_IOS_OIDC_CLIENT_ID="opencloud-ios"
WEBFINGER_IOS_OIDC_CLIENT_SCOPES="openid profile email groups offline_access"
WEBFINGER_DESKTOP_OIDC_CLIENT_ID="opencloud-desktop"
WEBFINGER_DESKTOP_OIDC_CLIENT_SCOPES="openid profile email groups offline_access"
OC_EXCLUDE_RUN_SERVICES="idp"
PROXY_ROLE_ASSIGNMENT_DRIVER="oidc"
GRAPH_ASSIGN_DEFAULT_USER_ROLE="false"
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  openCloud:
    environment:
      OC_OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      PROXY_OIDC_ACCESS_TOKEN_VERIFY_METHOD: 'jwt'
      PROXY_OIDC_REWRITE_WELLKNOWN: 'true'
      PROXY_AUTOPROVISION_ACCOUNTS: 'true'
      PROXY_AUTOPROVISION_CLAIM_USERNAME: 'preferred_username'
      PROXY_AUTOPROVISION_CLAIM_EMAIL: 'email'
      PROXY_AUTOPROVISION_CLAIM_DISPLAYNAME: 'name'
      PROXY_AUTOPROVISION_CLAIM_GROUPS: 'groups'
      PROXY_CSP_CONFIG_FILE_LOCATION: '/etc/opencloud/csp.yaml'
      WEBFINGER_WEB_OIDC_CLIENT_ID: 'opencloud'
      WEBFINGER_WEB_OIDC_CLIENT_SCOPES: 'openid profile email groups offline_access'
      WEBFINGER_ANDROID_OIDC_CLIENT_ID: 'opencloud-android'
      WEBFINGER_ANDROID_OIDC_CLIENT_SCOPES: 'openid profile email groups offline_access'
      WEBFINGER_IOS_OIDC_CLIENT_ID: 'opencloud-ios'
      WEBFINGER_IOS_OIDC_CLIENT_SCOPES: 'openid profile email groups offline_access'
      WEBFINGER_DESKTOP_OIDC_CLIENT_ID: 'opencloud-desktop'
      WEBFINGER_DESKTOP_OIDC_CLIENT_SCOPES: 'openid profile email groups offline_access'
      OC_EXCLUDE_RUN_SERVICES: 'idp'
      PROXY_ROLE_ASSIGNMENT_DRIVER: 'oidc'
      GRAPH_ASSIGN_DEFAULT_USER_ROLE: 'false'
```

### Files

The following files must be configured.

#### Content Security Policy

Create the CSP configuration file and save it next to `opencloud.yaml`:

```yaml {title="csp.yaml"}
directives:
  child-src:
    - '''self'''
  connect-src:
    - '''self'''
    - 'blob:'
    - 'https://${COMPANION_DOMAIN|companion.opencloud.test}${TRAEFIK_PORT_HTTPS}/'
    - 'wss://${COMPANION_DOMAIN|companion.opencloud.test}${TRAEFIK_PORT_HTTPS}/'
    - 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
    - 'https://raw.githubusercontent.com/opencloud-eu/awesome-apps/'
    - 'https://update.opencloud.eu/'
    - 'https://tile.openstreetmap.org/'
  default-src:
    - '''none'''
  font-src:
    - '''self'''
  frame-ancestors:
    - '''self'''
  frame-src:
    - '''self'''
    - 'blob:'
    - 'https://embed.diagrams.net/'
    - 'https://${COLLABORA_DOMAIN|collabora.opencloud.test}${TRAEFIK_PORT_HTTPS}/'
    - 'https://${EURO_OFFICE_DOMAIN|euro-office.opencloud.test}${TRAEFIK_PORT_HTTPS}/'
    - 'https://docs.opencloud.eu'
    - 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
  img-src:
    - '''self'''
    - 'data:'
    - 'blob:'
    - 'https://raw.githubusercontent.com/opencloud-eu/awesome-apps/'
    - 'https://tile.openstreetmap.org/'
    - 'https://${COLLABORA_DOMAIN|collabora.opencloud.test}${TRAEFIK_PORT_HTTPS}/'
    - 'https://${EURO_OFFICE_DOMAIN|euro-office.opencloud.test}${TRAEFIK_PORT_HTTPS}/'
  manifest-src:
    - '''self'''
  media-src:
    - '''self'''
  object-src:
    - '''self'''
    - 'blob:'
  script-src:
    - '''self'''
    - '''unsafe-inline'''
    - 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
  style-src:
    - '''self'''
    - '''unsafe-inline'''
  worker-src:
    - '''self'''
    - 'blob:'
```

Refer to [csp.yaml](https://github.com/opencloud-eu/opencloud-compose/blob/main/config/opencloud/csp.yaml)

#### Proxy

When using an external IDP, you need to map groups and roles. Create the following file and save it next to `opencloud.yaml`:

In this example the `opencloud-admins` Authelia group maps to the `admin` role in [openCloud]. This is because of the
`role_claim` value set to `groups`, and the list of mappings in `role_mappings`. Examples also exist for the
`spaceadmin`, `user`, and `user-light` roles in [openCloud].

You should adapt this configuration to your needs.

```yaml {title="proxy.yaml"}
role_assignment:
    driver: 'oidc'
    oidc_role_mapper:
        role_claim: 'groups'
        role_mapping:
          - role_name: 'admin'
            claim_value: 'opencloud-admins'
          - role_name: 'spaceadmin'
            claim_value: 'opencloud-space-admins'
          - role_name: 'user'
            claim_value: 'opencloud-users'
          - role_name: 'user-light'
            claim_value: 'opencloud-guests'
```

## See Also

- [openCloud]
- [openCloud - Integrating external OpenID Connect Identity Providers](https://docs.opencloud.eu/docs/admin/configuration/authentication-and-user-management/external-idp/)

[Authelia]: https://www.authelia.com
[openCloud]: https://opencloud.eu
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
