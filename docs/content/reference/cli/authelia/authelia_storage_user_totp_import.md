---
title: "authelia storage user totp import"
description: "Reference for the authelia storage user totp import command."
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

## authelia storage user totp import

Perform imports of the TOTP configurations

### Synopsis

Perform imports of the TOTP configurations.

This subcommand allows importing TOTP configurations from the YAML format.

```
authelia storage user totp import <filename> [flags]
```

### Examples

```
authelia storage user totp import authelia.export.totp.yml
authelia storage user totp import --config config.yml authelia.export.totp.yml
authelia storage user totp import --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw authelia.export.totp.yml
```

### Options

```
  -h, --help   help for import
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

* [authelia storage user totp](authelia_storage_user_totp.md)	 - Manage TOTP configurations

