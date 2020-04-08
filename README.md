<p align="center">
  <img src="./docs/images/authelia-title.png" width="350" title="Authelia">
</p>

  [![Build](https://img.shields.io/buildkite/d6543d3ece3433f46dbe5fd9fcfaf1f68a6dbc48eb1048bc22/master?style=flat-square&color=brightgreen)](https://buildkite.com/authelia/authelia)
  [![Go Report Card](https://goreportcard.com/badge/github.com/authelia/authelia?logo=go&style=flat-square)](https://goreportcard.com/report/github.com/authelia/authelia)
  [![Docker Tag](https://img.shields.io/docker/v/authelia/authelia?logo=docker&style=flat-square&color=blue&sort=semver)](https://microbadger.com/images/authelia/authelia)
  [![Docker Size](https://img.shields.io/docker/image-size/authelia/authelia?logo=docker&style=flat-square&color=blue&sort=semver)](https://hub.docker.com/r/authelia/authelia/tags)
  [![GitHub Release](https://img.shields.io/github/release/authelia/authelia.svg?logo=github&style=flat-square&color=blue)](https://github.com/authelia/authelia/releases)
  [![AUR source version](https://img.shields.io/aur/version/authelia?logo=arch-linux&label=authelia&style=flat-square&color=blue)](https://aur.archlinux.org/packages/authelia/)
  [![AUR binary version](https://img.shields.io/aur/version/authelia-bin?logo=arch-linux&label=authelia-bin&style=flat-square&color=blue)](https://aur.archlinux.org/packages/authelia-bin/)
  [![AUR development version](https://img.shields.io/aur/version/authelia-git?logo=arch-linux&label=authelia-git&style=flat-square&color=blue)](https://aur.archlinux.org/packages/authelia-git/)
  [![License](https://img.shields.io/github/license/authelia/authelia?logo=apache&style=flat-square&color=blue)][Apache 2.0]
  [![Sponsor](https://img.shields.io/opencollective/all/authelia-sponsors?logo=Open%20Collective&label=financial%20contributors&style=flat-square&color=blue)](https://opencollective.com/authelia-sponsors)
  [![Matrix](https://img.shields.io/matrix/authelia:matrix.org?logo=matrix&style=flat-square&color=blue)](https://riot.im/app/#/room/#authelia:matrix.org)

**Authelia** is an open-source authentication and authorization server
providing 2-factor authentication and single sign-on (SSO) for your
applications via a web portal.
It acts as a companion of reverse proxies like [nginx], [Traefik] or [HAProxy] to let them know whether queries should pass through. Unauthenticated user are
redirected to Authelia Sign-in portal instead.

Documentation is available at https://docs.authelia.com.

The architecture is shown in the diagram below.

<p align="center" style="margin:50px">
  <img src="./docs/images/archi.png"/>
</p>

**BREAKING NEWS: Authelia v4 has been released!
Please read BREAKING.md if you want to migrate from v3 to v4. Otherwise, start fresh in v4 and enjoy!**

**Authelia** can be installed as a standalone service from the [AUR](https://aur.archlinux.org/packages/authelia/), using a [Static binary](https://github.com/authelia/authelia/releases/latest), [Docker]
or can also be deployed easily on [Kubernetes] leveraging ingress controllers and ingress configuration.

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
  * **[Security Key (U2F)](https://docs.authelia.com/features/2fa/security-key)** with [Yubikey].
  * **[Time-based One-Time password](https://docs.authelia.com/features/2fa/one-time-password)** with [Google Authenticator].
  * **[Mobile Push Notifications](https://docs.authelia.com/features/2fa/push-notifications)** with [Duo](https://duo.com/).
* Password reset with identity verification using email confirmation.
* Single-factor only authentication method available.
* Access restriction after too many authentication attempts.
* Fine-grained access control per subdomain, user, resource and network.
* Support of basic authentication for endpoints protected by single factor.
* Highly available using a remote database and Redis as a highly available KV store.
* Compatible with Kubernetes [ingress-nginx](https://github.com/kubernetes/ingress-nginx) controller out of the box.

For more details about the features, follow [Features](https://docs.authelia.com/features/).

## Proxy support

Authelia works in combination with [nginx], [Traefik] or [HAProxy]. It can be deployed on bare metal with
Docker or on top of [Kubernetes].

<p align="center">
  <img src="./docs/images/logos/nginx.logo.png" height="50"/>
  <img src="./docs/images/logos/traefik.logo.png" height="50"/>
  <img src="./docs/images/logos/haproxy.logo.png" height="50"/>  
  <img src="./docs/images/logos/kubernetes.logo.png" height="50"/> 
</p>

## Getting Started

You can start utilising Authelia with the provided `docker-compose` bundles:

#### [Local](https://docs.authelia.com/getting-started)
The Local compose bundle is intended to test Authelia without worrying about configuration.
It's meant to be used for scenarios where the server is not be exposed to the internet.
Domains will be defined in the local hosts file and self-signed certificates will be utilised.

#### [Lite](https://docs.authelia.com/deployment/deployment-lite)
The Lite compose bundle is intended for scenarios where the server will be exposed to the internet, domains and DNS will need to be setup accordingly and certificates will be generated through LetsEncrypt.
The Lite element refers to minimal external dependencies; File based user storage, SQLite based configuration storage. In this configuration, the service will not scale well.

#### [Full](https://docs.authelia.com/deployment/deployment-ha)
The Full compose bundle is intended for scenarios where the server will be exposed to the internet, domains and DNS will need to be setup accordingly and certificates will be generated through LetsEncrypt.
The Full element refers to a scalable setup which includes external dependencies; LDAP based user storage, Database based configuration storage (MariaDB, MySQL or Postgres). 

## Deployment

Now that you have tested **Authelia** and you want to try it out in your own infrastructure,
you can learn how to deploy and use it with [Deployment](https://docs.authelia.com/deployment/deployment-ha).
This guide will show you how to deploy it on bare metal as well as on
[Kubernetes](https://kubernetes.io/).

## Security

Security is taken very seriously here, therefore we follow the rule of responsible
disclosure and we encourage you to do so.

Would you like to report any vulnerability discovered in Authelia, please first contact
**clems4ever** on [Matrix](https://riot.im/app/#/room/#authelia:matrix.org) or by
[email](mailto:clement.michaud34@gmail.com).

For details about security measures implemented in Authelia, please follow
this [link](https://docs.authelia.com/security/measures.html).

## Breaking changes

See [BREAKING](./BREAKING.md).

## Contribute

If you want to contribute to Authelia, check the documentation available
[here](https://docs.authelia.com/contributing/).

## Contributors

Authelia exists thanks to all the people who contribute. [[Contribute](./CONTRIBUTING.md)].
<a href="https://github.com/authelia/authelia/graphs/contributors"><img src="https://opencollective.com/authelia-sponsors/contributors.svg?width=890" /></a>

## Backers

Thank you to all our backers! 🙏 [[Become a backer](https://opencollective.com/authelia-sponsors/contribute)] and help us sustain our community.
<a href="https://opencollective.com/authelia-sponsors#backers"><img src="https://opencollective.com/authelia-sponsors/backers.svg?width=890"></a>

## Sponsors

Support Authelia by becoming a sponsor. Your logo will show up here with a link to your website. [[Become a sponsor](https://opencollective.com/authelia-sponsors#sponsor)]

<a href="https://opencollective.com/authelia-sponsors/sponsor/0/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/0/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/1/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/1/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/2/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/2/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/3/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/3/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/4/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/4/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/5/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/5/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/6/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/6/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/7/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/7/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/8/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/8/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/9/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/9/avatar.svg"></a>

## License

**Authelia** is **licensed** under the **[Apache 2.0]** license. The terms of the license are detailed
in [LICENSE](./LICENSE).


[Apache 2.0]: https://www.apache.org/licenses/LICENSE-2.0
[TOTP]: https://en.wikipedia.org/wiki/Time-based_One-time_Password_Algorithm
[Security Key]: https://www.yubico.com/about/background/fido/
[Yubikey]: https://www.yubico.com/products/yubikey-hardware/yubikey4/
[auth_request]: https://nginx.org/en/docs/http/ngx_http_auth_request_module.html
[Google Authenticator]: https://play.google.com/store/apps/details?id=com.google.android.apps.authenticator2&hl=en
[config.template.yml]: ./config.template.yml
[nginx]: https://www.nginx.com/
[Traefik]: https://traefik.io/
[HAProxy]: https://www.haproxy.org/
[Docker]: https://docker.com/
[Kubernetes]: https://kubernetes.io/
