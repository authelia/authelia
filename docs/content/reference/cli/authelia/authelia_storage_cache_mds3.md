---
title: "authelia storage cache mds3"
description: "Reference for the authelia storage cache mds3 command."
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
  -c, --config strings                         configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings    list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --encryption-key string                  the storage encryption key to use
      --mysql.database string                  the MySQL database name (default "authelia")
      --mysql.host string                      the MySQL hostname
      --mysql.password string                  the MySQL password
      --mysql.port int                         the MySQL port (default 3306)
      --mysql.username string                  the MySQL username (default "authelia")
      --postgres.database string               the PostgreSQL database name (default "authelia")
      --postgres.host string                   the PostgreSQL hostname
      --postgres.password string               the PostgreSQL password
      --postgres.port int                      the PostgreSQL port (default 5432)
      --postgres.schema string                 the PostgreSQL schema name (default "public")
      --postgres.ssl.certificate string        the PostgreSQL ssl certificate file location
      --postgres.ssl.key string                the PostgreSQL ssl key file location
      --postgres.ssl.mode string               the PostgreSQL ssl mode (default "disable")
      --postgres.ssl.root_certificate string   the PostgreSQL ssl root certificate file location
      --postgres.username string               the PostgreSQL username (default "authelia")
      --sqlite.path string                     the SQLite database path
```

### SEE ALSO

* [authelia storage cache](authelia_storage_cache.md)	 - Manage storage cache
* [authelia storage cache mds3 delete](authelia_storage_cache_mds3_delete.md)	 - Delete WebAuthn MDS3 cache storage
* [authelia storage cache mds3 dump](authelia_storage_cache_mds3_dump.md)	 - Dump WebAuthn MDS3 cache storage
* [authelia storage cache mds3 status](authelia_storage_cache_mds3_status.md)	 - View WebAuthn MDS3 cache storage status
* [authelia storage cache mds3 update](authelia_storage_cache_mds3_update.md)	 - Update WebAuthn MDS3 cache storage

