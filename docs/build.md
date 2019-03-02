# Build

**Authelia** is written in Typescript and built with [Authelia scripts](docs/authelia-scripts.md).

In order to build **Authelia**, you need to make sure Node with version >= 8 and < 10 and NPM is
installed on your machine.

Then, run the following command to install the node modules:

    npm install

And, this command to build **Authelia** under dist/:

    npm run build

Then you can also build the Docker image with:

    npm run docker build

## Details

### Build

**Authelia** is made of two parts: the frontend and the backend.

The frontend is a [React](https://reactjs.org/) application written in Typescript and
the backend is an express application also written in Typescript.

### Tests

There are two kind of tests: unit tests and integration tests.

### Unit tests

To run the unit tests, run:

    npm run unittest

### Integration tests

Integration tests run with Mocha and are based on Selenium. They generally
require a complete environment made of several components like redis, mongo and a LDAP
to run.

In order to simplify the creation of such environments, Authelia comes with a concept of
[Suites] that basically act as virtual environments for running either
manual or integration tests.

Please read the documentation related to [Suites] in order to discover
how to run related tests.


[Suites]: ./suites.md