---
layout: default
title: Statelessness
parent: Features
nav_order: 7
---

# Statelessness

**Authelia** supports operating as a stateless application. This is incredibly important
when running in highly available deployments like you may see in platforms like Kubernetes.

## Stateful Considerations

There are some components within **Authelia** that may optionally be made stateful by using
certain providers. Examples of this are as follows:

### Session Provider

**Severity:** *BREAKING*.

**Solution:** Use a session provider other than memory (Redis).

If you do not configure an external provider for the session configuration
it stores the session in memory. This is unacceptable for the operation of
**Authelia** and is thus not supported for high availability.


### Storage Provider

**Severity:** *BREAKING*.

**Solution:** Use a session provider other than SQLite3 (MySQL, MariaDB, PostgreSQL).

Use of the local storage provider (SQLite3) is not supported in high availability setups
due to a design limitation with how SQLite3 operates. Use any of the other storage providers.


### Notification Provider

**Severity:** *HIGH*.

**Solution:** Use a notification provider other than file system (SMTP).

Use of the file system notification provider prevents users from several key tasks which heavily impact usability of
the system, and technically reduce security. Users will be unable to reset passwords or register new 2FA devices on
their own. The file system provider is not supported for high availability. 

### Authentication Provider

**Severity:** *MEDIUM (limiting)*.

**Solution:** Use an authentication provider other than file (LDAP), or distribute the file and disable password reset.

Use of the file authentication provider (YAML) is only partially supported with high availability setups. It's 
recommended if you don't use a stateless provider that you disable password reset and make sure the file is distributed 
to all instances. We do not support using the file type in these scenarios.