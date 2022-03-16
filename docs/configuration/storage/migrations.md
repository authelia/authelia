---
layout: default
title: Migrations
parent: Storage Backends
grand_parent: Configuration
nav_order: 5
---

Storage migrations are important for keeping your database compatible with Authelia. Authelia will automatically upgrade
your schema on startup. However, if you wish to use an older version of Authelia you may be required to manually
downgrade your schema with a version of Authelia that supports your current schema.

## Schema Version to Authelia Version map

This table contains a list of schema versions and the corresponding release of Authelia that shipped with that version.
This means all Authelia versions between two schema versions use the first schema version. 

For example for version pre1, it is used for all versions between it and the version 1 schema, so 4.0.0 to 4.32.2. In 
this instance if you wanted to downgrade to pre1 you would need to use an Authelia binary with version 4.33.0 or higher.

| Schema Version | Authelia Version |                                               Notes                                               |
|:--------------:|:----------------:|:-------------------------------------------------------------------------------------------------:|
|      pre1      |      4.0.0       |                   Downgrading to this version requires you use the --pre1 flag                    |
|       1        |      4.33.0      |                                 Initial migration managed version                                 |
|       2        |      4.34.0      | Webauthn - added webauthn_devices table, altered totp_config to include device created/used dates |
|       3        |      4.34.2      |     Webauthn - fix V2 migration kid column length and provide migration path for anyone on V2     |
