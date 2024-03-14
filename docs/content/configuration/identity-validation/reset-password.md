---
title: "Reset Password"
description: "Reset Password Identity Validation Configuration"
summary: "Authelia uses multiple methods to verify the identity of users to prevent a malicious user from performing actions on behalf of them. This section describes Reset Password method."
date: 2024-03-04T20:29:11+11:00
draft: false
images: []
menu:
  configuration:
    parent: "identity-validation"
weight: 105200
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The Reset Password Identity Validation implementation ensures that users cannot perform a reset password flow without
first ensuring the user is adequately identified. The settings below therefore can affect the level of security Authelia
provides to your users so they should be carefully considered.

This process is performed by issuing a HMAC signed JWT using a secret key only known by Authelia.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
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

The JWA used to sign the JWT. Must be HS256, HS384, or HS512.

### jwt_secret

{{< confkey type="string" required="yes" >}}

The secret used with the HMAC algorithm to sign the JWT.
