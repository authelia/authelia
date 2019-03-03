# Getting Started

**Authelia** can be tested in a matter of seconds with docker-compose based
on the latest image available on [Dockerhub].

## Pre-requisites

In order to test **Authelia**, we need to make sure that:
- **Docker** and **docker-compose** are installed on your computer.
- Ports 8080 and 8085 are not already used on your machine.
- Some subdomains of **example.com** redirect to your test infrastructure.

### Docker & docker-compose

Make sure you have **docker** and **docker-compose** installed on your
machine.
Here are the versions used for testing in Travis:

    $ docker --version
    Docker version 17.03.1-ce, build c6d412e

    $ docker-compose --version
    docker-compose version 1.14.0, build c7bdf9e

### Available port

Make sure you don't have anything listening on port 8080 and 8085.

The port 8080 will be our frontend load balancer serving both **Authelia**'s portal and the
applications we want to protect.

The port 8085 is serving a webmail used to receive emails sent by **Authelia**
to validate your identity when registering U2F or TOTP secrets or when
resetting your password.

### Subdomain aliases

In order to simulate the behavior of a DNS resolving some test subdomains of **example.com**
to your machine, we need to add the following lines to your **/etc/hosts**. It will alias the
subdomains so that nginx can redirect requests to the correct virtual host.

    127.0.0.1       home.example.com
    127.0.0.1       public.example.com
    127.0.0.1       dev.example.com
    127.0.0.1       admin.example.com
    127.0.0.1       mx1.mail.example.com
    127.0.0.1       mx2.mail.example.com
    127.0.0.1       single_factor.example.com
    127.0.0.1       login.example.com

## Deploy

To deploy **Authelia** using the latest image from [Dockerhub], run the
following command:

    npm install commander
    npm run scripts suites start dockerhub

A Suites is a virtual environment for running Authelia. If you want more details please
read the related [documentation](./suites.md).

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
  <img src="../images/first_factor.png" width="400">
</p>

You can use one of the users listed in [https://home.example.com:8080/](https://home.example.com:8080/).
The rights granted to each user and group is also provided there.

At some point, you'll be required to register your second factor, either
U2F or TOTP. Since your security is **Authelia**'s priority, it will send 
an email to the email address of the user to confirm the user identity.
Since we're running a test environment, we provide a fake webmail called
*MailCatcher* from which you can checkout the email and confirm
your identity.
The webmail is accessible from
[http://localhost:8085](http://localhost:8085).

**Note:** If you cannot deploy the fake webmail for any reason. You can
configure **Authelia** to use the filesystem notifier (option available
in [config.template.yml]) that will send the content of the email in a
file instead of sending an email. It is advised to not use this option
in production.

Enjoy!

[config.template.yml]: ../config.template.yml
[DockerHub]: https://hub.docker.com/r/clems4ever/authelia/
[Build]: ./build.md
