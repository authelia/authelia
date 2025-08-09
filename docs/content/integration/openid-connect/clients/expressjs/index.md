---
title: "Express.js"
description: "Integrating Express.js with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-18T11:00:43+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/expressjs/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Express.js | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Express.js with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Express.js]

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://express.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `expressjs-example`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
This is a developer guide which can be used to create your open application. It's worth noting that some applications
may also use the associated libraries so the configurations may be adaptable to those applications.
{{< /callout >}}

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Express.js] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'expressjs-example'
        client_name: 'Express.js App'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        require_pushed_authorization_requests: true
        redirect_uris:
          - 'https://express.{{< sitevar name="domain" nojs="example.com" >}}/callback'
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

To configure [Express.js] there is one method, using the [Generalized Instructions](#generalized-instructions).

#### Generalized Instructions

Because each project is different this guide just demonstrates how this is possible.

##### Project Initialization

```shell
mkdir authelia-example && cd authelia-example && npm init -y && npm install express express-openid-connect
```

##### Create The Application

This application example assumes you're proxying the Node service with a proxy handling TLS termination for
`https://express.{{< sitevar name="domain" nojs="example.com" >}}`.

```js {title="server.js"}
"use strict";

const express = require('express');
const { auth, requiresAuth } = require('express-openid-connect');
const { randomBytes } = require('crypto');

const app = express();

app.use(
  auth({
    authRequired: false,
    baseURL: `${process.env.APP_BASE_URL || 'https://express.{{< sitevar name="domain" nojs="example.com" >}}'}`,
    secret: process.env.SESSION_ENCRYPTION_SECRET || randomBytes(64).toString('hex'),
    clientID: process.env.OIDC_CLIENT_ID || 'expressjs-example',
    clientSecret: process.env.OIDC_CLIENT_SECRET || 'insecure_secret',
    clientAuthMethod: process.env.OIDC_CLIENT_AUTH_METHOD || 'client_secret_basic',
    issuerBaseURL: process.env.OIDC_ISSUER || 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}',
    pushedAuthorizationRequests: toBoolean(process.env.OIDC_PUSHED_AUTHORIZATION_REQUESTS, true),
    authorizationParams: {
      response_type: 'code',
      scope: process.env.OIDC_SCOPE || 'openid profile email groups',
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
        claims: {
          id_token: req.oidc.idToken,
          userinfo: userInfo,
        },
        scopes: req.oidc.scope,
      }, null, 2);

    res.send(`<html lang='en'><body><pre><code>${data}</code></pre></body></html>`);
  });
});

app.listen(3000, function () {
  console.log("Listening on port 3000")
});

function toBoolean(value, defaultValue) {
  switch (value) {
    case "true":
    case "TRUE":
    case "1":
      return true
    case "false":
    case "FALSE":
    case "0":
      return false
    default:
      return defaultValue
  }
}
```

##### #nvironment Variables

To configure [Express.js] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

###### Standard

```shell {title=".env"}
APP_BASE_URL=https://express.{{< sitevar name="domain" nojs="example.com" >}}
# SESSION_ENCRYPTION_SECRET=
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=expressjs-example
OIDC_CLIENT_SECRET=insecure_secret
OIDC_PUSHED_AUTHORIZATION_REQUESTS=true
OIDC_CLIENT_AUTH_METHOD=client_secret_basic
OIDC_SCOPE=openid profile email groups
```

###### Docker Compose

```yaml {title="compose.yml"}
services:
  expressjs-example:
    environment:
      APP_BASE_URL: 'https://express.{{< sitevar name="domain" nojs="example.com" >}}'
      # SESSION_ENCRYPTION_SECRET: ''
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'expressjs-example'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_PUSHED_AUTHORIZATION_REQUESTS: 'true'
      OIDC_CLIENT_AUTH_METHOD: 'client_secret_basic'
      OIDC_SCOPE: 'openid profile email groups'
```

## See Also

- [express-openid-connect] API Reference Guide

[Express.js]: https://expressjs.com/
[express-openid-connect]: https://auth0.github.io/express-openid-connect/interfaces/ConfigParams.html
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
