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
  expiration: 3600 # 1 hour

  # The inactivity time in seconds before the session is reset.
  inactivity: 300 # 5 minutes

  # The remember me duration.
  # Value of 0 disables remember me.
  # Value is in seconds, or duration notation. See: https://docs.authelia.com/configuration/session.html#duration-notation
  # Longer periods are less secure. Currently the default is 1M or 1 month.
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

# Duration Notation

We have implemented a string based notation for configuration options that take a duration. This section describes its
usage.

**NOTE:** At the time of this writing, only remember_me_duration uses this value type. But we plan to change expiration
and inactivity.
 
The notation is comprised of a number which must be positive and not have leading zeros, followed by a letter
denoting the unit of time measurement. The table below describes the units of time and the associated letter.

|Unit   |Associated Letter|
|:-----:|:---------------:|
|Years  |y                |
|Months |M                |
|Weeks  |w                |
|Days   |d                |
|Hours  |h                |
|Minutes|m                |
|Seconds|s                |

Examples:
* 1 hour and 30 minutes: 90m
* 1 day: 1d
* 10 hours: 10h