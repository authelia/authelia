---
title: "Artifact Signing and Provenance"
description: "An overview of Authelia's Artifact Signing and Provenance."
summary: "An overview of Authelia's Artifact Signing and Provenance."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 440
toc: true
aliases:
  - /o/verify
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

As part of our passion for security and compliance we have adopted a number of practices that assist users in verifying
the integrity of the software they are running. This is an overview of these initiatives.

## Artifact Signing

Authelia uses a dedicated GPG key to sign distributed artifacts, ensuring authenticity and integrity.

The following information describes the key used to sign the artifacts:

- Key ID: `192085915BD608A458AC58DCE461FA1531286EEA`
- Key Fingerprint: `1920 8591 5BD6 08A4 58AC  58DC E461 FA15 3128 6EEA`
- Sub Key ID (Encryption): `7DBA42FED0069D5828A44079975E8FFC6876AFBB`
- Sub Key ID (Signing): `C387CC1B5FFC25E55F75F3E6A228F3BD04CC9652`
- Key Owners:
  - `Authelia Security <security@authelia.com>`
  - `Authelia Security <team@authelia.com>`

The public key can be obtained from the following locations:

- **Keyring File**:
  - Official Website: [authelia-security.gpg](https://www.authelia.com/keys/authelia-security.gpg)
- **Armored ASCII File**:
  - Official Website: [authelia-security.asc](https://www.authelia.com/keys/authelia-security.asc)
  - [Keybase](https://keybase.io/): [authelia/pgp_keys.asc](https://keybase.io/authelia/pgp_keys.asc)
- **Key Servers**:
  - [keys.openpgp.org](https://keys.openpgp.org/search?q=192085915BD608A458AC58DCE461FA1531286EEA)
  - [keyserver.ubuntu.com](https://keyserver.ubuntu.com/pks/lookup?search=192085915BD608A458AC58DCE461FA1531286EEA&fingerprint=on&op=index)

The following artifacts are signed with this key:

- **[Official Authelia Release Artifacts](https://github.com/authelia/authelia/releases)**
- **[Debian Packages](../../integration/deployment/bare-metal.md#debian)**
- **[APT Repository](../../integration/deployment/bare-metal.md#apt-repository)**
- **[SLSA Provenance](#slsa-provenance)**

## SLSA Provenance

In addition to artifact signatures, Authelia generates and signs **[SLSA Provenance](https://slsa.dev/)** for its
builds.

**Provenance** is metadata that describes how an artifact was built. For example, what source code, build steps, and
environment were used. This helps users and systems verify that the software was built in a trustworthy and repeatable
way.

Authelia’s provenance conforms to **[SLSA Build Level 3](https://slsa.dev/spec/v1.1/levels)**, meaning it is generated
automatically by the build system and signed with our GPG key to prevent tampering.
