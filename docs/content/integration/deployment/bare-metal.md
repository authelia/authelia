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

We publish `.deb` packages with our [releases] which can be installed
on most Debian based operating systems.

### APT Repository

In addition to the `.deb` packages we also have an [APT Repository](https://apt.authelia.com).

```shell
sudo apt install ca-certificates curl
sudo curl -fsSL https://apt.authelia.com/organization/signing.asc -o /usr/share/keyrings/authelia.asc
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/authelia.asc] https://apt.authelia.com/stable/debian/debian all main" | \
  sudo tee /etc/apt/sources.list.d/authelia.list > /dev/null
```

## Nix

Using the Nix package manager Authelia is available via the `https://nixos.org/channels/nixpkgs-unstable` channel.

```shell
$ nix-channel --add https://nixos.org/channels/nixpkgs-unstable
$ nix-channel --update
$ nix-env -iA nixpkgs.authelia
```

## FreeBSD

In addition to the [binaries](#binaries) we publish, [FreshPorts](https://www.freshports.org/www/authelia/) offer a
package.

We publish an [rc.d](https://docs.freebsd.org/en/articles/rc-scripting/) service script file:

* {{< github-link path="authelia-fb-rc.d" >}}

## Binaries

We publish binaries with our [releases] which can be installed on many operating systems.

[releases]: https://github.com/authelia/authelia/releases
[systemd]: https://systemd.io/
