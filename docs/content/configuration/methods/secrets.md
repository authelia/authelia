---
title: "Secrets"
description: "Using the Secrets Configuration Method."
summary: "Authelia allows providing configuration via secrets method. This section describes how to implement this."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 101400
toc: true
aliases:
  - '/c/secrets'
  - '/docs/configuration/secrets.html'
  - '/configuration/secrets/'
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Configuration of *Authelia* requires several secrets and passwords. Even if they can be set in the configuration file or
standard environment variables, the recommended way to set secrets is to use this configuration method as described below.

See the [security](#security) section for more information.

## Filters

In addition to the documented methods below, the configuration files can be passed through templating filters. These
filters can be used to inject or modify content within the file. Specifically the `fileContent` function can be used to
retrieve content of a file, and `nindent` can be used to add a new line and indent the content of that file.

Take the following example:

```yaml {title="configuration.yml"}
authentication_backend:
  ldap:
    address: 'ldap://{{ env "SERVICES_SERVER" }}'
    tls:
      private_key: |
        {{- fileContent "./test_resources/example_filter_rsa_private_key" | nindent 8 }}
```

When considering the `address` the value from the environment variable `SERVICES_SERVER` are used in place of the content
starting at the `{{` and `}}`, which indicate the start and end of the template content.

When considering the `private_key` the start of a templated section also has a `-` which removes the whitespace before
the template section which starts the template content just after the `|` above it. The `fileContent` function reads the
content of the `./test_resources/example_filter_rsa_private_key` file (relative to the Authelia working directory), and
the `nindent` function adds a new line and indents every line in the file by `8` characters. Note the `|` between
`nindent` and `fileContent` passes the output of `fileContent` function to the `nindent` function.

For more information on [File Filters](files.md#file-filters) including how to enable them, see the
[File Filters](files.md#file-filters) guide.

## Layers

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
While this method is the third layer of the layered configuration model as described by the
[introduction](introduction.md#layers), this layer is special in as much as *Authelia* will not start if you define
a secret as well as any other configuration method.
{{< /callout >}}

For example if you define `jwt_secret` in the [files method](files.md) and/or `AUTHELIA_JWT_SECRET` in the
[environment method](environment.md), as well as the `AUTHELIA_JWT_SECRET_FILE`, this will cause the aforementioned error.

## Security

This method is a slight improvement over the security of the other methods as it allows you to easily separate your
configuration in a logically secure way.

## Environment variables

A secret value can be loaded by *Authelia* when the configuration key ends with one of the following words: `key`,
`secret`, `password`, `token` or `certificate_chain`.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
It is not possible to configure several sections using environment variables or secrets. The sections affected are all
lists of objects. These include but may not be limited to the rules section in access control, the clients section in
the OpenID Connect 1.0 Provider, the cookies section of in session, and the authz section in the server endpoints. See
[ADR2](../../reference/architecture-decision-log/2.md) for more information.
{{< /callout >}}

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

{{% table-config-keys secrets="true" %}}

[server.tls.key]: ../miscellaneous/server.md#key
[duo_api.integration_key]: ../second-factor/duo.md#integration_key
[duo_api.secret_key]: ../second-factor/duo.md#secret_key
[session.secret]: ../session/introduction.md#secret
[session.redis.password]: ../session/redis.md#password
[session.redis.tls.certificate_chain]: ../session/redis.md#tls
[session.redis.tls.private_key]: ../session/redis.md#tls
[session.redis.high_availability.sentinel_password]: ../session/redis.md#sentinel_password
[storage.encryption_key]: ../storage/introduction.md#encryption_key
[storage.mysql.password]: ../storage/mysql.md#password
[storage.mysql.tls.certificate_chain]: ../storage/mysql.md#tls
[storage.mysql.tls.private_key]: ../storage/mysql.md#tls
[storage.postgres.password]: ../storage/postgres.md#password
[storage.postgres.tls.certificate_chain]: ../storage/postgres.md#tls
[storage.postgres.tls.private_key]: ../storage/postgres.md#tls
[storage.postgres.ssl.key]: ../storage/postgres.md
[notifier.smtp.password]: ../notifications/smtp.md#password
[notifier.smtp.tls.certificate_chain]: ../notifications/smtp.md#tls
[notifier.smtp.tls.private_key]: ../notifications/smtp.md#tls
[authentication_backend.ldap.password]: ../first-factor/ldap.md#password
[authentication_backend.ldap.tls.certificate_chain]: ../first-factor/ldap.md#tls
[authentication_backend.ldap.tls.private_key]: ../first-factor/ldap.md#tls
[identity_providers.oidc.hmac_secret]: ../identity-providers/openid-connect/provider.md#hmac_secret
[identity_validation.reset_password.jwt_secret]: ../identity-validation/reset-password.md#jwt_secret

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

See [this article](https://blog.diogomonica.com/2017/03/27/why-you-shouldnt-use-env-variables-for-secret-data/) for reasons
why setting them via the file counterparts is highly encouraged.

## Examples

See the [Docker Integration](../../integration/deployment/docker.md) and
[Kubernetes Integration](../../integration/kubernetes/secrets.md) guides for examples of secrets.
