---
title: "authelia storage user identifiers generate"
description: "Reference for the authelia storage user identifiers generate command."
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

## authelia storage user identifiers generate

Generate opaque identifiers in bulk

### Synopsis

Generate opaque identifiers in bulk.

This subcommand allows various options for generating the opaque identifies for users in bulk.

```
authelia storage user identifiers generate [flags]
```

### Examples

```
authelia storage user identifiers generate --users john,mary
authelia storage user identifiers generate --users john,mary --services openid
authelia storage user identifiers generate --users john,mary --services openid --sectors=",example.com,test.com"
authelia storage user identifiers generate --users john,mary --services openid --sectors=",example.com,test.com" --config config.yml
authelia storage user identifiers generate --users john,mary --services openid --sectors=",example.com,test.com" --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw
```

### Options

```
  -h, --help               help for generate
      --sectors strings    The list of sectors to generate identifiers for
      --services strings   The list of services to generate the opaque identifiers for, valid values are: openid (default [openid])
      --users strings      The list of users to generate the opaque identifiers for
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

* [authelia storage user identifiers](authelia_storage_user_identifiers.md)	 - Manage user opaque identifiers

