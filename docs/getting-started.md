# Getting Started

**Authelia** can be tested in a matter of seconds with Docker and docker-compose.

In order to deploy the current version of Authelia locally, run the following
command and follow the instructions of bootstrap.sh:

    $ source bootstrap.sh

Then, start the *Standalone* [suite].

    $ authelia-scripts suites setup Standalone

A [suite] is kind of a virtual environment for running Authelia in a complete ecosystem.
If you want more details please read the related [documentation](./suites.md).

## Test it!

After few seconds the services should be running and you should be able to
visit [https://home.example.com:8080/](https://home.example.com:8080/).

When accessing the login page, since this is a test environment a
self-signed certificate exception should appear, it has to be trusted
before you can get to the home page.
The certificate must also be trusted for each subdomain, therefore it is
normal to see this exception several times.

Below is what the login page looks like after you accepted all exceptions:

<p align="center">
  <img src="../docs/images/1FA.png" width="400">
</p>

You can use one of the users listed in
[https://home.example.com:8080/](https://home.example.com:8080/).
The rights granted to each user and group is also provided in the page as
a list of rules.

At some point, you'll be required to register your second factor device.
Since your security is **Authelia**'s priority, it will send 
an email to the email address of the user to confirm the user identity.
Since you are running a test environment, a fake webmail called
*MailCatcher* has been deployed for you to check out the email and
confirm your identity.
The webmail is accessible at
[http://mail.example.com:8080](http://mail.example.com:8080).

Enjoy!

## FAQ

### What version of Docker and docker-compose should I use?

Here are the versions used for testing in Travis:

    $ docker --version
    Docker version 17.03.1-ce, build c6d412e

    $ docker-compose --version
    docker-compose version 1.14.0, build c7bdf9e

### How am I supposed to access the subdomains of example.com?

In order to test Authelia, Authelia fakes your browser by adding entries
in /etc/hosts when you first source the bootstrap.sh script.

### What should I do if I want to contribute?

You can refer to the dedicated documentation [here](./build-and-dev.md).

[config.template.yml]: ../config.template.yml
[DockerHub]: https://hub.docker.com/r/authelia/authelia/
[suite]: ./suites.md