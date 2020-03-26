---
layout: default
title: Getting Started
nav_order: 2
---

# Getting Started

## Docker Compose

**Authelia** can be deployed as a local setup not requiring SQL server, Redis cluster
or an LDAP server. In some cases, like protecting personal projects/websites, it can be
fine to use this setup but beware that this setup is non-resilient to failures so it should
be used at your own risk.

The setup is called local since it reduces the number of components in the architecture to
only two: a reverse proxy such as Nginx, Traefik or HAProxy and Authelia.

Connection details to Redis are optional. If not provided, sessions will be stored in
memory instead. This has the inconvenience of logging out users every time Authelia restarts.

## Steps

- `git clone https://github.com/authelia/authelia.git`
- `cd authelia/compose/local`
- `sudo ./setup.sh`
- `docker-compose up -d`

You can now visit the following locations; replace example.com with the domain you specified in the setup script:
- https://public.example.com - Bypasses Authelia
- https://traefik.example.com - Secured with Authelia one-factor authentication
- https://secure.example.com - Secured with Authelia two-factor authentication

Once you have registered an OTP device, the link to generate your QR code will be in `compose/local/authelia/notifications.txt`.
`grep "<a href=" compose/local/authelia/notifications.txt` 

## Reverse Proxy

Documentation for deploying a reverse proxy collaborating with Authelia is available
[here](./supported-proxies/index.md).

## FAQ

### Can you give more details on why this is not suitable for production environments?

This documentation gives instructions that will make **Authelia** non
resilient to failures and non scalable by preventing you from running multiple
instances of the application. This means that **Authelia** won't be able to distribute
the load across multiple servers and it will prevent failover in case of a
crash or an hardware issue. Moreover, users will be logged out every time
Authelia restarts.

## Development workflow

**Authelia** and its development workflow can be tested with Docker and docker-compose on Linux.

In order to deploy the current version of Authelia locally, run the following command and
follow the instructions of bootstrap.sh:

    $ source bootstrap.sh

Then, start the *Standalone* [suite].

    $ authelia-scripts suites setup Standalone

A [suite] is kind of a virtual environment for running Authelia in a complete ecosystem.
If you want more details please read the related [documentation](./contributing/suites.md).

### Test it!

After few seconds the services should be running and you should be able to
visit [https://home.example.com:8080/](https://home.example.com:8080/).

When accessing the login page, since this is a test environment a
self-signed certificate exception should appear, it has to be trusted
before you can get to the home page.
The certificate must also be trusted for each subdomain, therefore it is
normal to see this exception several times.

Below is what the login page looks like after you accepted all exceptions:

<p align="center">
  <img src="./images/1FA.png" width="400">
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

Here are the versions used for testing in Buildkite:

    $ docker --version
    Docker version 19.03.5, build 633a0ea838

    $ docker-compose --version
    docker-compose version 1.24.1, build unknown

### How can I serve my application under example.com?

Don't worry, you don't need to own the domain *example.com* to test Authelia.
Copy the following lines in your /etc/hosts.

    192.168.240.100 home.example.com
    192.168.240.100 login.example.com
    192.168.240.100 singlefactor.example.com
    192.168.240.100 public.example.com
    192.168.240.100 secure.example.com
    192.168.240.100 mail.example.com
    192.168.240.100 mx1.mail.example.com

`192.168.240.100` is the IP attributed by Docker to the reverse proxy. Once done
you can access the listed sub-domains from your browser and they will target
the reverse proxy.

### What should I do if I want to contribute?

You can refer to the dedicated documentation [here](./contributing/index.md).

[suite]: ./contributing/suites.md