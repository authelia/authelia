---
title: "authelia storage encryption change-key"
description: "Reference for the authelia storage encryption change-key command."
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

## authelia storage encryption change-key

Changes the encryption key

### Synopsis

Changes the encryption key.

This subcommand allows you to change the encryption key of an Authelia SQL database.

```
authelia storage encryption change-key [flags]
```

### Examples

```
authelia storage encryption change-key --config config.yml --new-encryption-key 0e95cb49-5804-4ad9-be82-bb04a9ddecd8
authelia storage encryption change-key --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --new-encryption-key 0e95cb49-5804-4ad9-be82-bb04a9ddecd8 --postgres.address tcp://postgres:5432 --postgres.password autheliapw
```

### Options

```
  -h, --help                        help for change-key
      --new-encryption-key string   the new key to encrypt the data with
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

* [authelia storage encryption](authelia_storage_encryption.md)	 - Manage storage encryption

