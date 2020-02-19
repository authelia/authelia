---
layout: default
title: Home
nav_order: 1
has_children: true
---

# Home

*It has never been so easy to secure your applications with Single Sign-On
and Two-Factor.*


With **Authelia** you can login once and get access to all your web apps
safely from the Web thanks to two-factor authentication.

<p align="center">
  <img src="./images/1FA.png" width="400" />
</p>

**Authelia** is an open source authentication and authorization server
protecting modern web applications by collaborating with reverse proxies
such as NGINX, Traefik and HAProxy. Consequently, no code is required to
protect your apps.

<p align="center" style="margin:50px">
  <img src="./images/archi.png"/>
</p>

Multiple 2-factor methods are available for satisfying every users.

* Time-based One-Time passwords with [Google Authenticator].
* Security Keys with tokens like [Yubikey].
* Push notifications on your mobile using [Duo].

**Authelia** is available as Docker images, static binaries and AUR packages
so that you can test it in minutes. Let's begin with the
[Getting Started](./getting-started).


## Authelia...

* is not an OAuth or OpenID Connect provider yet.
* is not a SAML provider yet.
<<<<<<< HEAD
* does not support support authentication against an OAuth or OpenID Connect provider.
* does not support using hardware devices as single factor.
* does not allow provide a PAM module yet.
=======
* does not support authentication against an OAuth or OpenID Connect provider.
* does not support authentication against a SAML provider.
* does not support using hardware devices as single factor.
* does not provide a PAM module yet.
>>>>>>> [WIP] Use 'Just-the-docs' jekyll theme to organize documentation.


[Duo]: https://duo.com/
[Yubikey]: https://www.yubico.com/products/yubikey-hardware/yubikey4/
<<<<<<< HEAD
[Google Authenticator]: https://google-authenticator.com/
=======
[Google Authenticator]: https://google-authenticator.com/
>>>>>>> [WIP] Use 'Just-the-docs' jekyll theme to organize documentation.
