**Authelia** is an open-source authentication and authorization server licensed under the MIT license. Authelia brings 2-factor authentication and single sign-on to secure web applications and ease authentication.

## Features summary

* Two-factor authentication using either 
**[TOTP] - Time-Base One Time password -** or **[U2F] - Universal 2-Factor -** 
as 2nd factor.
* Password reset with identity verification using email.
* Single and two factors authentication methods available. 
* Access restriction after too many authentication attempts.
* User-defined access control per subdomain and resource.
* Support of [basic authentication] for endpoints protected by single factor.
* High-availability using a highly-available distributed database and KV store.


[TOTP]: https://en.wikipedia.org/wiki/Time-based_One-time_Password_Algorithm
[U2F]: https://www.yubico.com/about/background/fido/