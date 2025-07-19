---
title: "authelia storage bans user add"
description: "Reference for the authelia storage bans user add command."
lead: ""
date: 2025-02-23T22:10:30+11:00
draft: false
images: []
weight: 905
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## authelia storage bans user add

Adds user bans

### Synopsis

Adds user bans.

This subcommand allows adding user bans to the regulation system.

```
authelia storage bans user add <user> [flags]
```

### Examples

```
authelia storage bans user add --help
```

### Options

```
  -d, --duration string   the duration for the ban (default "1 day")
  -h, --help              help for add
  -p, --permanent         makes the ban effectively permanent
  -r, --reason string     includes a reason for the ban
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --encryption-key string                 the storage encryption key to use
      --mysql.address string                  the MySQL server address (default "tcp://127.0.0.1:3306")
      --mysql.database string                 the MySQL database name (default "authelia")
      --mysql.password string                 the MySQL password
      --mysql.username string                 the MySQL username (default "authelia")
      --postgres.address string               the PostgreSQL server address (default "tcp://127.0.0.1:5432")
      --postgres.database string              the PostgreSQL database name (default "authelia")
      --postgres.password string              the PostgreSQL password
      --postgres.schema string                the PostgreSQL schema name (default "public")
      --postgres.username string              the PostgreSQL username (default "authelia")
      --sqlite.path string                    the SQLite database path
```

### SEE ALSO

* [authelia storage bans user](authelia_storage_bans_user.md)	 - Manages user bans

