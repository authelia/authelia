---
title: "authelia storage bans user"
description: "Reference for the authelia storage bans user command."
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

## authelia storage bans user

Manages user bans

### Synopsis

Manages user bans.

This subcommand allows listing, creating, and revoking user bans from the regulation system.

### Examples

```
authelia storage bans user --help
```

### Options

```
  -h, --help   help for user
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

* [authelia storage bans](authelia_storage_bans.md)	 - Manages user and ip bans
* [authelia storage bans user add](authelia_storage_bans_user_add.md)	 - Adds user bans
* [authelia storage bans user list](authelia_storage_bans_user_list.md)	 - Lists user bans
* [authelia storage bans user revoke](authelia_storage_bans_user_revoke.md)	 - Revokes user bans

