---
title: "Secrets"
description: "Using the Secrets Configuration Method."
lead: "Authelia allows providing configuration via secrets method. This section describes how to implement this."
date: 2020-02-29T01:43:59+01:00
draft: false
images: []
menu:
  configuration:
    parent: "methods"
weight: 101400
toc: true
aliases:
  - /c/secrets
  - /docs/configuration/secrets.html
---

Configuration of *Authelia* requires several secrets and passwords. Even if they can be set in the configuration file or
standard environment variables, the recommended way to set secrets is to use this configuration method as described below.

See the [security](#security) section for more information.

## Layers

*__Important Note:__* While this method is the third layer of the layered configuration model as described by the
[introduction](introduction.md#layers), this layer is special in as much as *Authelia* will not start if you define
a secret as well as any other configuration method.

For example if you define `jwt_secret` in the [files method](files.md) and/or `AUTHELIA_JWT_SECRET` in the
[environment method](environment.md), as well as the `AUTHELIA_JWT_SECRET_FILE`, this will cause the aforementioned error.

## Security

This method is a slight improvement over the security of the other methods as it allows you to easily separate your
configuration in a logically secure way.

## Environment variables

A secret value can be loaded by *Authelia* when the configuration key ends with one of the following words: `key`,
`secret`, `password`, or `token`.

If you take the expected environment variable for the configuration option with the `_FILE` suffix at the end. The value
of these environment variables must be the path of a file that is readable by the Authelia process, if they are not,
*Authelia* will fail to load. Authelia will automatically remove the newlines from the end of the files contents.

For instance the LDAP password can be defined in the configuration
at the path __authentication_backend.ldap.password__, so this password
could alternatively be set using the environment variable called
__AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE__.

Here is the list of the environment variables which are considered secrets and can be defined. Please note that only
secrets can be loaded into the configuration if they end with one of the suffixes above, you can set the value of any
other configuration using the environment but instead of loading a file the value of the environment variable is used.

|                  Configuration Key                  |                   Environment Variable                   |
|:---------------------------------------------------:|:--------------------------------------------------------:|
|                  [server.tls.key]                   |               AUTHELIA_SERVER_TLS_KEY_FILE               |
|                    [jwt_secret]                     |                 AUTHELIA_JWT_SECRET_FILE                 |
|                [duo_api.secret_key]                 |             AUTHELIA_DUO_API_SECRET_KEY_FILE             |
|                  [session.secret]                   |               AUTHELIA_SESSION_SECRET_FILE               |
|              [session.redis.password]               |           AUTHELIA_SESSION_REDIS_PASSWORD_FILE           |
| [session.redis.high_availability.sentinel_password] | AUTHELIA_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE  |
|              [storage.encryption_key]               |           AUTHELIA_STORAGE_ENCRYPTION_KEY_FILE           |
|              [storage.mysql.password]               |           AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE           |
|             [storage.postgres.password]             |         AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE          |
|              [notifier.smtp.password]               |           AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE           |
|       [authentication_backend.ldap.password]        |    AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE    |
|    [identity_providers.oidc.issuer_private_key]     | AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE |
|        [identity_providers.oidc.hmac_secret]        |    AUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE     |

[server.tls.key]: ../miscellaneous/server.md#key
[jwt_secret]: ../miscellaneous/introduction.md#jwt_secret
[duo_api.secret_key]: ../second-factor/duo.md#secret_key
[session.secret]: ../session/introduction.md#secret
[session.redis.password]: ../session/redis.md#password
[session.redis.high_availability.sentinel_password]: ../session/redis.md#sentinel_password
[storage.encryption_key]: ../storage/introduction.md#encryption_key
[storage.mysql.password]: ../storage/mysql.md#password
[storage.postgres.password]: ../storage/postgres.md#password
[notifier.smtp.password]: ../notifications/smtp.md#password
[authentication_backend.ldap.password]: ../first-factor/ldap.md#password
[identity_providers.oidc.issuer_private_key]: ../identity-providers/open-id-connect.md#issuer_private_key
[identity_providers.oidc.hmac_secret]: ../identity-providers/open-id-connect.md#hmac_secret


## Secrets in configuration file

If for some reason you decide on keeping the secrets in the configuration file, it is strongly recommended that you
ensure the permissions of the configuration file are appropriately set so that other users or processes cannot access
this file. Generally the UNIX permissions that are appropriate are 0600.

## Secrets exposed in an environment variable

In all versions 4.30.0+ you can technically set secrets using the environment variables without the `_FILE` suffix by
setting the value to the value you wish to set in configuration, however we strongly urge people not to use this option
and instead use the file-based secrets above.

Prior to implementing file secrets the only way you were able to define secret values was either via configuration or
via environment variables in plain text.

See [this article](https://diogomonica.com/2017/03/27/why-you-shouldnt-use-env-variables-for-secret-data/) for reasons
why setting them via the file counterparts is highly encouraged.

## Examples

See the [Docker Integration](../../integration/deployment/docker.md) and
[Kubernetes Integration](../../integration/kubernetes/secrets.md) guides for examples of secrets.
