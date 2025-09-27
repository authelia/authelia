---
title: "authelia storage"
description: "Reference for the authelia storage command."
lead: ""
date: 2025-08-01T16:23:47+10:00
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

## authelia storage

Manage the Authelia storage

### Synopsis

Manage the Authelia storage.

This subcommand has several methods to interact with the Authelia SQL Database. This allows doing several advanced
operations which would be much harder to do manually.


### Examples

```
authelia storage --help
```

### Options

```
      --encryption-key string      the storage encryption key to use
  -h, --help                       help for storage
      --mysql.address string       the MySQL server address (default "tcp://127.0.0.1:3306")
      --mysql.database string      the MySQL database name (default "authelia")
      --mysql.password string      the MySQL password
      --mysql.username string      the MySQL username (default "authelia")
      --postgres.address string    the PostgreSQL server address (default "tcp://127.0.0.1:5432")
      --postgres.database string   the PostgreSQL database name (default "authelia")
      --postgres.password string   the PostgreSQL password
      --postgres.schema string     the PostgreSQL schema name (default "public")
      --postgres.username string   the PostgreSQL username (default "authelia")
      --sqlite.path string         the SQLite database path
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia](authelia.md)	 - authelia untagged-unknown-dirty (master, unknown)
* [authelia storage bans](authelia_storage_bans.md)	 - Manages user and ip bans
* [authelia storage cache](authelia_storage_cache.md)	 - Manage storage cache
* [authelia storage encryption](authelia_storage_encryption.md)	 - Manage storage encryption
* [authelia storage logs](authelia_storage_logs.md)	 - Manage various types of logs
* [authelia storage migrate](authelia_storage_migrate.md)	 - Perform or list migrations
* [authelia storage schema-info](authelia_storage_schema-info.md)	 - Show the storage information
* [authelia storage user](authelia_storage_user.md)	 - Manages user settings

