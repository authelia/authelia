# Suites

Authelia is a single component in interaction with many others. Consequently, testing the features
is not as easy as we might think. Consequently, a suite is kind of a virtual environment for Authelia, 
it allows to create an environment made of components such as nginx, redis or mongo in which Authelia can
run and be tested.

This abstraction allows to prepare an environment for manual testing during development and also to
craft and run integration tests.

## Start a suite.

Starting a suite called *simple* is done with the following command:

    npm run scripts suites start simple

It will start the suite and block until you hit ctrl-c to stop the suite.

## Run tests of a suite

### Run tests of running suite

If you are already running a suite with the previous command, you can simply type:

    npm run scripts test

and this will run the tests related to the running suite.

### Run tests of non-running suite

However, if no suite is running and you still want to test a particular suite like *complete*.
You can do so with the next command:

    npm run scripts test complete

This command will run the tests for the *complete* suite using the built version of Authelia that
should be located in *dist*.

WARNING: Authelia must be built before running this command.

### Run all tests of all suites

Running all tests is as easy as making sure that there is no running suite and typing:

    npm run scripts test


## Create a suite

Creating a suite is as easy as creating a new directory with at least two files:

* **environment.ts** - It defines the setup and teardown phases when creating the environment. The *setup*
phase is the phase when the required components will be spawned and Authelia will start while the *teardown*
is executed when the suite is destroyed (ctrl-c hit by the user).
* **test.ts** - It defines a set of tests to run in the virtual environment of the suite.