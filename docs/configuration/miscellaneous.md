---
layout: default
title: Miscellaneous
parent: Configuration
nav_order: 5
---

# Miscellaneous

Here are the main customizable options in Authelia.

## Host & Port
`optional: true`

Defines the address to listen on.

    host: 0.0.0.0
    port: 9091

## Logs level

`optional: true`

Defines the level of logs used by Authelia. This level can be set to
`trace`, `debug`, `info`.

    logs_level: debug


## JWT Secret

`optional: false`

Defines the secret used to craft JWT tokens leveraged by the identity
verification process

    jwt_secret: v3ry_important_s3cr3t

## Default redirection URL

`optional: true`

The default redirection URL is the URL where users are redirected when Authelia
cannot detect the target URL where the user was heading.

In a normal authentication workflow, a user tries to access a website and she
gets redirected to the sign-in portal in order to authenticate. Since the user
initially targeted a website, the portal knows where the user was heading and
can redirect her after the authentication process.
However, when a user visits the sign in portal directly, the portal considers
the targeted website is the portal. In that case and if the default redirection URL
is configured, the user is redirected to that URL. If not defined, the user is not
redirected after authentication.