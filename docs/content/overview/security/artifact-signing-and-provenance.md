---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Artifact Signing and Provenance"
description: "An overview of Authelia's artifact signing with GPG keys and SLSA Build Level 3 provenance for verifying the integrity and authenticity of official releases."
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

The [SLSA Provenance] covers the release artifacts i.e. those ending with `.tar.gz` and `.deb`. For releases after
v4.39.28 it also covers the container images published to [Docker Hub](https://hub.docker.com/r/authelia/authelia) and
the [GitHub Container Registry](https://github.com/authelia/authelia/pkgs/container/authelia).

### Verification

You can verify the [SLSA Provenance] using the [GitHub CLI](https://cli.github.com/) and its
[`gh attestation verify`](https://cli.github.com/manual/gh_attestation_verify) command. Below is an example verifying
all the Authelia release tarballs (add or remove artifacts depending on your requirements) for a specific version:

```shell
V=v{{% latest %}}
for F in authelia-${V}-{linux-{amd64,arm,arm64,amd64-musl,arm-musl,arm64-musl},freebsd-amd64,public_html}.tar.gz authelia.intoto.jsonl; do
  curl -fsSLO https://github.com/authelia/authelia/releases/download/${V}/${F}
done
for F in authelia-${V}-{linux-{amd64,arm,arm64,amd64-musl,arm-musl,arm64-musl},freebsd-amd64,public_html}.tar.gz; do
  gh attestation verify ${F} --bundle authelia.intoto.jsonl --repo authelia/authelia
done
```

The `--bundle` flag verifies against the `authelia.intoto.jsonl` file attached to the release, so this works without
running `gh auth login`. Without it `gh` fetches the attestation from GitHub instead, which requires you to be signed
in.

For a stricter check, add `--signer-workflow authelia/authelia/.github/workflows/slsa-provenance.yml` to require that
the attestation was produced by Authelia's provenance workflow, and `--source-ref refs/tags/${V}` to require that it was
produced for that release's tag.

When its output is not a terminal, `gh attestation verify` prints nothing on success and exits with status `0`.

#### Container Images

The container images are verified by referencing them with an `oci://` URI. The attestations are attached to the images
in both registries, so the `--bundle-from-oci` flag fetches them from the registry rather than from GitHub. Below is an
example verifying the Authelia v{{% latest %}} container image on both registries:

```shell
V={{% latest %}}
gh attestation verify oci://docker.io/authelia/authelia:${V} --bundle-from-oci --repo authelia/authelia
gh attestation verify oci://ghcr.io/authelia/authelia:${V} --bundle-from-oci --repo authelia/authelia
```

The `gh` CLI must be able to authenticate with the registry it reads from, for example with `docker login`.

The tag refers to the multi-platform image index. The image for each platform is also attested individually, but only
through GitHub, so verifying it by digest requires running `gh auth login` and omitting the `--bundle-from-oci` flag:

```shell
gh attestation verify oci://ghcr.io/authelia/authelia@sha256:<digest> --repo authelia/authelia
```

The `--signer-workflow` and `--source-ref refs/tags/v${V}` flags described above apply to container images as well.

#### Older Releases

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
The process above applies to releases after v4.39.28. Releases v4.39.10 to v4.39.28 were signed by the
[slsa-github-generator](https://github.com/slsa-framework/slsa-github-generator), with the exception of v4.39.21 which
has no provenance.
{{< /callout >}}

Releases signed by the generator can be verified with the same command and two additional flags. The `--signer-repo`
flag is required because the provenance was signed by the generator's repository, and the `--predicate-type` flag
because it uses SLSA provenance v0.2:

```shell
gh attestation verify authelia-${V}-linux-amd64.tar.gz --bundle authelia.intoto.jsonl --repo authelia/authelia \
  --signer-repo slsa-framework/slsa-github-generator --predicate-type https://slsa.dev/provenance/v0.2
```

Example output for v4.39.28:

```text
Loaded digest sha256:e8ad9ff965cbf3b93138945e13cb4b7874c999c254772b2bb77952863177699c for file://authelia-v4.39.28-linux-amd64.tar.gz
Loaded 1 attestations from authelia.intoto.jsonl

The following policy criteria will be enforced:
- Predicate type must match:................ https://slsa.dev/provenance/v0.2
- Source Repository Owner URI must match:... https://github.com/authelia
- Source Repository URI must match:......... https://github.com/authelia/authelia
- Subject Alternative Name must match regex: (?i)^https://github\.com/slsa-framework/slsa-github-generator/
- OIDC Issuer must match:................... https://token.actions.githubusercontent.com

✓ Verification succeeded!

The following 1 attestation matched the policy criteria

- Attestation #1
  - Build repo:..... authelia/authelia
  - Build workflow:. .github/workflows/slsa.yml@refs/tags/v4.39.28
  - Signer repo:.... slsa-framework/slsa-github-generator
  - Signer workflow: .github/workflows/generator_generic_slsa3.yml@f7dd8c54c2067bafc12ca7a55595d5ee9b75204a
```

[SLSA Provenance]: https://slsa.dev/
