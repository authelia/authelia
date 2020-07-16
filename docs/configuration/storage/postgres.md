---
layout: default
title: PostgreSQL
parent: Storage backends
grand_parent: Configuration
nav_order: 3
---

# PostgreSQL

```yaml
storage:
  postgres:
    host: 127.0.0.1
    port: 5432
    database: authelia
    username: authelia
    # Password can also be set using a secret: https://docs.authelia.com/configuration/secrets.html
    password: mypassword
    sslmode: disable
```

## SSL Mode

SSL mode configures how to handle SSL connections with Postgres. 
Valid options are 'disable', 'require', 'verify-ca', or 'verify-full'.
See the [PostgreSQL Documentation](https://www.postgresql.org/docs/12/libpq-ssl.html)
or [pgx - PostgreSQL Driver and Toolkit Documentation](https://pkg.go.dev/github.com/jackc/pgx?tab=doc) 
for more information.

## Loading a password from a secret instead of inside the configuration

Password can also be defined using a [secret](../secrets.md).