# http-two-factor
[![license](https://img.shields.io/github/license/mashape/apistatus.svg?maxAge=2592000)][MIT License]

**http-two-factor** is the simplest to set up HTTP 2-factor authentication server. It is compatible with NGINX auth_request module and is used in production to secure internal services in a swarm cluster.

## Getting started

This project is docker-enabled so that you can deploy and test it very quickly. 
Before starting, make sure you don't have anything listening on port 8080. Then, type the following command to build and deploy the services:

    docker-compose build
    docker-compose up -d

After few seconds the services should be running and you should be able to visit: http://localhost:8080/.

### LDAP authentication
An LDAP server has been deployed with the following credentials: **admin/password**.

### TOTP verification
You can use Google Authenticator for the verification of the TOTP token. You can either enter the base32 secret key or scan the QR code in Google Authenticator and the application should start generating verification tokens.

Test secret key: GRWGIJS6IRHVEODVNRCXCOBMJ5AGC6ZE

![secret-key](https://github.com/clems4ever/http-two-factor/raw/master/secret-key.png)

## Contributing to http-two-factor

Follow [contributing](CONTRIBUTING.md) file.

## License

http-2-factor is **licensed** under the **[MIT License]**. The terms of the license are as follows:

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
    
