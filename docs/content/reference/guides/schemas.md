---
title: "Schemas"
description: "A reference guide on the schemas provided by Authelia"
summary: "This section contains reference documentation for Authelia's schemas."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 220
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The following reference material documents some of the schemas we publish.

## JSON Schema

Authelia publishes several [JSON Schema] documents. These files are published in the following URL format (with the URL
also being the schema ID):

```
https://www.authelia.com/schemas/<version>/json-schema/<name>.json
```

These schemas can be added to the top of a YAML file using the following format:

```yaml {title="configuration.yml"}
# yaml-language-server: $schema=https://www.authelia.com/schemas/<version>/json-schema/<name>.json

example: 'this is just an example'
```

Where:

1. The `<version>` placeholder is in the format `v<major>.<minor>` i.e. for version 4.38.1 the `<version>` is replaced
   by `v4.38`.
2. The `<name>` placeholder replaced by the name of the individual [JSON Schema] below.
3. The following special meta versions exist:
   1. The `latest` version refers to the latest released version of Authelia.
   2. The `next` version refers to the latest commit to the master branch.


### Configuration

**Schema Name:** `configuration`

The [JSON Schema] document for the main [configuration file](../../configuration/methods/files.md).

### Users Database

**Schema Name:** `user-database`

The [JSON Schema] document for the [users database configuration file](passwords.md#user--password-file).

### TOTP Export

**Schema Name:** `exports.totp`

The [JSON Schema] document for the [TOTP export file](../cli/authelia/authelia_storage_user_totp_export.md).

### WebAuthn Export

**Schema Name:** `exports.webauthn`

The [JSON Schema] document for the [WebAuthn export file](../cli/authelia/authelia_storage_user_webauthn_export.md).

### Identifiers Export

**Schema Name:** `exports.identifiers`

The [JSON Schema] document for the [Identifiers export file](../cli/authelia/authelia_storage_user_identifiers_export.md).

[JSON Schema]: https://json-schema.org/
