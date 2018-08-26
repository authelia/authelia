# Features in details

## First factor using a LDAP server

**Authelia** uses an LDAP server as the backend for storing credentials.
When authentication is needed, the user is redirected to the login page which
corresponds to the first factor. **Authelia** tries to bind the username and
password against the configured LDAP backend.

You can find an example of the configuration of the LDAP backend in
[config.template.yml].

<p align="center">
  <img src="../images/second_factor.png" width="400">
</p>


## Second factor with TOTP

In **Authelia**, you can register a per user TOTP (Time-Based One Time
Password) secret before being being able to authenticate. Click on the
register button and check the email **Authelia** sent to your email address
to validate your identity.

Confirm your identity by clicking on **Continue** and you'll get redirected
on a page where your secret will be displayed  in QRCode and Base32 formats.
You can use [Google Authenticator] to store it and get the generated tokens.

<p align="center">
  <img src="../images/totp.png" width="400">
</p>

## Second factor with U2F security keys

**Authelia** also offers authentication using U2F (Universal 2-Factor) devices
like [Yubikey](Yubikey) USB security keys. U2F is one of the most secure
authentication protocol and is already available for Google, Facebook, Github
accounts and more.

Like TOTP, U2F requires you register your security key before authenticating. 
To do so, click on the register button. This will send a link to the 
user email address. 
Confirm your identity by clicking on **Continue** and you'll be asked to
touch the token of your device to register. Upon successful registration,
you can authenticate using your U2F device by simply touching the token.

Easy, right?!

<p align="center">
  <img src="./images/u2f.png" width="400">
</p>

## Password reset

With **Authelia**, you can also reset your password in no time. Click on the 
**Forgot password?** link in the login page, provide the username of the user
requiring a password reset and **Authelia** will send an email a confirmation
email to the user email address.

Proceed with the password reset form and validate to reset your password.

<p align="center">
  <img src="../images/reset_password.png" width="400">
</p>

## Access Control

With **Authelia**, you can define your own access control rules for finely
restricting user access to some resources and subdomains. Those rules are
defined and fully documented in the configuration file. They can apply to
users, groups or everyone.

Check out [config.template.yml] to see how they are defined.

## Single factor authentication

**Authelia** allows you to customize the authentication method to use for each 
subdomain. The supported methods are either "single_factor" or "two_factor". 
Please check [config.template.yml] to see an example of configuration.

It is also possible to use [basic authentication] to access a resource 
protected by a single factor.

## Session management with Redis

When your users authenticate against Authelia, sessions are stored in a
Redis key/value store. You can specify your own Redis instance in
[config.template.yml].

[basic authentication]: https://en.wikipedia.org/wiki/Basic_access_authentication
[config.template.yml]: https://github.com/clems4ever/authelia/blob/master/config.template.yml
[Google Authenticator]: https://play.google.com/store/apps/details?id=com.google.android.apps.authenticator2&hl=en
[Yubikey]: https://www.yubico.com/products/yubikey-hardware/yubikey4/
