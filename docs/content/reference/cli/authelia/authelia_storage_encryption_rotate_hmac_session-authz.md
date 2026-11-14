---
title: "authelia storage encryption rotate hmac session-authz"
description: "Reference for the authelia storage encryption rotate hmac session-authz command."
lead: ""
date: 2026-11-14T10:00:00+11:00
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

## authelia storage encryption rotate hmac session-authz

Rotate the session authz HMAC key

### Synopsis

Rotate the session authz HMAC key.

This subcommand allows rotation of the HMAC key used to derive the authz cookie of a session, which is the cookie the authorization endpoints accept. No table is truncated by this subcommand.

The key is only loaded when Authelia starts, so every running instance must be restarted for the rotation to take effect. Users are not signed out by this rotation as their sessions are unaffected, however after the restart every authz cookie issued before it is no longer accepted by the authorization endpoints, so the user is redirected to the portal which issues a new one.

```
authelia storage encryption rotate hmac session-authz [flags]
```

### Examples

```
authelia storage encryption rotate hmac session-authz --config config.yml
authelia storage encryption rotate hmac session-authz --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw
```

### Options

```
  -f, --force   force the rotation without confirmation
  -h, --help    help for session-authz
```

### Options inherited from parent commands

```
  -c, --config strings                                   configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.filters strings                           list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --config.filters.template.delimiter.left string    sets the left delimiter for the 'template' filter
      --config.filters.template.delimiter.right string   sets the right delimiter for the 'template' filter
      --config.filters.values strings                    file paths of values files (.yml, .yaml, .json, .toml) to utilize with configuration file filters; files are loaded in order with later files deep-merged on top, for more information run 'authelia -h authelia filters'
      --encryption-key string                            the storage encryption key to use
      --mysql.address string                             the MySQL server address (default "tcp://127.0.0.1:3306")
      --mysql.database string                            the MySQL database name (default "authelia")
      --mysql.password string                            the MySQL password
      --mysql.username string                            the MySQL username (default "authelia")
      --postgres.address string                          the PostgreSQL server address (default "tcp://127.0.0.1:5432")
      --postgres.database string                         the PostgreSQL database name (default "authelia")
      --postgres.password string                         the PostgreSQL password
      --postgres.schema string                           the PostgreSQL schema name (default "public")
      --postgres.username string                         the PostgreSQL username (default "authelia")
      --sqlite.path string                               the SQLite database path
```

### SEE ALSO

* [authelia storage encryption rotate hmac](authelia_storage_encryption_rotate_hmac.md)	 - Rotate HMAC keys

