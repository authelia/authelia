---
title: "Integration Suites"
description: "Integration Suites."
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

__Authelia__ is a single component in interaction with many others in a complete ecosystem. Consequently, testing the
features is not as easy as we might think. In order to solve this problem, Authelia came up with the concept of suite
which is a kind of virtual environment for Authelia and a set of tests. A suite can setup components such as NGINX,
Redis or MariaDB in which __Authelia__ can run and be tested.

This abstraction allows to prepare an environment for manual testing during development and also to craft and run
integration tests efficiently.

## Start a suite

Starting a suite called *Standalone* is done with the following command:

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

### Sites and Applications

All sites are hosted on the address `192.168.240.100:8080`. This list is not comprehensive and may change over time.
You can see a full list of the configured host entries by looking at
[bootstrap.go](https://github.com/authelia/authelia/blob/master/cmd/authelia-scripts/cmd/bootstrap.go). For an idea
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

## Remote Debugging

The Authelia Suites run via [delve] and can be remotely debugged. You can connect to the debugger on the address
`192.168.240.50:2345`.

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

However, if no suite is running yet and you just want to run the tests of a specific suite like *HighAvailability*, you
can do so with the next command:

```bash
authelia-scripts suites test HighAvailability
```

## Create a new suite

Creating a suite is as easy. Let's take the example of the __Standalone__ suite:

* [internal/suites/suite_standalone.go](https://github.com/authelia/authelia/blob/master/internal/suites/suite_standalone.go) - It
  defines the setup and teardown phases. It likely uses docker compose to setup the ecosystem. This file also defines
  the timeouts.
* [internal/suites/suite_standalone_test.go](https://github.com/authelia/authelia/blob/master/internal/suites/suite_standalone_test.go)
  - It defines the set of tests to run against the suite.
* [internal/suites/Standalone](https://github.com/authelia/authelia/tree/master/internal/suites/Standalone) directory - It contains
  resources required by the suite and likely mounted in the containers.

A suite can also be much more complex like setting up a complete Kubernetes ecosystem. You can check the Kubernetes
suite as example.

[delve]: https://github.com/go-delve/delve
