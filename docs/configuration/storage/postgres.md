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
    schema: public
    username: authelia
    password: mypassword
    ssl:
      mode: disable
      root_certificate: /path/to/root_cert.pem
      certificate: /path/to/cert.pem
      key: /path/to/key.pem
```

## Options

### encryption_key
See the [encryption_key docs](./index.md#encryption_key).

### host
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
required: yes
{: .label .label-config .label-red }
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
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: yes
{: .label .label-config .label-red }
</div>

The database name on the database server that the assigned [user](#username) has access to for the purpose of
**Authelia**.

### schema
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: public
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The database schema name to use on the database server that the assigned [user](#username) has access to for the purpose
of **Authelia**. By default this is the public schema.

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

### ssl

#### mode
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

#### root_certificate
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: no
{: .label .label-config .label-green }
</div>

The optional location of the root certificate file encoded in the PEM format for validation purposes.

#### certificate
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: no
{: .label .label-config .label-green }
</div>

The optional location of the certificate file encoded in the PEM format for validation purposes.

#### key
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: no
{: .label .label-config .label-green }
</div>

The optional location of the key file encoded in the PEM format for authentication purposes.
