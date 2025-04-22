
---
title: "Microsoft SQL"
description: "Microsoft SQL Configuration"
summary: "The Microsoft SQL storage provider which supports Microsoft SQL Server."
date: 2025-04-20T10:12:09+10:00
draft: false
images: []
weight: 107600
toc: true
aliases: []
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Version support

See the [Microsoft SQL Server Database Integration](../../reference/integrations/database-integrations.md#microsoft-sql-server) reference
guide for supported version information.

## Configuration

{{< config-alert-example >}}

```yaml
storage:
  encryption_key: 'a_very_important_secret'
  mssql:
    address: 'tcp://127.0.0.1:3306'
    database: 'authelia'
    username: 'authelia'
    password: 'mypassword'
    instance: ''
    timeout: '5s'
    tls:
      server_name: 'mssql.example.com'
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
        -----BEGIN PRIVATE KEY-----
        ...
        -----END PRIVATE KEY-----
```

## Options

This section describes the individual configuration options.

### encryption_key

See the [encryption_key docs](introduction.md#encryption_key).

### address

{{< confkey type="string" syntax="address" required="yes" default="tcp://127.0.0.1:1433" >}}

Configures the address for the Microsoft SQL Server. The address itself is a connector and the scheme must either be
the `unix` scheme or one of the `tcp` schemes.

__Examples:__

```yaml
storage:
  mssql:
    address: 'tcp://127.0.0.1:3306'
```

```yaml
storage:
  mssql:
    address: 'tcp://[fd00:1111:2222:3333::1]:3306'
```

```yaml
storage:
  mssql:
    address: 'unix:///var/run/mssqld.sock'
```

### database

{{< confkey type="string" required="yes" >}}

The database name on the database server that the assigned [user](#username) has access to for the purpose of
__Authelia__.

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

### instance

{{< confkey type="string" default="" >}}

The instance name to connect to.

### schema

{{< confkey type="string" default="" >}}

The schema name to utilize.

### timeout

{{< confkey type="string,integer" syntax="duration" default="5 seconds" required="no" >}}

The SQL connection timeout.

### tls

{{< confkey type="structure" structure="tls" required="no" >}}

If defined enables connecting to [Microsoft SQL] over a TLS socket, and additionally controls the TLS connection
validation parameters.

[MySQL]: https://www.mysql.com/
[MariaDB]: https://mariadb.org/
