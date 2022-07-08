---
title: "Environment"
description: "Using the Environment Variable Configuration Method."
lead: "Authelia has a layered configuration model. This section describes how to implement the environment configuration."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "methods"
weight: 101300
toc: true
---

Environment variables are applied after the configuration file meaning anything specified as part of the environment
overrides the configuration files.

*__Please Note:__ It is not possible to configure the access control rules section or OpenID Connect identity provider
clients section using environment variables at this time.*

## Prefix

The environment variables must be prefixed with `AUTHELIA_`. All environment variables that start with this prefix must
be for configuration. Any supplied environment variables that have this prefix and are not meant for configuration will
likely result in an error or even worse misconfiguration.

### Kubernetes

Please see the
[Kubernetes Integration: Enable Service Links](../../integration/kubernetes/introduction/index.md#enable-service-links)
documentation for specific requirements for using *Authelia* with Kubernetes.

## Mapping

Configuration options are mapped by their name. Levels of indentation / subkeys are replaced by underscores.

For example this YAML configuration:

```yaml
log:
  level: info
server:
  read_buffer_size: 4096
```

Can be replaced by this environment variable configuration:

```bash
AUTHELIA_LOG_LEVEL=info
AUTHELIA_SERVER_READ_BUFFER_SIZE=4096
```
