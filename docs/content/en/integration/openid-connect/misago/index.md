---
title: "Misago"
description: "Integrating Misago with the Authelia OpenID Connect Provider."
lead: ""
date: 2023-03-04T13:20:00+00:00
draft: false
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 620
toc: true
community: true
---

## Tested Versions [#](https://www.authelia.com/integration/openid-connect/misago//#tested-versions)

-   [Authelia](https://www.authelia.com)
    -   [v4.37.5](https://github.com/authelia/authelia/releases/tag/v4.37.5)
-   [Misago](https://github.com/rafalp/Misago)
    -   [misago-image v0.29.1](https://github.com/tetricky/misago-image/releases/tag/v0.29.1)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `[https://misago.example.com](https://misago.example.com)`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `misago`
* __Client Secret:__ `misago_client_secret`


## Configuration

### Application

To configure [Misago](https://misago-project.org/) to utilize Authelia as an [OpenID Connect 1.0](https://www.authelia.com/integration/openid-connect/introduction/) Provider:

1.  sign in to your forum's admin panel
2.  go to settings and click on "OAuth 2"

![Settings](https://misago-project.org/media/attachments/4a/f5/b0r8pRvbPRmj4BR4JHyCCL4JIm4Afu66zjRKZtfG5nBCLbk91QmD6C8UF7lHSKpf/zrzut-ekranu-202.png "oauth2")

3.  Basic Settings (for now leave 'Enable Oauth2 Client:' set to `No`)
    1.  Provider name: `authelia`
    2.  Client ID: `misago`
    3.  Client Secret: `misago_client_secret`

![Basic Settings](https://misago-project.org/a/XKgsgFihsa1Q9kO0mh0xY1x2C5XvpLGI4CaB4kIJoXQbhvzKPfls3dBM2ogsSNIs/572/?shva=1 "Basic Settings")

4.  Initializing Login
    1.  Login form URL: `https://auth.example.com/api/oidc/authorization`
    2.  Scopes: `openid profile email`

![Initializing Login](https://misago-project.org/a/UWWsRH9jz8clGQBWBt89dEaKdJAqVgl732QIT1NYlzj251r2WeYCs35nfHc2yF4Y/571/?shva=1 "Initializing Login")

5.  Retrieving access token
    1.  Access token retrieval URL: `https://auth.example.com/api/oidc/token`
    2.  Request method: `POST`
    3.  JSON path to access token: `access_token`

![Retrieving access token](https://misago-project.org/a/W6bIv37kiWEwvE9DQw0kzy6xlNfzITz8MjV8uliDAOKfMuxPTKOd6pBPkhWjDIP2/570/?shva=1 "Retrieving access token")

6.  Retrieving user data
    1.  User data URL: `https://auth.example.com/api/oidc/userinfo`
    2.  Request method: `GET`
    3.  Access token location: `Query string`
    4.  Access token name: `access_token`


![Retrieving user data](https://misago-project.org/a/4rVurbRGSWxq9qqNLi1GbWvIsEgi5V4JPzIolNREA2CcOF3Ay4POyAHkUg3s6Bc6/569/?shva=1 "Retrieving user data")

7.  User JSON mappings
    1.  User ID path: `sub`
    2.  User name path: `name`
    3.  User e-mail path: `email`

![User JSON mappings](https://misago-project.org/a/gXwi50GWQTYTDKIWASFTQMJNWE6tDbjSRG8TlpqHJQe5bkrkh2B0fQp2nZYoxuLI/568/?shva=1 "User JSON mappings")

Save the settings and set up the authelia configuration.

### Authelia

The following YAML configuration is an example **Authelia** [client configuration](https://www.authelia.com/configuration/identity-providers/open-id-connect/#clients) for use with [Misago](https://misago-project.org/) which will operate with the above example:

for allowed origin:
```yaml
identity_providers:
  oidc:
      allowed_origins:
        - https://misago.example.com
      allowed_origins_from_client_redirect_uris: true
```

and for the client section:
```yaml
    clients:
      - id: misago
        secret: <misago_client_secret>
        public: false
        authorization_policy: two_factor
        scopes:
          - openid
          - profile
          - email
        redirect_uris:
          - https://misago.example.com/oauth2/complete/
        grant_types:
          - refresh_token
          - authorization_code
        response_types:
          - code
        response_modes:
          - form_post
          - query
          - fragment
        userinfo_signing_algorithm: none
```

Restart Authelia to apply to new configuration and check for any errors in the log

### Complete Application Settings

Assuming all is well, you can return to the Misago Oauth2 Settings page, in the admin panel:

8.  Basic Settings:
    1. Enable Oauth2 Client: `Yes`
    
Saving the settings should now activate Oauth2 login to Misago as a client from your Authelia instance.

---
## See Also

-   [Misago](https://misago-project.org/) [OAuth 2 client configuration guide](https://misago-project.org/t/oauth-2-client-configuration-guide/1147/)
    

