---
title: "PostgreSQL"
description: "PostgreSQL Configuration"
summary: "The PostgreSQL storage provider."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 107400
toc: true
aliases:
  - /docs/configuration/storage/postgres.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Version support

See the [PostgreSQL Database Integration](../../reference/integrations/database-integrations.md#postgresql) reference
guide for supported version information.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
storage:
  encryption_key: 'a_very_important_secret'
  postgres:
    address: 'tcp://127.0.0.1:5432'
    database: 'authelia'
    schema: 'public'
    username: 'authelia'
    password: 'mypassword'
    timeout: '5s'
    tls:
      server_name: 'postgres.example.com'
      skip_verify: false
      minimum_version: 'TLS1.2'
      maximum_version: 'TLS1.3'
      certificate_chain: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
      private_key: |
        -----BEGIN RSA PRIVATE KEY-----
        ...
        -----END RSA PRIVATE KEY-----
```

## Options

This section describes the individual configuration options.

### encryption_key

See the [encryption_key docs](introduction.md#encryption_key).

### address

{{< confkey type="string" syntax="address" required="yes" >}}

Configures the address for the PostgreSQL Server. The address itself is a connector and the scheme must either be
the `unix` scheme or one of the `tcp` schemes.

__Examples:__

```yaml {title="configuration.yml"}
storage:
  postgres:
    address: 'tcp://127.0.0.1:5432'
```

```yaml {title="configuration.yml"}
storage:
  postgres:
    address: 'tcp://[fd00:1111:2222:3333::1]:5432'
```

```yaml {title="configuration.yml"}
storage:
  postgres:
    address: 'unix:///var/run/postgres.sock'
```

### database

{{< confkey type="string" required="yes" >}}

The database name on the database server that the assigned [user](#username) has access to for the purpose of
__Authelia__.

### schema

{{< confkey type="string" default="public" required="no" >}}

The database schema name to use on the database server that the assigned [user](#username) has access to for the purpose
of __Authelia__. By default this is the public schema.

### username

{{< confkey type="string" required="yes" >}}

The username paired with the password used to connect to the database.

### password

{{< confkey type="string" required="yes" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The password paired with the [username](#username) used to connect to the database.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters and the user password is changed to this value.

### timeout

{{< confkey type="string,integer" syntax="duration" default="5 seconds" required="no" >}}

The SQL connection timeout.

### tls

{{< confkey type="structure" structure="tls" required="no" >}}

If defined enables connecting to [PostgreSQL] over a TLS socket, and additionally controls the TLS connection
validation parameters.

[PostgreSQL]: https://www.postgresql.org/
