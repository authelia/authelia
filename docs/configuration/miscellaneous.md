---
layout: default
title: Miscellaneous
parent: Configuration
nav_order: 3
---

# Miscellaneous

Here are the main customizable options in Authelia.

## Host & Port

`optional: true`

Defines the address and port to listen on.

```yaml
host: 0.0.0.0
port: 9091
```

## TLS

`optional: true`

Authelia can use TLS. Provide the certificate and the key with the
following configuration options:

```yaml
tls_key: /config/ssl/key.pem
tls_cert: /config/ssl/cert.pem
```

## Log

### Log level

`optional: true`

Defines the level of logs used by Authelia. This level can be set to
`trace`, `debug` or `info`. When setting log_level to trace, you will
generate a large amount of log entries and expose the /debug/vars and
/debug/pprof/ endpoints which should not be enabled in production.

```yaml
log_level: debug
```

### Log file path

`optional: true`

Logs can be stored in a file when file path is provided. Otherwise logs
are written to standard output.

```yaml
log_file_path: /config/authelia.log
```


## JWT Secret

`optional: false`

Defines the secret used to craft JWT tokens leveraged by the identity
verification process. This can also be defined using a [secret](./secrets.md).

```yaml
jwt_secret: v3ry_important_s3cr3t
```

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
