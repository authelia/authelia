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

You can verify the artifact signature using the gpg tool. Below is an example of verifying the Authelia releases:

{{< envTabs "Verify Signatures" >}}
{{< envTab "4.39.11+" >}}
```shell
# Download checksums and signature
curl -fsSL \
  -O https://github.com/authelia/authelia/releases/download/v4.39.11/authelia-v4.39.11-linux-amd64.tar.gz \
  -O https://github.com/authelia/authelia/releases/download/v4.39.11/checksums.sha256 \
  -O https://github.com/authelia/authelia/releases/download/v4.39.11/checksums.sha256.sig

# Verify signature and checksums
gpg --verify checksums.sha256.sig checksums.sha256 && sha256sum -c checksums.sha256
```
{{< /envTab >}}
{{< envTab "Pre 4.39.11" >}}
```shell
gpg --verify authelia-v4.39.10-linux-amd64.tar.gz.sha256.sig authelia-v4.39.10-linux-amd64.tar.gz.sha256 && \
 echo "$(cat authelia-v4.39.10-linux-amd64.tar.gz.sha256)  authelia-v4.39.10-linux-amd64.tar.gz" | sha256sum -c
```
{{< /envTab >}}
{{< /envTabs >}}

Note: We adjusted the format of checksums in 4.39.11 to make verification easier.

Example output:
```text
gpg: Signature made Mon 15 Sep 2025 02:09:56 AM PDT
gpg:                using RSA key C387CC1B5FFC25E55F75F3E6A228F3BD04CC9652
gpg:                issuer "security@authelia.com"
gpg: Good signature from "Authelia Security <security@authelia.com>" [unknown]
gpg:                 aka "Authelia Security <team@authelia.com>" [unknown]
gpg: WARNING: This key is not certified with a trusted signature!
gpg:          There is no indication that the signature belongs to the owner.
Primary key fingerprint: 1920 8591 5BD6 08A4 58AC  58DC E461 FA15 3128 6EEA
     Subkey fingerprint: C387 CC1B 5FFC 25E5 5F75  F3E6 A228 F3BD 04CC 9652
authelia-v4.39.10-linux-amd64.tar.gz: OK
```

## SLSA Provenance

In addition to artifact signatures, Authelia generates and signs **[SLSA Provenance]** for its
builds.

**Provenance** is metadata that describes how an artifact was built. For example, what source code, build steps, and
environment were used. This helps users and systems verify that the software was built in a trustworthy and repeatable
way.

Authelia’s provenance conforms to **[SLSA Build Level 3](https://slsa.dev/spec/v1.1/levels#build-l3)**.

The [SLSA Provenance] covers the release artifacts i.e. those ending with `.tar.gz` and `.deb`.

You can verify the [SLSA Provenance] using the [slsa-verifier](https://github.com/slsa-framework/slsa-verifier). Below
is an example verifying the FreeBSD amd64 and Linux amd64 (musl) Authelia v4.39.8 release tarballs:

```shell
V=v4.39.8
for F in authelia-${V}-{linux-{amd64,arm,arm64,amd64-musl,arm-musl,arm64-musl},freebsd-amd64,public_html}.tar.gz authelia.intoto.jsonl; do
  curl -fsSLO https://github.com/authelia/authelia/releases/download/${V}/${F}
done
slsa-verifier verify-artifact authelia-${V}-{linux-{amd64,arm,arm64,amd64-musl,arm-musl,arm64-musl},freebsd-amd64,public_html}.tar.gz --provenance-path authelia.intoto.jsonl --source-uri "github.com/authelia/authelia"
```

Example output:

```text
Verified build using builder "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v2.1.0" at commit 5d90442e07cc695c61036ac1a539c0b942ebc71d
Verifying artifact authelia-v4.39.8-freebsd-amd64.tar.gz: PASSED

Verified build using builder "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v2.1.0" at commit 5d90442e07cc695c61036ac1a539c0b942ebc71d
Verifying artifact authelia-v4.39.8-linux-amd64-musl.tar.gz: PASSED

Verified build using builder "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v2.1.0" at commit 5d90442e07cc695c61036ac1a539c0b942ebc71d
Verifying artifact authelia-v4.39.8-linux-amd64.tar.gz: PASSED

PASSED: SLSA verification passed
```

[SLSA Provenance]: https://slsa.dev/
