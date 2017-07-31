# Authelia

  [![license](https://img.shields.io/github/license/mashape/apistatus.svg?maxAge=2592000)][MIT License]
  [![Build](https://travis-ci.org/clems4ever/authelia.svg?branch=master)](https://travis-ci.org/clems4ever/authelia)
  [![Gitter](https://img.shields.io/gitter/room/badges/shields.svg)](https://gitter.im/authelia/general?utm_source=share-link&utm_medium=link&utm_campaign=share-link)

**Authelia** is a complete HTTP 2-factor authentication server for proxies like 
nginx. It has been made to work with nginx [auth_request] module and is currently 
used in production to secure internal services in a small docker swarm cluster.

# Table of Contents
1. [Features summary](#features-summary)
2. [Deployment](#deployment)
    1. [With NPM](#with-npm)
    2. [With Docker](#with-docker)
3. [Getting started](#getting-started)
    1. [Pre-requisites](#pre-requisites)
    2. [Run it!](#run-it)
4. [Features in details](#features-in-details)
    1. [First factor with LDAP and ACL](#first-factor-with-ldap-and-acl)
    2. [Second factor with TOTP](#second-factor-with-totp)
    3. [Second factor with U2F security keys](#second-factor-with-u2f-security-keys)
    4. [Password reset](#password-reset)
    5. [Access control](#access-control)
    6. [Session management with Redis](#session-management-with-redis)
4. [Documentation](#documentation)
    1. [Authelia configuration](#authelia-configuration)
    1. [API documentation](#api-documentation)
5. [Contributing to Authelia](#contributing-to-authelia)
6. [License](#license)

---

## Features summary
* Two-factor authentication using either 
**[TOTP] - Time-Base One Time password -** or **[U2F] - Universal 2-Factor -** 
as 2nd factor.
* Password reset with identity verification by sending links to user email 
address.
* Access restriction after too many authentication attempts.
* Session management using Redis key/value store.

## Deployment

If you don't have any LDAP and/or nginx setup yet, I advise you to follow the 
[Getting Started](#Getting-started) section. That way, you can test it right away 
without even configure anything.

Otherwise here are the available steps to deploy **Authelia** on your machine given 
your configuration file is **/path/to/your/config.yml**. Note that you can create your 
own the configuration file from **config.template.yml** at the root of the repo.

### With NPM

    npm install -g authelia
    authelia /path/to/your/config.yml

### With Docker

    docker pull clems4ever/authelia
    docker run -v /path/to/your/config.yml:/etc/authelia/config.yml -v /path/to/data/dir:/var/lib/authelia clems4ever/authelia

where **/path/to/data/dir** is the directory where all user data will be stored.

## Getting started

The provided example is docker-based so that you can deploy and test it very 
quickly.

### Pre-requisites

#### npm
Make sure you have npm and node installed on your computer.

#### Docker
Make sure you have **docker** and **docker-compose** installed on your machine.
For your information, here are the versions that have been used for testing:

    docker --version

gave *Docker version 17.03.1-ce, build c6d412e*.

    docker-compose --version

gave *docker-compose version 1.14.0, build c7bdf9e*.

#### Available port
Make sure you don't have anything listening on port 8080.

#### Subdomain aliases

Add the following lines to your **/etc/hosts** to alias multiple subdomains so that nginx can redirect request to the correct virtual host.

    127.0.0.1       public.test.local
    127.0.0.1       secret.test.local
    127.0.0.1       secret1.test.local
    127.0.0.1       secret2.test.local
    127.0.0.1       home.test.local
    127.0.0.1       mx1.mail.test.local
    127.0.0.1       mx2.mail.test.local
    127.0.0.1       auth.test.local

### Run it!
    
Deploy **Authelia** example with the following command:

    npm install --only=dev
    ./node_modules/.bin/grunt build-dist
    ./scripts/example/deploy-example.sh

After few seconds the services should be running and you should be able to visit 
[https://home.test.local:8080/](https://home.test.local:8080/).

When accessing the login page, a self-signed certificate exception should appear, 
it has to be trusted before you can get to the target page. The certificate
must be trusted for each subdomain, therefore it is normal to see the exception
 several times.

Below is what the login page looks like:

<img src="https://raw.githubusercontent.com/clems4ever/authelia/master/images/first_factor.png" width="400">

## Features in details

### First factor with LDAP and ACL 
An LDAP server has been deployed for you with the following credentials and
access control list:

- **john / password** is in the admin group and has access to the secret from
any subdomain.
- **bob / password** is in the dev group and has access to the secret from
  - [secret.test.local](https://secret.test.local:8080/secret.html) 
  - [secret2.test.local](https://secret2.test.local:8080/secret.html)
  - [home.test.local](https://home.test.local:8080/secret.html)
  - [\*.mail.test.local](https://mx1.mail.test.local:8080/secret.html)
- **harry / password** is not in a group but has rules giving him has access to 
 the secret from 
  - [secret1.test.local](https://secret1.test.local:8080/secret.html)
  - [home.test.local](https://home.test.local:8080/secret.html)

You can use them in the login page. If everything is ok, the second factor 
page should appear as shown below. Otherwise you'll get an error message notifying
your credentials are wrong.


<img src="https://raw.githubusercontent.com/clems4ever/authelia/master/images/second_factor.png" width="400">


### Second factor with TOTP
In **Authelia**, you need to register a per user TOTP (Time-Based One Time Password) secret before 
authenticating. To do that, you need to click on the register button. It will 
send a link to the user email address. Since this is an example, no email will 
be sent, the link is rather delivered in the file 
**./notifications/notification.txt**. Paste the link in your browser and you'll get 
your secret in QRCode and Base32 formats. You can use 
[Google Authenticator] 
to store them and get the generated tokens with the app.

<img src="https://raw.githubusercontent.com/clems4ever/authelia/master/images/totp.png" width="400">

### Second factor with U2F security keys
**Authelia** also offers authentication using U2F (Universal 2-Factor) devices like [Yubikey](Yubikey) 
USB security keys. U2F is one of the most secure authentication protocol and is 
already available for Google, Facebook, Github accounts and more.

Like TOTP, U2F requires you register your security key before authenticating. 
To do so, click on the register button. This will send a link to the 
user email address. Since this is an example, no email will be sent, the 
link is rather delivered in the file **./notifications/notification.txt**. Paste 
the link in your browser and you'll be asking to touch the token of your device 
to register. Upon successful registration, you can authenticate using your U2F 
device by simply touching the token. Easy, right?!

<img src="https://raw.githubusercontent.com/clems4ever/authelia/master/images/u2f.png" width="400">

### Password reset
With **Authelia**, you can also reset your password in no time. Click on the 
**Forgot password?** link in the login page, provide the username of the user requiring 
a password reset and **Authelia** will send an email with an link to the user 
email address. For the sake of the example, the email is delivered in the file 
**./notifications/notification.txt**.
Paste the link in your browser and you should be able to reset the password.

<img src="https://raw.githubusercontent.com/clems4ever/authelia/master/images/reset_password.png" width="400">

### Access Control
With **Authelia**, you can define your own access control rules for restricting 
the user access to some subdomains. Those rules are defined in the
configuration file and can be set either for everyone, per-user or per-group policies. 
Check out the *config.template.yml* to see how they are defined.

### Session management with Redis
When your users authenticate against Authelia, sessions are stored in a Redis key/value store. You can specify your own Redis instance in the [configuration file](#authelia-configuration).

## Documentation
### Authelia configuration
The configuration of the server is defined in the file 
**configuration.template.yml**. All the details are documented there.
You can specify another configuration file by giving it as first argument of 
**Authelia**.

    authelia config.custom.yml

### API documentation
There is a complete API documentation generated with 
[apiDoc](http://apidocjs.com/) and embedded in the repo under the **doc/** 
directory. Simply open index.html locally to watch it.

## Contributing to Authelia
Follow [contributing](CONTRIBUTORS.md) file.

## License
**Authelia** is **licensed** under the **[MIT License]**. The terms of the license are as follows:

    The MIT License (MIT)

    Copyright (c) 2016 - Clement Michaud

    Permission is hereby granted, free of charge, to any person obtaining a copy
    of this software and associated documentation files (the "Software"), to deal
    in the Software without restriction, including without limitation the rights
    to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
    copies of the Software, and to permit persons to whom the Software is
    furnished to do so, subject to the following conditions:

    The above copyright notice and this permission notice shall be included in
    all copies or substantial portions of the Software.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
    IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
    AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
    WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
    CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.


[MIT License]: https://opensource.org/licenses/MIT
[TOTP]: https://en.wikipedia.org/wiki/Time-based_One-time_Password_Algorithm
[U2F]: https://www.yubico.com/about/background/fido/
[Yubikey]: https://www.yubico.com/products/yubikey-hardware/yubikey4/
[auth_request]: http://nginx.org/en/docs/http/ngx_http_auth_request_module.html
[Google Authenticator]: https://play.google.com/store/apps/details?id=com.google.android.apps.authenticator2&hl=en

