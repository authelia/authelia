---
title: "Express.js"
description: "Integrating Express.js with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-18T11:00:43+10:00
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
* [Express.js]

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://express.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `Express.js`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Express.js] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'Express.js'
        client_name: 'Express.js App'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        require_pushed_authorization_requests: true
        redirect_uris:
          - 'https://express.example.com/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Express.js] to utilize Authelia as an [OpenID Connect 1.0] Provider:

#### Project Initialization

```bash
mkdir authelia-example && cd authelia-example && npm init -y && npm install express express-openid-connect
```

#### Create The Application

This application example assumes you're proxying the Node service with a proxy handling TLS termination for
`https://express.example.com`.

```js {title="server.js"}
"use strict";

const express = require('express');
const { auth, requiresAuth } = require('express-openid-connect');
const { randomBytes } = require('crypto');

const app = express();

app.use(
  auth({
    authRequired: false,
    baseURL: `${process.env.APP_BASE_URL || 'https://express.example.com'}/callback`,
    secret: process.env.SESSION_ENCRYPTION_SECRET || randomBytes(64).toString('hex'),
    clientID: process.env.OIDC_CLIENT_ID || 'Express.js',
    clientSecret: process.env.OIDC_CLIENT_SECRET || 'insecure_secret',
    clientAuthMethod: 'client_secret_basic',
    issuerBaseURL: process.env.OIDC_ISSUER || 'https://auth.example.com',
    pushedAuthorizationRequests: true,
    authorizationParams: {
      response_type: 'code',
      scope: 'openid profile email groups',
    },
  })
);

app.get('/', requiresAuth(), (req, res) => {
  req.oidc.fetchUserInfo().then((userInfo) => {
    const data = JSON.stringify(
      {
        accessToken: req.oidc.accessToken,
        refreshToken: req.oidc.refreshToken,
        idToken: req.oidc.idToken,
        claims: req.oidc.idTokenClaims,
        scopes: req.oidc.scope,
        userInfo,
      }, null, 2);

    res.send(`<html lang='en'><body><pre><code>${data}</code></pre></body></html>`);
  });
});

app.listen(3000, function () {
  console.log("Listening on port 3000")
});
```

Environment Example:

```env
APP_BASE_URL=https://express.example.com
SESSION_ENCRYPTION_SECRET=
OIDC_ISSUER=https://auth.example.com
OIDC_CLIENT_ID=Express.js
OIDC_CLIENT_SECRET=insecure_secret
```

## See Also

- [express-openid-connect] API Reference Guide

[Express.js]: https://Express.js.com/
[express-openid-connect]: https://auth0.github.io/express-openid-connect/interfaces/ConfigParams.html
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
