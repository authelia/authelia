---
title: "openCloud"
description: "Integrating openCloud with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-07-17T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/openCloud/'
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
  - [v4.39.20](https://github.com/authelia/authelia/releases/tag/v4.39.20)
- [openCloud]
  - [v7.2.2](https://github.com/opencloud-eu/opencloud/releases/tag/v7.2.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://opencloud.{{< sitevar name="domain" nojs="example.com" >}}`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
- __Client ID:__
  - Web Application: `OpenCloudWeb`
  - Android App: `OpenCloudAndroid`
  - iOS App: `OpenCloudIOS`
  - Desktop client: `OpenCloudDesktop`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia
#### Limitations

* When using Authelia as an external IDP, the refresh token lifespan must be increased manually. This workaround should no longer be required once Authelia supports dynamic client registration (DCR), see the [Authelia OpenID Connect roadmap](https://www.authelia.com/roadmap/active/openid-connect-1.0-provider/#beta-8)

* When using the desktop client with Authelia as IDP:

  * If running behind Nginx, disable the common exploit protection rule for Authelia, as it interferes with the authentication flow.
  * The desktop client WebFinger integration is currently incomplete (pull request https://github.com/opencloud-eu/desktop/pull/847).

    The `groups` scope must be manually added to the authorization link when setting up the desktop client.

    Example:

    ```
    https://<authelia-domain>/api/oidc/authorization?response_type=code&client_id=<client_id>&redirect_uri=<redirect_uri>&code_challenge=<code_challenge>&code_challenge_method=S256&scope=<scope>&prompt=<prompt>&state=<state>
    ```

    The default scope is:

    ```
    scope=openid%20offline_access%20email%20profile
    ```

    Add `groups%20` manually at the beginning of the scope:

    ```
    scope=groups%20openid%20offline_access%20email%20profile
    ```

    {{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
  The desktop client integration with Authelia is currently not production ready.

  The current implementation is intended for one-time use cases, such as migrating files. It should not be considered a stable long-term desktop client setup until the WebFinger integration is completed.
  {{< /callout >}}





The following YAML configuration is an example __Authelia__ [client configuration] for use with
[openCloud] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    # Extend the access and refresh token lifespan from the default 30m to work around ownCloud client re-authentication prompts every few hours.
    # TODO It should be possible to remove this once Authelia supports dynamic client registration (DCR).
    # Note: ownCloud's built-in IDP uses a value of 30d.
    lifespans:
      custom:
        openCloud:
          access_token: '2 days'
          refresh_token: '3 days' # use 30 if external IDP (e.g Authelia)
    cors:
      endpoints:
        - 'authorization'
        - 'token'
        - 'revocation'
        - 'introspection'
        - 'userinfo'

    clients:
      - client_id: 'OpenCloudWeb'
        client_name: 'openCloud'
        lifespan: 'openCloud'
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
          - 'https://opencloud.example.com/'
          - 'https://opencloud.example.com/oidc-callback.html'
          - 'https://opencloud.example.com/oidc-silent-redirect.html'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
      
      - client_id: 'OpenCloudAndroid'
        client_name: 'openCloud (Android)'
        public: true
        lifespan: 'openCloud'
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

      - client_id: 'OpenCloudDesktop'
        client_name: 'openCloud (Desktop Client)'
        public: true
        lifespan: 'openCloud'
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

      - client_id: 'OpenCloudIOS'
        client_name: 'openCloud (iOS)'
        public: true
        lifespan: 'openCloud'
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

#### Environment Variables

To configure [openCloud] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
  OC_OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
  PROXY_OIDC_ACCESS_TOKEN_VERIFY_METHOD=jwt
  PROXY_OIDC_REWRITE_WELLKNOWN=true
  PROXY_AUTOPROVISION_ACCOUNTS=true
  PROXY_AUTOPROVISION_CLAIM_USERNAME=preferred_username
  PROXY_AUTOPROVISION_CLAIM_EMAIL=email
  PROXY_AUTOPROVISION_CLAIM_DISPLAYNAME=name
  PROXY_AUTOPROVISION_CLAIM_GROUPS=groups
  PROXY_CSP_CONFIG_FILE_LOCATION=/etc/opencloud/csp.yaml

  # Configure the clients
  WEBFINGER_WEB_OIDC_CLIENT_ID=OpenCloudWeb
  WEBFINGER_WEB_OIDC_CLIENT_SCOPES="openid profile email groups offline_access"

  WEBFINGER_ANDROID_OIDC_CLIENT_ID=OpenCloudAndroid
  WEBFINGER_ANDROID_OIDC_CLIENT_SCOPES="openid profile email groups offline_access"

  WEBFINGER_IOS_OIDC_CLIENT_ID=OpenCloudIOS
  WEBFINGER_IOS_OIDC_CLIENT_SCOPES="openid profile email groups offline_access"

  WEBFINGER_DESKTOP_OIDC_CLIENT_ID=OpenCloudDesktop
  WEBFINGER_DESKTOP_OIDC_CLIENT_SCOPES="openid profile email groups offline_access"

  # When using external IdP (e.g. Authelia), disable the internal IdP service to avoid conflicts
  OC_EXCLUDE_RUN_SERVICES=idp
  PROXY_ROLE_ASSIGNMENT_DRIVER=oidc
  GRAPH_ASSIGN_DEFAULT_USER_ROLE=false
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  oics:
    environment:
      OC_OIDC_ISSUER: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
      PROXY_OIDC_ACCESS_TOKEN_VERIFY_METHOD: jwt
      PROXY_OIDC_REWRITE_WELLKNOWN: true
      PROXY_AUTOPROVISION_ACCOUNTS: true
      PROXY_AUTOPROVISION_CLAIM_USERNAME: preferred_username
      PROXY_AUTOPROVISION_CLAIM_EMAIL: email
      PROXY_AUTOPROVISION_CLAIM_DISPLAYNAME: name
      PROXY_AUTOPROVISION_CLAIM_GROUPS: groups
      PROXY_CSP_CONFIG_FILE_LOCATION: /etc/opencloud/csp.yaml

      # Configure the clients
      WEBFINGER_WEB_OIDC_CLIENT_ID: OpenCloudWeb
      WEBFINGER_WEB_OIDC_CLIENT_SCOPES: openid profile email groups offline_access

      WEBFINGER_ANDROID_OIDC_CLIENT_ID: OpenCloudAndroid
      WEBFINGER_ANDROID_OIDC_CLIENT_SCOPES: openid profile email groups offline_access

      WEBFINGER_IOS_OIDC_CLIENT_ID: OpenCloudIOS
      WEBFINGER_IOS_OIDC_CLIENT_SCOPES: openid profile email groups offline_access

      WEBFINGER_DESKTOP_OIDC_CLIENT_ID: OpenCloudDesktop
      WEBFINGER_DESKTOP_OIDC_CLIENT_SCOPES: openid profile email groups offline_access

      # When using external IdP (e.g. Authelia), disable the internal IdP service to avoid conflicts
      OC_EXCLUDE_RUN_SERVICES: idp
      PROXY_ROLE_ASSIGNMENT_DRIVER: oidc
      GRAPH_ASSIGN_DEFAULT_USER_ROLE: false
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
    - "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"
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
    # In contrary to bash and docker the default is given after the | character
    - 'https://${COLLABORA_DOMAIN|collabora.opencloud.test}${TRAEFIK_PORT_HTTPS}/'
    - 'https://${EURO_OFFICE_DOMAIN|euro-office.opencloud.test}${TRAEFIK_PORT_HTTPS}/'
    # This is needed for the external-sites web extension when embedding sites
    - 'https://docs.opencloud.eu'
    - "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"

  img-src:
    - '''self'''
    - 'data:'
    - 'blob:'
    - 'https://raw.githubusercontent.com/opencloud-eu/awesome-apps/'
    - 'https://tile.openstreetmap.org/'
    # In contrary to bash and docker the default is given after the | character
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
    - "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"
  style-src:
    - '''self'''
    - '''unsafe-inline'''
  worker-src:
    - "'self'"
    - 'blob:'
```
Refer to [csp.yaml](https://github.com/opencloud-eu/opencloud-compose/blob/main/config/opencloud/csp.yaml)


#### Proxy
When using an external IDP, you need to map groups and roles. Create the following file and save it next to `opencloud.yaml`:

```yaml {title="proxy.yaml"}
role_assignment:
    driver: oidc
    oidc_role_mapper:
        role_claim: groups
        role_mapping:
          - role_name: admin
            claim_value: myAdminRole
          - role_name: spaceadmin
            claim_value: mySpaceAdminRole
          - role_name: user
            claim_value: myUserRole
          - role_name: user-light
            claim_value: myGuestRole
```

## See Also

- [openCloud]
- [openCloud - Integrating external OpenID Connect Identity Providers](https://docs.opencloud.eu/docs/admin/configuration/authentication-and-user-management/external-idp/)

[Authelia]: https://www.authelia.com
[openCloud]: https://opencloud.eu
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
