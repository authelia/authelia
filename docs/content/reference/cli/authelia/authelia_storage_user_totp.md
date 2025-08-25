---
title: "authelia storage user totp"
description: "Reference for the authelia storage user totp command."
lead: ""
date: 2024-03-14T06:00:14+11:00
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

## authelia storage user totp

Manage TOTP configurations

### Synopsis

Manage TOTP configurations.

This subcommand allows deleting, exporting, and creating user TOTP configurations.

### Examples

```
authelia storage user totp --help
```

### Options

```
  -h, --help   help for totp
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

* [authelia storage user](authelia_storage_user.md)	 - Manages user settings
* [authelia storage user totp delete](authelia_storage_user_totp_delete.md)	 - Delete a TOTP configuration for a user
* [authelia storage user totp export](authelia_storage_user_totp_export.md)	 - Perform exports of the TOTP configurations
* [authelia storage user totp generate](authelia_storage_user_totp_generate.md)	 - Generate a TOTP configuration for a user
* [authelia storage user totp import](authelia_storage_user_totp_import.md)	 - Perform imports of the TOTP configurations

