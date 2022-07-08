---
title: "Environment"
description: "How to configure your development environment."
lead: "This section covers the environment we recommend for development."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  contributing:
    parent: "development"
weight: 220
toc: true
---

__Authelia__ and its development workflow can be tested with [Docker] and [Docker Compose] on Linux.

## Setup

In order to build and contribute to __Authelia__, you need to make sure the following are installed in your environment:

* [go] *(v1.18 or greater)*
* [Docker]
* [Docker Compose]
* [Node.js] *(v16 or greater)*
* [pnpm]

The additional tools are recommended:

* [golangci-lint]
* [goimports-reviser]
* [yamllint]
* Either the [VSCodium] or [GoLand] IDE

## Scripts

There is a scripting context provided with __Authelia__ which can easily be configured. It allows running integration
[suites] and various other tasks. Read more about it in the [authelia-scripts](reference-authelia-scripts.md) reference
guide.

## FAQ

### Do you support development under Windows or OSX?

At the present time this is not officially supported. Some of the maintainers utilize Windows however running suites
under Windows or OSX is not something that is currently possible to do easily. As such we recommend utilizing Linux.

### What version of Docker and docker-compose should I use?

We have no firm recommendations on the version to use but we actively use the latest versions available to us in the
distributions of our choice. As long as it's a modern version it should be sufficient for the development environment.

### How can I serve my application under example.com?

Don't worry, you don't need to own the domain `example.com` to test Authelia. Copy the following lines in
your `/etc/hosts`:

```text
192.168.240.100 home.example.com
192.168.240.100 login.example.com
192.168.240.100 singlefactor.example.com
192.168.240.100 public.example.com
192.168.240.100 secure.example.com
192.168.240.100 mail.example.com
192.168.240.100 mx1.mail.example.com
```

The IP address `192.168.240.100` is the IP attributed by [Docker] to the reverse proxy. Once added you can access the
listed subdomains from your browser, and they will be served by the reverse proxy.

[suites]: ./integration-suites.md
[Buildkite]: https://buildkite.com/
[React]: https://reactjs.org/
[go]: https://go.dev/dl/
[Node.js]: https://nodejs.org/en/download/
[pnpm]: https://pnpm.io/installation
[Docker]: https://docs.docker.com/get-docker/
[Docker Compose]: https://docs.docker.com/compose/install/
[golangci-lint]: https://golangci-lint.run/usage/install/
[goimports-reviser]: https://github.com/incu6us/goimports-reviser#install
[yamllint]: https://yamllint.readthedocs.io/en/stable/quickstart.html
[VSCodium]: https://vscodium.com/
[GoLand]: https://www.jetbrains.com/go/
