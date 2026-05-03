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
  - [v4.39.19](https://github.com/authelia/authelia/releases/tag/v4.39.19)
- [Jellyfin]
  - [v10.10.7](https://github.com/jellyfin/jellyfin/releases/tag/v10.10.7)

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

To prepare Jellyfin to utilize Authelia as an [OpenID Connect 1.0] Provider, the [Jellyfin SSO-Auth Plugin], a plugin that is not
a part of the standard Jellyfin distribution, needs to be installed.

**Add the repository for the SSO-Auth plugin and install the plugin**

1. Visit the [Jellyfin] Administration Dashboard.
2. Visit the `Plugins` section.
3. Visit the `Repositories` tab.
4. Click the `+` to add a repository.
5. Enter the following details:
   - Repository Name: `Jellyfin SSO-Auth`
   - Repository URL: `https://raw.githubusercontent.com/9p4/jellyfin-plugin-sso/manifest-release/manifest.json`
6. Click `Save`.
8. Click `Ok` to confirm the repository installation.
9. Go Back to  the `Plugins` tab.
10. Select __All__ plugins and the __Other__ Cateegory, find `SSO-Auth` and select it.
11. Click `Install`.
12. Click `Ok` to confirm the plugin installation.
13. Once installed restart [Jellyfin].

To configure [Jellyfin's SSO-Auth Plugin] there are two methods, using the [Configuration File](#configuration-file), or using the
[Web GUI](#web-gui).

#### Configuration File
To configure [Jellyfin] to utilize Authelia as an [OpenID Connect 1.0] Provider via a configuration file, use the following instructions:

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
          <SchemeOverride>https</SchemeOverride>
        </PluginConfiguration>
      </value>
    </item>
  </OidConfigs>
</PluginConfiguration>
```

1. Using the above XML create a file called `SSO-Auth.xml` and place it in the proper directory based on where [Jellyfin]
is running.
Linux: /var/lib/jellyfin/plugins/configurations/SSO-Auth.xml
Docker: /config/plugins/configurations/SSO-Auth.xml (likely via a volume mount)
Windows: %ProgramData%\Jellyfin\Server\plugins\configurations\SSO-Auth.xml or %UserProfile%\AppData\Local\jellyfin\plugins\configurations\SSO-Auth.xml

2. Restart [Jellyfin]

To test if your Jellyfin server properly loaded the SSO-Auth configuration file:
1. Visit the [Jellyfin] Administration Dashboard
2. Visit the `Plugins` Section
3. Click the `SSO-Auth` plugin.
4. Click ⚙ Settings button.

If you see **authelia** in __Name of OID Provider__, or it is selectable via the drop down, your plugin configuration file is being processed correctly.

If not double check the path and the permissions of the file, on Linux the **jellyfin** user needs to be able to read the file.
Linux:
> sudo chown jellyfin:jellyfin /var/lib/jellyfin/plugins/configurations/SSO-Auth.xml

#### Web GUI
To configure [Jellyfin] to utilize Authelia as an [OpenID Connect 1.0] Provider via the Web GUI, use the following instructions:

1. Visit the [Jellyfin] Administration Dashboard.
2. Visit the `Plugins` section.
3. Click the `SSO-Auth` plugin.
4. Click ⚙ Settings button.
5. Add a provider.
6. Configure the following options:
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
    - Scheme Override: `https`
7. All other options may remain unchecked or unconfigured.
8. Click `Save`.
9. To log in visit `https://jellyfin.{{< sitevar name="domain" nojs="example.com" >}}/sso/OID/start/authelia`.
10. Follow the [Jellyfin SSO-Auth Plugin] documentation on how to create a button on the [Jellyfin] login page.

#### Add a Login Button

1. Visit the [Jellyfin] Administration Dashboard.
2. Visit the `Branding` section.
3. Add the following HTML code into the `Login disclaimer` section.
```html
<form action="https://jellyfin.{{< sitevar name="domain" nojs="example.com" >}}/sso/OID/start/authelia">
  <button class="raised block emby-button button-submit">
    https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}
  </button>
</form>
```
4. Add the following CSS code into the `Custom CSS Code` section.
```css
a.raised.emby-button {
  padding: 0.9em 1em;
  color: inherit !important;
}

.disclaimerContainer {
  display: block;
}
```

## See Also

- [Jellyfin SSO-Auth Plugin] Repository

[Authelia]: https://www.authelia.com
[Jellyfin]: https://jellyfin.org/
[Jellyfin SSO-Auth Plugin]: https://github.com/9p4/jellyfin-plugin-sso
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
