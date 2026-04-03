---
title: "Database Integrations"
description: "A database integration reference guide"
summary: "This section contains a database integration reference guide for Authelia."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 320
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

We generally recommend using [PostgreSQL] for a database. If high availability is not a consideration we also support
[SQLite3].

It is also a general recommendation that if you're using [PostgreSQL], [MySQL], or [MariaDB]; that you do not
automatically upgrade the major/minor version of these databases, and pin the image tag so at most the patch version
is updated. For example for database version `x.y.z` only the `z` should change, `x` and `y` should remain the same.

It is also generally recommended that you do not rely on automatic update tools to perform this action
unless you are sure they shut down the container properly (i.e. with a graceful stop).

While this guide exists and it contains some guidance on managing the database being used, it is by no means exhaustive
or intended as such and users should refer to the database vendors documentation.

## PostgreSQL

The only current support criteria for [PostgreSQL] at present is that the version you're using is supported by the
[PostgreSQL] developers. See [Vendor Supported Versions](#vendor-supported-versions) more information.

We generally perform integration testing against the latest supported version of [PostgreSQL] and that is generally the
recommended version for new installations.

### Vendor Supported Versions

See the [PostgreSQL Versioning Policy](https://www.postgresql.org/support/versioning/) for information on the versions
and platforms that are currently supported by this vendor.

## MySQL

[MySQL] and [MariaDB] are both supported as part of the [MySQL] implementation. This is generally discouraged as
[PostgreSQL] is widely considered as a significantly better database engine. If you choose to go with [MySQL], we
recommend specifically using the [MariaDB] backend.

[MySQL] comes with some rigid support requirements in addition to the standard requirements for us supporting a third
party.

1. Must both support the `InnoDB` engine and this engine must be the default engine.
2. Must support the `utf8mb4` charset.
3. Must support the `utf8mb4_unicode_520_ci` collation.
4. Must support maximum index size of no less than 2048 bytes. The default maximum index size for the InnoDB engine is
   3072 bytes on:
    1. [MySQL] [8.0](https://dev.mysql.com/doc/refman/8.0/en/innodb-limits.html) or later.
    2. [MySQL] [5.7](https://dev.mysql.com/doc/refman/5.7/en/innodb-limits.html) or later provided:
       1. The [innodb_large_prefix](#innodb-large-prefixes) option is **_ON_**.
    3. [MariaDB] [10.3](https://mariadb.com/kb/en/innodb-system-variables/#innodb_large_prefix) or later.
5. Must support ANSI standard time behaviors. See [ANSI standard time behaviors](#ansi-standard-time-behaviors).

We generally perform integration testing against the latest supported version of [MySQL] and [MariaDB], and the latest
supported version of [MariaDB] is generally the recommended version for new installations.

### Specific Notes

#### InnoDB Large Prefixes

This can be configured in the [MySQL] configuration file by setting the `innodb_large_prefix` option to on.
According to the [Oracle] documentation this is the default behavior in
[MySQL] [5.7](https://dev.mysql.com/doc/refman/5.7/en/innodb-parameters.html#sysvar_innodb_large_prefix) and it can't be
turned off in [MySQL] [8.0](https://dev.mysql.com/doc/refman/8.0/en/innodb-limits.html) or in [MariaDB] 10.3 and later.

```cnf
[mysqld]
innodb_large_prefix = ON
```

#### ANSI standard time behaviors

This can be configured in the [MySQL] configuration file by setting the `explicit_defaults_for_timestamp` value to on.
According to the [Oracle] documentation this is the default behavior in
[MySQL] [5.7](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_explicit_defaults_for_timestamp)
and [MySQL] [8.0](https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_explicit_defaults_for_timestamp).
This is however not the default behavior in
[MariaDB](https://mariadb.com/kb/en/server-system-variables/#explicit_defaults_for_timestamp) before 10.10.

```cnf
[mysqld]
explicit_defaults_for_timestamp = ON
```

#### Upgrades

[MySQL] and [MariaDB] have several standard but important system databases named `mysql`, `sys`, and
`performance_schema`. These databases are outside the scope and not intended for individual applications to manage as
they are system databases used by [MySQL] and [MariaDB] internally.

These servers/engines may successfully start when these databases are incompatible with your particular [MySQL] or
[MariaDB] version, but may raise errors when you attempt to use particular features of the database. This may lead a
user to believe the server/engine is functioning correctly when it is in fact running with a potentially badly corrupted
schema.

The risk here is that the database may run for an extended period of time unnoticed and may be getting more and more
corrupt with no visible signs until it's no longer recoverable. This makes it critically important users do not neglect
this operation or ensure it's happening.

While some [MySQL] or [MariaDB] containers will do this automatically  or give users an option to perform this
automatically, it is strongly recommended that this process is manually done and only done **_after_** doing a backup of
all databases on the server as is the recommendation from both [MySQL] and [MariaDB].

It is your responsibility to ensure these tables are upgraded as per the
[mysql_upgrade](https://dev.mysql.com/doc/refman/8.0/en/mysql-upgrade.html) and
[mariadb_upgrade](https://mariadb.com/kb/en/mysql_upgrade/) documentation.

### Vendor Supported Versions

#### MariaDB Vendor Supported Versions

See the [MariaDB Server Releases](https://mariadb.com/kb/en/mariadb-server-release-dates/) for information on the
versions and platforms that are currently supported by this vendor.

#### MySQL Vendor Supported Versions

See the [MySQL Supported Platforms](https://www.mysql.com/support/supportedplatforms/database.html) for information on
the versions and platforms that are currently supported by this vendor.

[PostgreSQL]: https://www.postgresql.org/
[MySQL]: https://www.mysql.com/
[MariaDB]: https://mariadb.org/
[SQLite3]: https://www.sqlite.org/index.html
[Oracle]: https://www.oracle.com/
