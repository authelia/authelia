---
title: "Environment"
description: "Using the Environment Variable Configuration Method."
summary: "Authelia has a layered configuration model. This section describes how to implement the environment configuration."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 101300
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Environment variables are applied after the configuration file meaning anything specified as part of the environment
overrides the configuration files.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
It is not possible to configure several sections using environment variables or secrets. The sections affected are all
lists of objects. These include but may not be limited to the rules section in access control, the clients section in
the OpenID Connect 1.0 Provider, the cookies section of in session, and the authz section in the server endpoints. See
[ADR2](../../reference/architecture-decision-log/2.md) for more information.
{{< /callout >}}

## Prefix

The environment variables must be prefixed with `AUTHELIA_`. All environment variables that start with this prefix must
be for configuration. Any supplied environment variables that have this prefix and are not meant for configuration will
likely result in an error or even worse misconfiguration.

### Kubernetes

Please see the
[Kubernetes Integration: Enable Service Links](../../integration/kubernetes/introduction.md#enable-service-links)
documentation for specific requirements for using *Authelia* with Kubernetes.

## Mapping

Configuration options are mapped by their name. Levels of indentation / subkeys are replaced by underscores.

For example this YAML configuration:

```yaml {title="configuration.yml"}
log:
  level: 'info'
server:
  buffers:
    read: 4096
```

Can be replaced by this environment variable configuration:

```bash
AUTHELIA_LOG_LEVEL=info
AUTHELIA_SERVER_BUFFERS_READ=4096
```

## Environment Variables

{{% table-config-keys secrets="false" %}}
