---
title: "Statelessness"
description: "Statelessness is the ability for a system to operate without an in-memory state. A crash could result in loss of the in-memory state causing a bad user experience."
summary: "Statelessness is the ability for a system to operate without an in-memory state. A crash could result in loss of the in-memory state causing a bad user experience."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 390
toc: true
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

__Solution:__ Use a stateless provider, i.e. other than file (LDAP).

__Potential Workaround:__ You may be able to use the file provider in a highly available setup provided all features
which perform stateful actions related to the YAML file (like writing to it) are disabled and a solution to ensure the
file is properly distributed to all instances.

This features which perform stateful actions includes but is not limited to:

1. Changing Passwords.
2. Resetting Passwords.
3. Watching.

While this is theoretically supported in as much as it should work if you do everything correctly we do not officially
endorse or support for this architecture. We are also unlikely to provide direct tooling for this footgun in our
deployment technologies such as the helm chart due to the complications it may introduce.
