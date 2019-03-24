# Security Keys (U2F)

**Authelia** also offers authentication using Security Keys supporting U2F
like [Yubikey](Yubikey) USB devices. U2F is one of the most secure
authentication protocol and is already available for Google, Facebook, Github
accounts and more.

The protocol requires your security key being enrolled before authenticating.

<p align="center">
  <img src="../../images/2factor_u2f.png" width="400">
</p>

To do so, select the *Security Key* method in the second factor page and click
on the *register new device* link. This will send a link to the 
user email address. This e-mail will likely be sent to https://mail.example.com:8080/
if you're testing Authelia and you've not configured anything.

Confirm your identity by clicking on **Continue** and you'll be asked to
touch the token of your security key to enroll.

<p align="center">
  <img src="../../images/u2f.png" width="400">
</p>

Upon successful registration, you can authenticate using your security key by simply
touching the token again.

Easy, right?!

## FAQ

### Why don't I have access to the *Security Key* option?

U2F protocol is a new protocol that is only supported by recent browser
and must even be enabled on some of them like Firefox. Please be sure
your browser supports U2F and that the feature is enabled to make the
option available in **Authelia**.

[Yubikey]: https://www.yubico.com/products/yubikey-hardware/yubikey4/
