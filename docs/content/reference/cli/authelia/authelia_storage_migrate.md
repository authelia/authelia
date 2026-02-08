---
title: "authelia storage migrate"
description: "Reference for the authelia storage migrate command."
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

## authelia storage migrate

Perform or list migrations

### Synopsis

Perform or list migrations.

This subcommand handles schema migration tasks.

### Examples

```
authelia storage migrate --help
```

### Options

```
  -h, --help   help for migrate
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

* [authelia storage](authelia_storage.md)	 - Manage the Authelia storage
* [authelia storage migrate down](authelia_storage_migrate_down.md)	 - Perform a migration down
* [authelia storage migrate history](authelia_storage_migrate_history.md)	 - Show migration history
* [authelia storage migrate list-down](authelia_storage_migrate_list-down.md)	 - List the down migrations available
* [authelia storage migrate list-up](authelia_storage_migrate_list-up.md)	 - List the up migrations available
* [authelia storage migrate up](authelia_storage_migrate_up.md)	 - Perform a migration up

