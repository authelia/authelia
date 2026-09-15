---
title: "authelia storage clean oauth2"
description: "Reference for the authelia storage clean oauth2 command."
lead: ""
date: 2026-09-10T23:38:38+10:00
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

## authelia storage clean oauth2

Removes stale OpenID Connect 1.0 sessions

### Synopsis

Removes stale OpenID Connect 1.0 sessions.

A consent session is stale once it has expired and the user has already responded to it. Removing it also removes
the authorization code, access token, refresh token, PKCE, device code, and OpenID Connect sessions which reference
it. A session the user has never responded to is never removed as a flow may still be in progress against it.

```
authelia storage clean oauth2 [flags]
```

### Examples

```
authelia storage clean oauth2
authelia storage clean oauth2 --dry-run
authelia storage clean oauth2 --before '30 days'
```

### Options

```
      --before string   only remove sessions which expired and were responded to before this duration ago
      --dry-run         report what would be removed without removing anything
  -h, --help            help for oauth2
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

* [authelia storage clean](authelia_storage_clean.md)	 - Removes stale rows from the database

