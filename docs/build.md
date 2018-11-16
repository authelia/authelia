# Build

**Authelia** is written in Typescript and built with [Grunt](https://gruntjs.com/).

In order to build **Authelia**, you need to make sure Node v8 and NPM is
installed on your machine.

Then, run the following command install the node modules:

    npm install

And, this command to build **Authelia** under dist/:

    ./node_modules/.bin/grunt build

## Details

### Build

**Authelia** is made of two components: the client and the server.

The client is written in Typescript and uses jQuery. It is built as part of
the global `build` Grunt command.

The server is written in Typescript. It is also built as part of the global `build`
Grunt command.

### Tests

Grunt also handles the commands to run the tests. There are several type of
tests for **Authelia**: unit tests for the server, the client and a shared
library and an integration test suite testing both components together.

The unit tests are written with Mocha while integration tests are using
Cucumber and Mocha.

### Unit tests

To run the client unit tests, run:

    ./node_modules/.bin/grunt test-client

To run the server unit tests, run:

    ./node_modules/.bin/grunt test-server

To run the shared library unit tests, run:

    ./node_modules/.bin/grunt test-shared

### Integration tests

Integration tests are mainly based on Selenium so they
need a complete environment to be run.

You can start by making sure **Authelia** is built with:

    grunt build

and the docker image is built with:

    ./scripts/example-commit/dc-example.sh build

Then, start the environment with:

    ./scripts/example-commit/dc-example.sh up -d

And run the tests with:

    ./node_modules/.bin/grunt test-int

Note: the Cucumber tests are hard to maintain and will therefore
be refactored to use Mocha instead.
