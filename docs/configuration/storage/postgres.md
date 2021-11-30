---
layout: default
title: PostgreSQL
parent: Storage Backends
grand_parent: Configuration
nav_order: 3
---

# PostgreSQL

The PostgreSQL storage provider.

## Configuration

```yaml
storage:
  encryption_key: a_very_important_secret
  postgres:
    host: 127.0.0.1
    port: 5432
    database: authelia
    username: authelia
    password: mypassword
    sslmode: disable
```

## Options

### encryption_key
See the [encryption_key docs](./index.md#encryption_key).

### host
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: localhost
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The database server host.

If utilising an IPv6 literal address it must be enclosed by square brackets and quoted:
```yaml
host: "[fd00:1111:2222:3333::1]"
```

### port
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: 5432
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The port the database server is listening on.

### database

The database name on the database server that the assigned [user](#username) has access to for the purpose of
**Authelia**.

### username
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: yes
{: .label .label-config .label-red }
</div>

The username paired with the password used to connect to the database.

### password
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: yes
{: .label .label-config .label-red }
</div>

The password paired with the username used to connect to the database. Can also be defined using a
[secret](../secrets.md) which is also the recommended way when running as a container.

### timeout
<div markdown="1">
type: duration
{: .label .label-config .label-purple }
default: 5s
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The SQL connection timeout.

### sslmode
<div markdown="1">
type: string
{: .label .label-config .label-purple }
default: disable
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

SSL mode configures how to handle SSL connections with Postgres.
Valid options are 'disable', 'require', 'verify-ca', or 'verify-full'.
See the [PostgreSQL Documentation](https://www.postgresql.org/docs/12/libpq-ssl.html)
or [pgx - PostgreSQL Driver and Toolkit Documentation](https://pkg.go.dev/github.com/jackc/pgx?tab=doc)
for more information.
