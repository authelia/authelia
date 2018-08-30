<p align="center">
  <img src="images/authelia-title.png" width="350" title="Authelia">
</p>

  [![license](https://img.shields.io/github/license/mashape/apistatus.svg?maxAge=2592000)][MIT License]
  [![Build](https://travis-ci.org/clems4ever/authelia.svg?branch=master)](https://travis-ci.org/clems4ever/authelia)
  [![Known Vulnerabilities](https://snyk.io/test/github/clems4ever/authelia/badge.svg?targetFile=package.json)](https://snyk.io/test/github/clems4ever/authelia?targetFile=package.json)
  [![Gitter](https://img.shields.io/gitter/room/badges/shields.svg)](https://gitter.im/authelia/general?utm_source=share-link&utm_medium=link&utm_campaign=share-link)
  [![Donate](https://img.shields.io/badge/Donate-PayPal-orange.svg)](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=clement%2emichaud34%40gmail%2ecom&lc=FR&item_name=Authelia&currency_code=EUR&bn=PP%2dDonationsBF%3abtn_donate_SM%2egif%3aNonHosted)

**Authelia** is an open-source authentication and authorization providing
 2-factor authentication and single sign-on (SSO) for your applications.
It acts as a companion of reverse proxies by handling authentication and
authorization requests.

**Authelia** can be installed as a standalone service using Docker or NPM
but can also be deployed easily on Kubernetes. On the latest, one can
leverage ingress configuration to set up authentication and authorizations
for specific services in only few seconds.

<p align="center">
  <img src="images/first_factor.png" width="400" />
  <img src="images/second_factor.png" width="400" />
</p>

## Features summary

Here is the list of the main available features:

* **[U2F] - Universal 2-Factor -** support with [Yubikey].
* **[TOTP] - Time-Base One Time password -** support with [Google Authenticator].
* Password reset with identity verification using email.
* Single-factor only authentication method available.
* Access restriction after too many authentication attempts.
* User-defined access control per subdomain and resource.
* Support of [basic authentication] for endpoints protected by single factor.
* High-availability using distributed database and KV store.
* Compatible with Kubernetes ingress-nginx controller out of the box.

For more details about the features, follow [Features](./docs/features.md).

## Getting Started

Follow [Getting Started](./docs/getting_started.md).

## Security

If you want more information about the security measures applied by
**Authelia** and some tips on how to set up **Authelia** in a secure way,
refer to [Security](./docs/security.md).

## Deployment

To learn how to deploy **Authelia** or use it on Kubernetes, please follow
[Deployment](./docs/deployment.md).

## Build Authelia

Follow [Build](./docs/build.md).

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## Contributors

See the list of contributors in [CONTRIBUTORS.md](CONTRIBUTORS.md).

## Donation

Wanna see more features? Then fuel me with a few beers!

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donate_SM.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=clement%2emichaud34%40gmail%2ecom&lc=FR&item_name=Authelia&currency_code=EUR&bn=PP%2dDonationsBF%3abtn_donate_SM%2egif%3aNonHosted)

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
[config.template.yml]: https://github.com/clems4ever/authelia/blob/master/config.template.yml
