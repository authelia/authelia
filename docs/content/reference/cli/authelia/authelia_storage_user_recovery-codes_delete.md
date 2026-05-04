---
title: "authelia storage user recovery-codes delete"
description: "Reference for the authelia storage user recovery-codes delete command."
lead: ""
date: 2026-05-03T10:30:35+10:00
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

## authelia storage user recovery-codes delete

Soft-delete (revoke) all of a user's recovery codes

### Synopsis

Soft-delete (revoke) all of a user's recovery codes. The user will need to generate a new batch from the Two-Factor Authentication settings page or via the generate subcommand to use recovery codes again.

```
authelia storage user recovery-codes delete [username] [flags]
```

### Options

```
  -h, --help   help for delete
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

* [authelia storage user recovery-codes](authelia_storage_user_recovery-codes.md)	 - Manage user second-factor recovery codes

