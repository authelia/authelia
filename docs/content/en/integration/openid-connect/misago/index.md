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

## Tested Versions

* [Authelia](https://www.authelia.com)
  * [v4.37.5](https://github.com/authelia/authelia/releases/tag/v4.37.5)
* [Misago](https://github.com/rafalp/Misago)
  * [misago-image v0.29.1](https://github.com/tetricky/misago-image/releases/tag/v0.29.1)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://misago.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `misago`
* __Client Secret:__ `insecure_secret`


## Configuration

### Application

To configure [Misago] to utilize Authelia as an [OpenID Connect 1.0](https://www.authelia.com/integration/openid-connect/introduction/) Provider:

1. Sign in to the [Misago] Admin Panel
2. Visit `Settings` and click `OAuth 2`
3. Configure the Following:
    1. Basic settings:
        1. Provider name: `authelia`
        2. Client ID: `misago`
        3. Client Secret: `insecure_secret`
    2. Initializing Login:
        1. Login form URL: `https://auth.example.com/api/oidc/authorization`
        2. Scopes: `openid profile email`
    3. Retrieving access token:
        1. Access token retrieval URL: `https://auth.example.com/api/oidc/token`
        2. Request method: `POST`
        3. JSON path to access token: `access_token`
    4. Retrieving user data:
        1. User data URL: `https://auth.example.com/api/oidc/userinfo`
        2. Request method: `GET`
        3. Access token location: `Query string`
        4. Access token name: `access_token`
    5. User JSON mappings:
        1. User ID path: `sub`
        2. User name path: `name`
        3. User e-mail path: `email`
4. Save the configuration

{{< figure src="misago-step-2.png" alt="Settings" width="736" style="padding-right: 10px" >}}

{{< figure src="misago-step-3-1.png" alt="Basic Settings" width="736" style="padding-right: 10px" >}}

{{< figure src="misago-step-3-2.png" alt="Initializing Login" width="736" style="padding-right: 10px" >}}

{{< figure src="misago-step-3-3.png" alt="Retrieving access token" width="736" style="padding-right: 10px" >}}

{{< figure src="misago-step-3-4.png" alt="Retrieving user data" width="736" style="padding-right: 10px" >}}

{{< figure src="misago-step-3-5.png" alt="User JSON mappings" width="736" style="padding-right: 10px" >}}

### Authelia

The following YAML configuration is an example **Authelia** [client configuration](https://www.authelia.com/configuration/identity-providers/open-id-connect/#clients) for use with [Misago] which will operate with the above example:

```yaml
    clients:
      - id: misago
        secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: two_factor
        scopes:
          - openid
          - profile
          - email
        redirect_uris:
          - https://misago.example.com/oauth2/complete/
        grant_types:
          - authorization_code
        response_types:
          - code
        response_modes:
          - query
        userinfo_signing_algorithm: none
```

---
## See Also

-   [Misago] [OAuth 2 Client Configuration guide](https://misago-project.org/t/oauth-2-client-configuration-guide/1147/)

[Misago]: https://misago-project.org/
