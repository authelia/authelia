---
title: "Network"
description: "Network Definitions Configuration"
summary: "Authelia allows configuring reusable network definitions."
date: 2024-10-16T22:30:41+11:00
draft: false
images: []
weight: 199100
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The network section configures named network lists.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
definitions:
  network:
    network_name:
      - '192.168.1.0/24'
      - '192.168.2.20'
```

## Options

This section describes the individual configuration options. The configuration for this section is incredibly basic,
effectively it's key value pairs, where the key is the name used elsewhere in the configuration and the value is a list
of network addresses.

These definitions are used as [Access Control Networks](../security/access-control.md#networks) and
[OpenID Connect 1.0 Authorization Policy Networks](../identity-providers/openid-connect/provider.md#networks).

### key

The key is the name of the policy. In the example above the key is `network_name` and is the value which must be used
in other areas of the configuration to reference it.

### value

{{< confkey type="string" syntax="network" required="yes" >}}

The values which represent the CIDR notation of the IP's this definition applies to. In the example the value is a list
which contains `192.168.1.0/24` and `192.168.2.20`.
