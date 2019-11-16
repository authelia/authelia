# Build and dev

**Authelia** is written in Go and comes with a dedicated CLI called [authelia-scripts](./authelia-scripts.md)
which is provided after running `source bootstrap.sh`. This CLI provides many useful tools to help you during
development.

In order to build and contribute to **Authelia**, you need to make sure Node with version >= 8 and < 12,
Go v1.13 and Docker is installed on your machine.

## Build

**Authelia** is made of two parts: the frontend and the backend.

The frontend is a [React](https://reactjs.org/) application written in Typescript and
the backend is Go application.

The following command builds **Authelia** under dist/:

    authelia-scripts build

Or you can also build the Alpine-based official Docker image with:

    authelia-scripts docker build

## Development

In order to ease development, Authelia uses the concept of [suites]. This is
a kind of virutal environment for **Authelia**, it allows you to run **Authelia** in a complete
ecosystem, develop and test your patches. A hot-reload feature has been implemented so that
you can test your changes right after the file has been saved.

The next command will start the suite called [basic](../test/suites/basic/README.md): 

    authelia-scripts suites start basic

Then, edit the code and observe how **Authelia** is automatically updated.

### Unit tests

To run the unit tests written, run:

    authelia-scripts unittest

### Integration tests

Integration tests run with Mocha and are based on Selenium. They generally
require a complete environment made of several components like redis, a SQL server and a
LDAP to run. That's why [suites] have been created. At this point, the *basic* suite should
already be running and you can run the tests related to this suite with the following
command:

    authelia-scripts suites test

You don't need to start the suite before testing it. Given your environment is not running
any suite, just use the following command to test the basic suite.

    authelia-scripts suites test basic

The suite will be spawned, tests will be run and then the suite will be teared down
automatically.


[suites]: ./suites.md
