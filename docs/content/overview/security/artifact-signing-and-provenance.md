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

### Verification

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
While prior releases can be verified, this _specific_ process only applies after the 4.39.11 release.
{{< /callout >}}

You can verify the artifact signature using the gpg tool. Below is an example verifying the Linux
amd64 musl and glibc Authelia v{{% latest %}} release tarballs (add or remove artifacts depending on your requirements):

```shell
# Download checksums and signature
curl -fsSL \
  -O https://github.com/authelia/authelia/releases/download/v{{% latest %}}/authelia-v{{% latest %}}-linux-amd64-musl.tar.gz \
  -O https://github.com/authelia/authelia/releases/download/v{{% latest %}}/authelia-v{{% latest %}}-linux-amd64.tar.gz \
  -O https://github.com/authelia/authelia/releases/download/v{{% latest %}}/checksums.sha256 \
  -O https://github.com/authelia/authelia/releases/download/v{{% latest %}}/checksums.sha256.sig

# Verify signature and checksums
gpg --verify checksums.sha256.sig checksums.sha256 && sha256sum --ignore-missing -c checksums.sha256
```

Example output:

```text
gpg: Signature made Wed 08 Oct 2025 19:00:48 AEDT
gpg:                using RSA key C387CC1B5FFC25E55F75F3E6A228F3BD04CC9652
gpg:                issuer "security@authelia.com"
gpg: Good signature from "Authelia Security <security@authelia.com>" [unknown]
gpg:                 aka "Authelia Security <team@authelia.com>" [unknown]
gpg: WARNING: The key's User ID is not certified with a trusted signature!
gpg:          There is no indication that the signature belongs to the owner.
Primary key fingerprint: 1920 8591 5BD6 08A4 58AC  58DC E461 FA15 3128 6EEA
     Subkey fingerprint: C387 CC1B 5FFC 25E5 5F75  F3E6 A228 F3BD 04CC 9652
authelia-v{{% latest %}}-linux-amd64-musl.tar.gz: OK
authelia-v{{% latest %}}-linux-amd64.tar.gz: OK
```

Note: The above warning from GPG is expected if you have not manually trusted the Authelia's gpg key. This is unnecessary for the purposes of verifying the integrity and authenticity of Authelia releases.

## SLSA Provenance

In addition to artifact signatures, Authelia generates and signs **[SLSA Provenance]** for its
builds.

**Provenance** is metadata that describes how an artifact was built. For example, what source code, build steps, and
environment were used. This helps users and systems verify that the software was built in a trustworthy and repeatable
way.

Authelia’s provenance conforms to **[SLSA Build Level 3](https://slsa.dev/spec/v1.1/levels#build-l3)**. Which means that "forging the provenance or evading verification requires exploiting a vulnerability that is beyond the capabilities of most adversaries."

The [SLSA Provenance] covers the release artifacts i.e. those ending with `.tar.gz` and `.deb` and does not include built docker images.

### Verification

You can verify the [SLSA Provenance] using the [slsa-verifier](https://github.com/slsa-framework/slsa-verifier). Below
is an example verifying all the Authelia release tarballs (add or
remove artifacts depending on your requirements) for a specific version:

```shell
V=v{{% latest %}}
for F in authelia-${V}-{linux-{amd64,arm,arm64,amd64-musl,arm-musl,arm64-musl},freebsd-amd64,public_html}.tar.gz authelia.intoto.jsonl; do
  curl -fsSLO https://github.com/authelia/authelia/releases/download/${V}/${F}
done
slsa-verifier verify-artifact authelia-${V}-{linux-{amd64,arm,arm64,amd64-musl,arm-musl,arm64-musl},freebsd-amd64,public_html}.tar.gz --provenance-path authelia.intoto.jsonl --source-uri "github.com/authelia/authelia"
```

Example output:

```text
Verified build using builder "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v2.1.0" at commit 5d90442e07cc695c61036ac1a539c0b942ebc71d
Verifying artifact authelia-v{{% latest %}}-freebsd-amd64.tar.gz: PASSED

Verified build using builder "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v2.1.0" at commit 5d90442e07cc695c61036ac1a539c0b942ebc71d
Verifying artifact authelia-v{{% latest %}}-linux-amd64-musl.tar.gz: PASSED

Verified build using builder "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v2.1.0" at commit 5d90442e07cc695c61036ac1a539c0b942ebc71d
Verifying artifact authelia-v{{% latest %}}-linux-amd64.tar.gz: PASSED

PASSED: SLSA verification passed
```

[SLSA Provenance]: https://slsa.dev/
