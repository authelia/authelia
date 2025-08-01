---
title: "authelia storage cache mds3"
description: "Reference for the authelia storage cache mds3 command."
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

## authelia storage cache mds3

Manage WebAuthn MDS3 cache storage

### Synopsis

Manage WebAuthn MDS3 cache storage.

This subcommand allows management of the WebAuthn MDS3 cache storage.

### Examples

```
authelia storage cache mds3 --help
```

### Options

```
  -h, --help   help for mds3
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

* [authelia storage cache](authelia_storage_cache.md)	 - Manage storage cache
* [authelia storage cache mds3 delete](authelia_storage_cache_mds3_delete.md)	 - Delete WebAuthn MDS3 cache storage
* [authelia storage cache mds3 dump](authelia_storage_cache_mds3_dump.md)	 - Dump WebAuthn MDS3 cache storage
* [authelia storage cache mds3 status](authelia_storage_cache_mds3_status.md)	 - View WebAuthn MDS3 cache storage status
* [authelia storage cache mds3 update](authelia_storage_cache_mds3_update.md)	 - Update WebAuthn MDS3 cache storage

