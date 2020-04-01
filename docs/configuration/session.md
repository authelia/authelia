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

  # The domain to protect.
  # Note: the login portal must also be a subdomain of that domain.
  domain: example.com

  # The remember me duration
  # Value of 0 disables remember me
  # Value is in seconds, or duration notation
  remember_me_duration: "1"

  # The redis connection details (optional)
  # If not provided, sessions will be stored in memory
  redis:
    host: 127.0.0.1
    port: 6379
    # This secret can also be set using the env variables AUTHELIA_SESSION_REDIS_PASSWORD
    password: authelia
```

# Duration Notation

We have implemented a string based notation for configuration options that take a duration. This section describes its
usage.

**NOTE:** At the time of this writing, only remember_me_duration uses this value type.
 
The notation is comprised of a number and a unit, and if more than one of these combinations exist it combines
them additively. For example `1h10M` or 1 hour and 10 minutes is the same as `10M1h`, both equal a total of 70 minutes.
The table below describes the units of time and the associated letter.

|Unit   |Associated Letter|
|:-----:|:---------------:|
|Years  |y                |
|Months |m                |
|Weeks  |w                |
|Days   |d                |
|Hours  |h                |
|Minutes|M                |
|Seconds|s                |