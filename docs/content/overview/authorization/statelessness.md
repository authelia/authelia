---
title: "Statelessness"
description: "Statelessness is the ability for a system to operate without an in-memory state. A crash could result in loss of the in-memory state causing a bad user experience."
summary: "Statelessness is the ability for a system to operate without an in-memory state. A crash could result in loss of the in-memory state causing a bad user experience."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 390
toc: false
aliases:
  - /t/statelessness
  - /docs/features/statelessness.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__ supports operating as a stateless application. This is incredibly important when running in highly
available deployments like you may see in platforms like Kubernetes.

## Stateful Considerations

There are some components within __Authelia__ that may optionally be made stateful by using certain providers. Examples
of this are as follows:

### Session Provider

__Severity:__ *BREAKING*.

__Solution:__ Use a session provider other than memory (Redis).

If you do not configure an external provider for the session configuration
it stores the session in memory. This is unacceptable for the operation of
__Authelia__ and is thus not supported for high availability.

### Storage Provider

__Severity:__ *BREAKING*.

__Solution:__ Use a storage provider other than SQLite3 (MySQL, MariaDB, PostgreSQL).

Use of the local storage provider (SQLite3) is not supported in high availability setups
due to a design limitation with how SQLite3 operates. Use any of the other storage providers.

### Notification Provider

__Severity:__ *HIGH*.

__Solution:__ Use a notification provider other than file system (SMTP).

Use of the file system notification provider prevents users from several key tasks which heavily impact usability of
the system, and technically reduce security. Users will be unable to reset passwords or register new 2FA devices on
their own. The file system provider is not supported for high availability.

### Authentication Provider

__Severity:__ *MEDIUM (limiting)*.

__Solution:__ Use an authentication provider other than file (LDAP), or distribute the file and disable password reset.

Use of the file authentication provider (YAML) is only partially supported with high availability setups. It's
recommended if you don't use a stateless provider that you disable password reset and make sure the file is distributed
to all instances. We do not support using the file type in these scenarios.
