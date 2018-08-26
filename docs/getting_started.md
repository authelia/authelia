# Getting started

**Authelia** can be tested in a matter of seconds with docker-compose based
on the latest image available on [Dockerhub] or by building the latest version
from the sources and use it in docker-compose.

## Pre-requisites

In order to test **Authelia**, we need to make sure that:
- **Docker** and **docker-compose** are installed.
- Some ports are open for listening on your machine.
- Some subdomains redirect to your machine to simulate the fact that some
applications you want to protect are served by some subdomains of
**example.com** on your machine.

### Docker & docker-compose

Make sure you have **docker** and **docker-compose** installed on your
machine.
Here are the versions used for testing in Travis:

    docker --version

gave *Docker version 17.03.1-ce, build c6d412e*.

    docker-compose --version

gave *docker-compose version 1.14.0, build c7bdf9e*.

### Available port

Make sure you don't have anything listening on port 8080 and 8085.

The port 8080 will be used by nginx to serve **Authelia** and the applications
we want to protect with **Authelia**.

The port 8085 is serving a webmail used to receive emails sent by **Authelia**
to validate your identity when registering U2F or TOTP secrets or when
resetting your password.

### Subdomain aliases

Make sure the following subdomains redirect to your machine by adding the
following lines to your **/etc/hosts**. It will alias the subdomains so that
nginx can redirect requests to the correct virtual host.

    127.0.0.1       home.example.com
    127.0.0.1       public.example.com
    127.0.0.1       dev.example.com
    127.0.0.1       admin.example.com
    127.0.0.1       mx1.mail.example.com
    127.0.0.1       mx2.mail.example.com
    127.0.0.1       single_factor.example.com
    127.0.0.1       login.example.com

## From Dockerhub

To deploy **Authelia** using the latest image from [Dockerhub], run the
following command:

    ./scripts/example-dockerhub/deploy-example.sh

## From source

To deploy **Authelia** from source, follow the [Build] manual and run the
following commands:

    ./scripts/example-commit/deploy-example.sh

## Test it!

After few seconds the services should be running and you should be able to
visit [https://home.example.com:8080/](https://home.example.com:8080/).

When accessing the login page, a self-signed certificate exception should
appear, it has to be trusted before you can get to the home page.
The certificate must also be trusted for each subdomain, therefore it is
normal to see this exception several times.

Below is what the login page looks like:

<p align="center">
  <img src="../images/first_factor.png" width="400">
</p>

At some point, you'll be required to register a secret for setting up
the second factor. **Authelia** will send an email to the user email
address to confirm the user identity. In order to receive it, visit the
webmail at [http://localhost:8085](http://localhost:8085).

**Note:** If you cannot deploy the fake webmail for any reason. You can
configure **Authelia** to use the filesystem notifier (option available
in [config.template.yml]) that will send the content of the email in a
file instead of sending an email. It is advised to use this option
for testing only.

Enjoy!

[DockerHub]: https://hub.docker.com/r/clems4ever/authelia/
[Build]: ./docs/build.md
