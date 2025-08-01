---
title: "authelia storage user webauthn delete"
description: "Reference for the authelia storage user webauthn delete command."
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

## authelia storage user webauthn delete

Delete a WebAuthn credential

### Synopsis

Delete a WebAuthn credential.

This subcommand allows deleting a WebAuthn credential directly from the database.

```
authelia storage user webauthn delete [username] [flags]
```

### Examples

```
authelia storage user webauthn delete john --all
authelia storage user webauthn delete john --all --config config.yml
authelia storage user webauthn delete john --all --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw
authelia storage user webauthn delete john --description Primary
authelia storage user webauthn delete john --description Primary --config config.yml
authelia storage user webauthn delete john --description Primary --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw
authelia storage user webauthn delete --kid abc123
authelia storage user webauthn delete --kid abc123 --config config.yml
authelia storage user webauthn delete --kid abc123 --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw
```

### Options

```
      --all                  delete all of the users WebAuthn credentials
      --description string   delete a users WebAuthn credential by description
  -h, --help                 help for delete
      --kid string           delete a users WebAuthn credential by key id
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

