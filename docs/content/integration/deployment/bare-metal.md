---
title: "Bare-Metal"
description: "Deploying Authelia on Bare-Metal."
summary: "Authelia can be deployed on Bare-Metal as long as it sits behind a proxy."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 250
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

There are several ways to achieve this, as *Authelia* runs as a daemon. We do not provide specific examples for running
*Authelia* as a service excluding the [systemd unit](#systemd) files.

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## systemd

We publish two example [systemd] unit files:

* {{< github-link path="authelia.service" >}}
* {{< github-link path="authelia@.service" >}}

## Arch Linux

In addition to the [binaries](#binaries) we publish, we also publish an
[AUR Package](https://aur.archlinux.org/packages/authelia).

## Debian

We publish Debian Packages (`.deb`) with our [releases] which can be installed
on most Debian based operating systems.

### Signing

Both the Debian Packages and the [APT Repository](#apt-repository) are signed using the signing architecture described
in [Artifact Signing and Provenance Overview](../../overview/security/artifact-signing-and-provenance.md).

### APT Repository

In addition to the [Debian](#debian) Packages we also have an APT Repository. The steps to add it are noted below.

#### Add the APT Repository

Add the required packages and download the repository key which is described in more detail in the
[Artifact Signing and Provenance Overview](../../overview/security/artifact-signing-and-provenance.md):

```shell
sudo apt install ca-certificates curl gnupg
sudo curl -fsSL https://www.authelia.com/keys/authelia-security.gpg -o /usr/share/keyrings/authelia-security.gpg
```

Verify the downloaded key:

```shell
gpg --no-default-keyring --keyring /usr/share/keyrings/authelia-security.gpg --list-keys --with-subkey-fingerprint
```

Example output showing the correct Key IDs:

```text
/usr/share/keyrings/authelia-security.gpg
-----------------------------------------
pub   rsa4096 2025-06-27 [SC]
      192085915BD608A458AC58DCE461FA1531286EEA
uid           [ unknown] Authelia Security <security@authelia.com>
uid           [ unknown] Authelia Security <team@authelia.com>
sub   rsa2048 2025-06-27 [E] [expires: 2033-06-25]
      7DBA42FED0069D5828A44079975E8FFC6876AFBB
sub   rsa2048 2025-06-27 [SA] [expires: 2033-06-25]
      C387CC1B5FFC25E55F75F3E6A228F3BD04CC9652
```

Add the repository to `sources.list.d`:

```shell
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/authelia-security.gpg] https://apt.authelia.com stable main" | \
  sudo tee /etc/apt/sources.list.d/authelia.list > /dev/null
```

Update the cache and install:

```shell
sudo apt update && sudo apt install authelia
```

## Nix

Using the Nix package manager Authelia is available via the `https://nixos.org/channels/nixpkgs-unstable` channel. It
should be noted that this channel is both unstable and this is a third party package.

```shell
nix-channel --add https://nixos.org/channels/nixpkgs-unstable
nix-channel --update
nix-env -iA nixpkgs.authelia
```

## FreeBSD

In addition to the [binaries](#binaries) we publish, [FreshPorts](https://www.freshports.org/www/authelia/) offer a
third party package.

We publish an [rc.d](https://docs.freebsd.org/en/articles/rc-scripting/) service script file:

* {{< github-link path="authelia-fb-rc.d" >}}

## Binaries

We publish binaries with our [releases] which can be installed on many operating systems.

[releases]: https://github.com/authelia/authelia/releases
[systemd]: https://systemd.io/
