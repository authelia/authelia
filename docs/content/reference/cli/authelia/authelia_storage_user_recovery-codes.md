---
title: "authelia storage user recovery-codes"
description: "Reference for the authelia storage user recovery-codes command."
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

## authelia storage user recovery-codes

Manage user second-factor recovery codes

### Synopsis

Manage user second-factor recovery codes such as listing, generating a new batch, deleting all codes for a user, or showing per-user status.

### Options

```
  -h, --help   help for recovery-codes
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
* [authelia storage user recovery-codes delete](authelia_storage_user_recovery-codes_delete.md)	 - Soft-delete (revoke) all of a user's recovery codes
* [authelia storage user recovery-codes generate](authelia_storage_user_recovery-codes_generate.md)	 - Generate a new batch of recovery codes for a user
* [authelia storage user recovery-codes list](authelia_storage_user_recovery-codes_list.md)	 - List the recovery code rows for a user
* [authelia storage user recovery-codes status](authelia_storage_user_recovery-codes_status.md)	 - Show the recovery codes status for a user

