<p align="center">
  <img src="./docs/images/authelia-title.png" width="350" title="Authelia">
</p>

  [![license](https://img.shields.io/badge/license-Apache%202.0-green.svg)][Apache 2.0]
  [![Build](https://travis-ci.org/authelia/authelia.svg?branch=master)](https://travis-ci.org/authelia/authelia)
  [![Gitter](https://img.shields.io/gitter/room/badges/shields.svg)](https://gitter.im/authelia/general?utm_source=share-link&utm_medium=link&utm_campaign=share-link)

**Authelia** is an open-source authentication and authorization server
providing 2-factor authentication and single sign-on (SSO) for your
applications via a web portal.
It acts as a companion of reverse proxies like [nginx] or [Traefik] to tell them wether queries should pass through. Unauthenticated user are
redirected to Authelia Sign-in portal instead.

The architecture is shown in the diagram below.

<p align="center" style="margin:50px">
  <img src="./docs/images/archi.png"/>
</p>

**BREAKING NEWS: Authelia v4 has been released!
Please read BREAKING.md if you want to migrate from v3 to v4. Otherwise, start fresh in v4 and enjoy!**

**Authelia** can be installed as a standalone service using Docker or NPM
but can also be deployed easily on [Kubernetes] leveraging ingress controllers and ingress configuration.

<p align="center">
  <img src="./docs/images/logos/kubernetes.logo.png" height="100"/>
  <img src="./docs/images/logos/docker.logo.png" width="100">
</p>

Here is what Authelia's portal looks like

<p align="center">
  <img src="./docs/images/1FA.png" width="400" />
  <img src="./docs/images/2FA-METHODS.png" width="400" />
</p>

## Features summary

Here is the list of the main available features:

* Several kind of second factor:
  * **[Security Key (U2F)](./docs/2factor/security-key.md)** with [Yubikey].
  * **[Time-based One-Time password](./docs/2factor/time-based-one-time-password.md)** with [Google Authenticator].
  * **[Mobile Push Notifications](./docs/2factor/duo-push-notifications.md)** with [Duo](https://duo.com/).
* Password reset with identity verification using email confirmation.
* Single-factor only authentication method available.
* Access restriction after too many authentication attempts.
* Fine-grained access control per subdomain, user, resource and network.
* Support of basic authentication for endpoints protected by single factor.
* Highly available using a remote database and Redis as a highly available KV store.
* Compatible with Kubernetes [ingress-nginx](https://github.com/kubernetes/ingress-nginx) controller out of the box.

For more details about the features, follow [Features](./docs/features.md).

## Proxy support

Authelia works in combination with [nginx] or [Traefik] and [HAProxy]. It can be deployed on bare metal with
Docker or directly in [Kubernetes].

<p align="center">
  <img src="./docs/images/logos/nginx.logo.png" height="50"/>
  <img src="./docs/images/logos/traefik.logo.png" height="50"/>
  <img src="./docs/images/logos/haproxy.logo.png" height="50"/>  
  <img src="./docs/images/logos/kubernetes.logo.png" height="50"/> 
</p>

## Getting Started

You can start off with

    git clone https://github.com/authelia/authelia.git && cd authelia
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

See [CHANGELOG.md](./CHANGELOG.md) and [BREAKING.md](./BREAKING.md).

## Contribute

Anybody willing to contribute to the project either with code, 
documentation, security reviews or whatever, are very welcome to issue
or review pull requests and take part to discussions in
[Gitter](https://gitter.im/authelia/general?utm_source=share-link&utm_medium=link&utm_campaign=share-link).

I am very grateful to contributors for their contributions to the project. Don't hesitate, be the next!

## Build Authelia

If you want to contribute with code, you should follow the documentation explaining how to [build](./docs/build-and-dev.md) the application.

## License

**Authelia** is **licensed** under the **[Apache 2.0]** license. The terms of the license are detailed
in [LICENSE](./LICENSE).


[Apache 2.0]: https://www.apache.org/licenses/LICENSE-2.0
[TOTP]: https://en.wikipedia.org/wiki/Time-based_One-time_Password_Algorithm
[Security Key]: https://www.yubico.com/about/background/fido/
[Yubikey]: https://www.yubico.com/products/yubikey-hardware/yubikey4/
[auth_request]: http://nginx.org/en/docs/http/ngx_http_auth_request_module.html
[Google Authenticator]: https://play.google.com/store/apps/details?id=com.google.android.apps.authenticator2&hl=en
[config.template.yml]: ./config.template.yml
[nginx]: https://www.nginx.com/
[Traefik]: https://traefik.io/
[HAproxy]: http://www.haproxy.org/
[Kubernetes]: https://kubernetes.io/