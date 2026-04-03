---
title: "Migrations"
description: "Storage Migrations"
summary: "A migration ."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 107200
toc: true
aliases:
  - /docs/configuration/storage/migrations.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
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
|      pre1      |      4.0.0       |          Downgrading to this version requires you use the --pre1 flag on Authelia 4.37.2           |
|       1        |      4.33.0      |                                 Initial migration managed version                                  |
|       2        |      4.34.0      | WebAuthn - added webauthn_devices table, altered totp_config to include device created/used dates  |
|       3        |      4.34.2      |     WebAuthn - fix V2 migration kid column length and provide migration path for anyone on V2      |
|       4        |      4.35.0      |             Added OpenID Connect 1.0 storage tables and opaque user identifier tables              |
|       5        |      4.35.1      | Fixed the oauth2_consent_session table to accept NULL subjects for users who are not yet signed in |
|       6        |      4.37.0      |        Adjusted the OpenID Connect 1.0 tables to allow pre-configured consent improvements         |
|       7        |      4.37.3      |       Fixed some schema inconsistencies most notably the MySQL/MariaDB Engine and Collation        |
|       8        |      4.38.0      |                          OpenID Connect 1.0 Pushed Authorization Requests                          |
|       9        |      4.38.0      | Fix a PostgreSQL NOT NULL constraint issue on the `aaguid` column of the `webauthn_devices` table  |
|       10       |      4.38.0      |   Fix constraints on the `oauth2_access_token_session` table for the `client credentials` grant    |
|       11       |      4.38.0      |             Adjust constraints for JWT Profile for OAuth 2.0 Access Tokens ([RFC9068])             |
|       12       |      4.38.0      |                        WebAuthn adjustments for multi-cookie domain changes                        |
|       13       |      4.38.0      |                   One-Time Password for Identity Verification via Email Changes                    |
|       14       |      4.38.0      |                                    Revoke Reset Password Token                                     |
|       15       |      4.38.0      |                         Time-based One-Time Password security enhancement                          |
|       16       |      4.39.0      |                                OAuth 2.0 Allow Consent Subject NULL                                |
|       17       |      4.39.0      |                                OpenID Connect 1.0 Claims Parameter                                 |
|       18       |      4.39.0      |                                     OAuth 2.0 Device Code Flow                                     |
|       19       |      4.39.0      |                                         WebAuthn Passkeys                                          |
|       20       |      4.39.0      |                                         Regulation Rework                                          |
|       21       |      4.39.1      |                                MySQL Specific Fix for WebAuthn MDS                                 |
|       22       |      4.39.2      |                OAuth 2.0 Consent Session Expiration Time instead of Subject Binding                |
|       23       |     4.39.12      |                            OAuth 2.0 Device Code Flow Null Constraints                             |

[RFC9068]: https://datatracker.ietf.org/doc/html/rfc9068
