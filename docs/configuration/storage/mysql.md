---
layout: default
title: MySQL
parent: Storage backends
grand_parent: Configuration
nav_order: 2
---

# MySQL

The MySQL storage provider.

## Configuration

```yaml
storage:
  mysql:
    host: 127.0.0.1
    port: 3306
    database: authelia
    username: authelia
    password: mypassword
```

## Options

### host

The database server host.

If utilising an IPv6 literal address it must be enclosed by square brackets and quoted:
```yaml
host: "[fd00:1111:2222:3333::1]"
```

### port

The port the database server is listening on.

### database

The database name on the database server that the assigned [user](#username) has access to for the purpose of
**Authelia**.

### username

The username paired with the password used to connect to the database.

### password

The password paired with the username used to connect to the database. Can also be defined using a
[secret](../secrets.md) which is also the recommended way when running as a container.