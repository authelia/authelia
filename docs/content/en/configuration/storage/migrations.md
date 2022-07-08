---
title: "Migrations"
description: "Storage Migrations"
lead: "A migration ."
date: 2021-11-23T20:45:38+11:00
draft: false
images: []
menu:
  configuration:
    parent: "storage"
weight: 106200
toc: true
aliases:
  - /docs/configuration/storage/migrations.html
---

Storage migrations are important for keeping your database compatible with Authelia. Authelia will automatically upgrade
your schema on startup. However, if you wish to use an older version of Authelia you may be required to manually
downgrade your schema with a version of Authelia that supports your current schema.

## Schema Version to Authelia Version map

This table contains a list of schema versions and the corresponding release of Authelia that shipped with that version.
This means all Authelia versions between two schema versions use the first schema version.

For example for version pre1, it is used for all versions between it and the version 1 schema, so 4.0.0 to 4.32.2. In
this instance if you wanted to downgrade to pre1 you would need to use an Authelia binary with version 4.33.0 or higher.

| Schema Version | Authelia Version |                                               Notes                                                |
|:--------------:|:----------------:|:--------------------------------------------------------------------------------------------------:|
|      pre1      |      4.0.0       |                    Downgrading to this version requires you use the --pre1 flag                    |
|       1        |      4.33.0      |                                 Initial migration managed version                                  |
|       2        |      4.34.0      | WebAuthn - added webauthn_devices table, altered totp_config to include device created/used dates  |
|       3        |      4.34.2      |     WebAuthn - fix V2 migration kid column length and provide migration path for anyone on V2      |
|       4        |      4.35.0      |               Added OpenID Connect storage tables and opaque user identifier tables                |
|       5        |      4.35.1      | Fixed the oauth2_consent_session table to accept NULL subjects for users who are not yet signed in |
