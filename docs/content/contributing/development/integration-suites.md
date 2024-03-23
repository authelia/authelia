---
title: "Integration Suites"
description: "Integration Suites."
summary: "This section covers the build process and how to perform tests in development."
date: 2022-06-15T17:51:47+10:00
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

## Run tests of a suite

### Run tests of running suite

If a suite is already running, you can simply type the test command that will run the test related to the currently
running suite:

```bash
authelia-scripts suites test
```

### Run tests of non-running suite

However, if no suite is running yet and you just want to run the tests of a specific suite like *HighAvailability*, you
can do so with the next command:

```bash
authelia-scripts suites test HighAvailability
```

### Run all tests of all suites

Running all tests is easy. Make sure that no suite is already running and run:

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

## Create a suite

Creating a suite is as easy. Let's take the example of the __Standalone__ suite:

* __suite_standalone.go__ - It defines the setup and teardown phases. It likely uses docker-compose to setup the
  ecosystem. This file also defines the timeouts.
* __suite_standalone_test.go__ - It defines the set of tests to run against the suite.
* __Standalone__ directory - It contains resources required by the suite and likely mounted in the containers.

A suite can also be much more complex like setting up a complete Kubernetes ecosystem. You can check the Kubernetes
suite as example.
