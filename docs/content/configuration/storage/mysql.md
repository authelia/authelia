---
title: "MySQL"
description: "MySQL Configuration"
summary: "The MySQL storage provider which supports both MySQL and MariaDB."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 107600
toc: true
aliases:
  - /docs/configuration/storage/mariadb.html
  - /docs/configuration/storage/mysql.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Version support

See the [MySQL Database Integration](../../reference/integrations/database-integrations.md#mysql) reference
guide for supported version information.

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
storage:
  encryption_key: 'a_very_important_secret'
  mysql:
    address: 'tcp://127.0.0.1:3306'
    database: 'authelia'
    username: 'authelia'
    password: 'mypassword'
    timeout: '5s'
    tls:
      server_name: 'mysql.{{< sitevar name="domain" nojs="example.com" >}}'
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

{{< confkey type="string" syntax="address" required="yes" >}}

Configures the address for the MySQL/MariaDB Server. The address itself is a connector and the scheme must either be
the `unix` scheme or one of the `tcp` schemes.

__Examples:__

```yaml {title="configuration.yml"}
storage:
  mysql:
    address: 'tcp://127.0.0.1:3306'
```

```yaml {title="configuration.yml"}
storage:
  mysql:
    address: 'tcp://[fd00:1111:2222:3333::1]:3306'
```

```yaml {title="configuration.yml"}
storage:
  mysql:
    address: 'unix:///var/run/mysqld.sock'
```

### database

{{< confkey type="string" required="yes" >}}

The database name on the database server that the assigned [user](#username) has access to for the purpose of
__Authelia__.

### username

{{< confkey type="string" required="yes" >}}

The username paired with the password used to connect to the database.

### password

{{< confkey type="string" required="situational" secret="yes" >}}

The password paired with the [username](#username) used to connect to the database.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters and the user password is changed to this value.

### timeout

{{< confkey type="string,integer" syntax="duration" default="5 seconds" required="no" >}}

The SQL connection timeout.

### tls

{{< confkey type="structure" structure="tls" required="no" >}}

If defined enables connecting over a TLS socket and additionally controls the TLS connection
verification parameters for the [MySQL] or [MariaDB] server.

By default Authelia uses the system certificate trust for TLS certificate verification of TLS connections and the
[certificates_directory](../miscellaneous/introduction.md#certificates_directory) global option can be used to augment
this.

[MySQL]: https://www.mysql.com/
[MariaDB]: https://mariadb.org/
