---
layout: default
title: Migration
parent: Configuration
nav_order: 6
---

This section documents changes in the configuration which may require manual migration by the administrator. Typically
this only occurs when a configuration key is renamed or moved to a more appropriate location.

## Format

The migrations are formatted in a table with the old key and the new key. Periods indicate a different section which can
be represented in YAML as a dictionary i.e. it's indented.

In our table `server.host` with a value of `0.0.0.0` is represented in YAML like this:

```yaml
server:
  host: 0.0.0.0
```

## Policy
Our deprecation policy for configuration keys is 3 minor versions. For example if a configuration option is deprecated
in version 4.30.0, it will remain as a warning for 4.30.x, 4.31.x, and 4.32.x; then it will become a fatal error in
4.33.0+. 

## Migrations

### 4.33.0
The options deprecated in version [4.30.0](#4300) have been fully removed as per our deprecation policy and warnings
logged for users.

### 4.30.0
The following changes occurred in 4.30.0:

|Previous Key |New Key               |
|:-----------:|:--------------------:|
|host         |server.host           |
|port         |server.port           |
|tls_key      |server.tls.key        |
|tls_cert     |server.tls.certificate|
|log_level    |log.level             |
|log_file_path|log.file_path         |
|log_format   |log.format            |

_**Please Note:** you can no longer define secrets for providers that you are not using. For example if you're using the 
[filesystem notifier](./notifier/filesystem.md) you must ensure that the `AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE` 
environment variable or other environment variables set. This also applies to other providers like 
[storage](./storage/index.md) and [authentication backend](./authentication/index.md)._

#### Kubernetes 4.30.0

_**Please Note:** if you're using Authelia with Kubernetes and are not using the provided [helm chart](https://charts.authelia.com)
you will be required to set the following option in your PodSpec. Keeping in mind this example is for a Pod, not for
a Deployment, StatefulSet, or DaemonSet; you will need to adapt the `enableServiceLinks` option to fit into the relevant
location depending on your needs._

```yaml
---
apiVersion: v1
kind: Pod
metadata:
  name: authelia
spec:
  enableServiceLinks: false
...
```

### 4.25.0

The following changes occurred in 4.25.0:

|Previous Key                                   |New Key                                        |
|:---------------------------------------------:|:---------------------------------------------:|
|authentication_backend.ldap.tls.skip_verify    |authentication_backend.ldap.tls.skip_verify    |
|authentication_backend.ldap.minimum_tls_version|authentication_backend.ldap.tls.minimum_version|
|notifier.smtp.disable_verify_cert              |notifier.smtp.tls.skip_verify                  |
|notifier.smtp.trusted_cert                     |certificates_directory                         |

_**Please Note:** `certificates_directory` is not a direct replacement for the `notifier.smtp.trusted_cert`, instead
of being the path to a specific file it is a path to a directory containing certificates trusted by Authelia. This
affects other services like LDAP as well._

### 4.7.0

The following changes occurred in 4.7.0:

|Previous Key|New Key  |
|:----------:|:-------:|
|logs_level  |log_level|
|logs_file   |log_file |

_**Please Note:** The new keys also changed in [4.30.0](#4.30.0) so you will need to update them to the new values if you
are using [4.30.0](#4.30.0) or newer instead of the new keys listed here._
