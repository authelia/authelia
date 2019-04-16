<p align="center">
  <img src="images/authelia-title.png" width="350" title="Authelia">
</p>

  [![license](https://img.shields.io/github/license/mashape/apistatus.svg?maxAge=2592000)][MIT License]
  [![Build](https://travis-ci.org/clems4ever/authelia.svg?branch=master)](https://travis-ci.org/clems4ever/authelia)
  [![Known Vulnerabilities](https://snyk.io/test/github/clems4ever/authelia/badge.svg?targetFile=package.json)](https://snyk.io/test/github/clems4ever/authelia?targetFile=package.json)
  [![Gitter](https://img.shields.io/gitter/room/badges/shields.svg)](https://gitter.im/authelia/general?utm_source=share-link&utm_medium=link&utm_campaign=share-link)
  [![Donate](https://img.shields.io/badge/Donate-PayPal-orange.svg)](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=clement%2emichaud34%40gmail%2ecom&lc=FR&item_name=Authelia&currency_code=EUR&bn=PP%2dDonationsBF%3abtn_donate_SM%2egif%3aNonHosted)

**Authelia** is an open-source authentication and authorization server
providing 2-factor authentication and single sign-on (SSO) for your
applications.
It acts as a companion of reverse proxies by handling authentication and
authorization requests.

**Authelia** can be installed as a standalone service using Docker or NPM
but can also be deployed easily on Kubernetes. On the latest, one can
leverage ingress configuration to set up authentication and authorizations
for specific services in only few seconds.

<p align="center">
  <img src="images/first_factor.png" width="400" />
  <img src="images/use-another-method.png" width="400" />
</p>

## Features summary

Here is the list of the main available features:

* Several kind of second factor:
  * **[Security Key (U2F)](./docs/2factor/security-key.md)** with [Yubikey].
  * **[Time-based One-Time password](./docs/2factor/time-based-one-time-password.md)** with [Google Authenticator].
  * **[Mobile Push Notifications](./docs/2factor/duo-push-notifications.md)** with [Duo](https://duo.com/).
* Password reset with identity verification using email.
* Single-factor only authentication method available.
* Access restriction after too many authentication attempts.
* Fine-grained access control per subdomain, user, resource and network.
* Support of [basic authentication] for endpoints protected by single factor.
* High-availability using distributed database and KV store.
* Compatible with Kubernetes ingress-nginx controller out of the box.

For more details about the features, follow [Features](./docs/features.md).

## Getting Started

You can start off with

    git clone https://github.com/clems4ever/authelia.git && cd authelia
    source bootstrap.sh

If you want to go further, please read [Getting Started](./docs/getting-started.md).

## Deployment

Now that you have tested **Authelia** and you want to try it out in your own infrastructure,
you can learn how to deploy and use it with [Deployment](./docs/deployment-production.md).
This guide will show you how to deploy it on bare metal as well as on
[Kubernetes](https://kubernetes.io/).

## Security

If you want more information about the security measures applied by
**Authelia** and some tips on how to set up **Authelia** in a secure way,
refer to [Security](./docs/security.md).

## Changelog & Breaking changes

See [CHANGELOG.md](CHANGELOG.md) and [BREAKING.md](BREAKING.md).

## Contribute

Anybody willing to contribute to the project either with code, 
documentation, security reviews or whatever, are very welcome to issue
or review pull requests and take part to discussions in
[Gitter](https://gitter.im/authelia/general?utm_source=share-link&utm_medium=link&utm_campaign=share-link).

We are already greatful to contributors listed in
[CONTRIBUTORS.md](CONTRIBUTORS.md) for their contributions to the project.
Be the next in the list!

## Build Authelia

If you want to contribute with code, you should follow the documentation explaining how to [build](./docs/build.md) the application.

## Donation

Wanna see more features? Then fuel us with a few beers!

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donate_SM.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=clement%2emichaud34%40gmail%2ecom&lc=FR&item_name=Authelia&currency_code=EUR&bn=PP%2dDonationsBF%3abtn_donate_SM%2egif%3aNonHosted)

## License

**Authelia** is **licensed** under the **[Apache 2.0]** license. The terms of the license are detailed
in [LICENSE](LICENSE).


[Apache 2.0]: https://www.apache.org/licenses/LICENSE-2.0
[TOTP]: https://en.wikipedia.org/wiki/Time-based_One-time_Password_Algorithm
[Security Key]: https://www.yubico.com/about/background/fido/
[Yubikey]: https://www.yubico.com/products/yubikey-hardware/yubikey4/
[auth_request]: http://nginx.org/en/docs/http/ngx_http_auth_request_module.html
[Google Authenticator]: https://play.google.com/store/apps/details?id=com.google.android.apps.authenticator2&hl=en
[config.template.yml]: https://github.com/clems4ever/authelia/blob/master/config.template.yml
