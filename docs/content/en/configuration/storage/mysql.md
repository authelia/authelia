---
title: "MySQL"
description: "MySQL Configuration"
lead: "The MySQL storage provider which supports both MySQL and MariaDB."
date: 2022-03-20T12:52:27+11:00
lastmod: 2022-06-03T10:43:55+10:00
draft: false
images: []
menu:
  configuration:
    parent: "storage"
weight: 107600
toc: true
---

## Version support

When using MySQL or MariaDB we recommend using the latest version that is officially supported by the MySQL or MariaDB
developers. We also suggest checking out [PostgreSQL](postgres.md) as an alternative.

The oldest versions that have been tested are MySQL 5.7 and MariaDB 10.6.

If using MySQL 5.7 or MariaDB 10.6 you may be required to adjust the `explicit_defaults_for_timestamp` setting. This
will be evident when the container starts with an error similar to `Error 1067: Invalid default value for 'exp'`. You
can adjust this setting in the mysql.cnf file like so:

```cnf
[mysqld]
explicit_defaults_for_timestamp = 1
```

## Configuration

```yaml
storage:
  encryption_key: a_very_important_secret
  mysql:
    host: 127.0.0.1
    port: 3306
    database: authelia
    username: authelia
    password: mypassword
    timeout: 5s
```

## Options

### encryption_key

See the [encryption_key docs](introduction.md#encryption_key).

### host

{{< confkey type="string" default="localhost" required="no" >}}

The database server host.

If utilising an IPv6 literal address it must be enclosed by square brackets and quoted:

```yaml
host: "[fd00:1111:2222:3333::1]"
```

### port

{{< confkey type="integer" default="3306" required="no" >}}

The port the database server is listening on.

### database

{{< confkey type="string" required="yes" >}}

The database name on the database server that the assigned [user](#username) has access to for the purpose of
__Authelia__.

### username

{{< confkey type="string" required="yes" >}}

The username paired with the password used to connect to the database.

### password

{{< confkey type="string" required="yes" >}}

The password paired with the username used to connect to the database. Can also be defined using a
[secret](../methods/secrets.md) which is also the recommended way when running as a container.

We recommend generating a random string with 64 characters or more for this purposes which can be done by following the
[Generating a Random Alphanumeric String](../miscellaneous/guides.md#generating-a-random-alphanumeric-string)
guide.

### timeout

{{< confkey type="duration" default="5s" required="no" >}}

The SQL connection timeout.
