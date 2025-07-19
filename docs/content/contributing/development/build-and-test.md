---
title: "Building and Testing"
description: "Building and Testing Authelia."
summary: "This section covers the build process and how to perform tests in development."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 240
toc: true
aliases:
  - /docs/contributing/build-and-dev.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__ is built a [React] frontend user portal bundled in a [Go] application which acts as a basic web server for
the [React] assets and a dedicated API.

The GitHub repository comes with a CLI dedicated to developers called
[authelia-scripts](../../reference/cli/authelia-scripts/authelia-scripts.md) which can be setup by looking at
the [authelia-scripts reference guide](../../reference/cli/authelia-scripts/authelia-scripts.md).

In order to build and contribute to __Authelia__, you need to make sure that you have looked at the
[Environment](environment.md) guide to configure your development environment.

## Get started

In order to ease development, __Authelia__ uses the concept of [suites] to run Authelia from source code so that your
patches are included. This is a kind of virtual environment running __Authelia__ in a complete ecosystem
(LDAP, Redis, SQL server). Note that __Authelia__ is hot-reloaded in the environment so that your patches are instantly
included.

The next command starts the suite called *Standalone*:

```bash
authelia-scripts suites setup Standalone
```

Most of the suites are using docker compose to bootstrap the environment. Therefore, you can check the logs of all
application by running the following command on the component you want to monitor.

```bash
docker logs authelia-authelia-backend-1 -f
```

Then, edit the code and observe how __Authelia__ is automatically reloaded.

### Unit tests

To run the unit tests, run:

```bash
authelia-scripts unittest
```

### Integration tests

Integration tests are located under the `internal/suites` directory and are based on Selenium. A suite is a combination
of environment and tests. Executing a suite therefore means starting the environment, running the tests and tearing down
the environment. Each step can be run independently:

```bash {title="List the Available Suites"}
authelia-scripts suites list
```

```bash
Standalone
DuoPush
LDAP
Traefik
```

```bash {title="Start the environment of the Standalone suite"}
authelia-scripts suites setup Standalone
```

```bash {title="Run the tests related to the currently running suite"}
authelia-scripts suites test
```

```bash {title="Tear down the environment of the currently running suite"}
authelia-scripts suites teardown Standalone
```

In order to test all suites (approx 30 minutes), you need to make sure there is no currently running sui te and then you
should run:

```bash
authelia-scripts suites test
```

Also, you don't need to start the suite before testing it. Given you're not running any suite, just use the following
command to test the *Standalone* suite.

```bash
authelia-scripts suites test Standalone
```

The suite will be spawned, tests will be run and then the suite will be torn down automatically.

## Manually Building

### Binary

If you want to manually build the binary from source you will require the open source software described in the
[Development Environment](./environment.md#setup) documentation. Then you can follow the below steps on Linux (you may
have to adapt them on other systems).

Clone the Repository:

```bash
git clone https://github.com/authelia/authelia.git
```

Download the Dependencies:

```bash
cd authelia && go mod download
cd web && pnpm install
cd ..
```

Build the Web Frontend:

```bash
cd web && pnpm build
cd ..
cp -r api internal/server/public_html/api
```

Build the Binary (with debug symbols):

```bash
CGO_ENABLED=1 CGO_CPPFLAGS="-D_FORTIFY_SOURCE=2 -fstack-protector-strong" CGO_LDFLAGS="-Wl,-z,relro,-z,now" \
go build -ldflags "-linkmode=external" -trimpath -buildmode=pie -o authelia ./cmd/authelia
```

Build the Binary (without debug symbols):

```bash
CGO_ENABLED=1 CGO_CPPFLAGS="-D_FORTIFY_SOURCE=2 -fstack-protector-strong" CGO_LDFLAGS="-Wl,-z,relro,-z,now" \
go build -ldflags "-linkmode=external -s -w" -trimpath -buildmode=pie -o authelia ./cmd/authelia
```

[suites]: ./integration-suites.md
[React]: https://reactjs.org/
[go]: https://go.dev/dl/
[Node.js]: https://nodejs.org/en/download/
[Docker]: https://docs.docker.com/get-docker/
[Docker Compose]: https://docs.docker.com/compose/install/
[golangci-lint]: https://golangci-lint.run/usage/install/
[goimports-reviser]: https://github.com/incu6us/goimports-reviser#install
