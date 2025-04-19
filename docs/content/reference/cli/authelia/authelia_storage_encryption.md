---
title: "authelia storage encryption"
description: "Reference for the authelia storage encryption command."
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

## authelia storage encryption

Manage storage encryption

### Synopsis

Manage storage encryption.

This subcommand allows management of the storage encryption.

### Examples

```
authelia storage encryption --help
```

### Options

```
  -h, --help   help for encryption
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --encryption-key string                 the storage encryption key to use
      --mssql.address string                  the MSSQL address (default "tcp://127.0.0.1:1433")
      --mssql.database string                 the MSSQL database name (default "authelia")
      --mssql.instance string                 the MSSQL instance name
      --mssql.password string                 the MSSQL password
      --mssql.schema string                   the MSSQL schema
      --mssql.username string                 the MSSQL username (default "authelia")
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
* [authelia storage encryption change-key](authelia_storage_encryption_change-key.md)	 - Changes the encryption key
* [authelia storage encryption check](authelia_storage_encryption_check.md)	 - Checks the encryption key against the database data

