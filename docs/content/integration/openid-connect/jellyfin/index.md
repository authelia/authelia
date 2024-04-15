---
title: "Jellyfin"
description: "Integrating Jellyfin with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-12T21:54:41+10:00
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
    * [v4.38.6](https://github.com/authelia/authelia/releases/tag/v4.38.6)
* [Jellyfin]
    * [10.8.13](https://github.com/jellyfin/jellyfin/releases/tag/v10.8.13)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://jellyfin.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `jellyfin`
* __Client Secret:__ `insecure_secret`

## Configuration

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
          - 'https://jellyfin.example.com/sso/OID/redirect/authelia'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

_**Important Note:** This configuration assumes [Jellyfin] administrators are part of the `jellyfin-admins` group, and
[Jellyfin] users are part of the `jellyfin-users` group. Depending on your specific group configuration, you will have
to adapt the `AdminRoles` and `Roles` nodes respectively. Alternatively you may elect to create a new authorization
policy in [provider authorization policies] then utilize that policy as the [client authorization policy]._

[client authorization policy]: ../../../configuration/identity-providers/openid-connect/clients.md#authorization_policy
[provider authorization policies]: ../../../configuration/identity-providers/openid-connect/provider.md#authorization_policies

To configure [Jellyfin] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Visit the [Jellyfin] Administration Dashboard.

2. Visit the `Plugins` section.

3. Visit the `Repositories` tab.

4. Click the `+` to add a repository.

5. Enter the following details:
   1. Repository Name: `Jellyfin SSO`
   2. Repository URL: `https://raw.githubusercontent.com/9p4/jellyfin-plugin-sso/manifest-release/manifest.json`

6. Click `Save`.

7. Click `Ok` to confirm the repository installation.

8. Visit the `Catalog` tab.

9. Select `SSO Authentication` from the `Authentication` section.

10. Click `Install`.

11. Click `Ok` to confirm the plugin installation.

12. Once installed restart [Jellyfin].

13. Complete steps 1 and 2 again.

14. Click the `SSO-Auth` plugin.

15. Add a provider with the following settings:

    1. Name of the OID Provider: `Authelia`

    2. OID Endpoint: `https://auth.example.com`

    3. OpenID Client ID: `jellyfin`

    4. OID Secret: `insecure_secret`

    5. Enabled: Checked

    6. Enable Authorization by Plugin: Checked

    7. Enable All Folders: Checked

    8. Roles: `jellyfin-users`

    9. Admin Roles: `jellyfin-admins`

    10. Role Claim: `groups`

    11. Request Additional Scopes: `groups`

    12. Set default username claim: `preferred_username`

16. All other options may remain unchecked or unconfigured.

17. Click `Save`.

Alternatively you can utilize the following configuration XML:

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
          <OidEndpoint>https://auth.example.com</OidEndpoint>
          <OidClientId>jellyfin</OidClientId>
          <OidSecret>insecure_secret</OidSecret>
          <Enabled>true</Enabled>
          <EnableAuthorization>true</EnableAuthorization>
          <EnableAllFolders>true</EnableAllFolders>
          <EnabledFolders />
          <AdminRoles>
            <string>admins</string>
          </AdminRoles>
          <Roles>
            <string>users</string>
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

## See Also

* [Jellyfin SSO Plugin] Repository

[Authelia]: https://www.authelia.com
[Jellyfin]: https://jellyfin.org/
[Jellyfin SSO Plugin]: https://github.com/9p4/jellyfin-plugin-sso
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
