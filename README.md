# two-factor-auth-server

  [![license](https://img.shields.io/github/license/mashape/apistatus.svg?maxAge=2592000)][MIT License]
  [![Build](https://travis-ci.org/clems4ever/two-factor-auth-server.svg?branch=master)](https://travis-ci.org/clems4ever/two-factor-auth-server)

**two-factor-auth-server** is the simplest to set up HTTP 2-factor authentication server. It is compatible with NGINX auth_request module and is used in production to secure internal services in a swarm cluster.

## Getting started

This project is docker-enabled so that you can deploy and test it very quickly. 
Before starting, make sure you don't have anything listening on port 8080. Then, type the following command to build and deploy the services:

    docker-compose build
    docker-compose up -d

After few seconds the services should be running and you should be able to visit [http://localhost:8080/](http://localhost:8080/) and access the login page:

![login-page](https://raw.githubusercontent.com/clems4ever/two-factor-auth-server/master/images/login.png)

### LDAP authentication
An LDAP server has been deployed with the following credentials: **admin/password**.

### TOTP verification
You can use Google Authenticator for the verification of the TOTP token. You can either enter the base32 secret key or scan the QR code in Google Authenticator and the application should start generating verification tokens.

Test secret key: GRWGIJS6IRHVEODVNRCXCOBMJ5AGC6ZE

![secret-key](https://raw.githubusercontent.com/clems4ever/two-factor-auth-server/master/images/secret-key.png)

## Documentation
two-factor-auth-server provides a way to log in using LDAP credentials and TOTP tokens. When the user is logged in,
the server generates a JSON web token with an expiry date that the user must keep in the *access_token* cookie.

### Endpoints
Here are the available endpoints:

| Endpoint        | Method    | Description                                                       |
|-----------------|-----------|-------------------------------------------------------------------|
| /login          | GET       | Serve a static webpage for login                                  |
| /logout         | GET       | Logout the current session if logged in                           |
| /_auth          | GET       | Verify whether the user is logged in                              |
| /_auth          | POST      | Generate an access token to store in *access_token* cookie        |

### Parameters
And the parameters:

| Endpoint           | Parameters                                                | Returns                          |
|--------------------|-----------------------------------------------------------|----------------------------------|
| /login             | None                                                      | Login static page                |
| /logout            | None                                                      | Redirect to *redirect* parameter |
| /_auth (GET)       | *access_token* cookie containing the JSON web token       | @204 or @401                     |
| /_auth (POST)      | { password: 'abc', username: 'user', token: '0982'}       | @200 with access_token or @401   |

## Contributing to two-factor-auth-server
Follow [contributing](CONTRIBUTORS.md) file.

## License
two-factor-auth-server is **licensed** under the **[MIT License]**. The terms of the license are as follows:

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
    
