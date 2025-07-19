---
title: "Environment"
description: "How to configure your development environment."
summary: "This section covers the environment we recommend for development."
date: 2024-03-14T06:00:14+11:00
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

## Prerequisites

In order to build and contribute to __Authelia__, you need to make sure the following are installed in your environment:

* General:
  * [git]
  * [bash]
  * A modern Linux distribution, Windows and macOS are not officially supported.
* Backend Development:
  * [go]:
    * Minimum is *v1.24.3*.
    * The toolchain version noted in [go.mod](https://github.com/authelia/authelia/blob/master/go.mod#L5) is the
      officially supported version.
    * We will not officially support old versions of go generally without a very good reason.
  * [gcc]
  * [gomock]
* Frontend Development:
  * [Node.js] *(v22.15.0 or greater)*.
  * [pnpm] *(v10.10.0 or greater)*.
* Integration Suites:
  * [Docker] *(v28.1.1 or greater)*.
  * [Docker Compose] *(v2.36.0 or greater)*
  * [chromium]
  * [delve]

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
We only support docker being installed as a system package. It must not be installed via Snappy or other similar tools.
If you wish to use these other tools we will not be able to provide support.
{{< /callout >}}

These additional tools are recommended:

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
 [internal/suites/common/pki/ca/ca.public.crt](https://github.com/authelia/authelia/blob/master/internal/suites/common/pki/ca/ca.public.crt).

## Context

There is a development context that can be loaded using [bootstrap.sh]. All that you need to do is satisfy the
prerequisites and run the below command (or an equivalent of):

```bash
source bootstrap.sh
```

This context gives you access to the following commands:

- [authelia-scripts](../../reference/cli/authelia-scripts/authelia-scripts.md) - used to perform various development
  focused tasks such as building the binary, building a docker image, setting up
  [integration suites](integration-suites.md), or performing tests
- [authelia-gen](../../reference/cli/authelia-gen/authelia-gen.md) - used to perform code generation, we recommend
  running this command just before you commit whenever you make changes
- authelia-suites - used to manage suites, generally we recommend using `authelia-scripts` instead

## Frequently Asked Questions

### Do you support development under Windows or macOS?

At the present time this is not officially supported. Some of the maintainers utilize Windows and/or macOS however
running suites under Windows or macOS is not something that is currently possible to do easily. As such we recommend
utilizing Linux.

### Why can't I use old versions of docker and docker compose?

We have geared all of our examples and suites based on modern versions of these products. If you decide to use an older
version then you're on your own.

### How can I run the suite test applications when they're hosted on arbitrary domains?

The [authelia-scripts bootstrap](../../reference/cli/authelia-scripts/authelia-scripts_bootstrap.md) subcommand handles
this for you and creates the relevant hosts entries. This is automatically executed when using [bootstrap.sh].

[suites]: ./integration-suites.md
[Buildkite]: https://buildkite.com/
[React]: https://reactjs.org/
[go]: https://go.dev/dl/
[gomock]: https://github.com/uber-go/mock
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
[bash]: https://www.gnu.org/software/bash/
[delve]: https://github.com/go-delve/delve
[core team]: https://www.authelia.com/information/about/#core-team
[bootstrap.sh]: https://github.com/authelia/authelia/blob/master/bootstrap.sh
