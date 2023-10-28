---
title: "Reset Password"
description: "Reset Password Identity Validation Configuration"
lead: "Authelia uses multiple methods to verify the identity of users to prevent a malicious user from performing actions on behalf of them. This section describes Reset Password method."
date: 2023-10-28T18:57:18+11:00
draft: false
images: []
menu:
  configuration:
    parent: "identity-validation"
weight: 105200
toc: true
---

The Reset Password Identity Validation implementation ensures that users cannot perform a reset password flow without
first ensuring the user is adequately identified. The settings below therefore can affect the level of security Authelia
provides to your users so they should be carefully considered.

This process is performed by issuing a HMAC signed JWT using a secret key only known by Authelia.

## Configuration

{{< config-alert-example >}}

```yaml
identity_validation:
  reset_password:
    expiration: '5 minutes'
    jwt_algorithm: 'HS256'
    jwt_secret: ''
```

## Options

This section describes the individual configuration options.

### expiration

{{< confkey type="string,integer" syntax="duration" default="5 minutes" required="no" >}}

The duration of time the emailed JWT is considered valid.

### jwt_algorithm

{{< confkey type="string" default="HS256" required="no" >}}

The JWA used to sign the JWT. Must be one of the HMAC backed JWA's.

### jwt_secret

{{< confkey type="string" required="yes" >}}

The secret used with the HMAC algorithm to sign the JWT.
