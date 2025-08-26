---
title: "authelia debug oidc claims"
description: "Reference for the authelia debug oidc claims command."
lead: ""
date: 2025-08-01T16:23:47+10:00
draft: false
images: []
weight: 905
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## authelia debug oidc claims

Perform a OpenID Connect 1.0 claims hydration debug operation

### Synopsis

Perform a OpenID Connect 1.0 claims hydration debug operation.

This subcommand allows checking an OpenID Connect 1.0 claims hydration scenario by providing certain information about a request.

```
authelia debug oidc claims <username> [flags]
```

### Examples

```
authelia debug oidc claims --help
```

### Options

```
      --claims strings         granted claims to use for this request
      --client-id string       arbitrary client id for the client (default "example")
      --grant-type string      grant type to use for this request (default "authorization_code")
  -h, --help                   help for claims
      --policy string          claims policy name to use
      --response-type string   response type to use for this request (default "code")
      --scopes strings         granted scopes to use for this request (default [openid,profile,email,phone,address,groups])
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia debug oidc](authelia_debug_oidc.md)	 - Perform a OpenID Connect 1.0 debug operation

