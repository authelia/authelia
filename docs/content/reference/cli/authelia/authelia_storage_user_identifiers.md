---
title: "authelia storage user identifiers"
description: "Reference for the authelia storage user identifiers command."
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

## authelia storage user identifiers

Manage user opaque identifiers

### Synopsis

Manage user opaque identifiers.

This subcommand allows performing various tasks related to the opaque identifiers for users.

### Examples

```
authelia storage user identifiers --help
```

### Options

```
  -h, --help   help for identifiers
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

* [authelia storage user](authelia_storage_user.md)	 - Manages user settings
* [authelia storage user identifiers add](authelia_storage_user_identifiers_add.md)	 - Add an opaque identifier for a user to the database
* [authelia storage user identifiers export](authelia_storage_user_identifiers_export.md)	 - Export the identifiers to a YAML file
* [authelia storage user identifiers generate](authelia_storage_user_identifiers_generate.md)	 - Generate opaque identifiers in bulk
* [authelia storage user identifiers import](authelia_storage_user_identifiers_import.md)	 - Import the identifiers from a YAML file

