---
title: "authelia storage bans ip"
description: "Reference for the authelia storage bans ip command."
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

## authelia storage bans ip

Manages ip bans

### Synopsis

Manages ip bans.

This subcommand allows listing, creating, and revoking ip bans from the regulation system.

### Examples

```
authelia storage bans ip --help
```

### Options

```
  -h, --help   help for ip
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

* [authelia storage bans](authelia_storage_bans.md)	 - Manages user and ip bans
* [authelia storage bans ip add](authelia_storage_bans_ip_add.md)	 - Adds ip bans
* [authelia storage bans ip list](authelia_storage_bans_ip_list.md)	 - Lists ip bans
* [authelia storage bans ip revoke](authelia_storage_bans_ip_revoke.md)	 - Revokes ip bans

