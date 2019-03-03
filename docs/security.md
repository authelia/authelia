# Security

## Protection against cookie theft

Authelia uses two mechanism to protect against cookie theft:
1. session attribute `httpOnly` set to true make client-side code unable to
read the cookie.
2. session attribute `secure` ensure the cookie will never be sent over an
unsecure HTTP connections.

## Protection against multi-domain cookie attacks

Since Authelia uses multi-domain cookies to perform single sign-on, an
attacker who poisonned a user's DNS cache can easily retrieve the user's
cookies by making the user send a request to one of the attacker's IPs.

To mitigate this risk, it's advisable to only use HTTPS connections with valid
certificates and enforce it with HTTP Strict Transport Security ([HSTS]) so
that the attacker must also require the certificate to retrieve the cookies.

Note that using [HSTS] has consequences. That's why you should read the blog
post nginx has written on [HSTS].

## Content-Security-Policy

Authelia's portal is protected against XSS using the content
security policy mechanism that is documented
[here](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP). This protection
will reject untrusted payloads threatening your users during the authentication
workflow.

## More protections measures with Nginx

You can also apply the following headers to your nginx configuration for
improving security. Please read the documentation of those headers before
applying them blindly.

```
# We don't want any credentials / TOTP secret key / QR code to be cached by
# the client
add_header Cache-Control "no-store";
add_header Pragma "no-cache";

# Clickjacking / XSS protection

# We don't want Authelia's login page to be rendered within a <frame>, 
# <iframe> or <object> from an external website.
add_header X-Frame-Options "SAMEORIGIN";

# Block pages from loading when they detect reflected XSS attacks.
add_header X-XSS-Protection "1; mode=block";
```

## Helmet

To improve even more the security, [Helmet] has been added to **Authelia**.

## Contributing

If you find possible vulnerabilities or threats, do not hesitate to contribute
either by writing a test case demonstrating the possible attack and if
possible some solutions to prevent it or submit a PR.

[HSTS]: https://www.nginx.com/blog/http-strict-transport-security-hsts-and-nginx/
[Helmet]: https://helmetjs.github.io/
