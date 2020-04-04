---
layout: default
title: Session
parent: Configuration
nav_order: 9
---

# Session

**Authelia** relies on session cookies to authenticate users. When the user visits
a website of the protected domain `example.com` for the first time, Authelia detects
that there is no cookie for that user. Consequently, Authelia redirects the user
to the login portal through which the user should authenticate to get a cookie which
is valid for `*.example.com`, meaning all websites of the domain.
At the next request, Authelia receives the cookie associated to the authenticated user
and can then order the reverse proxy to let the request pass through to the application.

## Configuration

```yaml
session:
  # The name of the session cookie. (default: authelia_session).
  name: authelia_session

  # The secret to encrypt the session cookie.
  # This secret can also be set using the env variables AUTHELIA_SESSION_SECRET
  secret: unsecure_session_secret

  # The time in seconds before the cookie expires and session is reset.
  expiration: 1h

  # The inactivity time in seconds before the session is reset.
  inactivity: 5m

  # The remember me duration.
  # Value of 0 disables remember me.
  # Value is in seconds, or duration notation. See: https://docs.authelia.com/configuration/index.html#duration-notation-format
  # Longer periods are considered less secure because a stolen cookie will last longer giving attackers more time to spy
  # or attack. Currently the default is 1M or 1 month.
  remember_me_duration:  1M

  # The domain to protect.
  # Note: the login portal must also be a subdomain of that domain.
  domain: example.com

  # The redis connection details (optional)
  # If not provided, sessions will be stored in memory
  redis:
    host: 127.0.0.1
    port: 6379
    # This secret can also be set using the env variables AUTHELIA_SESSION_REDIS_PASSWORD
    password: authelia
```

### Security

Configuration of this section has an impact on security. You should read notes in
[security measures](../security/measures.md#session-security) for more information.

### Duration Notation

The configuration parameters expiration, inactivity, and remember_me_duration use duration notation. See the documentation
for [duration notation format](index.md#duration-notation-format) for more information.