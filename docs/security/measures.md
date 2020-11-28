---
layout: default
title: Security Measures
parent: Security
nav_order: 1
---

# Security Measures

## Protection against cookie theft

Authelia uses two mechanisms to protect against cookie theft:
1. session attribute `httpOnly` set to true make client-side code unable to
read the cookie.
2. session attribute `secure` ensure the cookie will never be sent over an
insecure HTTP connections.

## Protection against multi-domain cookie attacks

Since Authelia uses multi-domain cookies to perform single sign-on, an
attacker who poisoned a user's DNS cache can easily retrieve the user's
cookies by making the user send a request to one of the attacker's IPs.

To mitigate this risk, it's advisable to only use HTTPS connections with valid
certificates and enforce it with HTTP Strict Transport Security ([HSTS]) so
that the attacker must also require the certificate to retrieve the cookies.

Note that using [HSTS] has consequences. That's why you should read the blog
post nginx has written on [HSTS].

## Protection against username enumeration

Authelia adaptively delays authentication attempts based on the mean (average) of the
previous 10 successful attempts, and a small random interval to make it even harder to
determine if the attempt was successful. On start it is assumed that the last 10 attempts
took 1000ms, this quickly grows or shrinks to the correct value over time regardless of the
authentication backend.

The cost of this is low since in the instance of a user not existing it just sleeps to delay
the login. Lastly the absolute minimum time authentication can take is 250ms. Both of these measures
also have the added effect of creating an additional delay for all authentication attempts reducing
the likelihood a password can be brute-forced even if regulation settings are too permissive.

## Protections against password cracking (File authentication provider)

Authelia implements a variety of measures to prevent an attacker cracking passwords if they
somehow obtain the file used by the file authentication provider, this is unrelated to LDAP auth.

First and foremost Authelia only uses very secure hashing algorithms with sane and secure defaults.
The first and default hashing algorithm we use is Argon2id which is currently considered
the most secure hashing algorithm. We also support SHA512, which previously was the default.

Secondly Authelia uses salting with all hashing algorithms. These salts are generated with a random
string generator, which is seeded every time it's used by a cryptographically secure 1024bit prime number.
This ensures that even if an attacker obtains the file, each password has to be brute forced individually.

Lastly Authelia's implementation of Argon2id is highly tunable. You can tune the key length, salt
used, iterations (time), parallelism, and memory usage. To read more about this please read how to
[configure](../configuration/authentication/file.md) file authentication.

## User profile and group membership always kept up-to-date (LDAP authentication provider)

Authelia by default refreshes the user's profile and membership every 5 minutes. Additionally, it
will invalidate any session where the user could not be retrieved from LDAP based on the user filter, for
example if they were deleted or disabled provided the user filter is set correctly. These updates occur when
a user accesses a resource protected by Authelia.

These protections can be [tuned](../configuration/authentication/ldap.md) according to your security policy
by changing refresh_interval, however we believe that 5 minutes is a fairly safe interval.

## Notifier security measures (SMTP)

By default the SMTP Notifier implementation does not allow connections that are not secure.
As such all connections require the following:

1. TLS Connection (STARTTLS or SMTPS) has been negotiated before authentication or sending emails (unauthenticated
connections require it as well)
2. Valid X509 Certificate presented to the client during the TLS handshake

There is an option to disable both of these security measures however they are
not recommended. You should only do this in a situation where you control all
networks between Authelia and the SMTP server. The following configuration options
exist to configure the security level:

### SMTPS vs STARTTLS

By default all connections start as plain text and are upgraded via STARTTLS. SMTPS is supported, however due to the
fact it was basically considered deprecated before the turn of the century, there is no way to configure it. It happens
automatically when a SMTP notifier is configured with the SMTPS port of 465.

### Configuration Option: disable_verify_cert

This is a YAML boolean type (true/false, y/n, 1/0, etc). This disables the X509 PKI
verification mechanism. We recommend using the trusted_cert option over this, as
disabling this security feature makes you vulnerable to MITM attacks.

### Configuration Option: disable_require_tls

This is a YAML boolean type (true/false, y/n, 1/0, etc). This disables the
requirement that all connections must be over TLS. This is only usable currently
with authentication disabled (comment the password) and as such is only an
option for SMTP servers that allow unauthenticated relay (bad practice).

### Configuration Option: trusted_cert

This is a YAML string type. This specifies the file location of a pub certificate
that can be used to validate the authenticity of a server with a self signed
certificate. This can either be the public cert of the certificate authority
used to sign the certificate or the public key itself. They must be in the PEM
format. The certificate is added in addition to the certificates trusted by the
host machine. If the certificate is invalid, inaccessible, or is otherwise not
configured; Authelia just uses the hosts certificates.

### Explanation
There are a few reasons for the security measures implemented:
1. Transmitting username's and passwords over plain-text is an obvious vulnerability
2. The emails generated by Authelia, if transmitted in plain-text could allow
an attacker to intercept a link used to setup 2FA; which reduces security
3. Not validating the identity of the server allows man-in-the-middle attacks

