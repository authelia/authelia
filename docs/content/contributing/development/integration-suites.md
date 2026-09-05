---
title: "Integration Suites"
description: "A guide to Authelia's integration suites which provide Docker-based virtual environments for manual testing and automated integration tests with Selenium."
summary: "This section covers the build process and how to perform tests in development."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 240
toc: true
aliases:
  - /docs/contributing/suites.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The following document assumes you've completed all the [prerequisites](environment.md#prerequisites) and have
bootstrapped the [Authelia Development Context](environment.md#context).
{{< /callout >}}

**Authelia** is a single component in interaction with many others in a complete ecosystem. Consequently, testing the
features is not as easy as we might think. In order to solve this problem, Authelia came up with the concept of suite
which is a kind of virtual environment for Authelia and a set of tests. A suite can setup components such as NGINX,
Redis or MariaDB in which **Authelia** can run and be tested.

This abstraction allows to prepare an environment for manual testing during development and also to craft and run
integration tests efficiently.

## Start a suite

Starting a suite called _Standalone_ is done with the following command:

```bash
authelia-scripts suites setup Standalone
```

This command deploys the environment of the suite.

## Accessing the Suite

The development suite has a standardized setup which makes it easy to interact with.

### IP Addresses

- Backend: 192.168.240.50
- Frontend: 192.168.240.100

The backend is the Authelia binary running in a docker container, the frontend is the webserver which hosts all of the
web frontends for each application.

These are the defaults and are what you get for local development. Anything that shares a single Docker daemon with
another suite run sets `SUITE_SLOT` so that it gets its own compose project and subnet, which changes the first three
octets of every address on this page. That is every CI agent, and every working tree beyond the first on a development
machine. See [Running Suites Concurrently](#running-suites-concurrently) and
[Environment Variables](#environment-variables) below.

### Sites and Applications

All sites are hosted on the address `${SUITE_SUBNET}.100:8080`, which is `192.168.240.100:8080` for local development.
This list is not comprehensive and may change over time. You can see a full list of the configured host entries by
looking at
[hosts.go](https://github.com/authelia/authelia/blob/master/internal/suites/hosts.go). For an idea
of the applications setup in a suite take a look at the `dockerEnvironment` var for the given suite. The file that
contains the `dockerEnvironment` var for a given suite is located in the
[internal/suites](https://github.com/authelia/authelia/tree/master/internal/suites) directory and has the name format
`suite_<name>.go` and does not end with `_test.go`. For example here is
[suite_standalone.go](https://github.com/authelia/authelia/blob/master/internal/suites/suite_standalone.go).

- Authelia: [https://login.example.com:8080](https://login.example.com:8080)
- Mailpit: [https://mail.example.com:8080](https://mail.example.com:8080)
- OpenID Connect 1.0 Testing Apps:
  - [https://oidc.example.com:8080](https://oidc.example.com:8080)
  - [https://oidc-public.example.com:8080](https://oidc-public.example.com:8080)
- Duo: [https://duo.example.com:8080](https://duo.example.com:8080)
- Kubernetes Dashboard: [https://kubernetes.example.com:8080](https://kubernetes.example.com:8080)
- Traefik Dashboard: [https://traefik.example.com:8080](https://traefik.example.com:8080)
- HAProxy: [https://haproxy.example.com:8080](https://haproxy.example.com:8080)
- Simple Test Applications:
  - [https://public.example.com:8080](https://public.example.com:8080)
  - [https://singlefactor.example.com:8080](https://singlefactor.example.com:8080)
  - [https://secure.example.com:8080](https://secure.example.com:8080)
  - [https://admin.example.com:8080](https://admin.example.com:8080)
  - [https://deny.example.com:8080](https://deny.example.com:8080)

## Environment Variables

The suites run several concurrent copies of themselves when the Docker daemon is shared. Every variable below is
optional and falls back to the value used for a single local run, so leaving them all unset gives the behavior
described above.

| Variable               | Default         | Purpose                                                                                                                                                                                                                                                                    |
| :--------------------- | :-------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `SUITE_SLOT`           | _unset_         | Slot number owned by this agent or working tree. When set, `bootstrap.sh` derives `COMPOSE_PROJECT_NAME`, `SUITE_SUBNET`, `LDAP_ADMIN_PORT`, `ENVOY_ADMIN_PORT` and, outside CI, `SUITE_TMP` and `SUITE_TMP_PATH` from it. A slotted shell also leaves `/etc/hosts` alone. |
| `SUITE_SLOT_AUTO`      | _unset_         | Set to `false` to stop `bootstrap.sh` allocating a slot for this working tree. Has no effect in CI, where the slot is supplied by the agent.                                                                                                                               |
| `COMPOSE_PROJECT_NAME` | `authelia`      | Compose project name. Also scopes the Traefik Docker provider so it only discovers its own containers.                                                                                                                                                                     |
| `SUITE_SUBNET`         | `192.168.240`   | First three octets of the suite network.                                                                                                                                                                                                                                   |
| `SUITE_TMP`            | `/tmp`          | Host directory bound into the suite containers, which always see it at `/tmp`. Give each agent or working tree its own directory, because everything at its top level apart from the agent's own working files is removed when a job finishes.                             |
| `SUITE_TMP_PATH`       | `/tmp`          | Path the test process itself reads and writes that same content through. In CI this stays `/tmp` because `SUITE_TMP` is bound there inside the agent; locally it is set to `SUITE_TMP`, since the test process runs on the host.                                           |
| `SUITE_IMAGE`          | `authelia:dist` | Image the backend runs.                                                                                                                                                                                                                                                    |
| `AGENT_CONTAINER`      | _unset_         | Name of the container the tests run in. When set, that container is attached to the suite network on setup and detached on teardown, so Chrome can reach the portal.                                                                                                       |

## Running Suites Concurrently

Several suites can run at once on one machine, one per git working tree. Sourcing `bootstrap.sh` in a working tree
allocates it a slot and derives everything that would otherwise collide from that number:

```bash
source bootstrap.sh
```

```console
[BOOTSTRAP] Using suite slot 2 for /home/user/authelia-feature
```

The slot is remembered per working tree, so it is the same on every subsequent source, and it is freed automatically
when the working tree is deleted. To inspect or manage the allocations:

```bash
authelia-scripts suites slot --list
authelia-scripts suites slot --release
```

A slotted shell deliberately does **not** touch `/etc/hosts`. The names in that file are machine wide but the addresses
behind them belong to one network, so only one working tree can own it. The tests do not need it: Chrome resolves the
suite domains from `--host-resolver-rules` and the Go clients resolve them when they dial, both from the same table in
[hosts.go](https://github.com/authelia/authelia/blob/master/internal/suites/hosts.go). This leaves `/etc/hosts`
pointing at the default `192.168.240.0/24` network, so an unslotted working tree stays browsable at the addresses
listed above while slotted ones run alongside it. To reach a slotted suite from your own browser, use its address
directly, for example `https://10.240.2.100:8080`.

Two things are still shared between concurrent runs and are worth knowing about:

- The pnpm store at `~/.local/share/pnpm/store` and the Go module and build caches at `${GOPATH}` are bound into the
  dev containers of every suite. They are safe for concurrent use but they do contend.
- Each suite in dev mode reserves roughly six CPUs and six gigabytes of memory across its backend, frontend and proxy
  containers, plus a Chrome process for the tests. Size the number of concurrent suites accordingly.

## Remote Debugging

The Authelia Suites run via [delve] and can be remotely debugged. You can connect to the debugger on the address
`${SUITE_SUBNET}.50:2345`, which is `192.168.240.50:2345` for an unslotted working tree.

Example connect command:

```bash
dlv connect 192.168.240.50:2345
```

## Run tests of a suite

### Run tests of running suite

If a suite is already running, you can simply type the test command that will run the test related to the currently
running suite:

```bash
authelia-scripts suites test
```

### Run tests in headless mode

As you might have noticed, the tests are run using chromedriver and selenium. It means that the tests open an instance
of Chrome that might interfere with your other activities. In order to run the tests in headless mode to avoid the
interference, use the following command:

```bash
authelia-scripts suites test --headless
```

### Run tests of non-running suite

However, if no suite is running yet and you just want to run the tests of a specific suite like _HighAvailability_, you
can do so with the next command:

```bash
authelia-scripts suites test HighAvailability
```

## Create a new suite

Creating a suite is as easy. Let's take the example of the **Standalone** suite:

- [internal/suites/suite_standalone.go](https://github.com/authelia/authelia/blob/master/internal/suites/suite_standalone.go) - It
  defines the setup and teardown phases. It likely uses docker compose to setup the ecosystem. This file also defines
  the timeouts.
- [internal/suites/suite_standalone_test.go](https://github.com/authelia/authelia/blob/master/internal/suites/suite_standalone_test.go)
  - It defines the set of tests to run against the suite.
- [internal/suites/Standalone](https://github.com/authelia/authelia/tree/master/internal/suites/Standalone) directory - It contains
  resources required by the suite and likely mounted in the containers.

A suite can also be much more complex like setting up a complete Kubernetes ecosystem. You can check the Kubernetes
suite as example.

[delve]: https://github.com/go-delve/delve
