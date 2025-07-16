---
title: "authelia storage logs auth prune"
description: "Reference for the authelia storage logs auth prune command."
lead: ""
date: 2022-06-15T17:51:47+10:00
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

## authelia storage logs auth prune

Prune old authentication logs

### Synopsis

Prune authentication logs based on age criteria.

This command helps manage the authentication_logs table which grows indefinitely.
Use this to remove authentication records older than a specified period to
prevent excessive database growth and improve performance.

```
authelia storage logs auth prune [flags]
```

### Examples

```
authelia storage logs auth prune --older-than 90d
authelia storage logs auth prune --older-than 6m --batch-size 50000
authelia storage logs auth prune --older-than 1y --dry-run
```

### Options

```
      --batch-size int      Number of records to delete per batch (prevents long-running commands) (default 20000)
      --dry-run             Show what would be deleted without actually deleting
  -h, --help                help for prune
      --older-than string   Delete logs older than this duration (e.g., 90d, 6m, 1y)
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

* [authelia storage logs auth](authelia_storage_logs_auth.md)	 - Manage authentication logs

