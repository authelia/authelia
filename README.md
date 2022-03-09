<p align="center">
  <img src="./docs/images/authelia-title.png" width="350" title="Authelia">
</p>

  [![Build](https://img.shields.io/buildkite/d6543d3ece3433f46dbe5fd9fcfaf1f68a6dbc48eb1048bc22/master?logo=buildkite&style=flat-square&color=brightgreen)](https://buildkite.com/authelia/authelia)
  [![Go Report Card](https://goreportcard.com/badge/github.com/authelia/authelia?logo=go&style=flat-square)](https://goreportcard.com/report/github.com/authelia/authelia)
  [![Docker Tag](https://img.shields.io/docker/v/authelia/authelia/latest?logo=docker&style=flat-square&color=blue&sort=semver)](https://hub.docker.com/r/authelia/authelia/tags)
  [![Docker Size](https://img.shields.io/docker/image-size/authelia/authelia/latest?logo=docker&style=flat-square&color=blue&sort=semver)](https://hub.docker.com/r/authelia/authelia/tags)
  [![GitHub Release](https://img.shields.io/github/release/authelia/authelia.svg?logo=github&style=flat-square&color=blue)](https://github.com/authelia/authelia/releases)
  [![AUR source version](https://img.shields.io/aur/version/authelia?logo=arch-linux&label=authelia&style=flat-square&color=blue)](https://aur.archlinux.org/packages/authelia/)
  [![AUR binary version](https://img.shields.io/aur/version/authelia-bin?logo=arch-linux&label=authelia-bin&style=flat-square&color=blue)](https://aur.archlinux.org/packages/authelia-bin/)
  [![AUR development version](https://img.shields.io/aur/version/authelia-git?logo=arch-linux&label=authelia-git&style=flat-square&color=blue)](https://aur.archlinux.org/packages/authelia-git/)
  [![License](https://img.shields.io/github/license/authelia/authelia?logo=apache&style=flat-square&color=blue)][Apache 2.0]
  [![Sponsor](https://img.shields.io/opencollective/all/authelia-sponsors?logo=Open%20Collective&label=financial%20contributors&style=flat-square&color=blue)](https://opencollective.com/authelia-sponsors)
  [![Discord](https://img.shields.io/discord/707844280412012608?label=discord&logo=discord&style=flat-square&color=blue)](https://discord.authelia.com)
  [![Matrix](https://img.shields.io/matrix/authelia-support:matrix.org?label=matrix&logo=matrix&style=flat-square&color=blue)](https://matrix.to/#/#support:authelia.com)

**Authelia** is an open-source authentication and authorization server providing two-factor authentication and single 
sign-on (SSO) for your applications via a web portal. It acts as a companion for reverse proxies like [nginx], [Traefik] 
or [HAProxy] to let them know whether requests should either be allowed or redirected to Authelia's portal for
authentication.

Documentation is available at https://www.authelia.com/docs.

The following is a simple diagram of the architecture:

<p align="center" style="margin:50px">
  <img src="./docs/images/archi.png"/>
</p>

**Authelia** can be installed as a standalone service from the [AUR](https://aur.archlinux.org/packages/authelia/), 
[APT](https://apt.authelia.com/stable/debian/packages/authelia/),
[FreeBSD Ports](https://svnweb.freebsd.org/ports/head/www/authelia/), or using a 
[Static binary](https://github.com/authelia/authelia/releases/latest), 
[.deb package]((https://github.com/authelia/authelia/releases/latest)), [Docker] or [Kubernetes] either manually or via
the Helm [Chart](https://charts.authelia.com) (beta) leveraging ingress controllers and ingress configurations. 

<p align="center">
  <img src="./docs/images/logos/kubernetes.png" height="100"/>
  <img src="./docs/images/logos/docker.logo.png" width="100">
</p>

Here is what Authelia's portal looks like:

<p align="center">
  <img src="./docs/images/1FA.png" width="400" />
  <img src="./docs/images/2FA-METHODS.png" width="400" />
</p>

## Features summary

This is a list of the key features of Authelia:

* Several second factor methods:
  * **[Security Keys](https://www.authelia.com/docs/features/2fa/security-key)** that support [FIDO2]&nbsp;[Webauthn] with devices like a [YubiKey].
  * **[Time-based One-Time password](https://www.authelia.com/docs/features/2fa/one-time-password)**
    with compatible authenticator applications.
  * **[Mobile Push Notifications](https://www.authelia.com/docs/features/2fa/push-notifications)** 
    with [Duo](https://duo.com/).
* Password reset with identity verification using email confirmation.
* Access restriction after too many invalid authentication attempts.
* Fine-grained access control using rules which match criteria like subdomain, user, user group membership, request uri, 
 request method, and network.
* Choice between one-factor and two-factor policies per-rule.
* Support of basic authentication for endpoints protected by the one-factor policy.
* Highly available using a remote database and Redis as a highly available KV store.
* Compatible with [Traefik](https://doc.traefik.io/traefik) out of the box using the
  [ForwardAuth](https://doc.traefik.io/traefik/middlewares/http/forwardauth/) middleware.
* Curated configuration from [LinuxServer](https://www.linuxserver.io/) via their 
  [Swag](https://docs.linuxserver.io/general/swag) container as well as a 
  [guide](https://blog.linuxserver.io/2020/08/26/setting-up-authelia/).
* Kubernetes Support:
  * Compatible with the [ingress-nginx](https://github.com/kubernetes/ingress-nginx), the 
    [Traefik Kubernetes CRD](https://doc.traefik.io/traefik/providers/kubernetes-crd/), and the 
    [Traefik Kubernetes Ingress](https://doc.traefik.io/traefik/providers/kubernetes-crd/) Kubernetes ingress
    controllers out of the box.
  * Beta support for installing via Helm using our [Charts](https://charts.authelia.com).
* Beta support for [OpenID Connect](https://www.authelia.com/docs/configuration/identity-providers/oidc.html).

For more details about the features, follow [Features](https://www.authelia.com/docs/features/).

If you want to know more about the roadmap, follow [Roadmap](https://www.authelia.com/docs/roadmap).

## Proxy support

Authelia works in combination with [nginx], [Traefik] or [HAProxy]. It can be deployed on bare metal with
Docker or on top of [Kubernetes].

<p align="center">
  <img src="./docs/images/logos/nginx.png" height="50"/>
  <img src="./docs/images/logos/traefik.png" height="50"/>
  <img src="./docs/images/logos/haproxy.png" height="50"/>  
  <img src="./docs/images/logos/kubernetes.png" height="50"/> 
</p>

***Help Wanted:*** Assistance would be appreciated in getting Authelia working with
[Caddy](https://caddyserver.com/) and [Envoy](https://www.envoyproxy.io/).

<p align="center">
  <img src="./docs/images/logos/caddy.png" height="50"/>
  <img src="./docs/images/logos/envoy.png" height="50"/>
</p>

## Getting Started

### docker-compose

The `docker-compose` bundles act as a starting point for anyone wanting to see Authelia in action. You will have to
customize them to your needs as they come with self-signed certificates.

#### [Local](https://www.authelia.com/docs/getting-started)
The Local compose bundle is intended to test Authelia without worrying about configuration.
It's meant to be used for scenarios where the server is not be exposed to the internet.
Domains will be defined in the local hosts file and self-signed certificates will be utilised.

#### [Lite](https://www.authelia.com/docs/deployment/deployment-lite)
The Lite compose bundle is intended for scenarios where the server will be exposed to the internet, domains and DNS will 
need to be setup accordingly and certificates will be generated through LetsEncrypt. The Lite element refers to minimal 
external dependencies; File based user storage, SQLite based configuration storage. In this configuration, the service 
will not scale well.

#### [Full](https://www.authelia.com/docs/deployment/deployment-ha)
The Full compose bundle is intended for scenarios where the server will be exposed to the internet, domains and DNS will 
need to be setup accordingly and certificates will be generated through LetsEncrypt. The Full element refers to a 
scalable setup which includes external dependencies; LDAP based user storage, Database based configuration storage 
(MariaDB, MySQL or Postgres). 

## Deployment

Now that you have tested **Authelia** and you want to try it out in your own infrastructure,
you can learn how to deploy and use it with [Deployment](https://www.authelia.com/docs/deployment/deployment-ha).
This guide will show you how to deploy it on bare metal as well as on
[Kubernetes](https://kubernetes.io/).

## Security

Authelia takes security very seriously. If you discover a vulnerability in Authelia, please see our 
[Security Policy](https://github.com/authelia/authelia/security/policy). 

For more information about [security](https://www.authelia.com/docs/security/) related matters, please read
[the documentation](https://www.authelia.com/docs/security/).

## Contact Options

Several contact options exist for our community, the primary one being [Matrix](#matrix). These are in addition to
[GitHub issues](https://github.com/authelia/authelia/issues) for creating a [new issue](https://github.com/authelia/authelia/issues/new/choose).

### Matrix

Community members are invited to join the [Matrix Space](https://matrix.to/#/#community:authelia.com) which includes both
the [Support Room](https://matrix.to/#/#support:authelia.com) and the [Contributing Room](https://matrix.to/#/#contributing:authelia.com).

- The core team members are identified as administrators in the Space and individual Rooms.
- All channels are linked to [Discord](#discord).

### Discord

Community members are invited to join the [Discord Server](https://discord.authelia.com).

- The core team members are identified by the <span style="color:#BA55D3;">**CORE TEAM**</span> role in Discord.
- The [#support] and [#contributing] channels are linked to [Matrix](#matrix).

### Email

You can contact the core team by email via [team@authelia.com](mailto:team@authelia.com). Please note the  
[security@authelia.com](mailto:security@authelia.com) is also available but is strictly reserved for [security] related
matters.

## Breaking changes

Since Authelia is still under active development, it is subject to breaking changes. It's recommended to pin a version 
tag instead of using the `latest` tag and reading the [release notes](https://github.com/authelia/authelia/releases) 
before upgrading. This is where you will find information about breaking changes and what you should do to overcome 
said changes.

## Why Open Source?

You might wonder why Authelia is open source while it adds a great deal of security and user experience to your 
infrastructure at zero cost. It is open source because we firmly believe that security should be available for all to 
benefit in the face of the battlefield which is the Internet, with near zero effort.

Additionally, keeping the code open source is a way to leave it auditable by anyone who is willing to contribute. This 
way, you can be confident that the product remains secure and does not act maliciously.

It's important to keep in mind Authelia is not directly exposed on the
Internet (your reverse proxies are) however, it's still the control plane for your internal security so take care of it!

## Contribute

If you want to contribute to Authelia, please read our [contribution guidelines](./CONTRIBUTING.md).

Authelia exists thanks to all the people who contribute so don't be shy, come chat with us on either [Matrix](#matrix) 
or [Discord](#discord) and start contributing too.

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://github.com/clems4ever"><img src="https://avatars.githubusercontent.com/u/3193257?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Clément Michaud</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=clems4ever" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=clems4ever" title="Documentation">📖</a> <a href="#ideas-clems4ever" title="Ideas, Planning, & Feedback">🤔</a> <a href="#maintenance-clems4ever" title="Maintenance">🚧</a> <a href="#question-clems4ever" title="Answering Questions">💬</a> <a href="https://github.com/authelia/authelia/pulls?q=is%3Apr+reviewed-by%3Aclems4ever" title="Reviewed Pull Requests">👀</a> <a href="https://github.com/authelia/authelia/commits?author=clems4ever" title="Tests">⚠️</a></td>
    <td align="center"><a href="https://github.com/nightah"><img src="https://avatars.githubusercontent.com/u/3339418?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Amir Zarrinkafsh</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=nightah" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=nightah" title="Documentation">📖</a> <a href="#ideas-nightah" title="Ideas, Planning, & Feedback">🤔</a> <a href="#maintenance-nightah" title="Maintenance">🚧</a> <a href="#question-nightah" title="Answering Questions">💬</a> <a href="https://github.com/authelia/authelia/pulls?q=is%3Apr+reviewed-by%3Anightah" title="Reviewed Pull Requests">👀</a> <a href="https://github.com/authelia/authelia/commits?author=nightah" title="Tests">⚠️</a></td>
    <td align="center"><a href="https://github.com/james-d-elliott"><img src="https://avatars.githubusercontent.com/u/3903683?v=4?s=100" width="100px;" alt=""/><br /><sub><b>James Elliott</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=james-d-elliott" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=james-d-elliott" title="Documentation">📖</a> <a href="#ideas-james-d-elliott" title="Ideas, Planning, & Feedback">🤔</a> <a href="#maintenance-james-d-elliott" title="Maintenance">🚧</a> <a href="#question-james-d-elliott" title="Answering Questions">💬</a> <a href="https://github.com/authelia/authelia/pulls?q=is%3Apr+reviewed-by%3Ajames-d-elliott" title="Reviewed Pull Requests">👀</a> <a href="https://github.com/authelia/authelia/commits?author=james-d-elliott" title="Tests">⚠️</a></td>
    <td align="center"><a href="https://github.com/n4kre"><img src="https://avatars.githubusercontent.com/u/14371127?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Antoine Favre</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3An4kre" title="Bug reports">🐛</a> <a href="#ideas-n4kre" title="Ideas, Planning, & Feedback">🤔</a></td>
    <td align="center"><a href="https://github.com/BankaiNoJutsu"><img src="https://avatars.githubusercontent.com/u/2241519?v=4?s=100" width="100px;" alt=""/><br /><sub><b>BankaiNoJutsu</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=BankaiNoJutsu" title="Code">💻</a> <a href="#design-BankaiNoJutsu" title="Design">🎨</a></td>
    <td align="center"><a href="https://github.com/p-rintz"><img src="https://avatars.githubusercontent.com/u/13933258?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Philipp Rintz</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=p-rintz" title="Documentation">📖</a></td>
    <td align="center"><a href="http://callanbryant.co.uk/"><img src="https://avatars.githubusercontent.com/u/208440?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Callan Bryant</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=naggie" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=naggie" title="Documentation">📖</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/ViViDboarder"><img src="https://avatars.githubusercontent.com/u/137025?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Ian</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ViViDboarder" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/FrozenDragoon"><img src="https://avatars.githubusercontent.com/u/5301673?v=4?s=100" width="100px;" alt=""/><br /><sub><b>FrozenDragoon</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=FrozenDragoon" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/vdot0x23"><img src="https://avatars.githubusercontent.com/u/40716069?v=4?s=100" width="100px;" alt=""/><br /><sub><b>vdot0x23</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=vdot0x23" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/alexw1982"><img src="https://avatars.githubusercontent.com/u/11628284?v=4?s=100" width="100px;" alt=""/><br /><sub><b>alexw1982</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=alexw1982" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/Sohalt"><img src="https://avatars.githubusercontent.com/u/2157287?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Sohalt</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Sohalt" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=Sohalt" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/Tedyst"><img src="https://avatars.githubusercontent.com/u/13637623?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Stoica Tedy</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Tedyst" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/Chemsmith"><img src="https://avatars.githubusercontent.com/u/9061024?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Dylan Smith</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Chemsmith" title="Code">💻</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/LukasK13"><img src="https://avatars.githubusercontent.com/u/24586740?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Lukas Klass</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=LukasK13" title="Documentation">📖</a></td>
    <td align="center"><a href="https://staiger.it/"><img src="https://avatars.githubusercontent.com/u/9325003?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Philipp Staiger</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=lippl" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=lippl" title="Documentation">📖</a> <a href="https://github.com/authelia/authelia/commits?author=lippl" title="Tests">⚠️</a></td>
    <td align="center"><a href="https://yaleman.org/"><img src="https://avatars.githubusercontent.com/u/168188?v=4?s=100" width="100px;" alt=""/><br /><sub><b>James Hodgkinson</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=yaleman" title="Documentation">📖</a></td>
    <td align="center"><a href="https://chris.smith.xyz/"><img src="https://avatars.githubusercontent.com/u/1979423?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Chris Smith</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=chris13524" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/mqmq0"><img src="https://avatars.githubusercontent.com/u/13240971?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Mihály</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=mqmq0" title="Documentation">📖</a></td>
    <td align="center"><a href="https://iret.xyz/"><img src="https://avatars.githubusercontent.com/u/6560655?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Silver Bullet</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=SilverBut" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/skenmy"><img src="https://avatars.githubusercontent.com/u/1454505?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Paul Williams</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=skenmy" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=skenmy" title="Tests">⚠️</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/ntimo"><img src="https://avatars.githubusercontent.com/u/6145026?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Timo</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ntimo" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/andrewkliskey"><img src="https://avatars.githubusercontent.com/u/44645768?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Andrew Kliskey</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=andrewkliskey" title="Documentation">📖</a></td>
    <td align="center"><a href="http://kristofmattei.be/"><img src="https://avatars.githubusercontent.com/u/864376?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Kristof Mattei</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Kristof-Mattei" title="Documentation">📖</a></td>
    <td align="center"><a href="https://www.zmiguel.me/"><img src="https://avatars.githubusercontent.com/u/4400540?v=4?s=100" width="100px;" alt=""/><br /><sub><b>ZMiguel Valdiviesso</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=zmiguel" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/akusei"><img src="https://avatars.githubusercontent.com/u/12972900?v=4?s=100" width="100px;" alt=""/><br /><sub><b>akusei</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=akusei" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=akusei" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/Peaches491"><img src="https://avatars.githubusercontent.com/u/494334?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Daniel Miller</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Peaches491" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/dustins"><img src="https://avatars.githubusercontent.com/u/14645?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Dustin Sweigart</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=dustins" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=dustins" title="Documentation">📖</a> <a href="https://github.com/authelia/authelia/commits?author=dustins" title="Tests">⚠️</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/rogue780"><img src="https://avatars.githubusercontent.com/u/247716?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Shawn Haggard</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=rogue780" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=rogue780" title="Tests">⚠️</a></td>
    <td align="center"><a href="https://github.com/kevynb"><img src="https://avatars.githubusercontent.com/u/4941215?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Kevyn Bruyere</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=kevynb" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/ducksecops"><img src="https://avatars.githubusercontent.com/u/25612094?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Daniel Sutton</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ducksecops" title="Code">💻</a></td>
    <td align="center"><a href="http://www.xenuser.org/"><img src="https://avatars.githubusercontent.com/u/2216868?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Valentin Höbel</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=xenuser" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/thehedgefrog"><img src="https://avatars.githubusercontent.com/u/38590447?v=4?s=100" width="100px;" alt=""/><br /><sub><b>thehedgefrog</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=thehedgefrog" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/ViRb3"><img src="https://avatars.githubusercontent.com/u/2650170?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Victor</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ViRb3" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/whiskerch"><img src="https://avatars.githubusercontent.com/u/35109315?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Chris Whisker</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=whiskerch" title="Documentation">📖</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/nasatome"><img src="https://avatars.githubusercontent.com/u/18271791?v=4?s=100" width="100px;" alt=""/><br /><sub><b>nasatome</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=nasatome" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/bbros-dev"><img src="https://avatars.githubusercontent.com/u/60454087?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Begley Brothers (Development)</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=bbros-dev" title="Documentation">📖</a></td>
    <td align="center"><a href="http://mikekusold.com/"><img src="https://avatars.githubusercontent.com/u/509966?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Mike Kusold</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=kusold" title="Code">💻</a></td>
    <td align="center"><a href="https://dzervas.gr/"><img src="https://avatars.githubusercontent.com/u/1029195?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Dimitris Zervas</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=dzervas" title="Documentation">📖</a></td>
    <td align="center"><a href="http://paypal.me/DHoung"><img src="https://avatars.githubusercontent.com/u/52870424?v=4?s=100" width="100px;" alt=""/><br /><sub><b>TheCatLady</b></sub></a><br /><a href="#ideas-TheCatLady" title="Ideas, Planning, & Feedback">🤔</a></td>
    <td align="center"><a href="https://lauri.vosandi.com/"><img src="https://avatars.githubusercontent.com/u/194685?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Lauri Võsandi</b></sub></a><br /><a href="#ideas-laurivosandi" title="Ideas, Planning, & Feedback">🤔</a></td>
    <td align="center"><a href="https://github.com/knnnrd"><img src="https://avatars.githubusercontent.com/u/5852381?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Kennard Vermeiren</b></sub></a><br /><a href="#ideas-knnnrd" title="Ideas, Planning, & Feedback">🤔</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/ThinkChaos"><img src="https://avatars.githubusercontent.com/u/4761135?v=4?s=100" width="100px;" alt=""/><br /><sub><b>ThinkChaos</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ThinkChaos" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=ThinkChaos" title="Documentation">📖</a> <a href="https://github.com/authelia/authelia/commits?author=ThinkChaos" title="Tests">⚠️</a></td>
    <td align="center"><a href="https://github.com/except"><img src="https://avatars.githubusercontent.com/u/26675576?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Hasan</b></sub></a><br /><a href="#security-except" title="Security">🛡️</a></td>
    <td align="center"><a href="http://blog.dchidell.com"><img src="https://avatars.githubusercontent.com/u/26146619?v=4?s=100" width="100px;" alt=""/><br /><sub><b>David Chidell</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=dchidell" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/mardom1"><img src="https://avatars.githubusercontent.com/u/32371724?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Marcel Marquardt</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3Amardom1" title="Bug reports">🐛</a></td>
    <td align="center"><a href="https://cdine.org"><img src="https://avatars.githubusercontent.com/u/127512?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Ian Gallagher</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=craSH" title="Documentation">📖</a></td>
    <td align="center"><a href="https://wuhanstudio.cc"><img src="https://avatars.githubusercontent.com/u/15157070?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Wu Han</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=wuhanstudio" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/lavih"><img src="https://avatars.githubusercontent.com/u/47455309?v=4?s=100" width="100px;" alt=""/><br /><sub><b>lavih</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=lavih" title="Documentation">📖</a></td>
  </tr>
  <tr>
    <td align="center"><a href="http://jonbayl"><img src="https://avatars.githubusercontent.com/u/30201351?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Jon B. </b></sub></a><br /><a href="#security-jonbayl" title="Security">🛡️</a></td>
    <td align="center"><a href="https://github.com/AlexGustafsson"><img src="https://avatars.githubusercontent.com/u/14974112?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Alex Gustafsson</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=AlexGustafsson" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=AlexGustafsson" title="Documentation">📖</a></td>
    <td align="center"><a href="https://www.aarsen.me/"><img src="https://avatars.githubusercontent.com/u/7805050?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Arsenović Arsen</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ArsenArsen" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=ArsenArsen" title="Tests">⚠️</a> <a href="#security-ArsenArsen" title="Security">🛡️</a></td>
    <td align="center"><a href="https://github.com/dakriy"><img src="https://avatars.githubusercontent.com/u/13756065?v=4?s=100" width="100px;" alt=""/><br /><sub><b>dakriy</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=dakriy" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/davama"><img src="https://avatars.githubusercontent.com/u/5359152?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Dave</b></sub></a><br /><a href="#userTesting-davama" title="User Testing">📓</a></td>
    <td align="center"><a href="https://github.com/nreymundo"><img src="https://avatars.githubusercontent.com/u/5833447?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Nicolas Reymundo</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=nreymundo" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/polandy"><img src="https://avatars.githubusercontent.com/u/3670670?v=4?s=100" width="100px;" alt=""/><br /><sub><b>polandy</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=polandy" title="Documentation">📖</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/you1996"><img src="https://avatars.githubusercontent.com/u/45292366?v=4?s=100" width="100px;" alt=""/><br /><sub><b>yossbg</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=you1996" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/mpdcampbell"><img src="https://avatars.githubusercontent.com/u/47434940?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Michael Campbell</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=mpdcampbell" title="Documentation">📖</a></td>
    <td align="center"><a href="https://sievenpiper.co"><img src="https://avatars.githubusercontent.com/u/1131882?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Justin Sievenpiper</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=jsievenpiper" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/kaysond"><img src="https://avatars.githubusercontent.com/u/1147328?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Aram Akhavan</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=kaysond" title="Documentation">📖</a></td>
    <td align="center"><a href="https://skhuf.net"><img src="https://avatars.githubusercontent.com/u/286341?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Shadow</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=shadow7412" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/tarioch"><img src="https://avatars.githubusercontent.com/u/2998148?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Patrick Ruckstuhl</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=tarioch" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/FineWolf"><img src="https://avatars.githubusercontent.com/u/203591?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Andrew Moore</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=FineWolf" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=FineWolf" title="Documentation">📖</a> <a href="https://github.com/authelia/authelia/commits?author=FineWolf" title="Tests">⚠️</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. 
Contributions of any kind welcome!

### Backers

Thank you to all our backers! 🙏 [Become a backer](https://opencollective.com/authelia-sponsors/contribute) and help us 
sustain our community. The money we currently receive is dedicated to bootstrap a bug bounty program to give us as many
eyes as we can to detect potential vulnerabilities. 
<a href="https://opencollective.com/authelia-sponsors#backers"><img src="https://opencollective.com/authelia-sponsors/backers.svg?width=890"></a>

### Sponsors

Any company can become a sponsor by donating or providing any benefit to the project or the team helping improve
Authelia.

***Help Wanted:*** We are actively looking for sponsorship to obtain either a code security audit, penetration testing,
or other audits related to improving the security of Authelia.

Companies contributing to Authelia will have a specical mention below. [[Become a sponsor](https://opencollective.com/authelia-sponsors#sponsor)]

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

#### Balto

Thank you to [<img src="./docs/images/logos/balto.svg" alt="Balto" width="32"> Balto](https://www.getbalto.com/) for
hosting our apt repository.

#### Digital Ocean

Thank you to [<img src="./docs/images/logos/digitalocean.svg" alt="Digital Ocean" width="32"> DigitalOcean](https://www.digitalocean.com/?from=Authelia) for
contributing on OpenCollective.

#### JetBrains

Thank you to [<img src="./docs/images/logos/jetbrains.svg" alt="JetBrains" width="32"> JetBrains](https://www.jetbrains.com/?from=Authelia)
for providing us with free licenses to their great tools.

* [<img src="./docs/images/logos/intellij-idea.svg" alt="IDEA" width="32"> IDEA](http://www.jetbrains.com/idea/)
* [<img src="./docs/images/logos/goland.svg" alt="GoLand" width="32"> GoLand](http://www.jetbrains.com/go/)
* [<img src="./docs/images/logos/webstorm.svg" alt="WebStorm" width="32"> WebStorm](http://www.jetbrains.com/webstorm/)

## License

**Authelia** is **licensed** under the **[Apache 2.0]** license. The terms of the license are detailed in 
[LICENSE](./LICENSE).

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fauthelia%2Fauthelia.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fauthelia%2Fauthelia?ref=badge_large)


[Apache 2.0]: https://www.apache.org/licenses/LICENSE-2.0
[TOTP]: https://en.wikipedia.org/wiki/Time-based_One-time_Password_Algorithm
[FIDO2]: https://www.yubico.com/authentication-standards/fido2/
[YubiKey]: https://www.yubico.com/products/yubikey-5-overview/
[Webauthn]: https://www.yubico.com/authentication-standards/webauthn/
[auth_request]: https://nginx.org/en/docs/http/ngx_http_auth_request_module.html
[config.template.yml]: ./config.template.yml
[nginx]: https://www.nginx.com/
[Traefik]: https://traefik.io/
[HAProxy]: https://www.haproxy.org/
[Docker]: https://docker.com/
[Kubernetes]: https://kubernetes.io/
[security]: https://github.com/authelia/authelia/security/policy
[#support]: https://discord.com/channels/707844280412012608/707844280412012612
[#contributing]: https://discord.com/channels/707844280412012608/804943261265297408
