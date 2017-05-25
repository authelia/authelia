# Authelia

  [![license](https://img.shields.io/github/license/mashape/apistatus.svg?maxAge=2592000)][MIT License]
  [![Build](https://travis-ci.org/clems4ever/authelia.svg?branch=master)](https://travis-ci.org/clems4ever/authelia)

**Authelia** is a complete HTTP 2-factor authentication server for proxies like 
nginx. It has been made to work with NGINX auth_request module and is currently 
used in production to secure internal services in a small docker swarm cluster.

## Features
* Two-factor authentication using either 
**[TOTP] - Time-Base One Time password -** or **[U2F] - Universal 2-Factor -** 
as 2nd factor.
* Password reset with identity verification by sending links to user email 
address.
* Access restriction after too many authentication attempts.

## Deployment

If you don't have any LDAP and nginx setup yet, I advise you to follow the 
Getting Started. That way, you will not require anything to start.

Otherwise here are the available steps to deploy on your machine.

### With NPM

    npm install -g authelia

### With Docker

    docker pull clems4ever/authelia

## Getting started

The provided example is docker-based so that you can deploy and test it very 
quickly. First clone the repo make sure you don't have anything listening on 
port 8080 before starting. 
Add the following lines to your /etc/hosts to simulate multiple subdomains

    127.0.0.1       secret.test.local
    127.0.0.1       secret1.test.local
    127.0.0.1       secret2.test.local
    127.0.0.1       home.test.local
    127.0.0.1       mx1.mail.test.local
    127.0.0.1       mx2.mail.test.local
    127.0.0.1       auth.test.local
    
Then, type the following command to build and deploy the services:

    docker-compose build
    docker-compose up -d

After few seconds the services should be running and you should be able to visit 
[https://home.test.local:8080/](https://home.test.local:8080/). 

Normally, a self-signed certificate exception should appear, it has to be 
accepted before getting to the login page:

![first-factor-page](https://raw.githubusercontent.com/clems4ever/authelia/master/images/first_factor.png)

### 1st factor: LDAP and ACL 
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

Type them in the login page and validate. Then, the second factor page should 
have appeared as shown below.

![second-factor-page](https://raw.githubusercontent.com/clems4ever/authelia/master/images/second_factor.png)


### 2nd factor: TOTP (Time-Base One Time Password)
In **Authelia**, you need to register a per user TOTP secret before 
authenticating. To do that, you need to click on the register button. It will 
send a link to the user email address. Since this is an example, no email will 
be sent, the link is rather delivered in the file 
./notifications/notification.txt. Paste the link in your browser and you'll get 
your secret in QRCode and Base32 formats. You can use 
[Google Authenticator](https://play.google.com/store/apps/details?id=com.google.android.apps.authenticator2&hl=en) 
to store them and get the generated tokens required during authentication.

![totp-secret](https://raw.githubusercontent.com/clems4ever/authelia/master/images/totp.png)

### 2nd factor: U2F (Universal 2-Factor) with security keys
**Authelia** also offers authentication using U2F devices like [Yubikey](Yubikey) 
USB security keys. U2F is one of the most secure authentication protocol and is 
already available for accounts on Google, Facebook, Github and more.

Like TOTP, U2F requires you register your security key before authenticating 
with it. To do so, click on the register button. This will send a link to the 
user email address. Since this is an example, no email will be sent, the 
link is rather delivered in the file ./notifications/notification.txt. Paste 
the link in your browser and you'll be asking to touch the token of your device 
to register it. You can now authenticate using your U2F device by simply 
touching the token.

![u2f-validation](https://raw.githubusercontent.com/clems4ever/authelia/master/images/u2f.png)

### Password reset
With **Authelia**, you can also reset your password in no time. Click on the 
according button in the login page, provide the username of the user requiring 
a password reset and **Authelia** will send an email with an link to the user 
email address. For the sake of the example, the email is delivered in the file 
./notifications/notification.txt.
Paste the link in your browser and you should be able to reset the password.

![reset-password](https://raw.githubusercontent.com/clems4ever/authelia/master/images/reset_password.png)

### Access Control
With **Authelia**, you can define your own access control rules for restricting 
the access to certain subdomains to your users. Those rules are defined in the
configuration file and can be either default, per-user or per-group policies. 
Check out the *config.template.yml* to see how they are defined.

## Documentation
### Configuration
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

