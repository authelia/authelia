---
layout: default
title: Security Keys
nav_order: 2
parent: Second Factor
grand_parent: Features
---

# Security Keys

**Authelia** supports hardware-based second factors leveraging security keys like
[Yubikeys](Yubikey).

Security keys are among the most secure second factor. This method is already
supported by many major applications and platforms like Google, Facebook, Github,
some banks, and much more...

<p align="center">
  <img src="../../images/yubikey.jpg" width="150">
</p>

Normally, the protocol requires your security key to be enrolled on each site before
being able to authenticate with it. Since Authelia provides Single Sign-On, your users
will need to enroll their device only once to get access to all your applications.

<p align="center">
  <img src="../../images/REGISTER-U2F.png" width="400">
</p>

After having successfully passed the first factor, select *Security Key* method and
click on *Not registered yet?* link. This will send you an email to verify your identity.

*NOTE: This e-mail has likely been sent to the mailbox at https://mail.example.com:8080/ if you're testing Authelia.*

Confirm your identity by clicking on **Register** and you'll be asked to
touch the token of your security key to complete the enrollment.

Upon successful enrollment, you can authenticate using your security key
by simply touching the token again when requested:

<p align="center">
  <img src="../../images/2FA-U2F.png" width="400">
</p>

Easy, right?!

## FAQ

### Why don't I have access to the *Security Key* option?

U2F protocol is a new protocol that is only supported by recent browsers
and might even be enabled on some of them. Please be sure your browser
supports U2F and that the feature is enabled to make the option
available in **Authelia**.

[Yubikey]: https://www.yubico.com/products/yubikey-hardware/yubikey4/
