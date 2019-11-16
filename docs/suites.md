# Suites

Authelia is a single component in interaction with many others. Consequently, testing the features
is not as easy as we might think. In order to solve this problem, Authelia came up with the concept of
suite which is a kind of virtual environment for Authelia, it allows to create an environment made of
components such as nginx, redis or mariadb in which Authelia can run and be tested.

This abstraction allows to prepare an environment for manual testing during development and also to
craft and run integration tests efficiently.

## Start a suite.

Starting a suite called *Standalone* is done with the following command:

    authelia-scripts suites setup Standalone

It will deploy the environment of the suite and block until you hit ctrl-c to stop the suite.

## Run tests of a suite

### Run tests of running suite

If a suite is already running, you can simply type:

    authelia-scripts suites test

and this will run the tests related to the running suite.

### Run tests of non-running suite

However, if no suite is running and you still want to test a particular suite like *HighAvailability*.
You can do so with the next command:

    authelia-scripts suites test HighAvailability

This command will run the tests for the *HighAvailability* suite. Beware that running tests of a
non-running suite implies the tests run against the distributable version of Authelia instead of
the current development version. If you made some patches, you must build the distributable version
before running the test command:

    # Build authelia before running the tests against the suite.
    authelia-scripts build
    authelia-scripts docker build

### Run all tests of all suites

Running all tests is easy. Make sure that no suite is already running and run:

    authelia-scripts suites test

Beware that the distributable version of Authelia is tested in that case. Don't
forget to build Authelia including your patches before running the command.


    # Build authelia before running the tests against the suite.
    authelia-scripts build
    authelia-scripts docker build

### Run tests in headless mode

In order to run the tests in headless mode, use the following command:

    authelia-scripts suites test --headless


## Create a suite

Creating a suite is as easy as creating a new directory with at least two files:

* **environment.ts** - It defines the setup and teardown phases when creating the environment. The *setup*
phase is the phase when the required components will be spawned and Authelia will start while the *teardown*
is executed when the suite is destroyed (ctrl-c hit by the user) or the tests are finished.
* **test.ts** - It defines a set of tests to run against the suite.