---
layout: default
title: Suites
parent: Contributing
nav_order: 3
---

# Suites

Authelia is a single component in interaction with many others in a complete
ecosystem. Consequently, testing the features is not as easy as we might
think. In order to solve this problem, Authelia came up with the concept of
suite which is a kind of virtual environment for Authelia and a set of tests.
A suite can setup components such as nginx, redis or mariadb in which
Authelia can run and be tested.

This abstraction allows to prepare an environment for manual testing during
development and also to craft and run integration tests efficiently.

## Start a suite.

Starting a suite called *Standalone* is done with the following command:

    $ authelia-scripts suites setup Standalone

This command deploys the environment of the suite.

## Run tests of a suite

### Run tests of running suite

If a suite is already running, you can simply type the test command
that will run the test related to the currently running suite:

    $ authelia-scripts suites test

### Run tests of non-running suite

However, if no suite is running yet and you just want to run the tests of a
specific suite like *HighAvailability*, you can do so with the next command:

    # Set up the env, run the tests and tear down the env
    $ authelia-scripts suites test HighAvailability

### Run all tests of all suites

Running all tests is easy. Make sure that no suite is already running and run:

    authelia-scripts suites test

### Run tests in headless mode

As you might have noticed, the tests are run using chromedriver and selenium. It means
that the tests open an instance of Chrome that might interfere with your other activities.
In order to run the tests in headless mode to avoid the interference, use the following
command:

    $ authelia-scripts suites test --headless


## Create a suite

Creating a suite is as easy. Let's take the example of the **Standalone** suite:

* **suite_standalone.go** - It defines the setup and teardown phases. It likely uses
docker-compose to setup the ecosystem. This file also defines the timeouts.
* **suite_standalone_test.go** - It defines the set of tests to run against the suite.
* **Standalone** directory - It contains resources required by the suite and likely
mounted in the containers.

A suite can also be much more complex like setting up a complete Kubernetes ecosystem.
You can check the Kubernetes suite as example.