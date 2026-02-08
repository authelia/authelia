---
title: "Jellyfin"
description: "Integrating Jellyfin with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-12T21:54:41+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/jellyfin/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Jellyfin | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Jellyfin with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.17](https://github.com/authelia/authelia/releases/tag/v4.38.17)
- [Jellyfin]
  - [v10.10.6](https://github.com/jellyfin/jellyfin/releases/tag/v10.10.6)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://jellyfin.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `jellyfin`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

The following example uses the [Jellyfin SSO Plugin] which is assumed to be installed when following this
section of the guide.

To install the [Jellyfin SSO Plugin] for [Jellyfin] via the Web GUI:

1. Visit the [Jellyfin] Administration Dashboard.
2. Visit the `Plugins` section.
3. Visit the `Repositories` tab.
4. Click the `+` to add a repository.
5. Enter the following details:
   - Repository Name: `Jellyfin SSO`
   - Repository URL: `https://raw.githubusercontent.com/9p4/jellyfin-plugin-sso/manifest-release/manifest.json`
6. Click `Save`.
7. Click `Ok` to confirm the repository installation.

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Jellyfin] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'jellyfin'
        client_name: 'Jellyfin'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://jellyfin.{{< sitevar name="domain" nojs="example.com" >}}/sso/OID/redirect/authelia'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This configuration assumes [Jellyfin](https://jellyfin.org/) administrators are part of the `jellyfin-admins` group, and
[Jellyfin](https://jellyfin.org/) users are part of the `jellyfin-users` group. Depending on your specific group configuration, you will have
to adapt the `AdminRoles` and `Roles` nodes respectively. Alternatively you may elect to create a new authorization
policy in [provider authorization policies](../../../configuration/identity-providers/openid-connect/provider.md#authorization_policies) then utilize that policy as the [client authorization policy](./../../configuration/identity-providers/openid-connect/clients.md#authorization_policy).
{{< /callout >}}

To configure [Jellyfin] there are two methods, using the [Configuration File](#configuration-file), or using the
[Web GUI](#web-gui).

However the following steps must be compelted via the UI first regardless of which option you choose:

1. Visit the `Catalog` tab.
2. Select `SSO Authentication` from the `Authentication` section.
3. Click `Install`.
4. Click `Ok` to confirm the plugin installation.
5. Once installed restart [Jellyfin].

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `SSO-Auth.xml`.
{{< /callout >}}

To configure [Jellyfin] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```xml {title="SSO-Auth.xml"}
<?xml version="1.0" encoding="utf-8"?>
<PluginConfiguration xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <SamlConfigs />
  <OidConfigs>
    <item>
      <key>
        <string>authelia</string>
      </key>
      <value>
        <PluginConfiguration>
          <OidEndpoint>https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}</OidEndpoint>
          <OidClientId>jellyfin</OidClientId>
          <OidSecret>insecure_secret</OidSecret>
          <Enabled>true</Enabled>
          <EnableAuthorization>true</EnableAuthorization>
          <EnableAllFolders>true</EnableAllFolders>
          <EnabledFolders />
          <AdminRoles>
            <string>jellyfin-admins</string>
          </AdminRoles>
          <Roles>
            <string>jellyfin-users</string>
            <string>jellyfin-admins</string>
          </Roles>
          <EnableFolderRoles>false</EnableFolderRoles>
          <EnableLiveTvRoles>false</EnableLiveTvRoles>
          <EnableLiveTv>false</EnableLiveTv>
          <EnableLiveTvManagement>false</EnableLiveTvManagement>
          <LiveTvRoles />
          <LiveTvManagementRoles />
          <FolderRoleMappings />
          <RoleClaim>groups</RoleClaim>
          <OidScopes>
            <string>groups</string>
          </OidScopes>
          <CanonicalLinks></CanonicalLinks>
          <DisableHttps>false</DisableHttps>
          <DoNotValidateEndpoints>false</DoNotValidateEndpoints>
          <DoNotValidateIssuerName>false</DoNotValidateIssuerName>
        </PluginConfiguration>
      </value>
    </item>
  </OidConfigs>
</PluginConfiguration>
```

#### Web GUI

To configure [Jellyfin] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit the [Jellyfin] Administration Dashboard.
2. Visit the `Plugins` section.
3. Click the `SSO-Auth` plugin.
4. Add a provider.
5. Configure the following options:
    - Name of the OID Provider: `authelia`
    - OID Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
    - OpenID Client ID: `jellyfin`
    - OID Secret: `insecure_secret`
    - Enabled: Checked
    - Enable Authorization by Plugin: Checked
    - Enable All Folders: Checked
    - Roles: `jellyfin-users`, `jellyfin-admins`
    - Admin Roles: `jellyfin-admins`
    - Role Claim: `groups`
    - Request Additional Scopes: `groups`
    - Set default username claim: `preferred_username`
6. All other options may remain unchecked or unconfigured.
7. Click `Save`.
8. To log in visit `https://jellyfin.{{< sitevar name="domain" nojs="example.com" >}}/sso/OID/start/authelia`.
9. Follow the [Jellyfin SSO Plugin] documentation on how to create a button on the [Jellyfin] login page.

## See Also

- [Jellyfin SSO Plugin] Repository

[Authelia]: https://www.authelia.com
[Jellyfin]: https://jellyfin.org/
[Jellyfin SSO Plugin]: https://github.com/9p4/jellyfin-plugin-sso
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
