---
title: "authelia storage user webauthn export"
description: "Reference for the authelia storage user webauthn export command."
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

## authelia storage user webauthn export

Perform exports of the WebAuthn credentials

### Synopsis

Perform exports of the WebAuthn credentials.

This subcommand allows exporting WebAuthn credentials to various formats.

```
authelia storage user webauthn export [flags]
```

### Examples

```
authelia storage user webauthn export
authelia storage user webauthn export --file authelia.export.webauthn.yml
authelia storage user webauthn export --config config.yml
authelia storage user webauthn export--encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw
```

### Options

```
  -f, --file string   The file name for the YAML export (default "authelia.export.webauthn.yml")
  -h, --help          help for export
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

* [authelia storage user webauthn](authelia_storage_user_webauthn.md)	 - Manage WebAuthn credentials

