---
layout: default
title: Deployment - Lite
parent: Deployment
---

# Lite Deployment

**Authelia** can be deployed as a lite setup not requiring any SQL server,
Redis cluster or LDAP server. In some cases, like protecting personal projects/websites,
it can be fine to use that setup but beware that this setup is non-resilient to failures
so it should be used at your own risk.

The setup is called lite since it reduces the number of components in the architecture to
only two: a reverse proxy such as Nginx, Traefik or HAProxy and Authelia.

## Reverse Proxy

Documentation for deploying a reverse proxy collaborating with Authelia is available
[here](./supported-proxies/index).

## Discard components

### Discard SQL server

It's possible to use a SQLite file instead of a SQL server as documented
[here](../configuration/storage/sqlite).

### Discard Redis

Connection details to Redis are optional. If not provided, sessions will
be stored in memory instead. This has the inconvenient of logging out users
every time Authelia restarts.

The documentation about session management is available
[here](../configuration/session).


### Discard LDAP

**Authelia** can use a file backend in order to store users instead of a
LDAP server or an Active Directory.

To use a file backend instead of a LDAP server, please follow the related
documentation [here](../configuration/authentication/file).

## FAQ

### Can you give more details on why this is not suitable for production environments?

This documentation gives instructions that will make **Authelia** non
resilient to failures and non scalable by preventing you from running multiple
instances of the application. This means that **Authelia** won't be able to distribute
the load across multiple servers and it will prevent failover in case of a
crash or an hardware issue. Moreover, users will be logged out every time
Authelia restarts.

### Why aren't all those steps automated?

We would really be more than happy to review any contribution with an Ansible playbook,
a Chef cookbook or whatever else to automate the process.
