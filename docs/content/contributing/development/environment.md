---
title: "Environment"
description: "How to configure your development environment."
summary: "This section covers the environment we recommend for development."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 220
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__ and its development workflow can be tested with [Docker] and [Docker Compose] on Linux.

## Setup

In order to build and contribute to __Authelia__, you need to make sure the following are installed in your environment:

* General:
  * [git]
* Backend Development:
  * [go] *(v1.21 or greater)*
  * [gcc]
  * [gomock]
* Frontend Development
  * [Node.js] *(v18 or greater)*
  * [pnpm]
* Integration Suites:
  * [Docker]
  * [Docker Compose]
  * [chromium]

The additional tools are recommended:

* [golangci-lint]
* [goimports-reviser]
* [yamllint]
* [VSCodium] or [GoLand]

## Certificate

Authelia utilizes a self-signed Root CA certificate for the development environment. This allows us to sign elements of
the CI process uniformly and only trust a single additional Root CA Certificate. The private key for this certificate is
maintained by the [Core Team] so if you need an additional certificate signed for this purpose please reach out to them.

While developing for Authelia you may also want to trust this Root CA. It is critical that you are aware of what this
means if you decide to do so.

1. It will allow us to generate trusted certificates for machines this is installed on.
2. If compromised there is no formal revocation process at this time as we are not a certified CA.
3. Trusting Root CA's is not necessary for the development process it only makes it smoother.
4. Trusting additional Root CA's for prolonged periods is not generally a good idea.

If you'd still like to trust the Root CA Certificate it's located (encoded as a PEM) in the main git repository at
 [/internal/suites/common/pki/ca/ca.public.crt](https://github.com/authelia/authelia/blob/master/internal/suites/common/pki/ca/ca.public.crt).

## Scripts

There is a scripting context provided with __Authelia__ which can easily be configured. It allows running integration
[suites] and various other tasks. Read more about it in the [authelia-scripts](reference-authelia-scripts.md) reference
guide.

## Frequently Asked Questions

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
[gomock]: https://github.com/golang/mock
[Node.js]: https://nodejs.org/en/download/
[pnpm]: https://pnpm.io/installation
[Docker]: https://docs.docker.com/get-docker/
[Docker Compose]: https://docs.docker.com/compose/install/
[golangci-lint]: https://golangci-lint.run/welcome/install/
[goimports-reviser]: https://github.com/incu6us/goimports-reviser#install
[yamllint]: https://yamllint.readthedocs.io/en/stable/quickstart.html
[VSCodium]: https://vscodium.com/
[GoLand]: https://www.jetbrains.com/go/
[chromium]: https://www.chromium.org/
[git]: https://git-scm.com/
[gcc]: https://gcc.gnu.org/