## Additional security

### Reset Password

It's possible to disable the reset password functionality and is recommended for anyone
wanting to increase security. See the [configuration](../configuration/authentication/index.md)
for information.

### Session security

We have a few options to configure the security of a session. The main and most important
one is the session secret. This is used to encrypt the session data when when stored in the
Redis key value database. This should be as random as possible.

Additionally you can configure the validity period of sessions. For example in a highly
security conscious domain you would probably want to set the session remember_me_duration
to 0 to disable this feature, and set an expiration of something like 2 hours and inactivity
of 10 minutes. This means the hard limit or the time the session will be destroyed no matter
what is 2 hours, and the soft limit or the time a user can be inactive for is 10 minutes.

### More protections measures with Nginx

You can also apply the following headers to your nginx configuration for
improving security. Please read the documentation of those headers before
applying them blindly.

```
# We don't want any credentials / TOTP secret key / QR code to be cached by
# the client
add_header Cache-Control "no-store";
add_header Pragma "no-cache";

# Clickjacking / XSS protection

# We don't want Authelia's login page to be rendered within a <frame>,
# <iframe> or <object> from an external website.
add_header X-Frame-Options "SAMEORIGIN";

# Block pages from loading when they detect reflected XSS attacks.
add_header X-XSS-Protection "1; mode=block";
```

[HSTS]: https://www.nginx.com/blog/http-strict-transport-security-hsts-and-nginx/

### More protections measures with fail2ban

If you are running fail2ban, adding a filter and jail for Authelia can reduce load on the application / web server. Fail2ban will ban IPs exceeding a threshold of repeated failed logins at the firewall level of your host.

If you are using Docker, the Authelia log file location has to be mounted from the host system to the container for fail2ban to access it.

Create a configuration file in the `filter.d` folder with the content below. In Debian-based systems the folder is typically located at `/etc/fail2ban/filter.d`.

```
# Fail2Ban filter for Authelia

# Make sure that the HTTP header "X-Forwarded-For" received by Authelia's backend
# only contains a single IP address (the one from the end-user), and not the proxy chain
# (it is misleading: usually, this is the purpose of this header).

# the failregex rule counts every failed login (wrong username or password) and failed TOTP entry as a failure
# the ignoreregex rule ignores debug, info and warning messages as all authentication failures are flagged as level=error by Authelia

[Definition]
failregex = ^.*Error while checking password for.*remote_ip=<HOST> stack.*
            ^.*Credentials are wrong for user .*remote_ip=<HOST> stack.*
            ^.*Wrong passcode during TOTP validation.*remote_ip=<HOST> stack.*

ignoreregex = ^.*level=debug.*
              ^.*level=info.*
              ^.*level=warning.*
```

Modify the `jail.local` file. In Debian-based systems the folder is typically located at `/etc/fail2ban/`. If the file does not exist, create it by copying the jail.conf `cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local`.
Add an Authelia entry to the "Jails" section of the file:
```
[authelia]
enabled = true
port = http,https,9091
filter = authelia
logpath = /path-to-your-authelia.log
maxretry = 3
bantime = 1d
findtime = 1d
chain = DOCKER-USER
```
If you are not using Docker remove the the line "chain = DOCKER-USER"

Finally, restart the fail2ban service.

## Container privilege de-escalation

Authelia will run as root by default, there are two options to run as a non-root user. The first option is to use the
Docker `--user` option on the command line or in docker-compose. The second option is to use the `PUID` and `PGID`
environment variables. An added benefit of using the environment variables is the mounted volumes ownership will automatically
be changed for you.

### Docker user option

With the Docker `--user` option, Docker will ensure Authelia is running as the user id and group id you specify.
In order to use this option, you will need to mount the `/config` volume to a directory on the host and set
the owner and group of that directory to the same user you supplied to docker. Running Authelia with `--user`
without mounting a volume to `/config` or incorrectly setting the host systems directory owner will cause Authelia
to exit immediately. The docker `--user` option will take precedence over the environment variables.

On the command line, you would create your Authelia data directory, change ownership to your non-root user
and run Authelia with `--user` set:
```
mkdir /authelia
chown user:group /authelia
docker run --user user:group -v /authelia:/config authelia/authelia:latest
```

As a docker-compose.yml file:
```
version: '3.8'
services:
  authelia:
    image: authelia/authelia
    container_name: authelia
    user: 1000:1000
    volumes:
      - ./authelia:/config
```

### PUID/PGID environment variables

If you choose to use the environment variables, the correct ownership will be applied automatically on startup of
the container, so there's no need to `chown` before running, to use this on the command line use the following:

```
docker run -e PUID=1000 -e PGID=1000 -v /authelia:/config authelia:authelia:latest
```

As a docker-compose.yml file:
```
version: '3.8'
services:
  authelia:
    image: authelia/authelia
    container_name: authelia
    environment:
      PUID: 1000
      PGID: 1000
    volumes:
      - ./authelia:/config
```
