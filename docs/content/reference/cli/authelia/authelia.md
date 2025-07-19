---
title: "authelia"
description: "Reference for the authelia command."
lead: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 900
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## authelia

authelia untagged-unknown-dirty (master, unknown)

### Synopsis

authelia untagged-unknown-dirty (master, unknown)

An open-source authentication and authorization server providing
two-factor authentication and single sign-on (SSO) for your
applications via a web portal.

General documentation is available at: https://www.authelia.com/
CLI documentation is available at: https://www.authelia.com/reference/cli/authelia/authelia/

```
authelia [flags]
```

### Examples

```
authelia --config /etc/authelia/config.yml --config /etc/authelia/access-control.yml
authelia --config /etc/authelia/config.yml,/etc/authelia/access-control.yml
authelia --config /etc/authelia/config/
```

### Options

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
  -h, --help                                  help for authelia
```

### SEE ALSO

* [authelia access-control](authelia_access-control.md)	 - Helpers for the access control system
* [authelia build-info](authelia_build-info.md)	 - Show the build information of Authelia
* [authelia config](authelia_config.md)	 - Perform config related actions
* [authelia crypto](authelia_crypto.md)	 - Perform cryptographic operations
* [authelia debug](authelia_debug.md)	 - Perform debug functions
* [authelia storage](authelia_storage.md)	 - Manage the Authelia storage
* [authelia validate-config](authelia_validate-config.md)	 - Check a configuration against the internal configuration validation mechanisms

