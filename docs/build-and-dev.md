# Build and dev

**Authelia** is written in Typescript and built with [Authelia scripts](./docs/authelia-scripts.md).

In order to build and contribute to **Authelia**, you need to make sure Node with version >= 8 and < 10
and NPM is installed on your machine.

## Build

**Authelia** is made of two parts: the frontend and the backend.

The frontend is a [React](https://reactjs.org/) application written in Typescript and
the backend is an express application also written in Typescript.


The following command builds **Authelia** under dist/:

    authelia-scripts build

And then you can also build the Docker image with:

    authelia-scripts docker build

## Development

In order to ease development, Authelia uses the concept of [suites]. This is
a kind of virutal environment for **Authelia**, it allows you to run **Authelia** in a complete
environment, develop and test your patches. A hot-reload feature has been implemented so that
you can test your changes in realtime.

The next command will start the suite called [basic](./test/suites/basic/README.md): 

    authelia-scripts suites start basic

Then, edit the code and observe how **Authelia** is automatically updated.

### Unit tests

To run the unit tests written in Mocha, run:

    authelia-scripts unittest

### Integration tests

Integration tests also run with Mocha and are based on Selenium. They generally
require a complete environment made of several components like redis, mongo and a LDAP
to run. That's why [suites] have been created. At this point, the *basic* suite should
already be running and you can run the tests related to this suite with the following
command:

    authelia-scripts suites test


[suites]: ./suites.md