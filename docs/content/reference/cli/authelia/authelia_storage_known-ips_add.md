---
title: "authelia storage known-ips add"
description: "Reference for the authelia storage known-ips add command."
lead: ""
date: 2026-04-02T15:48:21+11:00
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

## authelia storage known-ips add

Add a known IP address

### Synopsis

Add a known IP address.

This subcommand allows manually adding a known IP address for a user to the database.

```
authelia storage known-ips add <username> <ip> [flags]
```

### Examples

```
authelia storage known-ips add john 203.0.113.10
authelia storage known-ips add john 203.0.113.10 --expires 1h
authelia storage known-ips add john 203.0.113.10 --never-expires
authelia storage known-ips add john 203.0.113.10 --user-agent 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36'
```

### Options

```
      --expires duration    the duration after which the known IP address expires, relative to now (defaults to the configured default lifespan)
  -h, --help                help for add
      --never-expires       the known IP address never expires
      --user-agent string   the raw user agent string associated with the known IP address
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

* [authelia storage known-ips](authelia_storage_known-ips.md)	 - Manages known IP addresses

