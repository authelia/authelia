# Build and dev

**Authelia** is written in Go and comes with a dedicated CLI called [authelia-scripts](./authelia-scripts.md)
which is available after running `source bootstrap.sh`. This CLI provides many useful tools to help you during
development.

In order to build and contribute to **Authelia**, you need to make sure Go v1.13, Docker,
docker-compose and Node with version >= 8 and < 12 are installed on your machine.

## Get started

**Authelia** is made of Go application serving the API and a [React](https://reactjs.org/)
application for the portal.

In order to ease development, Authelia uses the concept of [suites] to run Authelia from source
code so that your patches are included. This is a kind of virtual environment running **Authelia**
in a complete ecosystem (LDAP, Redis, SQL server). Note that Authelia is hotreloaded in the
environment so that your patches are instantly included.

The next command starts the suite called *Standalone*:

    authelia-scripts suites setup Standalone

Most of the suites are using docker-compose to bootstrap the environment. Therefore, you
can check the logs of all application by running the following command on the component
you want to monitor.

    docker logs authelia_authelia-backend_1 -f

Then, edit the code and observe how **Authelia** is automatically reloaded.

### Unit tests

To run the unit tests, run:

    authelia-scripts unittest

### Integration tests

Integration tests are located under the `suites` directory based on Selenium.

    authelia-scripts suites test

You don't need to start the suite before testing it. Given you're not running
any suite, just use the following command to test the *Standalone* suite.

    authelia-scripts suites test Standalone

The suite will be spawned, tests will be run and then the suite will be teared down
automatically.


[suites]: ./suites.md
