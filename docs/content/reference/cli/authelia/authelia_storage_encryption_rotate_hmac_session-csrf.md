---
title: "authelia storage encryption rotate hmac session-csrf"
description: "Reference for the authelia storage encryption rotate hmac session-csrf command."
lead: ""
date: 2026-09-16T10:00:00+10:00
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

## authelia storage encryption rotate hmac session-csrf

Rotate the session CSRF HMAC key

### Synopsis

Rotate the session CSRF HMAC key.

This subcommand allows rotation of the HMAC key used to derive the CSRF token of a session from the CSRF secret held within it. No table is truncated by this subcommand.

The key is only loaded when Authelia starts, so every running instance must be restarted for the rotation to take effect. Users are not signed out by this rotation as their sessions are unaffected, however after the restart every CSRF token issued before it is rejected until the user agent obtains a new one, which occurs the next time it loads the portal.

```
authelia storage encryption rotate hmac session-csrf [flags]
```

### Examples

```
authelia storage encryption rotate hmac session-csrf --config config.yml
authelia storage encryption rotate hmac session-csrf --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw
```

### Options

```
  -f, --force   force the rotation without confirmation
  -h, --help    help for session-csrf
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

